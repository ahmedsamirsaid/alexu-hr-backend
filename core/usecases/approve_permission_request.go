package usecases

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ApprovePermissionRequestInput struct {
	ApprovalRequestUID string
	ActorUserID        int64
	ActorEmployeeUID   string
	Comments           *string
}

type ApprovePermissionRequestOutput struct {
	ApprovalRequest   *domain.ApprovalRequest
	PermissionRequest *domain.PermissionRequest
}

type ApprovePermissionRequestUseCase struct {
	db                   ports.DB
	permissionRepo       ports.PermissionRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	employeeRepo         ports.EmployeeRepository
	deptRepo             ports.DepartmentRepository
	shiftRepo            ports.ShiftRepository
	roleRepo             ports.RoleRepository
	userRepo             ports.UserRepository
	notificationService  ports.NotificationService
	auditor              audit.Auditor
}

func NewApprovePermissionRequestUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	roleRepo ports.RoleRepository,
	userRepo ports.UserRepository,
	notificationService ports.NotificationService,
	auditor audit.Auditor,
) *ApprovePermissionRequestUseCase {
	return &ApprovePermissionRequestUseCase{
		db:                   db,
		permissionRepo:       permissionRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		employeeRepo:         employeeRepo,
		deptRepo:             deptRepo,
		shiftRepo:            shiftRepo,
		roleRepo:             roleRepo,
		userRepo:             userRepo,
		notificationService:  notificationService,
		auditor:              auditor,
	}
}

func (uc *ApprovePermissionRequestUseCase) Execute(ctx context.Context, input ApprovePermissionRequestInput) (*ApprovePermissionRequestOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, input.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}
	if approvalRequest == nil {
		return nil, ErrApprovalRequestNotFound
	}
	if !approvalRequest.IsPending() {
		return nil, ErrRequestNotPending
	}

	step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, approvalRequest.ApprovalFlowUID, approvalRequest.CurrentStep)
	if err != nil {
		return nil, err
	}
	if step == nil {
		return nil, ErrApprovalFlowStepNotFound
	}

	requester, err := uc.employeeRepo.GetByUID(ctx, tx, approvalRequest.RequesterUID)
	if err != nil {
		return nil, err
	}
	if requester == nil {
		return nil, ErrEmployeeNotFound
	}

	// Authorize: when requester has no department we treat the role assignment as global.
	departmentUID := ""
	if requester.DepartmentUID != nil {
		departmentUID = *requester.DepartmentUID
	}
	authorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, tx, input.ActorUserID, step.RoleUID, departmentUID)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	currentStep := approvalRequest.CurrentStep
	action := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeApprove,
		&currentStep,
		input.ActorEmployeeUID,
		input.Comments,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return nil, err
	}

	isFinal := approvalRequest.IsAtFinalStep()
	approvalRequest.Approve()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	permission, err := uc.permissionRepo.GetByApprovalRequestUID(ctx, tx, approvalRequest.UID)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, ErrPermissionRequestNotFound
	}

	if isFinal {
		permission.SetDecided()
		if err := uc.permissionRepo.Update(ctx, tx, permission); err != nil {
			return nil, err
		}
	}

	// Capture next-step info before commit.
	var nextStepRoleUID, nextStepDepartmentUID string
	if !isFinal {
		next, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, approvalRequest.ApprovalFlowUID, approvalRequest.CurrentStep)
		if err == nil && next != nil {
			nextStepRoleUID = next.RoleUID
			if requester.DepartmentUID != nil {
				nextStepDepartmentUID = *requester.DepartmentUID
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Log audit event
	if isFinal {
		actorUser, _ := uc.userRepo.GetByEmployeeUID(context.Background(), uc.db, input.ActorEmployeeUID)
		actorRole := ""
		if actorUser != nil {
			roles, _ := uc.roleRepo.GetRolesForUser(context.Background(), uc.db, actorUser.ID)
			if len(roles) > 0 {
				actorRole = roles[0].Name
			}
		}
		actor, _ := uc.employeeRepo.GetByUID(context.Background(), uc.db, input.ActorEmployeeUID)
		actorName := ""
		if actor != nil {
			actorName = actor.Name
		}
		actionParams := map[string]interface{}{
			"Actor":          actorName,
			"ActorRole":      actorRole,
			"PermissionType": permission.Type.NameEN(),
			"Date":           permission.PermissionDate.Format("Jan 2, 2006"),
		}
		defer uc.auditor.Actor(input.ActorEmployeeUID).
			Did(audit.ActionApprove).
			On("permission_request", permission.UID).
			WithMeta("action_key", "audit.sentence.approve_permission").
			WithMeta("action_params", actionParams).
			Save(ctx)
	}

	if isFinal {
		// Recompute attendance for that day so the excuse is reflected immediately.
		go func() {
			_ = recomputeAttendanceForEmployeeDate(context.Background(), uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, uc.permissionRepo, requester.UID, permission.PermissionDate, nil)
		}()
		go uc.notifyRequesterApproved(requester.UID, permission)
	} else if nextStepRoleUID != "" {
		go uc.notifyNextStepApprovers(nextStepRoleUID, nextStepDepartmentUID, requester, permission)
	}

	return &ApprovePermissionRequestOutput{
		ApprovalRequest:   approvalRequest,
		PermissionRequest: permission,
	}, nil
}

func (uc *ApprovePermissionRequestUseCase) notifyRequesterApproved(employeeUID string, permission *domain.PermissionRequest) {
	if uc.notificationService == nil {
		return
	}
	user, err := uc.userRepo.GetByEmployeeUID(context.Background(), uc.db, employeeUID)
	if err != nil || user == nil {
		return
	}
	params := map[string]interface{}{
		"PermissionType": permission.Type.NameAR(),
	}
	data := ports.NotificationData{
		"type":       "permission_approved",
		"requestUid": permission.UID,
	}
	if _, err := uc.notificationService.SendToUser(user.UID, "notification.permission_approved.title", "notification.permission_approved.body", params, data); err != nil {
		slog.Error("approve_permission_request.notify", "error", err)
	}
}

func (uc *ApprovePermissionRequestUseCase) notifyNextStepApprovers(roleUID, departmentUID string, requester *domain.Employee, permission *domain.PermissionRequest) {
	if uc.notificationService == nil {
		return
	}
	var deptPtr *string
	if departmentUID != "" {
		deptPtr = &departmentUID
	}
	approvers, err := uc.roleRepo.GetUsersByRoleAndDepartment(context.Background(), uc.db, roleUID, deptPtr)
	if err != nil || len(approvers) == 0 {
		return
	}
	userUIDs := make([]string, len(approvers))
	for i, u := range approvers {
		userUIDs[i] = u.UID
	}
	params := map[string]interface{}{
		"PermissionType": permission.Type.NameAR(),
		"EmployeeName":   requester.Name,
	}
	data := ports.NotificationData{
		"type":       "pending_permission",
		"requestUid": permission.UID,
	}
	if _, err := uc.notificationService.SendToUsers(userUIDs, "notification.pending_permission.title", "notification.pending_permission.body", params, data); err != nil {
		slog.Error("approve_permission_request.notify_next", "error", err)
	}
}
