package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
)

type DeleteApprovalFlowStepUseCase struct {
	db       ports.DB
	stepRepo ports.ApprovalFlowStepRepository
	auditor  audit.Auditor
}

func NewDeleteApprovalFlowStepUseCase(
	db ports.DB,
	stepRepo ports.ApprovalFlowStepRepository,
	auditor audit.Auditor,
) *DeleteApprovalFlowStepUseCase {
	return &DeleteApprovalFlowStepUseCase{
		db:       db,
		stepRepo: stepRepo,
		auditor:  auditor,
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

	// Capture step information for audit metadata before deletion
	approvalFlowUID := step.ApprovalFlowUID
	stepOrder := step.StepOrder
	roleUID := step.RoleUID

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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s deleted approval flow step %d",
		actorName, stepOrder,
	)

	// Audit log after successful deletion with step information
	defer uc.auditor.From(ctx).
		Did(audit.ActionDeleteStep).
		On(audit.EntityApprovalFlowStep, uid).
		WithMeta("action", actionSentence).
		WithMeta("approval_flow_uid", approvalFlowUID).
		WithMeta("step_order", stepOrder).
		WithMeta("role_uid", roleUID).
		Save(ctx)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
