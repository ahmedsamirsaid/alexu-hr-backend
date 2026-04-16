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

type LeaveRecordRepository struct{}

func NewLeaveRecordRepository() *LeaveRecordRepository {
	return &LeaveRecordRepository{}
}

func (r *LeaveRecordRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRecord, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at
		FROM leave_records
		WHERE id = ?`

	return r.scanLeaveRecord(q.QueryRowContext(ctx, query, id))
}

func (r *LeaveRecordRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRecord, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at
		FROM leave_records
		WHERE uid = ?`

	return r.scanLeaveRecord(q.QueryRowContext(ctx, query, uid))
}

func (r *LeaveRecordRepository) Create(ctx context.Context, q ports.Querier, record *domain.LeaveRecord) error {
	query := `
		INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		record.UID, record.EmployeeID, record.LeaveTypeID,
		record.StartDate, record.EndDate, record.Days,
		record.RecordedAt, record.RecordedBy, record.Notes, record.LeaveRequestUID,
		record.CreatedAt, record.UpdatedAt)
	if err != nil {
		slog.Error("leave_record_repository.Create.exec_query", "error", err, "uid", record.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("leave_record_repository.Create.last_insert_id", "error", err, "uid", record.UID)
		return err
	}
	record.ID = id

	return nil
}

func (r *LeaveRecordRepository) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveRecord, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at
		FROM leave_records
		WHERE employee_id = ?
		ORDER BY start_date DESC`

	return r.queryLeaveRecords(ctx, q, query, employeeID)
}

func (r *LeaveRecordRepository) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeID int64, limit, offset int) ([]*domain.LeaveRecord, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at
		FROM leave_records
		WHERE employee_id = ?
		ORDER BY start_date DESC
		LIMIT ? OFFSET ?`

	return r.queryLeaveRecords(ctx, q, query, employeeID, limit, offset)
}

func (r *LeaveRecordRepository) CountByEmployee(ctx context.Context, q ports.Querier, employeeID int64) (int, error) {
	query := `SELECT COUNT(*) FROM leave_records WHERE employee_id = ?`

	var count int
	err := q.QueryRowContext(ctx, query, employeeID).Scan(&count)
	if err != nil {
		slog.Error("leave_record_repository.CountByEmployee.scan", "error", err, "employee_id", employeeID)
		return 0, err
	}
	return count, nil
}

func (r *LeaveRecordRepository) ListByEmployeeAndDateRange(ctx context.Context, q ports.Querier, employeeID int64, start, end time.Time) ([]*domain.LeaveRecord, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at
		FROM leave_records
		WHERE employee_id = ? AND start_date >= ? AND end_date <= ?
		ORDER BY start_date DESC`

	return r.queryLeaveRecords(ctx, q, query, employeeID, start, end)
}

func (r *LeaveRecordRepository) ListByEmployeeAndType(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64) ([]*domain.LeaveRecord, error) {
	query := `
		SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at
		FROM leave_records
		WHERE employee_id = ? AND leave_type_id = ?
		ORDER BY start_date DESC`

	return r.queryLeaveRecords(ctx, q, query, employeeID, leaveTypeID)
}

func (r *LeaveRecordRepository) queryLeaveRecords(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.LeaveRecord, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("leave_record_repository.queryLeaveRecords.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var records []*domain.LeaveRecord
	for rows.Next() {
		rec, err := r.scanLeaveRecordRow(rows)
		if err != nil {
			slog.Error("leave_record_repository.queryLeaveRecords.scan_row", "error", err)
			return nil, err
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_record_repository.queryLeaveRecords.rows_iteration", "error", err)
		return nil, err
	}

	return records, nil
}

func (r *LeaveRecordRepository) scanLeaveRecord(row *sql.Row) (*domain.LeaveRecord, error) {
	var rec domain.LeaveRecord
	var startDate, endDate, recordedAt, createdAt, updatedAt domain.Time
	err := row.Scan(
		&rec.ID, &rec.UID, &rec.EmployeeID, &rec.LeaveTypeID,
		&startDate, &endDate, &rec.Days,
		&recordedAt, &rec.RecordedBy, &rec.Notes, &rec.LeaveRequestUID,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("leave_record_repository.scanLeaveRecord.scan_row", "error", err)
		return nil, err
	}
	rec.StartDate = startDate.Time
	rec.EndDate = endDate.Time
	rec.RecordedAt = recordedAt.Time
	rec.CreatedAt = createdAt.Time
	rec.UpdatedAt = updatedAt.Time
	return &rec, nil
}

func (r *LeaveRecordRepository) scanLeaveRecordRow(rows *sql.Rows) (*domain.LeaveRecord, error) {
	var rec domain.LeaveRecord
	var startDate, endDate, recordedAt, createdAt, updatedAt domain.Time
	err := rows.Scan(
		&rec.ID, &rec.UID, &rec.EmployeeID, &rec.LeaveTypeID,
		&startDate, &endDate, &rec.Days,
		&recordedAt, &rec.RecordedBy, &rec.Notes, &rec.LeaveRequestUID,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	rec.StartDate = startDate.Time
	rec.EndDate = endDate.Time
	rec.RecordedAt = recordedAt.Time
	rec.CreatedAt = createdAt.Time
	rec.UpdatedAt = updatedAt.Time
	return &rec, nil
}

func (r *LeaveRecordRepository) ListAllPaginated(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter, limit, offset int) ([]*ports.LeaveRecordWithEmployee, error) {
	query := `
		SELECT lr.id, lr.uid, lr.employee_id, lr.leave_type_id, lr.start_date, lr.end_date, lr.days, lr.recorded_at, lr.recorded_by, lr.notes, lr.leave_request_uid, lr.created_at, lr.updated_at,
		       e.uid, e.name
		FROM leave_records lr
		JOIN employees e ON lr.employee_id = e.id
		WHERE 1=1`

	args := []any{}

	if filter.Search != "" {
		query += ` AND e.name LIKE ?`
		args = append(args, "%"+filter.Search+"%")
	}

	if filter.LeaveTypeID != nil {
		query += ` AND lr.leave_type_id = ?`
		args = append(args, *filter.LeaveTypeID)
	}

	if filter.StartDate != nil {
		query += ` AND lr.start_date >= ?`
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		query += ` AND lr.start_date <= ?`
		args = append(args, *filter.EndDate)
	}

	query += ` ORDER BY lr.start_date DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("leave_record_repository.ListAllPaginated.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var results []*ports.LeaveRecordWithEmployee
	for rows.Next() {
		var rec domain.LeaveRecord
		var startDate, endDate, recordedAt, createdAt, updatedAt domain.Time
		var employeeUID, employeeName string

		err := rows.Scan(
			&rec.ID, &rec.UID, &rec.EmployeeID, &rec.LeaveTypeID,
			&startDate, &endDate, &rec.Days,
			&recordedAt, &rec.RecordedBy, &rec.Notes, &rec.LeaveRequestUID,
			&createdAt, &updatedAt,
			&employeeUID, &employeeName)
		if err != nil {
			slog.Error("leave_record_repository.ListAllPaginated.scan_row", "error", err)
			return nil, err
		}

		rec.StartDate = startDate.Time
		rec.EndDate = endDate.Time
		rec.RecordedAt = recordedAt.Time
		rec.CreatedAt = createdAt.Time
		rec.UpdatedAt = updatedAt.Time

		results = append(results, &ports.LeaveRecordWithEmployee{
			LeaveRecord:  &rec,
			EmployeeUID:  employeeUID,
			EmployeeName: employeeName,
		})
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_record_repository.ListAllPaginated.rows_iteration", "error", err)
		return nil, err
	}

	return results, nil
}

func (r *LeaveRecordRepository) CountAll(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM leave_records lr
		JOIN employees e ON lr.employee_id = e.id
		WHERE 1=1`

	args := []any{}

	if filter.Search != "" {
		query += ` AND e.name LIKE ?`
		args = append(args, "%"+filter.Search+"%")
	}

	if filter.LeaveTypeID != nil {
		query += ` AND lr.leave_type_id = ?`
		args = append(args, *filter.LeaveTypeID)
	}

	if filter.StartDate != nil {
		query += ` AND lr.start_date >= ?`
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		query += ` AND lr.start_date <= ?`
		args = append(args, *filter.EndDate)
	}

	var count int
	err := q.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.Error("leave_record_repository.CountAll.scan", "error", err)
		return 0, err
	}
	return count, nil
}

func (r *LeaveRecordRepository) CountOnLeaveToday(ctx context.Context, q ports.Querier, date time.Time) (int, error) {
	// Count distinct employees who have a leave record where date falls between start_date and end_date
	query := `
		SELECT COUNT(DISTINCT employee_id)
		FROM leave_records
		WHERE date(?) BETWEEN date(start_date) AND date(end_date)`

	var count int
	err := q.QueryRowContext(ctx, query, date).Scan(&count)
	if err != nil {
		slog.Error("leave_record_repository.CountOnLeaveToday.scan", "error", err)
		return 0, err
	}
	return count, nil
}

func (r *LeaveRecordRepository) HasLeaveOnDate(ctx context.Context, q ports.Querier, employeeID int64, date time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM leave_records
		WHERE employee_id = ?
			AND date(?) BETWEEN date(start_date) AND date(end_date)`

	var count int
	err := q.QueryRowContext(ctx, query, employeeID, date).Scan(&count)
	if err != nil {
		slog.Error("leave_record_repository.HasLeaveOnDate.scan", "error", err, "employee_id", employeeID)
		return false, err
	}

	return count > 0, nil
}
