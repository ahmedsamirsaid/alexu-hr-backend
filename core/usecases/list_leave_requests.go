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
	LeaveType       *domain.LeaveType
	SubLeaveType    *domain.SubLeaveType
}

type ListLeaveRequestsOutput struct {
	Requests []*LeaveRequestWithStatus
	Total    int
}

type ListLeaveRequestsUseCase struct {
	db                  ports.DB
	leaveRequestRepo    ports.LeaveRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	leaveTypeRepo       ports.LeaveTypeRepository
}

func NewListLeaveRequestsUseCase(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
) *ListLeaveRequestsUseCase {
	return &ListLeaveRequestsUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
		leaveTypeRepo:       leaveTypeRepo,
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

		leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, lr.LeaveTypeUID)
		if err != nil {
			return nil, err
		}

		var subLeaveType *domain.SubLeaveType
		if lr.SubLeaveTypeUID != nil && *lr.SubLeaveTypeUID != "" {
			subLeaveType, err = uc.leaveTypeRepo.GetSubLeaveTypeByUID(ctx, uc.db, *lr.SubLeaveTypeUID)
			if err != nil {
				return nil, err
			}
		}

		result[i] = &LeaveRequestWithStatus{
			LeaveRequest:    lr,
			ApprovalRequest: approvalRequest,
			LeaveType:       leaveType,
			SubLeaveType:    subLeaveType,
		}
	}

	return &ListLeaveRequestsOutput{
		Requests: result,
		Total:    total,
	}, nil
}
