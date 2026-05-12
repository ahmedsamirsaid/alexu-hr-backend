package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
)

// Leadership role names that are valid manager targets (above department manager).
var leadershipManagerRoleNames = map[string]bool{
	"University President": true,
	"Vice President":       true,
	"Dean":                 true,
}

type SetEmployeeManagerInput struct {
	EmployeeUID string
	ManagerUID  *string // nil = remove manager (make root)
}

type SetEmployeeManagerUseCase struct {
	db          ports.DB
	empRepo     ports.EmployeeRepository
	roleRepo    ports.RoleRepository
	auditor     audit.Auditor
}

func NewSetEmployeeManagerUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *SetEmployeeManagerUseCase {
	return &SetEmployeeManagerUseCase{db: db, empRepo: empRepo, roleRepo: roleRepo, auditor: auditor}
}

func (uc *SetEmployeeManagerUseCase) Execute(ctx context.Context, input SetEmployeeManagerInput) error {
	employee, err := uc.empRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return err
	}
	if employee == nil {
		return ErrEmployeeNotFound
	}

	// Determine if the employee is ordinary (no non-employee role).
	roleNames, err := uc.roleRepo.GetRoleNamesByEmployeeUIDs(ctx, uc.db, []string{employee.UID})
	if err != nil {
		return err
	}
	if roleNames[employee.UID] == "" {
		return errors.New("cannot_change_manager_for_regular_employee")
	}

	// Validate the proposed manager.
	if input.ManagerUID != nil {
		mgr, err := uc.empRepo.GetByUID(ctx, uc.db, *input.ManagerUID)
		if err != nil {
			return err
		}
		if mgr == nil {
			return ErrEmployeeNotFound
		}
		if mgr.UID == employee.UID {
			return errors.New("employee_cannot_be_own_manager")
		}
		mgrRoles, err := uc.roleRepo.GetRoleNamesByEmployeeUIDs(ctx, uc.db, []string{mgr.UID})
		if err != nil {
			return err
		}
		if !leadershipManagerRoleNames[mgrRoles[mgr.UID]] {
			return errors.New("manager_must_have_leadership_role")
		}
	}

	employee.ManagerUID = input.ManagerUID
	if err := uc.empRepo.Update(ctx, uc.db, employee); err != nil {
		return err
	}

	defer uc.auditor.From(ctx).
		Did("set_manager").
		On("employee", employee.UID).
		WithMeta("manager_uid", input.ManagerUID).
		Save(ctx)

	return nil
}
