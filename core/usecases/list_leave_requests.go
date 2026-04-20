package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListLeaveRequestsInput struct {
	EmployeeUID           *string
	ManagedDepartmentUIDs []string
	Status                *domain.ApprovalRequestStatus
	Limit                 int
	Offset                int
}

type LeaveRequestWithStatus struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
}

type ListLeaveRequestsOutput struct {
	Requests []*LeaveRequestWithStatus
	Total    int
}

type ListLeaveRequestsUseCase struct {
	db                  ports.DB
	leaveRequestRepo    ports.LeaveRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
}

func NewListLeaveRequestsUseCase(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
) *ListLeaveRequestsUseCase {
	return &ListLeaveRequestsUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
	}
}

func (uc *ListLeaveRequestsUseCase) Execute(ctx context.Context, input ListLeaveRequestsInput) (*ListLeaveRequestsOutput, error) {
	filter := ports.LeaveRequestListFilter{
		EmployeeUID: input.EmployeeUID,
		Status:      input.Status,
	}

	if input.EmployeeUID == nil && input.ManagedDepartmentUIDs != nil {
		filter.DepartmentUIDs = append([]string(nil), input.ManagedDepartmentUIDs...)
	}

	requests, err := uc.leaveRequestRepo.List(ctx, uc.db, filter, input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.leaveRequestRepo.Count(ctx, uc.db, filter)
	if err != nil {
		return nil, err
	}

	// Enrich with approval request status
	result := make([]*LeaveRequestWithStatus, len(requests))
	for i, lr := range requests {
		approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, lr.ApprovalRequestUID)
		if err != nil {
			return nil, err
		}
		result[i] = &LeaveRequestWithStatus{
			LeaveRequest:    lr,
			ApprovalRequest: approvalRequest,
		}
	}

	return &ListLeaveRequestsOutput{
		Requests: result,
		Total:    total,
	}, nil
}
