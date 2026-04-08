package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type LeaveBalanceRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.LeaveBalance, error)
	GetByEmployeeAndTypeAndYear(ctx context.Context, q Querier, employeeID, leaveTypeID int64, year int) (*domain.LeaveBalance, error)
	Create(ctx context.Context, q Querier, balance *domain.LeaveBalance) error
	Update(ctx context.Context, q Querier, balance *domain.LeaveBalance) error
	ListByEmployee(ctx context.Context, q Querier, employeeID int64) ([]*domain.LeaveBalance, error)
	ListByEmployeeAndYear(ctx context.Context, q Querier, employeeID int64, year int) ([]*domain.LeaveBalance, error)
}
