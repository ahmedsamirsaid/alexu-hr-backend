package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/audit"
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
	empRepo  ports.EmployeeRepository
	auditor  audit.Auditor
}

func NewAssignDepartmentManagerUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	empRepo ports.EmployeeRepository,
	auditor audit.Auditor,
) *AssignDepartmentManagerUseCase {
	return &AssignDepartmentManagerUseCase{
		db:       db,
		deptRepo: deptRepo,
		userRepo: userRepo,
		roleRepo: roleRepo,
		empRepo:  empRepo,
		auditor:  auditor,
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

	// Get employee information for audit metadata
	var managerName string
	if user.EmployeeUID != nil {
		employee, err := uc.empRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
		if err == nil && employee != nil {
			managerName = employee.Name
		}
	}
	if managerName == "" {
		managerName = user.Phone
	}

	// Get the Department Manager role
	role, err := uc.roleRepo.GetByUID(ctx, uc.db, DepartmentManagerRoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("department_manager_role_not_found")
	}

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionParams := map[string]interface{}{
		"Actor":      actorName,
		"Manager":    managerName,
		"Department": department.NameEN,
	}

	// Audit log with manager information
	defer uc.auditor.From(ctx).
		Did("assign_manager").
		On(audit.EntityDepartment, input.DepartmentUID).
		WithMeta("action_key", "audit.sentence.assign_department_manager").
		WithMeta("action_params", actionParams).
		WithMeta("department_name", department.NameEN).
		WithMeta("manager_name", managerName).
		Save(ctx)

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
