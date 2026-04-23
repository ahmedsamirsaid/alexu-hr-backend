package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateApprovalFlowStepInput struct {
	UID       string
	StepOrder *int
	RoleUID   *string
}

type UpdateApprovalFlowStepOutput struct {
	Step *domain.ApprovalFlowStep
}

type UpdateApprovalFlowStepUseCase struct {
	db       ports.DB
	stepRepo ports.ApprovalFlowStepRepository
	roleRepo ports.RoleRepository
	auditor  audit.Auditor
}

func NewUpdateApprovalFlowStepUseCase(
	db ports.DB,
	stepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *UpdateApprovalFlowStepUseCase {
	return &UpdateApprovalFlowStepUseCase{
		db:       db,
		stepRepo: stepRepo,
		roleRepo: roleRepo,
		auditor:  auditor,
	}
}

func (uc *UpdateApprovalFlowStepUseCase) Execute(ctx context.Context, input UpdateApprovalFlowStepInput) (*UpdateApprovalFlowStepOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	step, err := uc.stepRepo.GetByUID(ctx, tx, input.UID)
	if err != nil {
		return nil, err
	}
	if step == nil {
		return nil, ErrApprovalFlowStepNotFound
	}

	// Capture old values for audit metadata
	oldStepOrder := step.StepOrder
	oldRoleUID := step.RoleUID

	// Build audit metadata with field changes
	auditBuilder := uc.auditor.From(ctx).Did(audit.ActionUpdateStep).On(audit.EntityApprovalFlowStep, input.UID)

	if input.StepOrder != nil {
		// Check for duplicate step order if changing
		if *input.StepOrder != step.StepOrder {
			existing, err := uc.stepRepo.GetByFlowAndStep(ctx, tx, step.ApprovalFlowUID, *input.StepOrder)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, ErrDuplicateStepOrder
			}
			auditBuilder.WithMeta("old_step_order", oldStepOrder).WithMeta("new_step_order", *input.StepOrder)
		}
		step.StepOrder = *input.StepOrder
	}

	if input.RoleUID != nil {
		// Verify role exists
		role, err := uc.roleRepo.GetByUID(ctx, tx, *input.RoleUID)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, ErrRoleNotFound
		}
		if *input.RoleUID != step.RoleUID {
			auditBuilder.WithMeta("old_role_uid", oldRoleUID).WithMeta("new_role_uid", *input.RoleUID).WithMeta("new_role_name", role.Name)
		}
		step.RoleUID = *input.RoleUID
	}

	if err := uc.stepRepo.Update(ctx, tx, step); err != nil {
		return nil, err
	}

	// Audit log after successful update
	defer auditBuilder.Save(ctx)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UpdateApprovalFlowStepOutput{Step: step}, nil
}
