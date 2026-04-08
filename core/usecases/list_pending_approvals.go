package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type PendingApprovalItem struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
	Employee        *domain.Employee
	LeaveType       *domain.LeaveType
}

type ListPendingApprovalsOutput struct {
	Items []*PendingApprovalItem
}

type ListPendingApprovalsUseCase struct {
	db                   ports.DB
	userRepo             ports.UserRepository
	employeeRepo         ports.EmployeeRepository
	leaveTypeRepo        ports.LeaveTypeRepository
	leaveRequestRepo     ports.LeaveRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	roleRepo             ports.RoleRepository
}

func NewListPendingApprovalsUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
) *ListPendingApprovalsUseCase {
	return &ListPendingApprovalsUseCase{
		db:                   db,
		userRepo:             userRepo,
		employeeRepo:         employeeRepo,
		leaveTypeRepo:        leaveTypeRepo,
		leaveRequestRepo:     leaveRequestRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		roleRepo:             roleRepo,
	}
}

func (uc *ListPendingApprovalsUseCase) Execute(ctx context.Context, userID int64) (*ListPendingApprovalsOutput, error) {
	// Get all pending approval requests
	pendingRequests, err := uc.approvalRequestRepo.ListPending(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	var result []*PendingApprovalItem
	for _, approvalReq := range pendingRequests {
		// Get the step configuration for current step
		step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, uc.db, approvalReq.ApprovalFlowUID, approvalReq.CurrentStep)
		if err != nil {
			return nil, err
		}
		if step == nil {
			continue // Skip if step not configured
		}

		// Get the requester to check their department
		requester, err := uc.employeeRepo.GetByUID(ctx, uc.db, approvalReq.RequesterUID)
		if err != nil {
			return nil, err
		}
		if requester == nil || requester.DepartmentUID == nil {
			continue // Skip if requester has no department
		}

		// Check if current user is authorized to approve
		isAuthorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, uc.db, userID, step.RoleUID, *requester.DepartmentUID)
		if err != nil {
			return nil, err
		}
		if !isAuthorized {
			continue // User cannot approve this request
		}

		// Get the leave request
		leaveReq, err := uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, uc.db, approvalReq.UID)
		if err != nil {
			return nil, err
		}
		if leaveReq == nil {
			continue // Skip if no leave request (shouldn't happen)
		}

		// Get leave type
		leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, leaveReq.LeaveTypeUID)
		if err != nil {
			return nil, err
		}

		result = append(result, &PendingApprovalItem{
			LeaveRequest:    leaveReq,
			ApprovalRequest: approvalReq,
			Employee:        requester,
			LeaveType:       leaveType,
		})
	}

	return &ListPendingApprovalsOutput{Items: result}, nil
}
