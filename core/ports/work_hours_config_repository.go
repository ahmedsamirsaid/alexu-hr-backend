package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type WorkHoursConfigRepository interface {
	Get(ctx context.Context, q Querier) (*domain.WorkHoursConfig, error)
	Upsert(ctx context.Context, q Querier, cfg *domain.WorkHoursConfig) error
}
