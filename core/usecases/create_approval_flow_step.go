package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CreateApprovalFlowStepInput struct {
	ApprovalFlowUID string
	StepOrder       int
	RoleUID         string
}

type CreateApprovalFlowStepOutput struct {
	Step *domain.ApprovalFlowStep
}

type CreateApprovalFlowStepUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
	stepRepo ports.ApprovalFlowStepRepository
	roleRepo ports.RoleRepository
}

func NewCreateApprovalFlowStepUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
	stepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
) *CreateApprovalFlowStepUseCase {
	return &CreateApprovalFlowStepUseCase{
		db:       db,
		flowRepo: flowRepo,
		stepRepo: stepRepo,
		roleRepo: roleRepo,
	}
}

func (uc *CreateApprovalFlowStepUseCase) Execute(ctx context.Context, input CreateApprovalFlowStepInput) (*CreateApprovalFlowStepOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Verify flow exists
	flow, err := uc.flowRepo.GetByUID(ctx, tx, input.ApprovalFlowUID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrApprovalFlowNotFound
	}

	// Verify role exists
	role, err := uc.roleRepo.GetByUID(ctx, tx, input.RoleUID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	// Check for duplicate step order
	existingStep, err := uc.stepRepo.GetByFlowAndStep(ctx, tx, input.ApprovalFlowUID, input.StepOrder)
	if err != nil {
		return nil, err
	}
	if existingStep != nil {
		return nil, ErrDuplicateStepOrder
	}

	step := domain.NewApprovalFlowStep(input.ApprovalFlowUID, input.StepOrder, input.RoleUID)

	if err := uc.stepRepo.Create(ctx, tx, step); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateApprovalFlowStepOutput{Step: step}, nil
}
