package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
)

type RemoveDepartmentManagerInput struct {
	DepartmentUID string
}

type RemoveDepartmentManagerUseCase struct {
	db       ports.DB
	deptRepo ports.DepartmentRepository
	roleRepo ports.RoleRepository
	userRepo ports.UserRepository
	empRepo  ports.EmployeeRepository
	auditor  audit.Auditor
}

func NewRemoveDepartmentManagerUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	roleRepo ports.RoleRepository,
	userRepo ports.UserRepository,
	empRepo ports.EmployeeRepository,
	auditor audit.Auditor,
) *RemoveDepartmentManagerUseCase {
	return &RemoveDepartmentManagerUseCase{
		db:       db,
		deptRepo: deptRepo,
		roleRepo: roleRepo,
		userRepo: userRepo,
		empRepo:  empRepo,
		auditor:  auditor,
	}
}

func (uc *RemoveDepartmentManagerUseCase) Execute(ctx context.Context, input RemoveDepartmentManagerInput) error {
	// Validate department exists
	department, err := uc.deptRepo.GetByUID(ctx, uc.db, input.DepartmentUID)
	if err != nil {
		return err
	}
	if department == nil {
		return ErrDepartmentNotFound
	}

	// Get current manager information for audit metadata (before removing)
	var managerUID string
	var managerName string
	currentManager, err := uc.roleRepo.GetDepartmentManager(ctx, uc.db, input.DepartmentUID)
	if err == nil && currentManager != nil {
		managerUID = currentManager.UID
		managerName = currentManager.Phone
		
		// Try to get employee name if available
		if currentManager.EmployeeUID != nil {
			employee, err := uc.empRepo.GetByUID(ctx, uc.db, *currentManager.EmployeeUID)
			if err == nil && employee != nil {
				managerName = employee.Name
			}
		}
	}

	// Audit log with manager information
	auditBuilder := uc.auditor.From(ctx).
		Did("remove_manager").
		On(audit.EntityDepartment, input.DepartmentUID)
	
	if managerUID != "" {
		auditBuilder = auditBuilder.
			WithMeta("manager_uid", managerUID).
			WithMeta("manager_name", managerName)
	}
	
	defer auditBuilder.Save(ctx)

	// Remove Department Manager role assignment for this department (idempotent)
	if err := uc.roleRepo.RemoveRoleFromUserForDepartment(ctx, uc.db, DepartmentManagerRoleUID, input.DepartmentUID); err != nil {
		return err
	}

	return nil
}
