package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type EmployeeProfileChangeRequestDetails struct {
	Request         *domain.EmployeeProfileChangeRequest
	ApprovalRequest *domain.ApprovalRequest
	Employee        *domain.Employee
	SubmittedBy     *domain.Employee
}

type GetEmployeeProfileChangeRequestUseCase struct {
	db                  ports.DB
	roleRepo            ports.RoleRepository
	changeRequestRepo   ports.EmployeeProfileChangeRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	employeeRepo        ports.EmployeeRepository
}

func NewGetEmployeeProfileChangeRequestUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	employeeRepo ports.EmployeeRepository,
) *GetEmployeeProfileChangeRequestUseCase {
	return &GetEmployeeProfileChangeRequestUseCase{
		db:                  db,
		roleRepo:            roleRepo,
		changeRequestRepo:   changeRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
		employeeRepo:        employeeRepo,
	}
}

func (uc *GetEmployeeProfileChangeRequestUseCase) Execute(ctx context.Context, actorUserID int64, requestUID string) (*EmployeeProfileChangeRequestDetails, error) {
	authorized, err := canSubmitOrViewEmployeeProfileChanges(ctx, uc.db, uc.roleRepo, actorUserID)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	request, err := uc.changeRequestRepo.GetByUID(ctx, uc.db, requestUID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, ErrEmployeeProfileChangeRequestNotFound
	}
	return uc.buildDetails(ctx, request)
}

func (uc *GetEmployeeProfileChangeRequestUseCase) buildDetails(ctx context.Context, request *domain.EmployeeProfileChangeRequest) (*EmployeeProfileChangeRequestDetails, error) {
	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, request.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, request.EmployeeUID)
	if err != nil {
		return nil, err
	}
	submittedBy, err := uc.employeeRepo.GetByUID(ctx, uc.db, request.SubmittedByEmployeeUID)
	if err != nil {
		return nil, err
	}
	return &EmployeeProfileChangeRequestDetails{
		Request:         request,
		ApprovalRequest: approvalRequest,
		Employee:        employee,
		SubmittedBy:     submittedBy,
	}, nil
}

type ListEmployeeProfileChangeRequestsUseCase struct {
	db                ports.DB
	roleRepo          ports.RoleRepository
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository
	getUC             *GetEmployeeProfileChangeRequestUseCase
}

func NewListEmployeeProfileChangeRequestsUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository,
	getUC *GetEmployeeProfileChangeRequestUseCase,
) *ListEmployeeProfileChangeRequestsUseCase {
	return &ListEmployeeProfileChangeRequestsUseCase{db: db, roleRepo: roleRepo, changeRequestRepo: changeRequestRepo, getUC: getUC}
}

func (uc *ListEmployeeProfileChangeRequestsUseCase) Execute(ctx context.Context, actorUserID int64, employeeUID string) ([]*EmployeeProfileChangeRequestDetails, error) {
	authorized, err := canSubmitOrViewEmployeeProfileChanges(ctx, uc.db, uc.roleRepo, actorUserID)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	requests, err := uc.changeRequestRepo.ListByEmployeeUID(ctx, uc.db, employeeUID)
	if err != nil {
		return nil, err
	}
	items := make([]*EmployeeProfileChangeRequestDetails, 0, len(requests))
	for _, request := range requests {
		item, err := uc.getUC.buildDetails(ctx, request)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

type ListPendingEmployeeProfileChangeRequestsUseCase struct {
	db                ports.DB
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository
	roleRepo          ports.RoleRepository
	getUC             *GetEmployeeProfileChangeRequestUseCase
}

func NewListPendingEmployeeProfileChangeRequestsUseCase(
	db ports.DB,
	changeRequestRepo ports.EmployeeProfileChangeRequestRepository,
	roleRepo ports.RoleRepository,
	getUC *GetEmployeeProfileChangeRequestUseCase,
) *ListPendingEmployeeProfileChangeRequestsUseCase {
	return &ListPendingEmployeeProfileChangeRequestsUseCase{
		db:                db,
		changeRequestRepo: changeRequestRepo,
		roleRepo:          roleRepo,
		getUC:             getUC,
	}
}

func (uc *ListPendingEmployeeProfileChangeRequestsUseCase) Execute(ctx context.Context, actorUserID int64) ([]*EmployeeProfileChangeRequestDetails, error) {
	authorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, uc.db, actorUserID, informationCenterRoleUID, "")
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrNotAuthorizedApprover
	}

	requests, err := uc.changeRequestRepo.ListPending(ctx, uc.db)
	if err != nil {
		return nil, err
	}
	items := make([]*EmployeeProfileChangeRequestDetails, 0, len(requests))
	for _, request := range requests {
		item, err := uc.getUC.buildDetails(ctx, request)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
