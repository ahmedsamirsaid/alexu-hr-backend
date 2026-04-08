package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type ApprovalRequestRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.ApprovalRequest, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.ApprovalRequest, error)
	Create(ctx context.Context, q Querier, request *domain.ApprovalRequest) error
	Update(ctx context.Context, q Querier, request *domain.ApprovalRequest) error

	// ListByRequester returns all approval requests for an employee
	ListByRequester(ctx context.Context, q Querier, requesterUID string) ([]*domain.ApprovalRequest, error)

	// ListPending returns all pending approval requests
	ListPending(ctx context.Context, q Querier) ([]*domain.ApprovalRequest, error)

	// ListPendingByFlowAndStep returns pending requests at a specific step for a flow
	ListPendingByFlowAndStep(ctx context.Context, q Querier, approvalFlowUID string, stepOrder int) ([]*domain.ApprovalRequest, error)

	// CountPendingByFlowAndStep counts pending requests at a specific step for a flow
	CountPendingByFlowAndStep(ctx context.Context, q Querier, approvalFlowUID string, stepOrder int) (int, error)
}
