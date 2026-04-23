package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CreateApprovalFlowInput struct {
	Code        string
	NameEN      string
	NameAR      *string
	Description *string
}

type CreateApprovalFlowOutput struct {
	Flow *domain.ApprovalFlow
}

type CreateApprovalFlowUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
	auditor  audit.Auditor
}

func NewCreateApprovalFlowUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
	auditor audit.Auditor,
) *CreateApprovalFlowUseCase {
	return &CreateApprovalFlowUseCase{
		db:       db,
		flowRepo: flowRepo,
		auditor:  auditor,
	}
}

func (uc *CreateApprovalFlowUseCase) Execute(ctx context.Context, input CreateApprovalFlowInput) (*CreateApprovalFlowOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check for duplicate code
	existing, err := uc.flowRepo.GetByCode(ctx, tx, input.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrApprovalFlowCodeExists
	}

	flow := domain.NewApprovalFlow(input.Code, input.NameEN, input.NameAR, input.Description)

	if err := uc.flowRepo.Create(ctx, tx, flow); err != nil {
		return nil, err
	}

	// Audit log after successful creation
	defer uc.auditor.From(ctx).Did(audit.ActionCreate).On(audit.EntityApprovalFlow, flow.UID).Save(ctx)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateApprovalFlowOutput{Flow: flow}, nil
}
