package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type AssignEmployeeDepartmentInput struct {
	EmployeeUID   string
	DepartmentUID string
	ShiftUID      *string
}

type AssignEmployeeDepartmentUseCase struct {
	db       ports.DB
	empRepo  ports.EmployeeRepository
	deptRepo ports.DepartmentRepository
}

func NewAssignEmployeeDepartmentUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
) *AssignEmployeeDepartmentUseCase {
	return &AssignEmployeeDepartmentUseCase{
		db:       db,
		empRepo:  empRepo,
		deptRepo: deptRepo,
	}
}

func (uc *AssignEmployeeDepartmentUseCase) Execute(ctx context.Context, input AssignEmployeeDepartmentInput) error {
	// Get and validate employee
	employee, err := uc.empRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return err
	}
	if employee == nil {
		return ErrEmployeeNotFound
	}

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

	// Update employee's department
	employee.DepartmentUID = &input.DepartmentUID
	if input.ShiftUID != nil {
		employee.ShiftUID = input.ShiftUID
	} else if employee.ShiftUID == nil && department.DefaultShiftUID != nil {
		employee.ShiftUID = department.DefaultShiftUID
	}
	if err := uc.empRepo.Update(ctx, uc.db, employee); err != nil {
		return err
	}

	return nil
}
