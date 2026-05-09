package usecases

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CancelPermissionRequestInput struct {
	PermissionRequestUID string
	ActorEmployeeUID     string
}

type CancelPermissionRequestUseCase struct {
	db                  ports.DB
	permissionRepo      ports.PermissionRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	approvalActionRepo  ports.ApprovalActionRepository
	employeeRepo        ports.EmployeeRepository
	deptRepo            ports.DepartmentRepository
	shiftRepo           ports.ShiftRepository
	userRepo            ports.UserRepository
	notificationService ports.NotificationService
	auditor             audit.Auditor
}

func NewCancelPermissionRequestUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	userRepo ports.UserRepository,
	notificationService ports.NotificationService,
	auditor audit.Auditor,
) *CancelPermissionRequestUseCase {
	return &CancelPermissionRequestUseCase{
		db:                  db,
		permissionRepo:      permissionRepo,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		employeeRepo:        employeeRepo,
		deptRepo:            deptRepo,
		shiftRepo:           shiftRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
		auditor:             auditor,
	}
}

func (uc *CancelPermissionRequestUseCase) Execute(ctx context.Context, input CancelPermissionRequestInput) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	request, err := uc.permissionRepo.GetByUID(ctx, tx, input.PermissionRequestUID)
	if err != nil {
		return err
	}
	if request == nil {
		return ErrPermissionRequestNotFound
	}
	if request.EmployeeUID != input.ActorEmployeeUID {
		return ErrNotRequestOwner
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, request.ApprovalRequestUID)
	if err != nil {
		return err
	}
	if approvalRequest == nil {
		return ErrApprovalRequestNotFound
	}

	wasApproved := approvalRequest.Status == domain.ApprovalRequestStatusApproved
	switch approvalRequest.Status {
	case domain.ApprovalRequestStatusPending, domain.ApprovalRequestStatusApproved:
		// allowed
	default:
		return ErrPermissionNotCancellable
	}

	approvalRequest.Cancel()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return err
	}

	request.SetDecided()
	if err := uc.permissionRepo.Update(ctx, tx, request); err != nil {
		return err
	}

	action := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeCancel,
		nil,
		input.ActorEmployeeUID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return err
	}

	// Get employee name for audit
	employee, _ := uc.employeeRepo.GetByUID(ctx, tx, input.ActorEmployeeUID)
	actorName := input.ActorEmployeeUID
	if employee != nil {
		actorName = employee.Name
	}

	// Audit log for permission request cancellation
	actionParams := map[string]interface{}{
		"Actor":          actorName,
		"PermissionType": request.Type.NameEN(),
		"Date":           request.PermissionDate.Format("Jan 2, 2006"),
	}
	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionCancel).
		On("permission_request", request.UID).
		WithMeta("action_key", "audit.sentence.cancel_permission").
		WithMeta("action_params", actionParams).
		WithMeta("permission_type", string(request.Type)).
		WithMeta("permission_date", request.PermissionDate.Format("2006-01-02")).
		WithMeta("old_status", "pending").
		WithMeta("new_status", "cancelled").
		WithMeta("approval_request_uid", approvalRequest.UID).
		Save(ctx)

	// Find last approver (if any) to notify on cancel-after-approval.
	var lastApproverUID string
	if wasApproved {
		actions, err := uc.approvalActionRepo.ListByRequest(ctx, tx, approvalRequest.UID)
		if err == nil {
			for i := len(actions) - 1; i >= 0; i-- {
				if actions[i].Action == domain.ApprovalActionTypeApprove {
					lastApproverUID = actions[i].ActorUID
					break
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if wasApproved {
		employeeUID := request.EmployeeUID
		date := request.PermissionDate
		go func() {
			_ = recomputeAttendanceForEmployeeDate(context.Background(), uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, uc.permissionRepo, employeeUID, date, nil)
		}()

		if lastApproverUID != "" && uc.notificationService != nil {
			go uc.notifyApproverCancelled(lastApproverUID, request)
		}
	}

	return nil
}

func (uc *CancelPermissionRequestUseCase) notifyApproverCancelled(approverEmployeeUID string, request *domain.PermissionRequest) {
	user, err := uc.userRepo.GetByEmployeeUID(context.Background(), uc.db, approverEmployeeUID)
	if err != nil || user == nil {
		return
	}
	params := map[string]interface{}{
		"PermissionType": ports.LocalizableString{Ar: request.Type.NameAR(), En: request.Type.NameEN()},
		"Date":           request.PermissionDate.Format("2006-01-02"),
	}
	data := ports.NotificationData{
		"type":       "permission_cancelled_after_approval",
		"requestUid": request.UID,
	}
	if _, err := uc.notificationService.SendToUser(user.UID, "notification.permission_cancelled.title", "notification.permission_cancelled.body", params, data); err != nil {
		slog.Error("cancel_permission_request.notify", "error", err)
	}
}
