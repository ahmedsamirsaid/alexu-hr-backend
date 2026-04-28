package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewCreateDepartmentUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	auditor audit.Auditor,
) *CreateDepartmentUseCase {
	return &CreateDepartmentUseCase{
		db:       db,
		deptRepo: deptRepo,
		auditor:  auditor,
	}
}

func (uc *CreateDepartmentUseCase) Execute(ctx context.Context, input CreateDepartmentInput) (*CreateDepartmentOutput, error) {
	existing, err := uc.deptRepo.GetByCode(ctx, uc.db, input.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDepartmentCodeExists
	}

	department := domain.NewDepartment(input.Code, input.NameEN, input.NameAR)
	department.DefaultShiftUID = input.DefaultShiftUID

	actionSentence := fmt.Sprintf("Department '%s' (code: %s) was created", input.NameEN, input.Code)

	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityDepartment, department.UID).
		WithMeta("action", actionSentence).
		WithMeta("code", input.Code).
		WithMeta("name_en", input.NameEN).
		WithMeta("new_state", map[string]interface{}{
			"uid":     department.UID,
			"code":    input.Code,
			"name_en": input.NameEN,
		}).
		Save(ctx)

	if err := uc.deptRepo.Create(ctx, uc.db, department); err != nil {
		return nil, err
	}

	return &CreateDepartmentOutput{Department: department}, nil
}
