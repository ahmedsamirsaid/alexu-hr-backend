package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type RemoveDepartmentManagerInput struct {
	DepartmentUID string
}

type RemoveDepartmentManagerUseCase struct {
	db       ports.DB
	deptRepo ports.DepartmentRepository
	roleRepo ports.RoleRepository
}

func NewRemoveDepartmentManagerUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	roleRepo ports.RoleRepository,
) *RemoveDepartmentManagerUseCase {
	return &RemoveDepartmentManagerUseCase{
		db:       db,
		deptRepo: deptRepo,
		roleRepo: roleRepo,
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

	// Remove Department Manager role assignment for this department (idempotent)
	if err := uc.roleRepo.RemoveRoleFromUserForDepartment(ctx, uc.db, DepartmentManagerRoleUID, input.DepartmentUID); err != nil {
		return err
	}

	return nil
}
