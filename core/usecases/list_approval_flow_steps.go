package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListApprovalFlowStepsOutput struct {
	Steps []*domain.ApprovalFlowStep
}

type ListApprovalFlowStepsUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
	stepRepo ports.ApprovalFlowStepRepository
}

func NewListApprovalFlowStepsUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
	stepRepo ports.ApprovalFlowStepRepository,
) *ListApprovalFlowStepsUseCase {
	return &ListApprovalFlowStepsUseCase{
		db:       db,
		flowRepo: flowRepo,
		stepRepo: stepRepo,
	}
}

func (uc *ListApprovalFlowStepsUseCase) Execute(ctx context.Context, flowUID string) (*ListApprovalFlowStepsOutput, error) {
	// Verify flow exists
	flow, err := uc.flowRepo.GetByUID(ctx, uc.db, flowUID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrApprovalFlowNotFound
	}

	steps, err := uc.stepRepo.ListByFlow(ctx, uc.db, flowUID)
	if err != nil {
		return nil, err
	}

	return &ListApprovalFlowStepsOutput{Steps: steps}, nil
}
