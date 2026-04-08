package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type LeaveTypeRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.LeaveType, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.LeaveType, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.LeaveType, error)
	List(ctx context.Context, q Querier, activeOnly bool) ([]*domain.LeaveType, error)
	SetActive(ctx context.Context, q Querier, uid string, isActive bool) error
}
