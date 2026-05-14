package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/banumusa/backend/core/audit"
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
	auditor              audit.Auditor
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
	auditor audit.Auditor,
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
		auditor:              auditor,
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

	// Fetch actor name and role name for audit action sentence
	actorName := input.ActorEmployeeUID
	if actor, err := uc.employeeRepo.GetByUID(ctx, tx, input.ActorEmployeeUID); err == nil && actor != nil {
		actorName = actor.Name
	}
	actorRoleName := "Approver"
	if role, err := uc.roleRepo.GetByUID(ctx, tx, step.RoleUID); err == nil && role != nil {
		actorRoleName = role.Name
	}

	// Fetch leave request and leave type for context
	leaveRequest, _ := uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, tx, approvalRequest.UID)
	leaveTypeName := "Leave"
	var leaveStartDate, leaveEndDate string
	leaveDays := 0
	leaveRequestUID := ""
	if leaveRequest != nil {
		leaveRequestUID = leaveRequest.UID
		leaveStartDate = leaveRequest.StartDate.Format("Jan 2, 2006")
		leaveEndDate = leaveRequest.EndDate.Format("Jan 2, 2006")
		leaveDays = leaveRequest.Days
		if lt, err := uc.leaveTypeRepo.GetByUID(ctx, tx, leaveRequest.LeaveTypeUID); err == nil && lt != nil {
			leaveTypeName = lt.NameEN
		}
	}

	// Build reason string for the sentence
	reasonStr := ""
	if input.Comments != nil && *input.Comments != "" {
		reasonStr = fmt.Sprintf(" Reason: %s", *input.Comments)
	}

	actionParams := map[string]interface{}{
		"Actor":     actorName,
		"ActorRole": actorRoleName,
		"Requester": requester.Name,
		"LeaveType": leaveTypeName,
		"StartDate": leaveStartDate,
		"EndDate":   leaveEndDate,
		"Days":      leaveDays,
		"Reason":    reasonStr,
	}

	auditBuilder := uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionReject).
		On(audit.EntityLeaveRequest, leaveRequestUID).
		WithMeta("action_key", "audit.sentence.reject_leave").
		WithMeta("action_params", actionParams).
		WithMeta("reason", input.Comments).
		WithMeta("requester_name", requester.Name).
		WithMeta("leave_type", leaveTypeName).
		WithMeta("start_date", leaveStartDate).
		WithMeta("end_date", leaveEndDate).
		WithMeta("days", leaveDays).
		WithMeta("old_status", "pending").
		WithMeta("new_status", "rejected").
		WithMeta("step_order", approvalRequest.CurrentStep)
	defer auditBuilder.Save(ctx)

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
	if leaveRequest != nil {
		leaveRequest.SetDecided()
		if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
			return nil, err
		}
	}

	// Get data for notification
	var requesterUID string
	var leaveTypeNameAR, leaveTypeNameEN string
	if leaveRequest != nil {
		requesterUID = approvalRequest.RequesterUID
		leaveType, _ := uc.leaveTypeRepo.GetByUID(ctx, tx, leaveRequest.LeaveTypeUID)
		if leaveType != nil {
			leaveTypeNameAR = leaveType.NameAR
			leaveTypeNameEN = leaveType.NameEN
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Send notification to requester (after commit, non-blocking)
	if requesterUID != "" {
		go uc.notifyRequesterRejected(requesterUID, leaveTypeNameAR, leaveTypeNameEN, leaveRequestUID)
	}

	return &RejectRequestOutput{
		ApprovalRequest: approvalRequest,
	}, nil
}

func (uc *RejectRequestUseCase) notifyRequesterRejected(employeeUID, leaveTypeNameAR, leaveTypeNameEN, requestUID string) {
	user, err := uc.userRepo.GetByEmployeeUID(context.Background(), uc.db, employeeUID)
	if err != nil || user == nil {
		slog.Debug("reject_request.notifyRequesterRejected.no_user", "employee_uid", employeeUID)
		return
	}

	params := map[string]interface{}{
		"LeaveTypeName": ports.LocalizableString{Ar: leaveTypeNameAR, En: leaveTypeNameEN},
	}
	data := ports.NotificationData{
		"type":       "request_rejected",
		"requestUid": requestUID,
	}

	_, err = uc.notificationService.SendToUser(
		user.UID,
		"notification.leave_rejected.title",
		"notification.leave_rejected.body",
		params,
		data,
	)
	if err != nil {
		slog.Error("reject_request.notifyRequesterRejected.send", "error", err)
	}
}
