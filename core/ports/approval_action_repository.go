package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type ApprovalActionRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.ApprovalAction, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.ApprovalAction, error)
	Create(ctx context.Context, q Querier, action *domain.ApprovalAction) error

	// ListByRequest returns all actions for an approval request, ordered by acted_at
	ListByRequest(ctx context.Context, q Querier, approvalRequestUID string) ([]*domain.ApprovalAction, error)
}
