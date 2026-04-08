package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListApprovalFlowsOutput struct {
	Flows []*domain.ApprovalFlow
}

type ListApprovalFlowsUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
}

func NewListApprovalFlowsUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
) *ListApprovalFlowsUseCase {
	return &ListApprovalFlowsUseCase{
		db:       db,
		flowRepo: flowRepo,
	}
}

func (uc *ListApprovalFlowsUseCase) Execute(ctx context.Context, activeOnly bool) (*ListApprovalFlowsOutput, error) {
	flows, err := uc.flowRepo.List(ctx, uc.db, activeOnly)
	if err != nil {
		return nil, err
	}

	return &ListApprovalFlowsOutput{Flows: flows}, nil
}
