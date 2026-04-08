package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrDepartmentNotFound = errors.New("department_not_found")

type ManagerInfo struct {
	User     *domain.User
	Employee *domain.Employee
}

type GetDepartmentOutput struct {
	Department *domain.Department
	Manager    *ManagerInfo
}

type GetDepartmentUseCase struct {
	db       ports.DB
	deptRepo ports.DepartmentRepository
	roleRepo ports.RoleRepository
	empRepo  ports.EmployeeRepository
}

func NewGetDepartmentUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	roleRepo ports.RoleRepository,
	empRepo ports.EmployeeRepository,
) *GetDepartmentUseCase {
	return &GetDepartmentUseCase{
		db:       db,
		deptRepo: deptRepo,
		roleRepo: roleRepo,
		empRepo:  empRepo,
	}
}

func (uc *GetDepartmentUseCase) Execute(ctx context.Context, uid string) (*GetDepartmentOutput, error) {
	department, err := uc.deptRepo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrDepartmentNotFound
	}

	// Get department manager (user with Department Manager role scoped to this department)
	user, err := uc.roleRepo.GetDepartmentManager(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}

	var managerInfo *ManagerInfo
	if user != nil {
		managerInfo = &ManagerInfo{User: user}
		// If user has an associated employee, fetch the employee details
		if user.EmployeeUID != nil {
			employee, err := uc.empRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
			if err == nil && employee != nil {
				managerInfo.Employee = employee
			}
		}
	}

	return &GetDepartmentOutput{
		Department: department,
		Manager:    managerInfo,
	}, nil
}
