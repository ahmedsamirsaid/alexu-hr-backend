package usecases

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type RejectPermissionRequestInput struct {
	ApprovalRequestUID string
	ActorUserID        int64
	ActorEmployeeUID   string
	Comments           *string
}

type RejectPermissionRequestUseCase struct {
	db                   ports.DB
	permissionRepo       ports.PermissionRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	employeeRepo         ports.EmployeeRepository
	roleRepo             ports.RoleRepository
	userRepo             ports.UserRepository
	notificationService  ports.NotificationService
	auditor              audit.Auditor
}

func NewRejectPermissionRequestUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	employeeRepo ports.EmployeeRepository,
	roleRepo ports.RoleRepository,
	userRepo ports.UserRepository,
	notificationService ports.NotificationService,
	auditor audit.Auditor,
) *RejectPermissionRequestUseCase {
	return &RejectPermissionRequestUseCase{
		db:                   db,
		permissionRepo:       permissionRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		employeeRepo:         employeeRepo,
		roleRepo:             roleRepo,
		userRepo:             userRepo,
		notificationService:  notificationService,
		auditor:              auditor,
	}
}

func (uc *RejectPermissionRequestUseCase) Execute(ctx context.Context, input RejectPermissionRequestInput) (*domain.ApprovalRequest, error) {
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
		domain.ApprovalActionTypeReject,
		&currentStep,
		input.ActorEmployeeUID,
		input.Comments,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return nil, err
	}

	approvalRequest.Reject()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	permission, err := uc.permissionRepo.GetByApprovalRequestUID(ctx, tx, approvalRequest.UID)
	if err != nil {
		return nil, err
	}
	if permission != nil {
		permission.SetDecided()
		if err := uc.permissionRepo.Update(ctx, tx, permission); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Log audit event
	if permission != nil {
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
		if input.Comments != nil && *input.Comments != "" {
			actionParams["Reason"] = *input.Comments
		}
		defer uc.auditor.Actor(input.ActorEmployeeUID).
			Did(audit.ActionReject).
			On("permission_request", permission.UID).
			WithMeta("action_key", "audit.sentence.reject_permission").
			WithMeta("action_params", actionParams).
			Save(ctx)
	}

	if permission != nil {
		go uc.notifyRequesterRejected(requester.UID, permission, input.Comments)
	}

	return approvalRequest, nil
}

func (uc *RejectPermissionRequestUseCase) notifyRequesterRejected(employeeUID string, permission *domain.PermissionRequest, comments *string) {
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
	if comments != nil && *comments != "" {
		params["Comments"] = *comments
	}
	data := ports.NotificationData{
		"type":       "permission_rejected",
		"requestUid": permission.UID,
	}
	if _, err := uc.notificationService.SendToUser(user.UID, "notification.permission_rejected.title", "notification.permission_rejected.body", params, data); err != nil {
		slog.Error("reject_permission_request.notify", "error", err)
	}
}
