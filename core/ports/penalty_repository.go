package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type PenaltyRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.Penalty, error)
	ListByEmployeeUID(ctx context.Context, q Querier, employeeUID string) ([]*domain.Penalty, error)
	Create(ctx context.Context, q Querier, penalty *domain.Penalty) error
	IsRemoved(ctx context.Context, q Querier, penaltyID int64) (bool, error)
	CreateRemoval(ctx context.Context, q Querier, removal *domain.PenaltyRemoval) error
	DeleteByEmployeeUID(ctx context.Context, q Querier, employeeUID string) error
}
