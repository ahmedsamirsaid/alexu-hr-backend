package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewAssignEmployeeDepartmentUseCase(
	db ports.DB,
	empRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	auditor audit.Auditor,
) *AssignEmployeeDepartmentUseCase {
	return &AssignEmployeeDepartmentUseCase{
		db:       db,
		empRepo:  empRepo,
		deptRepo: deptRepo,
		auditor:  auditor,
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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s assigned %s (Employee) to department '%s'",
		actorName, employee.Name, department.NameEN,
	)
	
	// Audit log with department information
	defer uc.auditor.From(ctx).
		Did("assign_department").
		On(audit.EntityEmployee, input.EmployeeUID).
		WithMeta("action", actionSentence).
		WithMeta("employee_name", employee.Name).
		WithMeta("department_uid", input.DepartmentUID).
		WithMeta("department_name", department.NameEN).
		WithMeta("old_department_uid", employee.DepartmentUID).
		WithMeta("new_department_uid", input.DepartmentUID).
		Save(ctx)

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
