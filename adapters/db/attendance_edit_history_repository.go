package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type AttendanceEditHistoryRepository struct{}

func NewAttendanceEditHistoryRepository() *AttendanceEditHistoryRepository {
	return &AttendanceEditHistoryRepository{}
}

func (r *AttendanceEditHistoryRepository) Create(ctx context.Context, q ports.Querier, history *domain.AttendanceEditHistory) error {
	query := `
		INSERT INTO attendance_edit_history (
			uid, attendance_record_uid, field_changed, old_value, new_value, reason, edited_by_uid, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now().UTC()
	}

	result, err := q.ExecContext(
		ctx,
		query,
		history.UID,
		history.AttendanceRecordUID,
		history.FieldChanged,
		history.OldValue,
		history.NewValue,
		history.Reason,
		history.EditedByUID,
		history.CreatedAt,
	)
	if err != nil {
		slog.Error("attendance_edit_history_repository.Create.exec_query", "error", err, "uid", history.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		history.ID = id
	}

	return nil
}

func (r *AttendanceEditHistoryRepository) ListByAttendanceRecordUID(ctx context.Context, q ports.Querier, attendanceRecordUID string) ([]*ports.AttendanceEditHistoryWithEditor, error) {
	query := `
		SELECT
			aeh.id,
			aeh.uid,
			aeh.attendance_record_uid,
			aeh.field_changed,
			aeh.old_value,
			aeh.new_value,
			aeh.reason,
			aeh.edited_by_uid,
			aeh.created_at,
			COALESCE(editor_employee.name, editor_user.phone, aeh.edited_by_uid) AS edited_by_name,
			editor_user.phone,
			editor_user.employee_uid
		FROM attendance_edit_history aeh
		LEFT JOIN users editor_user ON editor_user.uid = aeh.edited_by_uid
		LEFT JOIN employees editor_employee ON editor_employee.uid = editor_user.employee_uid
		WHERE aeh.attendance_record_uid = ?
		ORDER BY datetime(aeh.created_at) ASC, aeh.id ASC`

	rows, err := q.QueryContext(ctx, query, attendanceRecordUID)
	if err != nil {
		slog.Error("attendance_edit_history_repository.ListByAttendanceRecordUID.query", "error", err, "attendance_record_uid", attendanceRecordUID)
		return nil, err
	}
	defer rows.Close()

	items := make([]*ports.AttendanceEditHistoryWithEditor, 0)
	for rows.Next() {
		item, err := r.scanHistoryRow(rows)
		if err != nil {
			slog.Error("attendance_edit_history_repository.ListByAttendanceRecordUID.scan_row", "error", err, "attendance_record_uid", attendanceRecordUID)
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		slog.Error("attendance_edit_history_repository.ListByAttendanceRecordUID.rows_err", "error", err, "attendance_record_uid", attendanceRecordUID)
		return nil, err
	}

	return items, nil
}

func (r *AttendanceEditHistoryRepository) scanHistoryRow(rows *sql.Rows) (*ports.AttendanceEditHistoryWithEditor, error) {
	var history domain.AttendanceEditHistory
	var createdAt domain.Time
	var oldValue sql.NullString
	var newValue sql.NullString
	var reason sql.NullString
	var editedByName string
	var editedByPhone sql.NullString
	var editedByEmployeeUID sql.NullString

	if err := rows.Scan(
		&history.ID,
		&history.UID,
		&history.AttendanceRecordUID,
		&history.FieldChanged,
		&oldValue,
		&newValue,
		&reason,
		&history.EditedByUID,
		&createdAt,
		&editedByName,
		&editedByPhone,
		&editedByEmployeeUID,
	); err != nil {
		return nil, err
	}

	history.CreatedAt = createdAt.Time
	history.OldValue = nullableStringPtr(oldValue)
	history.NewValue = nullableStringPtr(newValue)
	history.Reason = nullableStringPtr(reason)

	return &ports.AttendanceEditHistoryWithEditor{
		History:            &history,
		EditedByName:       editedByName,
		EditedByPhone:      nullableStringPtr(editedByPhone),
		EditedByEmployeeUID: nullableStringPtr(editedByEmployeeUID),
	}, nil
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
