package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type ShiftRepository interface {
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.Shift, error)
	List(ctx context.Context, q Querier) ([]*domain.Shift, error)
	Upsert(ctx context.Context, q Querier, shift *domain.Shift) error
}
