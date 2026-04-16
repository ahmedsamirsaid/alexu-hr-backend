package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateDepartmentInput struct {
	UID             string
	Code            *string
	NameEN          *string
	NameAR          *string
	IsActive        *bool
	DefaultShiftUID *string
}

type UpdateDepartmentOutput struct {
	Department *domain.Department
}

type UpdateDepartmentUseCase struct {
	db       ports.DB
	deptRepo ports.DepartmentRepository
}

func NewUpdateDepartmentUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
) *UpdateDepartmentUseCase {
	return &UpdateDepartmentUseCase{
		db:       db,
		deptRepo: deptRepo,
	}
}

func (uc *UpdateDepartmentUseCase) Execute(ctx context.Context, input UpdateDepartmentInput) (*UpdateDepartmentOutput, error) {
	department, err := uc.deptRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrDepartmentNotFound
	}

	// Check if new code conflicts with existing department
	if input.Code != nil && *input.Code != department.Code {
		existing, err := uc.deptRepo.GetByCode(ctx, uc.db, *input.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrDepartmentCodeExists
		}
		department.Code = *input.Code
	}

	if input.NameEN != nil {
		department.NameEN = *input.NameEN
	}

	if input.NameAR != nil {
		department.NameAR = input.NameAR
	}

	if input.IsActive != nil {
		department.IsActive = *input.IsActive
	}

	if input.DefaultShiftUID != nil {
		department.DefaultShiftUID = input.DefaultShiftUID
	}

	if err := uc.deptRepo.Update(ctx, uc.db, department); err != nil {
		return nil, err
	}

	return &UpdateDepartmentOutput{Department: department}, nil
}
