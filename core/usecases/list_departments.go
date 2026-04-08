package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListDepartmentsOutput struct {
	Departments []*domain.Department
}

type ListDepartmentsUseCase struct {
	db      ports.DB
	deptRepo ports.DepartmentRepository
}

func NewListDepartmentsUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
) *ListDepartmentsUseCase {
	return &ListDepartmentsUseCase{
		db:       db,
		deptRepo: deptRepo,
	}
}

func (uc *ListDepartmentsUseCase) Execute(ctx context.Context, activeOnly bool) (*ListDepartmentsOutput, error) {
	departments, err := uc.deptRepo.List(ctx, uc.db, activeOnly)
	if err != nil {
		return nil, err
	}

	return &ListDepartmentsOutput{Departments: departments}, nil
}
