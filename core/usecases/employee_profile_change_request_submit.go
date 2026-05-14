package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type SubmitEmployeeProfileChangeRequestInput struct {
	EmployeeUID               string
	ActorUserID               int64
	ActorEmployeeUID          string
	RequestedFinancialGrade   *string
	RequestedIDCardValidUntil *time.Time
	RequestedMaritalStatus    *string
	Comments                  *string
	SetFinancialGrade         bool
	SetIDCardValidUntil       bool
	SetMaritalStatus          bool
}

type SubmitEmployeeProfileChangeRequestOutput struct {
	Request         *domain.EmployeeProfileChangeRequest
	ApprovalRequest *domain.ApprovalRequest
}

type SubmitEmployeeProfileChangeRequestUseCase struct {
	db                   ports.DB
	employeeRepo         ports.EmployeeRepository
	roleRepo             ports.RoleRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	changeRequestRepo    ports.EmployeeProfileChangeRequestRepository
	auditor              audit.Auditor
}

func NewSubmitEmployeeProfileChangeRequestUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	roleRepo ports.RoleRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository,
	auditor audit.Auditor,
) *SubmitEmployeeProfileChangeRequestUseCase {
	return &SubmitEmployeeProfileChangeRequestUseCase{
		db:                   db,
		employeeRepo:         employeeRepo,
		roleRepo:             roleRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		changeRequestRepo:    changeRequestRepo,
		auditor:              auditor,
	}
}

func (uc *SubmitEmployeeProfileChangeRequestUseCase) Execute(ctx context.Context, input SubmitEmployeeProfileChangeRequestInput) (*SubmitEmployeeProfileChangeRequestOutput, error) {
	if !input.SetFinancialGrade && !input.SetIDCardValidUntil && !input.SetMaritalStatus {
		return nil, ErrNoProfileChangesRequested
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	authorized, err := canSubmitOrViewEmployeeProfileChanges(ctx, tx, uc.roleRepo, input.ActorUserID)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	submitter, err := uc.employeeRepo.GetByUID(ctx, tx, input.ActorEmployeeUID)
	if err != nil {
		return nil, err
	}
	if submitter == nil {
		return nil, ErrEmployeeNotFound
	}

	hasPending, err := uc.changeRequestRepo.HasPendingForEmployee(ctx, tx, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrEmployeeProfileChangeAlreadyPending
	}

	flowStep, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, employeeProfileChangeApprovalFlowUID, 1)
	if err != nil {
		return nil, err
	}
	if flowStep == nil {
		return nil, ErrApprovalFlowStepNotFound
	}

	requestedFinancialGrade := employee.FinancialGrade
	if input.SetFinancialGrade {
		requestedFinancialGrade = trimStringPtr(input.RequestedFinancialGrade)
	}
	requestedIDCardValidUntil := employee.IDCardValidUntil
	if input.SetIDCardValidUntil {
		requestedIDCardValidUntil = input.RequestedIDCardValidUntil
	}
	requestedMaritalStatus := employee.MaritalStatus
	if input.SetMaritalStatus {
		requestedMaritalStatus = trimStringPtr(input.RequestedMaritalStatus)
	}

	if equalStringPointers(employee.FinancialGrade, requestedFinancialGrade) &&
		equalDatePointers(employee.IDCardValidUntil, requestedIDCardValidUntil) &&
		equalStringPointers(employee.MaritalStatus, requestedMaritalStatus) {
		return nil, ErrNoProfileChangesRequested
	}

	approvalRequest := domain.NewApprovalRequest(employeeProfileChangeApprovalFlowUID, input.ActorEmployeeUID, 1)
	if err := uc.approvalRequestRepo.Create(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	changeRequest := domain.NewEmployeeProfileChangeRequest(input.EmployeeUID, approvalRequest.UID, input.ActorEmployeeUID)
	changeRequest.CurrentFinancialGrade = cloneStringPtr(employee.FinancialGrade)
	changeRequest.CurrentIDCardValidUntil = cloneTimePtr(employee.IDCardValidUntil)
	changeRequest.CurrentMaritalStatus = cloneStringPtr(employee.MaritalStatus)
	changeRequest.RequestedFinancialGrade = cloneStringPtr(requestedFinancialGrade)
	changeRequest.RequestedIDCardValidUntil = cloneTimePtr(requestedIDCardValidUntil)
	changeRequest.RequestedMaritalStatus = cloneStringPtr(requestedMaritalStatus)
	changeRequest.Comments = trimStringPtr(input.Comments)

	if err := uc.changeRequestRepo.Create(ctx, tx, changeRequest); err != nil {
		return nil, err
	}

	action := domain.NewApprovalAction(approvalRequest.UID, domain.ApprovalActionTypeSubmit, nil, input.ActorEmployeeUID, changeRequest.Comments)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionSubmit).
		On(audit.EntityEmployeeProfileChangeRequest, changeRequest.UID).
		WithMeta("target_employee_uid", employee.UID).
		WithMeta("target_employee_name", employee.Name).
		WithMeta("submitted_by_employee_uid", submitter.UID).
		WithMeta("submitted_by_employee_name", submitter.Name).
		WithMeta("approval_request_uid", approvalRequest.UID).
		WithMeta("current_financial_grade", stringPtrValue(changeRequest.CurrentFinancialGrade)).
		WithMeta("requested_financial_grade", stringPtrValue(changeRequest.RequestedFinancialGrade)).
		WithMeta("current_id_card_valid_until", formatOptionalDate(changeRequest.CurrentIDCardValidUntil)).
		WithMeta("requested_id_card_valid_until", formatOptionalDate(changeRequest.RequestedIDCardValidUntil)).
		WithMeta("current_marital_status", stringPtrValue(changeRequest.CurrentMaritalStatus)).
		WithMeta("requested_marital_status", stringPtrValue(changeRequest.RequestedMaritalStatus)).
		WithMeta("comments", stringPtrValue(changeRequest.Comments)).
		Save(ctx)

	return &SubmitEmployeeProfileChangeRequestOutput{
		Request:         changeRequest,
		ApprovalRequest: approvalRequest,
	}, nil
}
