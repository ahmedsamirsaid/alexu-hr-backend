package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ApproveEmployeeProfileChangeRequestInput struct {
	RequestUID       string
	ActorUserID      int64
	ActorEmployeeUID string
	Comments         *string
}

type ApproveEmployeeProfileChangeRequestOutput struct {
	Request         *domain.EmployeeProfileChangeRequest
	ApprovalRequest *domain.ApprovalRequest
	Employee        *domain.Employee
}

type ApproveEmployeeProfileChangeRequestUseCase struct {
	db                   ports.DB
	employeeRepo         ports.EmployeeRepository
	changeRequestRepo    ports.EmployeeProfileChangeRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	roleRepo             ports.RoleRepository
	auditor              audit.Auditor
}

func NewApproveEmployeeProfileChangeRequestUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *ApproveEmployeeProfileChangeRequestUseCase {
	return &ApproveEmployeeProfileChangeRequestUseCase{
		db:                   db,
		employeeRepo:         employeeRepo,
		changeRequestRepo:    changeRequestRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		roleRepo:             roleRepo,
		auditor:              auditor,
	}
}

func (uc *ApproveEmployeeProfileChangeRequestUseCase) Execute(ctx context.Context, input ApproveEmployeeProfileChangeRequestInput) (*ApproveEmployeeProfileChangeRequestOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	request, err := uc.changeRequestRepo.GetByUID(ctx, tx, input.RequestUID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, ErrEmployeeProfileChangeRequestNotFound
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, request.ApprovalRequestUID)
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

	authorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, tx, input.ActorUserID, step.RoleUID, "")
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, request.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	oldFinancialGrade := cloneStringPtr(employee.FinancialGrade)
	oldIDCardValidUntil := cloneTimePtr(employee.IDCardValidUntil)
	oldMaritalStatus := cloneStringPtr(employee.MaritalStatus)

	employee.FinancialGrade = cloneStringPtr(request.RequestedFinancialGrade)
	employee.IDCardValidUntil = cloneTimePtr(request.RequestedIDCardValidUntil)
	employee.MaritalStatus = cloneStringPtr(request.RequestedMaritalStatus)

	if err := uc.employeeRepo.Update(ctx, tx, employee); err != nil {
		return nil, err
	}

	currentStep := approvalRequest.CurrentStep
	action := domain.NewApprovalAction(approvalRequest.UID, domain.ApprovalActionTypeApprove, &currentStep, input.ActorEmployeeUID, trimStringPtr(input.Comments))
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return nil, err
	}

	approvalRequest.Approve()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionApprove).
		On(audit.EntityEmployeeProfileChangeRequest, request.UID).
		WithMeta("employee_uid", employee.UID).
		WithMeta("approval_request_uid", approvalRequest.UID).
		WithMeta("old_financial_grade", stringPtrValue(oldFinancialGrade)).
		WithMeta("new_financial_grade", stringPtrValue(employee.FinancialGrade)).
		WithMeta("old_id_card_valid_until", formatOptionalDate(oldIDCardValidUntil)).
		WithMeta("new_id_card_valid_until", formatOptionalDate(employee.IDCardValidUntil)).
		WithMeta("old_marital_status", stringPtrValue(oldMaritalStatus)).
		WithMeta("new_marital_status", stringPtrValue(employee.MaritalStatus)).
		WithMeta("comments", stringPtrValue(trimStringPtr(input.Comments))).
		Save(ctx)

	return &ApproveEmployeeProfileChangeRequestOutput{
		Request:         request,
		ApprovalRequest: approvalRequest,
		Employee:        employee,
	}, nil
}

type RejectEmployeeProfileChangeRequestInput struct {
	RequestUID       string
	ActorUserID      int64
	ActorEmployeeUID string
	Comments         *string
}

type RejectEmployeeProfileChangeRequestUseCase struct {
	db                   ports.DB
	changeRequestRepo    ports.EmployeeProfileChangeRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	roleRepo             ports.RoleRepository
	auditor              audit.Auditor
}

func NewRejectEmployeeProfileChangeRequestUseCase(
	db ports.DB,
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *RejectEmployeeProfileChangeRequestUseCase {
	return &RejectEmployeeProfileChangeRequestUseCase{
		db:                   db,
		changeRequestRepo:    changeRequestRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		roleRepo:             roleRepo,
		auditor:              auditor,
	}
}

func (uc *RejectEmployeeProfileChangeRequestUseCase) Execute(ctx context.Context, input RejectEmployeeProfileChangeRequestInput) (*domain.ApprovalRequest, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	request, err := uc.changeRequestRepo.GetByUID(ctx, tx, input.RequestUID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, ErrEmployeeProfileChangeRequestNotFound
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, request.ApprovalRequestUID)
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

	authorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, tx, input.ActorUserID, step.RoleUID, "")
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	currentStep := approvalRequest.CurrentStep
	action := domain.NewApprovalAction(approvalRequest.UID, domain.ApprovalActionTypeReject, &currentStep, input.ActorEmployeeUID, trimStringPtr(input.Comments))
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return nil, err
	}

	approvalRequest.Reject()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionReject).
		On(audit.EntityEmployeeProfileChangeRequest, request.UID).
		WithMeta("approval_request_uid", approvalRequest.UID).
		WithMeta("comments", stringPtrValue(trimStringPtr(input.Comments))).
		Save(ctx)

	return approvalRequest, nil
}
