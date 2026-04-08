package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/ports"
)

const DepartmentManagerRoleUID = "role_department_manager"

var ErrDepartmentInactive = errors.New("department_inactive")

type AssignDepartmentManagerInput struct {
	DepartmentUID string
	UserUID       string
}

type AssignDepartmentManagerUseCase struct {
	db       ports.DB
	deptRepo ports.DepartmentRepository
	userRepo ports.UserRepository
	roleRepo ports.RoleRepository
}

func NewAssignDepartmentManagerUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
) *AssignDepartmentManagerUseCase {
	return &AssignDepartmentManagerUseCase{
		db:       db,
		deptRepo: deptRepo,
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (uc *AssignDepartmentManagerUseCase) Execute(ctx context.Context, input AssignDepartmentManagerInput) error {
	// Get and validate department
	department, err := uc.deptRepo.GetByUID(ctx, uc.db, input.DepartmentUID)
	if err != nil {
		return err
	}
	if department == nil {
		return ErrDepartmentNotFound
	}
	if !department.IsActive {
		return ErrDepartmentInactive
	}

	// Get and validate user
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Get the Department Manager role
	role, err := uc.roleRepo.GetByUID(ctx, uc.db, DepartmentManagerRoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("department_manager_role_not_found")
	}

	// Remove existing manager for this department (if any)
	if err := uc.roleRepo.RemoveRoleFromUserForDepartment(ctx, uc.db, DepartmentManagerRoleUID, input.DepartmentUID); err != nil {
		return err
	}

	// Assign the role to the new user scoped to this department
	departmentUID := input.DepartmentUID
	if err := uc.roleRepo.AssignRoleToUserWithDepartment(ctx, uc.db, user.ID, role.ID, &departmentUID); err != nil {
		return err
	}

	return nil
}
