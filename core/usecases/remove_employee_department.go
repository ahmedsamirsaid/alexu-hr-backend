package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type RemoveEmployeeDepartmentInput struct {
	EmployeeUID string
}

type RemoveEmployeeDepartmentUseCase struct {
	db      ports.DB
	empRepo ports.EmployeeRepository
}

func NewRemoveEmployeeDepartmentUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
) *RemoveEmployeeDepartmentUseCase {
	return &RemoveEmployeeDepartmentUseCase{
		db:      db,
		empRepo: empRepo,
	}
}

func (uc *RemoveEmployeeDepartmentUseCase) Execute(ctx context.Context, input RemoveEmployeeDepartmentInput) error {
	// Get and validate employee
	employee, err := uc.empRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return err
	}
	if employee == nil {
		return ErrEmployeeNotFound
	}

	// Remove department assignment (idempotent - no error if already null)
	employee.DepartmentUID = nil
	if err := uc.empRepo.Update(ctx, uc.db, employee); err != nil {
		return err
	}

	return nil
}
