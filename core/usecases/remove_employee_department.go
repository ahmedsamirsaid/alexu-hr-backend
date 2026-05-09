package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
)

type RemoveEmployeeDepartmentInput struct {
	EmployeeUID string
}

type RemoveEmployeeDepartmentUseCase struct {
	db      ports.DB
	empRepo ports.EmployeeRepository
	auditor audit.Auditor
}

func NewRemoveEmployeeDepartmentUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
	auditor audit.Auditor,
) *RemoveEmployeeDepartmentUseCase {
	return &RemoveEmployeeDepartmentUseCase{
		db:      db,
		empRepo: empRepo,
		auditor: auditor,
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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionParams := map[string]interface{}{
		"Actor":    actorName,
		"Employee": employee.Name,
	}

	// Audit log with department information
	defer uc.auditor.From(ctx).
		Did("remove_department").
		On(audit.EntityEmployee, input.EmployeeUID).
		WithMeta("action_key", "audit.sentence.remove_employee_department").
		WithMeta("action_params", actionParams).
		WithMeta("employee_name", employee.Name).
		Save(ctx)

	// Remove department assignment (idempotent - no error if already null)
	employee.DepartmentUID = nil
	if err := uc.empRepo.Update(ctx, uc.db, employee); err != nil {
		return err
	}

	return nil
}
