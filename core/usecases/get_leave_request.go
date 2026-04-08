package usecases

import (
	"context"
	"strings"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetLeaveRequestOutput struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
	Employee        *domain.Employee
	LeaveType       *domain.LeaveType
}

type GetLeaveRequestUseCase struct {
	db                  ports.DB
	leaveRequestRepo    ports.LeaveRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	employeeRepo        ports.EmployeeRepository
	leaveTypeRepo       ports.LeaveTypeRepository
}

func NewGetLeaveRequestUseCase(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
) *GetLeaveRequestUseCase {
	return &GetLeaveRequestUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
		employeeRepo:        employeeRepo,
		leaveTypeRepo:       leaveTypeRepo,
	}
}

func (uc *GetLeaveRequestUseCase) Execute(ctx context.Context, uid string) (*GetLeaveRequestOutput, error) {
	var leaveRequest *domain.LeaveRequest
	var err error

	// Try to find by leave request UID first
	leaveRequest, err = uc.leaveRequestRepo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}

	// If not found and UID looks like an approval request UID, try that
	if leaveRequest == nil && strings.HasPrefix(uid, "apr_") {
		leaveRequest, err = uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, uc.db, uid)
		if err != nil {
			return nil, err
		}
	}

	if leaveRequest == nil {
		return nil, ErrLeaveRequestNotFound
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, leaveRequest.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, leaveRequest.EmployeeUID)
	if err != nil {
		return nil, err
	}

	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, leaveRequest.LeaveTypeUID)
	if err != nil {
		return nil, err
	}

	return &GetLeaveRequestOutput{
		LeaveRequest:    leaveRequest,
		ApprovalRequest: approvalRequest,
		Employee:        employee,
		LeaveType:       leaveType,
	}, nil
}
