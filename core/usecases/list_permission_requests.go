package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListPermissionRequestsInput struct {
	Status         *domain.ApprovalRequestStatus
	Type           *domain.PermissionType
	EmployeeUID    *string
	DepartmentUIDs []string
	Page           int
	PageSize       int
}

type ListPermissionRequestsItem struct {
	PermissionRequest *domain.PermissionRequest
	ApprovalRequest   *domain.ApprovalRequest
	Employee          *domain.Employee
}

type ListPermissionRequestsOutput struct {
	Items      []*ListPermissionRequestsItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type ListPermissionRequestsUseCase struct {
	db                  ports.DB
	permissionRepo      ports.PermissionRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	employeeRepo        ports.EmployeeRepository
}

func NewListPermissionRequestsUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	employeeRepo ports.EmployeeRepository,
) *ListPermissionRequestsUseCase {
	return &ListPermissionRequestsUseCase{
		db:                  db,
		permissionRepo:      permissionRepo,
		approvalRequestRepo: approvalRequestRepo,
		employeeRepo:        employeeRepo,
	}
}

func (uc *ListPermissionRequestsUseCase) Execute(ctx context.Context, input ListPermissionRequestsInput) (*ListPermissionRequestsOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := ports.PermissionRequestListFilter{
		Status:         input.Status,
		Type:           input.Type,
		EmployeeUID:    input.EmployeeUID,
		DepartmentUIDs: input.DepartmentUIDs,
	}

	total, err := uc.permissionRepo.Count(ctx, uc.db, filter)
	if err != nil {
		return nil, err
	}

	requests, err := uc.permissionRepo.List(ctx, uc.db, filter, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*ListPermissionRequestsItem, 0, len(requests))
	for _, req := range requests {
		approval, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, req.ApprovalRequestUID)
		if err != nil {
			return nil, err
		}
		emp, err := uc.employeeRepo.GetByUID(ctx, uc.db, req.EmployeeUID)
		if err != nil {
			return nil, err
		}
		items = append(items, &ListPermissionRequestsItem{
			PermissionRequest: req,
			ApprovalRequest:   approval,
			Employee:          emp,
		})
	}

	return &ListPermissionRequestsOutput{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: ports.TotalPages(total, pageSize),
	}, nil
}
