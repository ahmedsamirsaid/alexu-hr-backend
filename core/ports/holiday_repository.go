package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type HolidayDefinitionRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.HolidayDefinition, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.HolidayDefinition, error)
	GetByDate(ctx context.Context, q Querier, date time.Time) ([]*domain.HolidayDefinition, error)
	ListByDateRange(ctx context.Context, q Querier, start, end time.Time) ([]*domain.HolidayDefinition, error)
	List(ctx context.Context, q Querier) ([]*domain.HolidayDefinition, error)
	Create(ctx context.Context, q Querier, def *domain.HolidayDefinition) error
}
