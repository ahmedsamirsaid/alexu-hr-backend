package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type IncentiveBonusRepository interface {
	ListByEmployeeUID(ctx context.Context, q Querier, employeeUID string) ([]*domain.IncentiveBonus, error)
	Create(ctx context.Context, q Querier, bonus *domain.IncentiveBonus) error
	DeleteByEmployeeUID(ctx context.Context, q Querier, employeeUID string) error
}
