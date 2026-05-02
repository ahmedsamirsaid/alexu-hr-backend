package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type DepartmentRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.Department, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.Department, error)
	GetByUIDs(ctx context.Context, q Querier, uids []string) ([]*domain.Department, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.Department, error)
	Create(ctx context.Context, q Querier, department *domain.Department) error
	Update(ctx context.Context, q Querier, department *domain.Department) error
	List(ctx context.Context, q Querier, activeOnly bool) ([]*domain.Department, error)
}
