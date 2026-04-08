package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type DeleteApprovalFlowStepUseCase struct {
	db       ports.DB
	stepRepo ports.ApprovalFlowStepRepository
}

func NewDeleteApprovalFlowStepUseCase(
	db ports.DB,
	stepRepo ports.ApprovalFlowStepRepository,
) *DeleteApprovalFlowStepUseCase {
	return &DeleteApprovalFlowStepUseCase{
		db:       db,
		stepRepo: stepRepo,
	}
}

func (uc *DeleteApprovalFlowStepUseCase) Execute(ctx context.Context, uid string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	step, err := uc.stepRepo.GetByUID(ctx, tx, uid)
	if err != nil {
		return err
	}
	if step == nil {
		return ErrApprovalFlowStepNotFound
	}

	// Check for pending requests at this step
	hasPending, err := uc.stepRepo.HasPendingRequestsAtStep(ctx, tx, uid)
	if err != nil {
		return err
	}
	if hasPending {
		return ErrStepHasPendingRequests
	}

	if err := uc.stepRepo.Delete(ctx, tx, uid); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
