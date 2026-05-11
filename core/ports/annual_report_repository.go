package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type AnnualReportRepository interface {
	ListByEmployeeUID(ctx context.Context, q Querier, employeeUID string) ([]*domain.AnnualReport, error)
	Create(ctx context.Context, q Querier, report *domain.AnnualReport) error
	DeleteByEmployeeUID(ctx context.Context, q Querier, employeeUID string) error
}
