package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type WeekendConfigRepository interface {
	List(ctx context.Context, q Querier) ([]*domain.WeekendConfig, error)
	GetWeekendDays(ctx context.Context, q Querier) ([]int, error)
}
