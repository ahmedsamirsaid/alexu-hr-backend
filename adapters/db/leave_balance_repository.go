package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LeaveBalanceRepository struct{}

func NewLeaveBalanceRepository() *LeaveBalanceRepository {
	return &LeaveBalanceRepository{}
}

func (r *LeaveBalanceRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveBalance, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at
		FROM leave_balances
		WHERE id = ?`

	return r.scanLeaveBalance(q.QueryRowContext(ctx, query, id))
}

func (r *LeaveBalanceRepository) GetByEmployeeAndTypeAndYear(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64, year int) (*domain.LeaveBalance, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at
		FROM leave_balances
		WHERE employee_id = ? AND leave_type_id = ? AND year = ?`

	return r.scanLeaveBalance(q.QueryRowContext(ctx, query, employeeID, leaveTypeID, year))
}

func (r *LeaveBalanceRepository) Create(ctx context.Context, q ports.Querier, balance *domain.LeaveBalance) error {
	query := `
		INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	balance.CreatedAt = now
	balance.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		balance.UID, balance.EmployeeID, balance.LeaveTypeID, balance.Year,
		balance.TotalDays, balance.UsedDays, balance.CreatedAt, balance.UpdatedAt)
	if err != nil {
		slog.Error("leave_balance_repository.Create.exec_query", "error", err, "uid", balance.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("leave_balance_repository.Create.last_insert_id", "error", err, "uid", balance.UID)
		return err
	}
	balance.ID = id

	return nil
}

func (r *LeaveBalanceRepository) Update(ctx context.Context, q ports.Querier, balance *domain.LeaveBalance) error {
	query := `
		UPDATE leave_balances
		SET total_days = ?, used_days = ?, updated_at = ?
		WHERE id = ?`

	balance.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		balance.TotalDays, balance.UsedDays, balance.UpdatedAt, balance.ID)
	if err != nil {
		slog.Error("leave_balance_repository.Update.exec_query", "error", err, "uid", balance.UID)
	}

	return err
}

func (r *LeaveBalanceRepository) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveBalance, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at
		FROM leave_balances
		WHERE employee_id = ?
		ORDER BY year DESC, leave_type_id`

	return r.queryLeaveBalances(ctx, q, query, employeeID)
}

func (r *LeaveBalanceRepository) ListByEmployeeAndYear(ctx context.Context, q ports.Querier, employeeID int64, year int) ([]*domain.LeaveBalance, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at
		FROM leave_balances
		WHERE employee_id = ? AND year = ?
		ORDER BY leave_type_id`

	return r.queryLeaveBalances(ctx, q, query, employeeID, year)
}

func (r *LeaveBalanceRepository) queryLeaveBalances(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.LeaveBalance, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("leave_balance_repository.queryLeaveBalances.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var balances []*domain.LeaveBalance
	for rows.Next() {
		b, err := r.scanLeaveBalanceRow(rows)
		if err != nil {
			slog.Error("leave_balance_repository.queryLeaveBalances.scan_row", "error", err)
			return nil, err
		}
		balances = append(balances, b)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_balance_repository.queryLeaveBalances.rows_iteration", "error", err)
		return nil, err
	}

	return balances, nil
}

func (r *LeaveBalanceRepository) scanLeaveBalance(row *sql.Row) (*domain.LeaveBalance, error) {
	var b domain.LeaveBalance
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&b.ID, &b.UID, &b.EmployeeID, &b.LeaveTypeID, &b.Year,
		&b.TotalDays, &b.UsedDays, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("leave_balance_repository.scanLeaveBalance.scan_row", "error", err)
		return nil, err
	}
	b.CreatedAt = createdAt.Time
	b.UpdatedAt = updatedAt.Time
	return &b, nil
}

func (r *LeaveBalanceRepository) scanLeaveBalanceRow(rows *sql.Rows) (*domain.LeaveBalance, error) {
	var b domain.LeaveBalance
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&b.ID, &b.UID, &b.EmployeeID, &b.LeaveTypeID, &b.Year,
		&b.TotalDays, &b.UsedDays, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	b.CreatedAt = createdAt.Time
	b.UpdatedAt = updatedAt.Time
	return &b, nil
}
