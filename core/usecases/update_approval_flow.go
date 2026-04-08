package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateApprovalFlowInput struct {
	UID         string
	NameEN      *string
	NameAR      *string
	Description *string
	IsActive    *bool
}

type UpdateApprovalFlowOutput struct {
	Flow *domain.ApprovalFlow
}

type UpdateApprovalFlowUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
}

func NewUpdateApprovalFlowUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
) *UpdateApprovalFlowUseCase {
	return &UpdateApprovalFlowUseCase{
		db:       db,
		flowRepo: flowRepo,
	}
}

func (uc *UpdateApprovalFlowUseCase) Execute(ctx context.Context, input UpdateApprovalFlowInput) (*UpdateApprovalFlowOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	flow, err := uc.flowRepo.GetByUID(ctx, tx, input.UID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrApprovalFlowNotFound
	}

	if input.NameEN != nil {
		flow.NameEN = *input.NameEN
	}
	if input.NameAR != nil {
		flow.NameAR = input.NameAR
	}
	if input.Description != nil {
		flow.Description = input.Description
	}
	if input.IsActive != nil {
		flow.IsActive = *input.IsActive
	}

	if err := uc.flowRepo.Update(ctx, tx, flow); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UpdateApprovalFlowOutput{Flow: flow}, nil
}
