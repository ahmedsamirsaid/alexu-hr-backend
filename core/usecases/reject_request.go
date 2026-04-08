package usecases

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type RejectRequestInput struct {
	ApprovalRequestUID string
	ActorUserID        int64
	ActorEmployeeUID   string
	Comments           *string
}

type RejectRequestOutput struct {
	ApprovalRequest *domain.ApprovalRequest
}

type RejectRequestUseCase struct {
	db                   ports.DB
	employeeRepo         ports.EmployeeRepository
	leaveRequestRepo     ports.LeaveRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	roleRepo             ports.RoleRepository
	notificationService  ports.NotificationService
	leaveTypeRepo        ports.LeaveTypeRepository
	userRepo             ports.UserRepository
}

func NewRejectRequestUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
	notificationService ports.NotificationService,
	leaveTypeRepo ports.LeaveTypeRepository,
	userRepo ports.UserRepository,
) *RejectRequestUseCase {
	return &RejectRequestUseCase{
		db:                   db,
		employeeRepo:         employeeRepo,
		leaveRequestRepo:     leaveRequestRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		roleRepo:             roleRepo,
		notificationService:  notificationService,
		leaveTypeRepo:        leaveTypeRepo,
		userRepo:             userRepo,
	}
}

func (uc *RejectRequestUseCase) Execute(ctx context.Context, input RejectRequestInput) (*RejectRequestOutput, error) {
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

	// Get the step configuration for current step
	step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, approvalRequest.ApprovalFlowUID, approvalRequest.CurrentStep)
	if err != nil {
		return nil, err
	}
	if step == nil {
		return nil, ErrApprovalFlowStepNotFound
	}

	// Get the requester to check their department
	requester, err := uc.employeeRepo.GetByUID(ctx, tx, approvalRequest.RequesterUID)
	if err != nil {
		return nil, err
	}
	if requester == nil || requester.DepartmentUID == nil {
		return nil, ErrNoDepartmentAssigned
	}

	// Check if actor is authorized to reject
	isAuthorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, tx, input.ActorUserID, step.RoleUID, *requester.DepartmentUID)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, ErrNotAuthorizedApprover
	}

	// Record the reject action
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

	// Update approval request status to rejected
	approvalRequest.Reject()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	// Update leave request decided_at
	leaveRequest, err := uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, tx, approvalRequest.UID)
	if err != nil {
		return nil, err
	}
	if leaveRequest != nil {
		leaveRequest.SetDecided()
		if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
			return nil, err
		}
	}

	// Get data for notification before commit
	var requesterUID string
	var leaveTypeName string
	var leaveRequestUID string
	if leaveRequest != nil {
		requesterUID = approvalRequest.RequesterUID
		leaveRequestUID = leaveRequest.UID
		leaveType, _ := uc.leaveTypeRepo.GetByUID(ctx, tx, leaveRequest.LeaveTypeUID)
		if leaveType != nil {
			leaveTypeName = leaveType.NameAR
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Send notification to requester (after commit, non-blocking)
	if requesterUID != "" {
		go uc.notifyRequesterRejected(requesterUID, leaveTypeName, leaveRequestUID)
	}

	return &RejectRequestOutput{
		ApprovalRequest: approvalRequest,
	}, nil
}

func (uc *RejectRequestUseCase) notifyRequesterRejected(employeeUID, leaveTypeName, requestUID string) {
	// Get the user linked to this employee
	user, err := uc.userRepo.GetByEmployeeUID(context.Background(), uc.db, employeeUID)
	if err != nil || user == nil {
		slog.Debug("reject_request.notifyRequesterRejected.no_user", "employee_uid", employeeUID)
		return
	}

	title := "تم رفض طلب الإجازة"
	body := "تم رفض طلب " + leaveTypeName
	data := ports.NotificationData{
		"type":       "request_rejected",
		"requestUid": requestUID,
	}

	_, err = uc.notificationService.SendToUser(user.UID, title, body, data)
	if err != nil {
		slog.Error("reject_request.notifyRequesterRejected.send", "error", err)
	}
}
