package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LeaveRequestRepository struct{}

func NewLeaveRequestRepository() *LeaveRequestRepository {
	return &LeaveRequestRepository{}
}

func (r *LeaveRequestRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRequest, error) {
	query := `
		SELECT id, uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at
		FROM leave_requests
		WHERE id = $1`

	return r.scanLeaveRequest(q.QueryRowContext(ctx, query, id))
}

func (r *LeaveRequestRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRequest, error) {
	query := `
		SELECT id, uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at
		FROM leave_requests
		WHERE uid = $1`

	return r.scanLeaveRequest(q.QueryRowContext(ctx, query, uid))
}

func (r *LeaveRequestRepository) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.LeaveRequest, error) {
	if len(uids) == 0 {
		return []*domain.LeaveRequest{}, nil
	}

	// Build IN clause with placeholders
	query := `
		SELECT id, uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at
		FROM leave_requests
		WHERE uid IN (`

	args := make([]any, len(uids))
	for i, uid := range uids {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
		args[i] = uid
	}
	query += ")"

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("leave_request_repository.GetByUIDs.query", "error", err, "count", len(uids))
		return nil, err
	}
	defer rows.Close()

	var requests []*domain.LeaveRequest
	for rows.Next() {
		req, err := r.scanLeaveRequestRow(rows)
		if err != nil {
			slog.Error("leave_request_repository.GetByUIDs.scan", "error", err)
			return nil, err
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_request_repository.GetByUIDs.rows_err", "error", err)
		return nil, err
	}

	return requests, nil
}
func (r *LeaveRequestRepository) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.LeaveRequest, error) {
	query := `
		SELECT id, uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at
		FROM leave_requests
		WHERE approval_request_uid = $1`

	return r.scanLeaveRequest(q.QueryRowContext(ctx, query, approvalRequestUID))
}

func (r *LeaveRequestRepository) Create(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	query := `
		INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id`

	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now
	if request.SubmittedAt.IsZero() {
		request.SubmittedAt = now
	}

	err := q.QueryRowContext(ctx, query,
		request.UID, request.EmployeeUID, request.LeaveTypeUID, request.SubLeaveTypeUID, request.OtherSubLeaveName,
		request.StartDate, request.EndDate, request.Days, request.Notes,
		request.StudyDestination, request.Assignment, request.AssignmentCountry, request.SpouseWorkCountry,
		request.SubmittedAt, request.DecidedAt, request.ApprovalRequestUID,
		request.CreatedAt, request.UpdatedAt).Scan(&request.ID)
	if err != nil {
		slog.Error("leave_request_repository.Create.exec_query", "error", err, "uid", request.UID, "employee_uid", request.EmployeeUID)
		return err
	}

	return nil
}

func (r *LeaveRequestRepository) Update(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	query := `
		UPDATE leave_requests
		SET leave_type_uid = $1, sub_leave_type_uid = $2, other_sub_leave_name = $3, start_date = $4, end_date = $5, days = $6, notes = $7,
		    study_destination = $8, assignment = $9, assignment_country = $10, spouse_work_country = $11,
		    decided_at = $12, updated_at = $13
		WHERE id = $14`

	request.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		request.LeaveTypeUID, request.SubLeaveTypeUID, request.OtherSubLeaveName, request.StartDate, request.EndDate, request.Days, request.Notes,
		request.StudyDestination, request.Assignment, request.AssignmentCountry, request.SpouseWorkCountry,
		request.DecidedAt, request.UpdatedAt, request.ID)
	if err != nil {
		slog.Error("leave_request_repository.Update.exec_query", "error", err, "id", request.ID, "uid", request.UID)
	}
	return err
}

func (r *LeaveRequestRepository) ListByEmployee(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.LeaveRequest, error) {
	query := `
		SELECT id, uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at
		FROM leave_requests
		WHERE employee_uid = $1
		ORDER BY submitted_at DESC`

	return r.queryLeaveRequests(ctx, q, query, employeeUID)
}

func (r *LeaveRequestRepository) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeUID string, limit, offset int) ([]*domain.LeaveRequest, error) {
	query := `
		SELECT id, uid, employee_uid, leave_type_uid, sub_leave_type_uid, other_sub_leave_name, start_date, end_date, days, notes, study_destination, assignment, assignment_country, spouse_work_country, submitted_at, decided_at, approval_request_uid, created_at, updated_at
		FROM leave_requests
		WHERE employee_uid = $1
		ORDER BY submitted_at DESC
		LIMIT $2 OFFSET $3`

	return r.queryLeaveRequests(ctx, q, query, employeeUID, limit, offset)
}

func (r *LeaveRequestRepository) CountByEmployee(ctx context.Context, q ports.Querier, employeeUID string) (int, error) {
	query := `SELECT COUNT(*) FROM leave_requests WHERE employee_uid = $1`
	var count int
	err := q.QueryRowContext(ctx, query, employeeUID).Scan(&count)
	if err != nil {
		slog.Error("leave_request_repository.CountByEmployee.scan_row", "error", err, "employee_uid", employeeUID)
	}
	return count, err
}

func (r *LeaveRequestRepository) HasOverlapping(ctx context.Context, q ports.Querier, employeeUID string, startDate, endDate time.Time, excludeUID *string) (bool, error) {
	// Check for overlapping pending or approved leave requests
	query := `
		SELECT COUNT(*)
		FROM leave_requests lr
		JOIN approval_requests ar ON lr.approval_request_uid = ar.uid
		WHERE lr.employee_uid = $1
		AND ar.status IN ('pending', 'approved')
		AND NOT (lr.end_date < $2 OR lr.start_date > $3)`

	args := []any{employeeUID, startDate, endDate}

	if excludeUID != nil {
		query += ` AND lr.uid != ` + nextPlaceholder(args)
		args = append(args, *excludeUID)
	}

	var count int
	err := q.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.Error("leave_request_repository.HasOverlapping.scan_row", "error", err, "employee_uid", employeeUID)
		return false, err
	}
	return count > 0, nil
}

func (r *LeaveRequestRepository) List(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter, limit, offset int) ([]*domain.LeaveRequest, error) {
	query := `
		SELECT lr.id, lr.uid, lr.employee_uid, lr.leave_type_uid, lr.sub_leave_type_uid, lr.other_sub_leave_name, lr.start_date, lr.end_date, lr.days, lr.notes, lr.study_destination, lr.assignment, lr.assignment_country, lr.spouse_work_country, lr.submitted_at, lr.decided_at, lr.approval_request_uid, lr.created_at, lr.updated_at
		FROM leave_requests lr
		JOIN approval_requests ar ON lr.approval_request_uid = ar.uid
		WHERE 1=1`

	args := []any{}

	if filter.Status != nil {
		query += ` AND ar.status = ` + nextPlaceholder(args)
		args = append(args, *filter.Status)
	}

	if filter.EmployeeUID != nil {
		query += ` AND lr.employee_uid = ` + nextPlaceholder(args)
		args = append(args, *filter.EmployeeUID)
	}

	query, args = applyLeaveRequestDepartmentScope(query, args, filter.DepartmentUIDs)
	query += ` ORDER BY lr.submitted_at DESC LIMIT ` + nextPlaceholder(args) + ` OFFSET ` + nextPlaceholder(append(args, nil))
	args = append(args, limit, offset)

	return r.queryLeaveRequests(ctx, q, query, args...)
}

func (r *LeaveRequestRepository) Count(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM leave_requests lr
		JOIN approval_requests ar ON lr.approval_request_uid = ar.uid
		WHERE 1=1`

	args := []any{}

	if filter.Status != nil {
		query += ` AND ar.status = ` + nextPlaceholder(args)
		args = append(args, *filter.Status)
	}

	if filter.EmployeeUID != nil {
		query += ` AND lr.employee_uid = ` + nextPlaceholder(args)
		args = append(args, *filter.EmployeeUID)
	}

	query, args = applyLeaveRequestDepartmentScope(query, args, filter.DepartmentUIDs)

	var count int
	err := q.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.Error("leave_request_repository.Count.scan_row", "error", err)
	}
	return count, err
}

func applyLeaveRequestDepartmentScope(query string, args []any, departmentUIDs []string) (string, []any) {
	if departmentUIDs == nil {
		return query, args
	}

	if len(departmentUIDs) == 0 {
		return query + ` AND 1=0`, args
	}

	startIndex := len(args) + 1
	for _, departmentUID := range departmentUIDs {
		args = append(args, departmentUID)
	}

	placeholders := buildPlaceholders(len(departmentUIDs), startIndex)
	query += ` AND EXISTS (
		SELECT 1
		FROM employees e
		WHERE e.uid = lr.employee_uid
		  AND e.department_uid IN (` + placeholders + `)
	)`

	return query, args
}

func (r *LeaveRequestRepository) queryLeaveRequests(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.LeaveRequest, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("leave_request_repository.queryLeaveRequests.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*domain.LeaveRequest
	for rows.Next() {
		req, err := r.scanLeaveRequestRow(rows)
		if err != nil {
			slog.Error("leave_request_repository.queryLeaveRequests.scan_row", "error", err)
			return nil, err
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_request_repository.queryLeaveRequests.rows_iteration", "error", err)
		return nil, err
	}
	return requests, nil
}

func (r *LeaveRequestRepository) scanLeaveRequest(row *sql.Row) (*domain.LeaveRequest, error) {
	var req domain.LeaveRequest
	var startDate, endDate, submittedAt, createdAt, updatedAt domain.Time
	var decidedAt domain.NullTime
	err := row.Scan(
		&req.ID, &req.UID, &req.EmployeeUID, &req.LeaveTypeUID, &req.SubLeaveTypeUID, &req.OtherSubLeaveName,
		&startDate, &endDate, &req.Days, &req.Notes, &req.StudyDestination, &req.Assignment, &req.AssignmentCountry, &req.SpouseWorkCountry,
		&submittedAt, &decidedAt, &req.ApprovalRequestUID,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("leave_request_repository.scanLeaveRequest.scan_row", "error", err)
		return nil, err
	}
	req.StartDate = startDate.Time
	req.EndDate = endDate.Time
	req.SubmittedAt = submittedAt.Time
	if decidedAt.Valid {
		req.DecidedAt = &decidedAt.Time
	}
	req.CreatedAt = createdAt.Time
	req.UpdatedAt = updatedAt.Time
	return &req, nil
}

func (r *LeaveRequestRepository) scanLeaveRequestRow(rows *sql.Rows) (*domain.LeaveRequest, error) {
	var req domain.LeaveRequest
	var startDate, endDate, submittedAt, createdAt, updatedAt domain.Time
	var decidedAt domain.NullTime
	err := rows.Scan(
		&req.ID, &req.UID, &req.EmployeeUID, &req.LeaveTypeUID, &req.SubLeaveTypeUID, &req.OtherSubLeaveName,
		&startDate, &endDate, &req.Days, &req.Notes, &req.StudyDestination, &req.Assignment, &req.AssignmentCountry, &req.SpouseWorkCountry,
		&submittedAt, &decidedAt, &req.ApprovalRequestUID,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	req.StartDate = startDate.Time
	req.EndDate = endDate.Time
	req.SubmittedAt = submittedAt.Time
	if decidedAt.Valid {
		req.DecidedAt = &decidedAt.Time
	}
	req.CreatedAt = createdAt.Time
	req.UpdatedAt = updatedAt.Time
	return &req, nil
}

func (r *LeaveRequestRepository) FindExpiredPending(ctx context.Context, q ports.Querier, graceDays int) ([]*domain.LeaveRequest, error) {
	// Find pending leave requests where start_date < (now - graceDays)
	query := `
		SELECT lr.id, lr.uid, lr.employee_uid, lr.leave_type_uid, lr.sub_leave_type_uid, lr.other_sub_leave_name, lr.start_date, lr.end_date, lr.days, lr.notes, lr.study_destination, lr.assignment, lr.assignment_country, lr.spouse_work_country, lr.submitted_at, lr.decided_at, lr.approval_request_uid, lr.created_at, lr.updated_at
		FROM leave_requests lr
		JOIN approval_requests ar ON lr.approval_request_uid = ar.uid
		WHERE ar.status = 'pending'
		AND lr.decided_at IS NULL
		AND lr.start_date::date < (CURRENT_DATE - $1 * INTERVAL '1 day')`

	return r.queryLeaveRequests(ctx, q, query, graceDays)
}
