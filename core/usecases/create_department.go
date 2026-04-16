package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrDepartmentCodeExists = errors.New("department_code_exists")

type CreateDepartmentInput struct {
	Code            string
	NameEN          string
	NameAR          *string
	DefaultShiftUID *string
}

type CreateDepartmentOutput struct {
	Department *domain.Department
}

type CreateDepartmentUseCase struct {
	db       ports.DB
	deptRepo ports.DepartmentRepository
}

func NewCreateDepartmentUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
) *CreateDepartmentUseCase {
	return &CreateDepartmentUseCase{
		db:       db,
		deptRepo: deptRepo,
	}
}

func (uc *CreateDepartmentUseCase) Execute(ctx context.Context, input CreateDepartmentInput) (*CreateDepartmentOutput, error) {
	// Check if code already exists
	existing, err := uc.deptRepo.GetByCode(ctx, uc.db, input.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDepartmentCodeExists
	}

	department := domain.NewDepartment(input.Code, input.NameEN, input.NameAR)
	department.DefaultShiftUID = input.DefaultShiftUID

	if err := uc.deptRepo.Create(ctx, uc.db, department); err != nil {
		return nil, err
	}

	return &CreateDepartmentOutput{Department: department}, nil
}
