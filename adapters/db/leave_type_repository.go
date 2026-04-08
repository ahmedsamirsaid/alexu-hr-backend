package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrLeaveTypeNotFound = errors.New("leave_type_not_found")

type LeaveTypeRepository struct{}

func NewLeaveTypeRepository() *LeaveTypeRepository {
	return &LeaveTypeRepository{}
}

func (r *LeaveTypeRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types
		WHERE id = ?`

	return r.scanLeaveType(q.QueryRowContext(ctx, query, id))
}

func (r *LeaveTypeRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types
		WHERE uid = ?`

	return r.scanLeaveType(q.QueryRowContext(ctx, query, uid))
}

func (r *LeaveTypeRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types
		WHERE code = ?`

	return r.scanLeaveType(q.QueryRowContext(ctx, query, code))
}

func (r *LeaveTypeRepository) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types`

	if activeOnly {
		query += ` WHERE is_active = 1`
	}

	query += ` ORDER BY id`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("leave_type_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var types []*domain.LeaveType
	for rows.Next() {
		lt, err := r.scanLeaveTypeRow(rows)
		if err != nil {
			slog.Error("leave_type_repository.List.scan_row", "error", err)
			return nil, err
		}
		types = append(types, lt)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_type_repository.List.rows_iteration", "error", err)
		return nil, err
	}

	return types, nil
}

func (r *LeaveTypeRepository) SetActive(ctx context.Context, q ports.Querier, uid string, isActive bool) error {
	query := `UPDATE leave_types SET is_active = ?, updated_at = datetime('now') WHERE uid = ?`

	result, err := q.ExecContext(ctx, query, isActive, uid)
	if err != nil {
		slog.Error("leave_type_repository.SetActive.exec_query", "error", err, "uid", uid)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("leave_type_repository.SetActive.rows_affected", "error", err, "uid", uid)
		return err
	}

	if rowsAffected == 0 {
		return ErrLeaveTypeNotFound
	}

	return nil
}

func (r *LeaveTypeRepository) scanLeaveType(row *sql.Row) (*domain.LeaveType, error) {
	var lt domain.LeaveType
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&lt.ID, &lt.UID, &lt.Code, &lt.NameEN, &lt.NameAR,
		&lt.DefaultBalance, &lt.MaxConsecutive,
		&lt.RecordingDeadlineDays, &lt.AdvanceNoticeDays, &lt.IsActive, &lt.ApprovalFlowUID,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("leave_type_repository.scanLeaveType.scan_row", "error", err)
		return nil, err
	}
	lt.CreatedAt = createdAt.Time
	lt.UpdatedAt = updatedAt.Time
	return &lt, nil
}

func (r *LeaveTypeRepository) scanLeaveTypeRow(rows *sql.Rows) (*domain.LeaveType, error) {
	var lt domain.LeaveType
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&lt.ID, &lt.UID, &lt.Code, &lt.NameEN, &lt.NameAR,
		&lt.DefaultBalance, &lt.MaxConsecutive,
		&lt.RecordingDeadlineDays, &lt.AdvanceNoticeDays, &lt.IsActive, &lt.ApprovalFlowUID,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	lt.CreatedAt = createdAt.Time
	lt.UpdatedAt = updatedAt.Time
	return &lt, nil
}
