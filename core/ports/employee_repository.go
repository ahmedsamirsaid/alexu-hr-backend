package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

// EmployeeListFilter defines optional filters for listing employees.
type EmployeeListFilter struct {
	Status       *domain.EmployeeStatus
	HireDateFrom *time.Time
	HireDateTo   *time.Time
}

type EmployeeRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.Employee, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.Employee, error)
	GetByUIDs(ctx context.Context, q Querier, uids []string) ([]*domain.Employee, error)
	Create(ctx context.Context, q Querier, employee *domain.Employee) error
	Update(ctx context.Context, q Querier, employee *domain.Employee) error

	// List returns all employees matching the optional filter.
	List(ctx context.Context, q Querier, filter *EmployeeListFilter) ([]*domain.Employee, error)

	// ExistingGovernmentIDs returns the subset of governmentIDs that already exist in the database.
	ExistingGovernmentIDs(ctx context.Context, q Querier, governmentIDs []string) ([]string, error)

	// ExistingMobiles returns the subset of mobiles that already exist in the database.
	ExistingMobiles(ctx context.Context, q Querier, mobiles []string) ([]string, error)

	// ExistingUniversityIDs returns the subset of universityIDs that already exist in the database.
	ExistingUniversityIDs(ctx context.Context, q Querier, universityIDs []string) ([]string, error)

	// Count returns the total number of employees.
	Count(ctx context.Context, q Querier) (int, error)
}
