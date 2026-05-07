package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetPermissionRequestInput struct {
	UID         string
	ActorUserID int64
}

type GetPermissionRequestOutput struct {
	PermissionRequest *domain.PermissionRequest
	ApprovalRequest   *domain.ApprovalRequest
	Employee          *domain.Employee
	// ActorCanApprove indicates whether the calling user is currently authorised to act on this
	// request — used by the UI to decide whether to show the Approve button. True for the
	// request's assigned approver, including the lone-UHR self-approval case.
	ActorCanApprove bool
}

type GetPermissionRequestUseCase struct {
	db                   ports.DB
	permissionRepo       ports.PermissionRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	employeeRepo         ports.EmployeeRepository
	roleRepo             ports.RoleRepository
}

func NewGetPermissionRequestUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	employeeRepo ports.EmployeeRepository,
	roleRepo ports.RoleRepository,
) *GetPermissionRequestUseCase {
	return &GetPermissionRequestUseCase{
		db:                   db,
		permissionRepo:       permissionRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		employeeRepo:         employeeRepo,
		roleRepo:             roleRepo,
	}
}

func (uc *GetPermissionRequestUseCase) Execute(ctx context.Context, input GetPermissionRequestInput) (*GetPermissionRequestOutput, error) {
	req, err := uc.permissionRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, ErrPermissionRequestNotFound
	}
	approval, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, req.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}
	emp, err := uc.employeeRepo.GetByUID(ctx, uc.db, req.EmployeeUID)
	if err != nil {
		return nil, err
	}

	actorCanApprove := false
	if approval != nil && approval.IsPending() && emp != nil && input.ActorUserID > 0 {
		step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, uc.db, approval.ApprovalFlowUID, approval.CurrentStep)
		if err != nil {
			return nil, err
		}
		if step != nil {
			deptUID := ""
			if emp.DepartmentUID != nil {
				deptUID = *emp.DepartmentUID
			}
			ok, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, uc.db, input.ActorUserID, step.RoleUID, deptUID)
			if err != nil {
				return nil, err
			}
			actorCanApprove = ok
		}
	}

	return &GetPermissionRequestOutput{
		PermissionRequest: req,
		ApprovalRequest:   approval,
		Employee:          emp,
		ActorCanApprove:   actorCanApprove,
	}, nil
}
