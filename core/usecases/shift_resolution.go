package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

func resolveEffectiveShift(
	ctx context.Context,
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	employeeUID string,
	groupDepartmentUID *string,
) (*domain.Shift, error) {
	if employeeRepo != nil {
		employee, err := employeeRepo.GetByUID(ctx, db, employeeUID)
		if err != nil {
			return nil, err
		}
		if employee != nil {
			if employee.ShiftUID != nil {
				shift, err := shiftRepo.GetByUID(ctx, db, *employee.ShiftUID)
				if err != nil {
					return nil, err
				}
				if shift != nil {
					return shift, nil
				}
			}

			if employee.DepartmentUID != nil {
				groupDepartmentUID = employee.DepartmentUID
			}
		}
	}

	if deptRepo != nil && groupDepartmentUID != nil {
		department, err := deptRepo.GetByUID(ctx, db, *groupDepartmentUID)
		if err != nil {
			return nil, err
		}
		if department != nil && department.DefaultShiftUID != nil {
			shift, err := shiftRepo.GetByUID(ctx, db, *department.DefaultShiftUID)
			if err != nil {
				return nil, err
			}
			if shift != nil {
				return shift, nil
			}
		}
	}

	return domain.NewDefaultShift(), nil
}
