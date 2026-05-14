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
		WHERE id = $1`

	return r.scanLeaveType(q.QueryRowContext(ctx, query, id))
}

func (r *LeaveTypeRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types
		WHERE uid = $1`

	return r.scanLeaveType(q.QueryRowContext(ctx, query, uid))
}

func (r *LeaveTypeRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types
		WHERE code = $1`

	return r.scanLeaveType(q.QueryRowContext(ctx, query, code))
}

func (r *LeaveTypeRepository) GetSubLeaveTypeByUID(ctx context.Context, q ports.Querier, uid string) (*domain.SubLeaveType, error) {
	query := `
		SELECT id, uid, leave_type_uid, name_en, name_ar, created_at, updated_at
		FROM sub_leave_types
		WHERE uid = $1`

	return r.scanSubLeaveType(q.QueryRowContext(ctx, query, uid))
}

func (r *LeaveTypeRepository) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.LeaveType, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive,
		       recording_deadline_days, advance_notice_days, is_active, approval_flow_uid, created_at, updated_at
		FROM leave_types`

	if activeOnly {
		query += ` WHERE is_active = true`
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

func (r *LeaveTypeRepository) ListSubLeaveTypesByLeaveTypeUID(ctx context.Context, q ports.Querier, leaveTypeUID string) ([]*domain.SubLeaveType, error) {
	query := `
		SELECT id, uid, leave_type_uid, name_en, name_ar, created_at, updated_at
		FROM sub_leave_types
		WHERE leave_type_uid = $1
		ORDER BY id`

	rows, err := q.QueryContext(ctx, query, leaveTypeUID)
	if err != nil {
		slog.Error("leave_type_repository.ListSubLeaveTypesByLeaveTypeUID.query", "error", err, "leave_type_uid", leaveTypeUID)
		return nil, err
	}
	defer rows.Close()

	var types []*domain.SubLeaveType
	for rows.Next() {
		subLeaveType, err := r.scanSubLeaveTypeRow(rows)
		if err != nil {
			slog.Error("leave_type_repository.ListSubLeaveTypesByLeaveTypeUID.scan_row", "error", err, "leave_type_uid", leaveTypeUID)
			return nil, err
		}
		types = append(types, subLeaveType)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_type_repository.ListSubLeaveTypesByLeaveTypeUID.rows_iteration", "error", err, "leave_type_uid", leaveTypeUID)
		return nil, err
	}

	return types, nil
}

func (r *LeaveTypeRepository) Update(ctx context.Context, q ports.Querier, leaveType *domain.LeaveType) error {
	query := `
		UPDATE leave_types
		SET code = $1, name_en = $2, name_ar = $3, default_balance = $4, max_consecutive = $5,
		    recording_deadline_days = $6, advance_notice_days = $7, is_active = $8, approval_flow_uid = $9,
		    updated_at = NOW()
		WHERE uid = $10`

	result, err := q.ExecContext(
		ctx,
		query,
		leaveType.Code,
		leaveType.NameEN,
		leaveType.NameAR,
		leaveType.DefaultBalance,
		leaveType.MaxConsecutive,
		leaveType.RecordingDeadlineDays,
		leaveType.AdvanceNoticeDays,
		leaveType.IsActive,
		leaveType.ApprovalFlowUID,
		leaveType.UID,
	)
	if err != nil {
		slog.Error("leave_type_repository.Update.exec_query", "error", err, "uid", leaveType.UID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("leave_type_repository.Update.rows_affected", "error", err, "uid", leaveType.UID)
		return err
	}

	if rowsAffected == 0 {
		return ErrLeaveTypeNotFound
	}

	updated, err := r.GetByUID(ctx, q, leaveType.UID)
	if err != nil {
		return err
	}
	if updated != nil {
		*leaveType = *updated
	}

	return nil
}

func (r *LeaveTypeRepository) SetActive(ctx context.Context, q ports.Querier, uid string, isActive bool) error {
	query := `UPDATE leave_types SET is_active = $1, updated_at = NOW() WHERE uid = $2`

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


func (r *LeaveTypeRepository) scanSubLeaveType(row *sql.Row) (*domain.SubLeaveType, error) {
	var subLeaveType domain.SubLeaveType
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&subLeaveType.ID, &subLeaveType.UID, &subLeaveType.LeaveTypeUID, &subLeaveType.NameEN, &subLeaveType.NameAR,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("leave_type_repository.scanSubLeaveType.scan_row", "error", err)
		return nil, err
	}
	subLeaveType.CreatedAt = createdAt.Time
	subLeaveType.UpdatedAt = updatedAt.Time
	return &subLeaveType, nil
}

func (r *LeaveTypeRepository) scanSubLeaveTypeRow(rows *sql.Rows) (*domain.SubLeaveType, error) {
	var subLeaveType domain.SubLeaveType
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&subLeaveType.ID, &subLeaveType.UID, &subLeaveType.LeaveTypeUID, &subLeaveType.NameEN, &subLeaveType.NameAR,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	subLeaveType.CreatedAt = createdAt.Time
	subLeaveType.UpdatedAt = updatedAt.Time
	return &subLeaveType, nil
}
