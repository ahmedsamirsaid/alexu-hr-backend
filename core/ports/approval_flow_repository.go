package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type ApprovalFlowRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.ApprovalFlow, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.ApprovalFlow, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.ApprovalFlow, error)
	Create(ctx context.Context, q Querier, flow *domain.ApprovalFlow) error
	Update(ctx context.Context, q Querier, flow *domain.ApprovalFlow) error
	List(ctx context.Context, q Querier, activeOnly bool) ([]*domain.ApprovalFlow, error)
}
