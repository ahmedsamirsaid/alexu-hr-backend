package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type ApprovalFlowStepRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.ApprovalFlowStep, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.ApprovalFlowStep, error)
	Create(ctx context.Context, q Querier, step *domain.ApprovalFlowStep) error
	Update(ctx context.Context, q Querier, step *domain.ApprovalFlowStep) error
	Delete(ctx context.Context, q Querier, uid string) error

	// ListByFlow returns all steps for an approval flow, ordered by step_order
	ListByFlow(ctx context.Context, q Querier, approvalFlowUID string) ([]*domain.ApprovalFlowStep, error)

	// CountByFlow returns the number of steps in an approval flow
	CountByFlow(ctx context.Context, q Querier, approvalFlowUID string) (int, error)

	// GetByFlowAndStep returns the step at a given order for a flow
	GetByFlowAndStep(ctx context.Context, q Querier, approvalFlowUID string, stepOrder int) (*domain.ApprovalFlowStep, error)

	// HasPendingRequestsAtStep checks if there are pending approval requests at this step
	HasPendingRequestsAtStep(ctx context.Context, q Querier, stepUID string) (bool, error)
}
