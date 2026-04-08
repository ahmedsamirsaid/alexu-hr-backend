package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type LeaveBalanceTransactionRepository interface {
	Create(ctx context.Context, q Querier, tx *domain.LeaveBalanceTransaction) error
	ListByBalance(ctx context.Context, q Querier, balanceID int64) ([]*domain.LeaveBalanceTransaction, error)
}
