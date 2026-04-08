package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LeaveBalanceTransactionRepository struct{}

func NewLeaveBalanceTransactionRepository() *LeaveBalanceTransactionRepository {
	return &LeaveBalanceTransactionRepository{}
}

func (r *LeaveBalanceTransactionRepository) Create(ctx context.Context, q ports.Querier, tx *domain.LeaveBalanceTransaction) error {
	query := `
		INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	tx.CreatedAt = time.Now()

	result, err := q.ExecContext(ctx, query,
		tx.UID, tx.BalanceID, tx.TransactionType, tx.Days,
		tx.LeaveRecordID, tx.Notes, tx.CreatedBy, tx.CreatedAt)
	if err != nil {
		slog.Error("leave_balance_transaction_repository.Create.exec_query", "error", err, "uid", tx.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("leave_balance_transaction_repository.Create.last_insert_id", "error", err, "uid", tx.UID)
		return err
	}
	tx.ID = id

	return nil
}

func (r *LeaveBalanceTransactionRepository) ListByBalance(ctx context.Context, q ports.Querier, balanceID int64) ([]*domain.LeaveBalanceTransaction, error) {
	query := `
		SELECT id, uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at
		FROM leave_balance_transactions
		WHERE balance_id = ?
		ORDER BY created_at`

	rows, err := q.QueryContext(ctx, query, balanceID)
	if err != nil {
		slog.Error("leave_balance_transaction_repository.ListByBalance.query", "error", err, "balance_id", balanceID)
		return nil, err
	}
	defer rows.Close()

	var txs []*domain.LeaveBalanceTransaction
	for rows.Next() {
		t, err := r.scanTransactionRow(rows)
		if err != nil {
			slog.Error("leave_balance_transaction_repository.ListByBalance.scan_row", "error", err, "balance_id", balanceID)
			return nil, err
		}
		txs = append(txs, t)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_balance_transaction_repository.ListByBalance.rows_iteration", "error", err, "balance_id", balanceID)
		return nil, err
	}

	return txs, nil
}

func (r *LeaveBalanceTransactionRepository) scanTransactionRow(rows *sql.Rows) (*domain.LeaveBalanceTransaction, error) {
	var t domain.LeaveBalanceTransaction
	var createdAt domain.Time
	err := rows.Scan(
		&t.ID, &t.UID, &t.BalanceID, &t.TransactionType, &t.Days,
		&t.LeaveRecordID, &t.Notes, &t.CreatedBy, &createdAt)
	if err != nil {
		return nil, err
	}
	t.CreatedAt = createdAt.Time
	return &t, nil
}
