package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewUpdateDepartmentUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	auditor audit.Auditor,
) *UpdateDepartmentUseCase {
	return &UpdateDepartmentUseCase{
		db:       db,
		deptRepo: deptRepo,
		auditor:  auditor,
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

	// Capture old values for audit metadata
	oldCode := department.Code
	oldNameEN := department.NameEN
	oldNameAR := ""
	if department.NameAR != nil {
		oldNameAR = *department.NameAR
	}
	oldIsActive := department.IsActive

	auditBuilder := uc.auditor.From(ctx).
		Did(audit.ActionUpdate).
		On(audit.EntityDepartment, input.UID).
		WithMeta("old_state", map[string]interface{}{
			"code":      oldCode,
			"name_en":   oldNameEN,
			"name_ar":   oldNameAR,
			"is_active": oldIsActive,
		})

	var changedFields []string

	if input.Code != nil && *input.Code != department.Code {
		existing, err := uc.deptRepo.GetByCode(ctx, uc.db, *input.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrDepartmentCodeExists
		}
		auditBuilder = auditBuilder.WithMeta("old_code", oldCode).WithMeta("new_code", *input.Code)
		changedFields = append(changedFields, fmt.Sprintf("code (%s → %s)", oldCode, *input.Code))
		department.Code = *input.Code
	}

	if input.NameEN != nil && *input.NameEN != department.NameEN {
		auditBuilder = auditBuilder.WithMeta("old_name_en", oldNameEN).WithMeta("new_name_en", *input.NameEN)
		changedFields = append(changedFields, fmt.Sprintf("name_en (%s → %s)", oldNameEN, *input.NameEN))
		department.NameEN = *input.NameEN
	}

	if input.NameAR != nil {
		newNameAR := ""
		if input.NameAR != nil {
			newNameAR = *input.NameAR
		}
		if newNameAR != oldNameAR {
			auditBuilder = auditBuilder.WithMeta("old_name_ar", oldNameAR).WithMeta("new_name_ar", newNameAR)
			changedFields = append(changedFields, "name_ar")
		}
		department.NameAR = input.NameAR
	}

	if input.IsActive != nil && *input.IsActive != department.IsActive {
		auditBuilder = auditBuilder.WithMeta("old_is_active", oldIsActive).WithMeta("new_is_active", *input.IsActive)
		status := "deactivated"
		if *input.IsActive {
			status = "activated"
		}
		changedFields = append(changedFields, status)
		department.IsActive = *input.IsActive
	}

	if input.DefaultShiftUID != nil {
		department.DefaultShiftUID = input.DefaultShiftUID
	}

	// Build human-readable action sentence
	actionSentence := fmt.Sprintf("Department '%s' was updated", department.NameEN)
	if len(changedFields) > 0 {
		actionSentence = fmt.Sprintf("Department '%s' was updated: ", department.NameEN)
		for i, f := range changedFields {
			if i > 0 {
				actionSentence += ", "
			}
			actionSentence += f
		}
	}
	auditBuilder = auditBuilder.WithMeta("action", actionSentence)

	defer auditBuilder.Save(ctx)

	if err := uc.deptRepo.Update(ctx, uc.db, department); err != nil {
		return nil, err
	}

	return &UpdateDepartmentOutput{Department: department}, nil
}
