package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type AttendanceRecordRepository struct{}

func NewAttendanceRecordRepository() *AttendanceRecordRepository {
	return &AttendanceRecordRepository{}
}

func (r *AttendanceRecordRepository) Create(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) (bool, error) {
	query := `
		INSERT INTO attendance_records (
			uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(device_uid, device_user_id, punched_at, punch_type) DO NOTHING`

	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		record.UID,
		record.EmployeeUID,
		record.DeviceUID,
		record.DeviceUserID,
		record.PunchedAt.Format(time.RFC3339Nano),
		record.PunchType,
		record.RawPayload,
		record.CreatedAt,
		record.UpdatedAt,
	)
	if err != nil {
		slog.Error("attendance_record_repository.Create.exec_query", "error", err, "uid", record.UID)
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if affected == 0 {
		return false, nil
	}

	id, err := result.LastInsertId()
	if err == nil {
		record.ID = id
	}

	return true, nil
}

func (r *AttendanceRecordRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.AttendanceRecord, error) {
	query := `
		SELECT id, uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
		FROM attendance_records
		WHERE uid = ?
		LIMIT 1`

	return r.scanRecord(q.QueryRowContext(ctx, query, uid))
}

func (r *AttendanceRecordRepository) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.AttendanceRecord, error) {
	if len(uids) == 0 {
		return []*domain.AttendanceRecord{}, nil
	}

	// Build IN clause with placeholders
	query := `
		SELECT id, uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
		FROM attendance_records
		WHERE uid IN (`

	args := make([]any, len(uids))
	for i, uid := range uids {
		if i > 0 {
			query += ", "
		}
		query += "?"
		args[i] = uid
	}
	query += ")"

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("attendance_record_repository.GetByUIDs.query", "error", err, "count", len(uids))
		return nil, err
	}
	defer rows.Close()

	var records []*domain.AttendanceRecord
	for rows.Next() {
		rec, err := r.scanRecordRow(rows)
		if err != nil {
			slog.Error("attendance_record_repository.GetByUIDs.scan", "error", err)
			return nil, err
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		slog.Error("attendance_record_repository.GetByUIDs.rows_err", "error", err)
		return nil, err
	}

	return records, nil
}

func (r *AttendanceRecordRepository) Update(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) error {
	query := `
		UPDATE attendance_records
		SET device_uid = ?, punched_at = ?, punch_type = ?, updated_at = ?
		WHERE uid = ?`

	record.UpdatedAt = time.Now()

	_, err := q.ExecContext(
		ctx,
		query,
		record.DeviceUID,
		record.PunchedAt.Format(time.RFC3339Nano),
		record.PunchType,
		record.UpdatedAt,
		record.UID,
	)
	if err != nil {
		slog.Error("attendance_record_repository.Update.exec_query", "error", err, "uid", record.UID)
		return err
	}

	return nil
}

func (r *AttendanceRecordRepository) ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	query := `
		SELECT id, uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
		FROM attendance_records
		WHERE date(punched_at) = date(?)`

	args := []any{date.Format("2006-01-02")}
	if employeeUID != nil {
		query += ` AND employee_uid = ?`
		args = append(args, *employeeUID)
	}
	query += ` ORDER BY punched_at ASC`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("attendance_record_repository.ListByDate.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	records := make([]*domain.AttendanceRecord, 0)
	for rows.Next() {
		record, err := r.scanRecordRow(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		slog.Error("attendance_record_repository.ListByDate.rows_err", "error", err)
		return nil, err
	}

	return records, nil
}

func (r *AttendanceRecordRepository) ListByDateRange(ctx context.Context, q ports.Querier, startDate, endDate time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	query := `
		SELECT id, uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
		FROM attendance_records
		WHERE datetime(punched_at) >= datetime(?)
			AND datetime(punched_at) <= datetime(?)`

	inclusiveEndDate := makeInclusiveEndDate(endDate)
	args := []any{startDate.Format(time.RFC3339), inclusiveEndDate.Format(time.RFC3339)}
	if employeeUID != nil {
		query += ` AND employee_uid = ?`
		args = append(args, *employeeUID)
	}
	query += ` ORDER BY punched_at ASC, id ASC`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("attendance_record_repository.ListByDateRange.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	records := make([]*domain.AttendanceRecord, 0)
	for rows.Next() {
		record, err := r.scanRecordRow(rows)
		if err != nil {
			slog.Error("attendance_record_repository.ListByDateRange.scan_row", "error", err)
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		slog.Error("attendance_record_repository.ListByDateRange.rows_err", "error", err)
		return nil, err
	}

	return records, nil
}

func (r *AttendanceRecordRepository) ListByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	baseQuery := `
		SELECT
			ar.id, ar.uid, ar.employee_uid, ar.device_uid, ar.device_user_id, ar.punched_at, ar.punch_type, ar.raw_payload, ar.created_at, ar.updated_at,
			e.name, e.department_uid, ad.name
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		INNER JOIN attendance_devices ad ON ad.uid = ar.device_uid
		WHERE e.department_uid = ?`

	whereClause, args := buildDepartmentAttendanceLogsWhere(departmentUID, filter)
	orderBy, err := buildDepartmentAttendanceLogsOrderBy(params)
	if err != nil {
		return nil, err
	}

	query := baseQuery + whereClause + orderBy + ` LIMIT ? OFFSET ?`
	args = append(args, params.PageSize, params.Offset())

	return r.queryAttendanceRecordsWithEmployee(ctx, q, query, args, "attendance_record_repository.ListByDepartmentUID", "department_uid", departmentUID)
}

func (r *AttendanceRecordRepository) CountByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	baseQuery := `
		SELECT COUNT(*)
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		WHERE e.department_uid = ?`

	whereClause, args := buildDepartmentAttendanceLogsWhere(departmentUID, filter)
	query := baseQuery + whereClause

	var total int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		slog.Error("attendance_record_repository.CountByDepartmentUID.query", "error", err, "department_uid", departmentUID)
		return 0, err
	}

	return total, nil
}

func (r *AttendanceRecordRepository) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	baseQuery := `
		SELECT
			ar.id, ar.uid, ar.employee_uid, ar.device_uid, ar.device_user_id, ar.punched_at, ar.punch_type, ar.raw_payload, ar.created_at, ar.updated_at,
			e.name, e.department_uid, ad.name
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		INNER JOIN attendance_devices ad ON ad.uid = ar.device_uid
		WHERE ar.employee_uid = ?`

	whereClause, args := buildEmployeeAttendanceLogsWhere(employeeUID, filter)
	orderBy, err := buildDepartmentAttendanceLogsOrderBy(params)
	if err != nil {
		return nil, err
	}

	query := baseQuery + whereClause + orderBy + ` LIMIT ? OFFSET ?`
	args = append(args, params.PageSize, params.Offset())

	return r.queryAttendanceRecordsWithEmployee(ctx, q, query, args, "attendance_record_repository.ListByEmployeeUID", "employee_uid", employeeUID)
}

func (r *AttendanceRecordRepository) CountByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	baseQuery := `
		SELECT COUNT(*)
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		WHERE ar.employee_uid = ?`

	whereClause, args := buildEmployeeAttendanceLogsWhere(employeeUID, filter)
	query := baseQuery + whereClause

	var total int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		slog.Error("attendance_record_repository.CountByEmployeeUID.query", "error", err, "employee_uid", employeeUID)
		return 0, err
	}

	return total, nil
}

func (r *AttendanceRecordRepository) ListDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	baseQuery := `
		SELECT
			date(ar.punched_at) AS attendance_date,
			ar.employee_uid,
			e.name,
			e.department_uid,
			MIN(CASE WHEN ar.punch_type IN ('check_in', 'unknown') THEN ar.punched_at END) AS check_in,
			(
			SELECT ar1.uid
			FROM attendance_records ar1
			WHERE ar1.employee_uid = ar.employee_uid
				AND date(ar1.punched_at) = date(ar.punched_at)
				AND ar1.punch_type IN ('check_in', 'unknown')
			ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
			LIMIT 1
			) AS check_in_log_uid,
			MAX(CASE WHEN ar.punch_type IN ('check_out', 'unknown') THEN ar.punched_at END) AS check_out,
						 		(
			SELECT ar2.uid
			FROM attendance_records ar2
			WHERE ar2.employee_uid = ar.employee_uid
				AND date(ar2.punched_at) = date(ar.punched_at)
				AND ar2.punch_type IN ('check_out', 'unknown')
			ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
			LIMIT 1
			) AS check_out_log_uid,
			(
				SELECT ad1.name
				FROM attendance_records ar1
				INNER JOIN attendance_devices ad1 ON ad1.uid = ar1.device_uid
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_device,
			(
				SELECT ad1.uid
				FROM attendance_records ar1
				INNER JOIN attendance_devices ad1 ON ad1.uid = ar1.device_uid
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_device_uid,
			(
				SELECT ad2.name
				FROM attendance_records ar2
				INNER JOIN attendance_devices ad2 ON ad2.uid = ar2.device_uid
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_device,
			(
				SELECT ad2.uid
				FROM attendance_records ar2
				INNER JOIN attendance_devices ad2 ON ad2.uid = ar2.device_uid
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_device_uid,
			(
				EXISTS (
					SELECT 1
					FROM audit_logs al
					WHERE al.entity_type = 'attendance_record' AND al.entity_uid = (
						SELECT ar1.uid
						FROM attendance_records ar1
						WHERE ar1.employee_uid = ar.employee_uid
							AND date(ar1.punched_at) = date(ar.punched_at)
							AND ar1.punch_type IN ('check_in', 'unknown')
						ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
						LIMIT 1
					)
				)
				OR EXISTS (
					SELECT 1
					FROM audit_logs al
					WHERE al.entity_type = 'attendance_record' AND al.entity_uid = (
						SELECT ar2.uid
						FROM attendance_records ar2
						WHERE ar2.employee_uid = ar.employee_uid
							AND date(ar2.punched_at) = date(ar.punched_at)
							AND ar2.punch_type IN ('check_out', 'unknown')
						ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
						LIMIT 1
					)
				)
			) AS has_edit_history
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		WHERE e.department_uid = ?`

	whereClause, args := buildDepartmentAttendanceLogsWhere(departmentUID, filter)
	orderBy, err := buildDailyAttendanceLogsOrderBy(params)
	if err != nil {
		return nil, err
	}

	query := baseQuery + whereClause + ` GROUP BY date(ar.punched_at), ar.employee_uid, e.name, e.department_uid
		HAVING check_in IS NOT NULL OR check_out IS NOT NULL` + orderBy + ` LIMIT ? OFFSET ?`
	args = append(args, params.PageSize, params.Offset())

	return r.queryDailyAttendanceGroups(ctx, q, query, args, "attendance_record_repository.ListDailyByDepartmentUID", "department_uid", departmentUID)
}

func (r *AttendanceRecordRepository) CountDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	baseQuery := `
		SELECT COUNT(*)
		FROM (
			SELECT 1
			FROM attendance_records ar
			INNER JOIN employees e ON e.uid = ar.employee_uid
			WHERE e.department_uid = ?`

	whereClause, args := buildDepartmentAttendanceLogsWhere(departmentUID, filter)
	query := baseQuery + whereClause + ` GROUP BY date(ar.punched_at), ar.employee_uid
		) grouped`

	var total int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		slog.Error("attendance_record_repository.CountDailyByDepartmentUID.query", "error", err, "department_uid", departmentUID)
		return 0, err
	}

	return total, nil
}

func (r *AttendanceRecordRepository) ListDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	baseQuery := `
		SELECT
			date(ar.punched_at) AS attendance_date,
			ar.employee_uid,
			e.name,
			e.department_uid,
			MIN(CASE WHEN ar.punch_type IN ('check_in', 'unknown') THEN ar.punched_at END) AS check_in,
			(
				SELECT ar1.uid
				FROM attendance_records ar1
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_log_uid,
			MAX(CASE WHEN ar.punch_type IN ('check_out', 'unknown') THEN ar.punched_at END) AS check_out,
			(
				SELECT ar2.uid
				FROM attendance_records ar2
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_log_uid,
			(
				SELECT ad1.name
				FROM attendance_records ar1
				INNER JOIN attendance_devices ad1 ON ad1.uid = ar1.device_uid
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_device,
			(
				SELECT ad1.uid
				FROM attendance_records ar1
				INNER JOIN attendance_devices ad1 ON ad1.uid = ar1.device_uid
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_device_uid,
			(
				SELECT ad2.name
				FROM attendance_records ar2
				INNER JOIN attendance_devices ad2 ON ad2.uid = ar2.device_uid
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_device,
			(
				SELECT ad2.uid
				FROM attendance_records ar2
				INNER JOIN attendance_devices ad2 ON ad2.uid = ar2.device_uid
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_device_uid,
			(
				EXISTS (
					SELECT 1
					FROM audit_logs al
					WHERE al.entity_type = 'attendance_record' AND al.entity_uid = (
						SELECT ar1.uid
						FROM attendance_records ar1
						WHERE ar1.employee_uid = ar.employee_uid
							AND date(ar1.punched_at) = date(ar.punched_at)
							AND ar1.punch_type IN ('check_in', 'unknown')
						ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
						LIMIT 1
					)
				)
				OR EXISTS (
					SELECT 1
					FROM audit_logs al
					WHERE al.entity_type = 'attendance_record' AND al.entity_uid = (
						SELECT ar2.uid
						FROM attendance_records ar2
						WHERE ar2.employee_uid = ar.employee_uid
							AND date(ar2.punched_at) = date(ar.punched_at)
							AND ar2.punch_type IN ('check_out', 'unknown')
						ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
						LIMIT 1
					)
				)
			) AS has_edit_history
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		WHERE ar.employee_uid = ?`

	whereClause, args := buildEmployeeAttendanceLogsWhere(employeeUID, filter)
	orderBy, err := buildDailyAttendanceLogsOrderBy(params)
	if err != nil {
		return nil, err
	}

	query := baseQuery + whereClause + ` GROUP BY date(ar.punched_at), ar.employee_uid, e.name, e.department_uid
		HAVING check_in IS NOT NULL OR check_out IS NOT NULL` + orderBy + ` LIMIT ? OFFSET ?`
	args = append(args, params.PageSize, params.Offset())

	return r.queryDailyAttendanceGroups(ctx, q, query, args, "attendance_record_repository.ListDailyByEmployeeUID", "employee_uid", employeeUID)
}

func (r *AttendanceRecordRepository) CountDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	baseQuery := `
		SELECT COUNT(*)
		FROM (
			SELECT 1
			FROM attendance_records ar
			INNER JOIN employees e ON e.uid = ar.employee_uid
			WHERE ar.employee_uid = ?`

	whereClause, args := buildEmployeeAttendanceLogsWhere(employeeUID, filter)
	query := baseQuery + whereClause + ` GROUP BY date(ar.punched_at), ar.employee_uid
		) grouped`

	var total int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		slog.Error("attendance_record_repository.CountDailyByEmployeeUID.query", "error", err, "employee_uid", employeeUID)
		return 0, err
	}

	return total, nil
}

func (r *AttendanceRecordRepository) ListDaily(ctx context.Context, q ports.Querier, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	baseQuery := `
		SELECT
			date(ar.punched_at) AS attendance_date,
			ar.employee_uid,
			e.name,
			e.department_uid,
			MIN(CASE WHEN ar.punch_type IN ('check_in', 'unknown') THEN ar.punched_at END) AS check_in,
			(
				SELECT ar1.uid
				FROM attendance_records ar1
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_log_uid,
			MAX(CASE WHEN ar.punch_type IN ('check_out', 'unknown') THEN ar.punched_at END) AS check_out,
			(
				SELECT ar2.uid
				FROM attendance_records ar2
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_log_uid,
			(
				SELECT ad1.name
				FROM attendance_records ar1
				INNER JOIN attendance_devices ad1 ON ad1.uid = ar1.device_uid
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_device,
			(
				SELECT ad1.uid
				FROM attendance_records ar1
				INNER JOIN attendance_devices ad1 ON ad1.uid = ar1.device_uid
				WHERE ar1.employee_uid = ar.employee_uid
					AND date(ar1.punched_at) = date(ar.punched_at)
					AND ar1.punch_type IN ('check_in', 'unknown')
				ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
				LIMIT 1
			) AS check_in_device_uid,
			(
				SELECT ad2.name
				FROM attendance_records ar2
				INNER JOIN attendance_devices ad2 ON ad2.uid = ar2.device_uid
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_device,
			(
				SELECT ad2.uid
				FROM attendance_records ar2
				INNER JOIN attendance_devices ad2 ON ad2.uid = ar2.device_uid
				WHERE ar2.employee_uid = ar.employee_uid
					AND date(ar2.punched_at) = date(ar.punched_at)
					AND ar2.punch_type IN ('check_out', 'unknown')
				ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
				LIMIT 1
			) AS check_out_device_uid,
			(
				EXISTS (
					SELECT 1
					FROM audit_logs al
					WHERE al.entity_type = 'attendance_record' AND al.entity_uid = (
						SELECT ar1.uid
						FROM attendance_records ar1
						WHERE ar1.employee_uid = ar.employee_uid
							AND date(ar1.punched_at) = date(ar.punched_at)
							AND ar1.punch_type IN ('check_in', 'unknown')
						ORDER BY datetime(ar1.punched_at) ASC, ar1.id ASC
						LIMIT 1
					)
				)
				OR EXISTS (
					SELECT 1
					FROM audit_logs al
					WHERE al.entity_type = 'attendance_record' AND al.entity_uid = (
						SELECT ar2.uid
						FROM attendance_records ar2
						WHERE ar2.employee_uid = ar.employee_uid
							AND date(ar2.punched_at) = date(ar.punched_at)
							AND ar2.punch_type IN ('check_out', 'unknown')
						ORDER BY datetime(ar2.punched_at) DESC, ar2.id DESC
						LIMIT 1
					)
				)
			) AS has_edit_history
		FROM attendance_records ar
		INNER JOIN employees e ON e.uid = ar.employee_uid
		WHERE 1 = 1`

	whereClause, args := buildAttendanceLogsWhere(filter)
	orderBy, err := buildDailyAttendanceLogsOrderBy(params)
	if err != nil {
		return nil, err
	}

	query := baseQuery + whereClause + ` GROUP BY date(ar.punched_at), ar.employee_uid, e.name, e.department_uid
		HAVING check_in IS NOT NULL OR check_out IS NOT NULL` + orderBy + ` LIMIT ? OFFSET ?`
	args = append(args, params.PageSize, params.Offset())

	return r.queryDailyAttendanceGroups(ctx, q, query, args, "attendance_record_repository.ListDaily", "scope", "all")
}

func (r *AttendanceRecordRepository) ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q ports.Querier, deviceUserID string) (*string, error) {
	query := `
		SELECT uid
		FROM employees
		WHERE university_id = ? OR government_id = ?
		LIMIT 1`

	var employeeUID string
	err := q.QueryRowContext(ctx, query, deviceUserID, deviceUserID).Scan(&employeeUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("attendance_record_repository.ResolveEmployeeUIDByDeviceUserID.scan", "error", err, "device_user_id", deviceUserID)
		return nil, err
	}

	return &employeeUID, nil
}

func (r *AttendanceRecordRepository) scanRecordRow(rows *sql.Rows) (*domain.AttendanceRecord, error) {
	var record domain.AttendanceRecord
	var punchedAt, createdAt, updatedAt domain.Time

	err := rows.Scan(
		&record.ID,
		&record.UID,
		&record.EmployeeUID,
		&record.DeviceUID,
		&record.DeviceUserID,
		&punchedAt,
		&record.PunchType,
		&record.RawPayload,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		slog.Error("attendance_record_repository.scanRecordRow.scan", "error", err)
		return nil, err
	}

	record.PunchedAt = punchedAt.Time
	record.CreatedAt = createdAt.Time
	record.UpdatedAt = updatedAt.Time
	return &record, nil
}

func (r *AttendanceRecordRepository) scanRecord(row *sql.Row) (*domain.AttendanceRecord, error) {
	var record domain.AttendanceRecord
	var punchedAt, createdAt, updatedAt domain.Time

	err := row.Scan(
		&record.ID,
		&record.UID,
		&record.EmployeeUID,
		&record.DeviceUID,
		&record.DeviceUserID,
		&punchedAt,
		&record.PunchType,
		&record.RawPayload,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("attendance_record_repository.scanRecord.scan_row", "error", err)
		return nil, err
	}

	record.PunchedAt = punchedAt.Time
	record.CreatedAt = createdAt.Time
	record.UpdatedAt = updatedAt.Time
	return &record, nil
}

func (r *AttendanceRecordRepository) queryAttendanceRecordsWithEmployee(ctx context.Context, q ports.Querier, query string, args []any, logKey, idKey, idValue string) ([]*ports.AttendanceRecordWithEmployee, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error(logKey+".query", "error", err, idKey, idValue)
		return nil, err
	}
	defer rows.Close()

	records := make([]*ports.AttendanceRecordWithEmployee, 0)
	for rows.Next() {
		var employeeName string
		var deviceName string
		var record domain.AttendanceRecord
		var punchedAt, createdAt, updatedAt domain.Time
		var deptUID sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.UID,
			&record.EmployeeUID,
			&record.DeviceUID,
			&record.DeviceUserID,
			&punchedAt,
			&record.PunchType,
			&record.RawPayload,
			&createdAt,
			&updatedAt,
			&employeeName,
			&deptUID,
			&deviceName,
		)
		if err != nil {
			slog.Error(logKey+".scan_row", "error", err, idKey, idValue)
			return nil, err
		}

		record.PunchedAt = punchedAt.Time
		record.CreatedAt = createdAt.Time
		record.UpdatedAt = updatedAt.Time

		var departmentUIDPtr *string
		if deptUID.Valid {
			departmentUIDPtr = &deptUID.String
		}

		records = append(records, &ports.AttendanceRecordWithEmployee{
			Record:        &record,
			EmployeeName:  employeeName,
			DepartmentUID: departmentUIDPtr,
			DeviceName:    deviceName,
		})
	}

	if err := rows.Err(); err != nil {
		slog.Error(logKey+".rows_err", "error", err, idKey, idValue)
		return nil, err
	}

	return records, nil
}

func (r *AttendanceRecordRepository) queryDailyAttendanceGroups(ctx context.Context, q ports.Querier, query string, args []any, logKey, idKey, idValue string) ([]*ports.DailyAttendanceGroup, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error(logKey+".query", "error", err, idKey, idValue)
		return nil, err
	}
	defer rows.Close()

	groups := make([]*ports.DailyAttendanceGroup, 0)
	for rows.Next() {
		var dateValue string
		var employeeUID string
		var employeeName string
		var departmentUID sql.NullString
		var checkIn sql.NullString
		var checkInLogUID sql.NullString
		var checkOut sql.NullString
		var checkOutLogUID sql.NullString
		var checkInDevice sql.NullString
		var checkInDeviceUID sql.NullString
		var checkOutDevice sql.NullString
		var checkOutDeviceUID sql.NullString
		var hasEditHistory bool

		if err := rows.Scan(&dateValue, &employeeUID, &employeeName, &departmentUID, &checkIn, &checkInLogUID, &checkOut, &checkOutLogUID, &checkInDevice, &checkInDeviceUID, &checkOutDevice, &checkOutDeviceUID, &hasEditHistory); err != nil {
			slog.Error(logKey+".scan_row", "error", err, idKey, idValue)
			return nil, err
		}

		dateParsed, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return nil, err
		}

		var departmentUIDPtr *string
		if departmentUID.Valid {
			departmentUIDPtr = &departmentUID.String
		}

		group := &ports.DailyAttendanceGroup{
			Date:           dateParsed,
			EmployeeUID:    employeeUID,
			EmployeeName:   employeeName,
			DepartmentUID:  departmentUIDPtr,
			HasEditHistory: hasEditHistory,
		}

		if checkIn.Valid {
			t, err := parseAttendanceTimestamp(checkIn.String)
			if err != nil {
				return nil, err
			}
			group.CheckIn = &t
		}
		if checkOut.Valid {
			t, err := parseAttendanceTimestamp(checkOut.String)
			if err != nil {
				return nil, err
			}
			group.CheckOut = &t
		}
		if checkInDevice.Valid {
			group.CheckInDevice = &checkInDevice.String
		}
		if checkInDeviceUID.Valid {
			group.CheckInDeviceUID = &checkInDeviceUID.String
		}
		if checkOutDevice.Valid {
			group.CheckOutDevice = &checkOutDevice.String
		}
		if checkOutDeviceUID.Valid {
			group.CheckOutDeviceUID = &checkOutDeviceUID.String
		}

		// populate the log UID fields returned by the query
		if checkInLogUID.Valid {
			v := checkInLogUID.String
			group.CheckInLogUID = &v
		}
		if checkOutLogUID.Valid {
			v := checkOutLogUID.String
			group.CheckOutLogUID = &v
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		slog.Error(logKey+".rows_err", "error", err, idKey, idValue)
		return nil, err
	}

	return groups, nil
}

func parseAttendanceTimestamp(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid attendance timestamp format: %s", value)
}

func buildDepartmentAttendanceLogsWhere(departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (string, []any) {
	clauses := []string{}
	args := []any{departmentUID}

	if filter.EmployeeUID != nil {
		clauses = append(clauses, "ar.employee_uid = ?")
		args = append(args, *filter.EmployeeUID)
	}
	if filter.EmployeeName != nil {
		mode := "contains"
		if filter.EmployeeNameMode != nil && *filter.EmployeeNameMode != "" {
			mode = *filter.EmployeeNameMode
		}

		switch mode {
		case "equals":
			clauses = append(clauses, "e.name = ?")
			args = append(args, *filter.EmployeeName)
		case "contains":
			clauses = append(clauses, "e.name LIKE ?")
			args = append(args, "%"+*filter.EmployeeName+"%")
		}
	}
	if filter.DeviceUID != nil {
		clauses = append(clauses, "ar.device_uid = ?")
		args = append(args, *filter.DeviceUID)
	}
	if filter.PunchType != nil {
		clauses = append(clauses, "ar.punch_type = ?")
		args = append(args, *filter.PunchType)
	}
	if filter.StartDate != nil {
		clauses = append(clauses, "datetime(ar.punched_at) >= datetime(?)")
		args = append(args, filter.StartDate.Format(time.RFC3339))
	}
	if filter.EndDate != nil {
		inclusiveEndDate := makeInclusiveEndDate(*filter.EndDate)
		clauses = append(clauses, "datetime(ar.punched_at) <= datetime(?)")
		args = append(args, inclusiveEndDate.Format(time.RFC3339))
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " AND " + strings.Join(clauses, " AND "), args
}

func buildAttendanceLogsWhere(filter ports.DepartmentAttendanceLogsFilter) (string, []any) {
	clauses := []string{}
	args := []any{}

	if filter.EmployeeUID != nil {
		clauses = append(clauses, "ar.employee_uid = ?")
		args = append(args, *filter.EmployeeUID)
	}
	if filter.EmployeeName != nil {
		mode := "contains"
		if filter.EmployeeNameMode != nil && *filter.EmployeeNameMode != "" {
			mode = *filter.EmployeeNameMode
		}

		switch mode {
		case "equals":
			clauses = append(clauses, "e.name = ?")
			args = append(args, *filter.EmployeeName)
		case "contains":
			clauses = append(clauses, "e.name LIKE ?")
			args = append(args, "%"+*filter.EmployeeName+"%")
		}
	}
	if filter.DeviceUID != nil {
		clauses = append(clauses, "ar.device_uid = ?")
		args = append(args, *filter.DeviceUID)
	}
	if filter.PunchType != nil {
		clauses = append(clauses, "ar.punch_type = ?")
		args = append(args, *filter.PunchType)
	}
	if filter.StartDate != nil {
		clauses = append(clauses, "datetime(ar.punched_at) >= datetime(?)")
		args = append(args, filter.StartDate.Format(time.RFC3339))
	}
	if filter.EndDate != nil {
		inclusiveEndDate := makeInclusiveEndDate(*filter.EndDate)
		clauses = append(clauses, "datetime(ar.punched_at) <= datetime(?)")
		args = append(args, inclusiveEndDate.Format(time.RFC3339))
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " AND " + strings.Join(clauses, " AND "), args
}

func buildEmployeeAttendanceLogsWhere(employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (string, []any) {
	clauses := []string{}
	args := []any{employeeUID}

	if filter.EmployeeName != nil {
		mode := "contains"
		if filter.EmployeeNameMode != nil && *filter.EmployeeNameMode != "" {
			mode = *filter.EmployeeNameMode
		}

		switch mode {
		case "equals":
			clauses = append(clauses, "e.name = ?")
			args = append(args, *filter.EmployeeName)
		case "contains":
			clauses = append(clauses, "e.name LIKE ?")
			args = append(args, "%"+*filter.EmployeeName+"%")
		}
	}
	if filter.DeviceUID != nil {
		clauses = append(clauses, "ar.device_uid = ?")
		args = append(args, *filter.DeviceUID)
	}
	if filter.PunchType != nil {
		clauses = append(clauses, "ar.punch_type = ?")
		args = append(args, *filter.PunchType)
	}
	if filter.StartDate != nil {
		clauses = append(clauses, "datetime(ar.punched_at) >= datetime(?)")
		args = append(args, filter.StartDate.Format(time.RFC3339))
	}
	if filter.EndDate != nil {
		inclusiveEndDate := makeInclusiveEndDate(*filter.EndDate)
		clauses = append(clauses, "datetime(ar.punched_at) <= datetime(?)")
		args = append(args, inclusiveEndDate.Format(time.RFC3339))
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " AND " + strings.Join(clauses, " AND "), args
}

func buildDepartmentAttendanceLogsOrderBy(params ports.ListParams) (string, error) {
	allowedColumns := map[string]string{
		"punchedAt":    "ar.punched_at",
		"employeeName": "e.name",
		"punchType":    "ar.punch_type",
	}

	column, ok := allowedColumns[params.SortBy]
	if !ok {
		return "", fmt.Errorf("invalid sort field: %s", params.SortBy)
	}

	order := "ASC"
	if params.SortOrder == ports.SortOrderDesc {
		order = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", column, order), nil
}

func buildDailyAttendanceLogsOrderBy(params ports.ListParams) (string, error) {
	allowedColumns := map[string]string{
		"date":         "attendance_date",
		"employeeName": "e.name",
	}

	column, ok := allowedColumns[params.SortBy]
	if !ok {
		return "", fmt.Errorf("invalid sort field: %s", params.SortBy)
	}

	order := "ASC"
	if params.SortOrder == ports.SortOrderDesc {
		order = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s, e.name ASC, ar.employee_uid ASC", column, order), nil
}

func makeInclusiveEndDate(value time.Time) time.Time {
	if value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 && value.Nanosecond() == 0 {
		return value.Add(24*time.Hour - time.Nanosecond)
	}

	return value
}
