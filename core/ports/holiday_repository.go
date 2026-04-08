package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type HolidayDefinitionRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.HolidayDefinition, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.HolidayDefinition, error)
	List(ctx context.Context, q Querier) ([]*domain.HolidayDefinition, error)
	Create(ctx context.Context, q Querier, def *domain.HolidayDefinition) error
}

type HolidayInstanceRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.HolidayInstance, error)
	GetByDefinitionAndYear(ctx context.Context, q Querier, definitionID int64, year int) (*domain.HolidayInstance, error)
	ListByYear(ctx context.Context, q Querier, year int) ([]*domain.HolidayInstance, error)
	ListByDateRange(ctx context.Context, q Querier, start, end time.Time) ([]*domain.HolidayInstance, error)
	Create(ctx context.Context, q Querier, instance *domain.HolidayInstance) error
	Update(ctx context.Context, q Querier, instance *domain.HolidayInstance) error
}
