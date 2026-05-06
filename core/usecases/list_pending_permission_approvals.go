package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type PendingPermissionApprovalItem struct {
	PermissionRequest *domain.PermissionRequest
	ApprovalRequest   *domain.ApprovalRequest
	Employee          *domain.Employee
}

type ListPendingPermissionApprovalsOutput struct {
	Items []*PendingPermissionApprovalItem
}

type ListPendingPermissionApprovalsUseCase struct {
	db                   ports.DB
	permissionRepo       ports.PermissionRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	employeeRepo         ports.EmployeeRepository
	roleRepo             ports.RoleRepository
}

func NewListPendingPermissionApprovalsUseCase(
	db ports.DB,
	permissionRepo ports.PermissionRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	employeeRepo ports.EmployeeRepository,
	roleRepo ports.RoleRepository,
) *ListPendingPermissionApprovalsUseCase {
	return &ListPendingPermissionApprovalsUseCase{
		db:                   db,
		permissionRepo:       permissionRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		employeeRepo:         employeeRepo,
		roleRepo:             roleRepo,
	}
}

func (uc *ListPendingPermissionApprovalsUseCase) Execute(ctx context.Context, userID int64) (*ListPendingPermissionApprovalsOutput, error) {
	pending, err := uc.approvalRequestRepo.ListPending(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	out := make([]*PendingPermissionApprovalItem, 0)
	for _, ar := range pending {
		if ar.ApprovalFlowUID != permissionDefaultFlowUID {
			continue
		}
		step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, uc.db, ar.ApprovalFlowUID, ar.CurrentStep)
		if err != nil {
			return nil, err
		}
		if step == nil {
			continue
		}
		requester, err := uc.employeeRepo.GetByUID(ctx, uc.db, ar.RequesterUID)
		if err != nil {
			return nil, err
		}
		if requester == nil {
			continue
		}
		departmentUID := ""
		if requester.DepartmentUID != nil {
			departmentUID = *requester.DepartmentUID
		}
		authorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, uc.db, userID, step.RoleUID, departmentUID)
		if err != nil {
			return nil, err
		}
		if !authorized {
			continue
		}
		permission, err := uc.permissionRepo.GetByApprovalRequestUID(ctx, uc.db, ar.UID)
		if err != nil {
			return nil, err
		}
		if permission == nil {
			continue
		}
		out = append(out, &PendingPermissionApprovalItem{
			PermissionRequest: permission,
			ApprovalRequest:   ar,
			Employee:          requester,
		})
	}
	return &ListPendingPermissionApprovalsOutput{Items: out}, nil
}
