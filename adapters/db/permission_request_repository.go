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

type PermissionRequestRepository struct{}

func NewPermissionRequestRepository() *PermissionRequestRepository {
	return &PermissionRequestRepository{}
}

const permissionRequestColumns = `id, uid, employee_uid, type, permission_date, start_time, end_time, reason, submitted_at, decided_at, approval_request_uid, created_at, updated_at`

func (r *PermissionRequestRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.PermissionRequest, error) {
	query := `SELECT ` + permissionRequestColumns + ` FROM permission_requests WHERE id = $1`
	return r.scanRow(q.QueryRowContext(ctx, query, id))
}

func (r *PermissionRequestRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.PermissionRequest, error) {
	query := `SELECT ` + permissionRequestColumns + ` FROM permission_requests WHERE uid = $1`
	return r.scanRow(q.QueryRowContext(ctx, query, uid))
}

func (r *PermissionRequestRepository) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.PermissionRequest, error) {
	query := `SELECT ` + permissionRequestColumns + ` FROM permission_requests WHERE approval_request_uid = $1`
	return r.scanRow(q.QueryRowContext(ctx, query, approvalRequestUID))
}

func (r *PermissionRequestRepository) Create(ctx context.Context, q ports.Querier, request *domain.PermissionRequest) error {
	query := `INSERT INTO permission_requests
		(uid, employee_uid, type, permission_date, start_time, end_time, reason, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now
	if request.SubmittedAt.IsZero() {
		request.SubmittedAt = now
	}

	err := q.QueryRowContext(ctx, query,
		request.UID,
		request.EmployeeUID,
		string(request.Type),
		request.PermissionDate.Format("2006-01-02"),
		request.StartTime,
		request.EndTime,
		request.Reason,
		request.SubmittedAt,
		request.DecidedAt,
		request.ApprovalRequestUID,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(&request.ID)
	if err != nil {
		slog.Error("permission_request_repository.Create.exec", "error", err, "uid", request.UID)
		return err
	}
	return nil
}

func (r *PermissionRequestRepository) Update(ctx context.Context, q ports.Querier, request *domain.PermissionRequest) error {
	query := `UPDATE permission_requests
		SET type = $1, permission_date = $2, start_time = $3, end_time = $4, reason = $5, decided_at = $6, updated_at = $7
		WHERE id = $8`

	request.UpdatedAt = time.Now()
	_, err := q.ExecContext(ctx, query,
		string(request.Type),
		request.PermissionDate.Format("2006-01-02"),
		request.StartTime,
		request.EndTime,
		request.Reason,
		request.DecidedAt,
		request.UpdatedAt,
		request.ID,
	)
	if err != nil {
		slog.Error("permission_request_repository.Update.exec", "error", err, "id", request.ID, "uid", request.UID)
	}
	return err
}

func (r *PermissionRequestRepository) ListApprovedForEmployeeOnDate(ctx context.Context, q ports.Querier, employeeUID string, date time.Time) ([]*domain.PermissionRequest, error) {
	query := `
		SELECT ` + qualifyColumns(permissionRequestColumns, "pr") + `
		FROM permission_requests pr
		JOIN approval_requests ar ON pr.approval_request_uid = ar.uid
		WHERE pr.employee_uid = $1
		  AND pr.permission_date = $2
		  AND ar.status = 'approved'
		ORDER BY (pr.start_time IS NOT NULL), pr.start_time, pr.id`

	return r.query(ctx, q, query, employeeUID, date.Format("2006-01-02"))
}

func (r *PermissionRequestRepository) ListApprovedForEmployeesOnDate(ctx context.Context, q ports.Querier, employeeUIDs []string, date time.Time) (map[string][]*domain.PermissionRequest, error) {
	if len(employeeUIDs) == 0 {
		return map[string][]*domain.PermissionRequest{}, nil
	}

	args := make([]any, 0, len(employeeUIDs)+1)
	for _, uid := range employeeUIDs {
		args = append(args, uid)
	}
	args = append(args, date.Format("2006-01-02"))

	placeholders := buildPlaceholders(len(employeeUIDs), 1)
	query := `
		SELECT ` + qualifyColumns(permissionRequestColumns, "pr") + `
		FROM permission_requests pr
		JOIN approval_requests ar ON pr.approval_request_uid = ar.uid
		WHERE pr.employee_uid IN (` + placeholders + `)
		  AND pr.permission_date = $` + fmt.Sprintf("%d", len(employeeUIDs)+1) + `
		  AND ar.status = 'approved'`

	rows, err := r.query(ctx, q, query, args...)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]*domain.PermissionRequest, len(employeeUIDs))
	for _, p := range rows {
		result[p.EmployeeUID] = append(result[p.EmployeeUID], p)
	}
	return result, nil
}

// HasOverlappingOnDate returns true if a same-day permission collides with the candidate window.
// Both candidate and existing rows now carry concrete HH:MM start/end values, so the comparison
// is a direct string compare: existingStart < candEnd AND existingEnd > candStart.
func (r *PermissionRequestRepository) HasOverlappingOnDate(ctx context.Context, q ports.Querier, employeeUID string, date time.Time, startTime, endTime string, excludeUID *string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM permission_requests pr
		JOIN approval_requests ar ON pr.approval_request_uid = ar.uid
		WHERE pr.employee_uid = $1
		  AND pr.permission_date = $2
		  AND ar.status IN ('pending','approved')
		  AND pr.start_time < $3
		  AND pr.end_time > $4`

	args := []any{
		employeeUID,
		date.Format("2006-01-02"),
		endTime,
		startTime,
	}

	if excludeUID != nil {
		query += ` AND pr.uid != ` + nextPlaceholder(args)
		args = append(args, *excludeUID)
	}

	var count int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		slog.Error("permission_request_repository.HasOverlappingOnDate.scan", "error", err, "employee_uid", employeeUID)
		return false, err
	}
	return count > 0, nil
}

func (r *PermissionRequestRepository) CountInWeek(ctx context.Context, q ports.Querier, employeeUID string, types []domain.PermissionType, weekStart, weekEnd time.Time, statuses []domain.ApprovalRequestStatus, excludeUID *string) (int, error) {
	if len(types) == 0 || len(statuses) == 0 {
		return 0, nil
	}

	args := make([]any, 0, len(types)+len(statuses)+4)
	args = append(args, employeeUID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
	
	typePlaceholders := buildPlaceholders(len(types), 4)
	for _, t := range types {
		args = append(args, string(t))
	}
	
	statusPlaceholders := buildPlaceholders(len(statuses), 4+len(types))
	for _, s := range statuses {
		args = append(args, string(s))
	}

	query := `
		SELECT COUNT(*)
		FROM permission_requests pr
		JOIN approval_requests ar ON pr.approval_request_uid = ar.uid
		WHERE pr.employee_uid = $1
		  AND pr.permission_date BETWEEN $2 AND $3
		  AND pr.type IN (` + typePlaceholders + `)
		  AND ar.status IN (` + statusPlaceholders + `)`

	if excludeUID != nil {
		query += ` AND pr.uid != $` + fmt.Sprintf("%d", len(args)+1)
		args = append(args, *excludeUID)
	}

	var count int
	if err := q.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		slog.Error("permission_request_repository.CountInWeek.scan", "error", err, "employee_uid", employeeUID)
		return 0, err
	}
	return count, nil
}

func (r *PermissionRequestRepository) List(ctx context.Context, q ports.Querier, filter ports.PermissionRequestListFilter, limit, offset int) ([]*domain.PermissionRequest, error) {
	query, args := r.buildListQuery(filter, false)
	query += ` ORDER BY pr.permission_date DESC, pr.submitted_at DESC LIMIT ` + nextPlaceholder(args) + ` OFFSET ` + nextPlaceholder(append(args, nil))
	args = append(args, limit, offset)
	return r.query(ctx, q, query, args...)
}

func (r *PermissionRequestRepository) Count(ctx context.Context, q ports.Querier, filter ports.PermissionRequestListFilter) (int, error) {
	query, args := r.buildListQuery(filter, true)
	var count int
	err := q.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.Error("permission_request_repository.Count.scan", "error", err)
	}
	return count, err
}

func (r *PermissionRequestRepository) buildListQuery(filter ports.PermissionRequestListFilter, asCount bool) (string, []any) {
	selectClause := `SELECT ` + qualifyColumns(permissionRequestColumns, "pr")
	if asCount {
		selectClause = `SELECT COUNT(*)`
	}

	query := selectClause + `
		FROM permission_requests pr
		JOIN approval_requests ar ON pr.approval_request_uid = ar.uid
		WHERE 1=1`

	args := []any{}

	if filter.Status != nil {
		query += ` AND ar.status = ` + nextPlaceholder(args)
		args = append(args, string(*filter.Status))
	}
	if filter.Type != nil {
		query += ` AND pr.type = ` + nextPlaceholder(args)
		args = append(args, string(*filter.Type))
	}
	if filter.EmployeeUID != nil {
		query += ` AND pr.employee_uid = ` + nextPlaceholder(args)
		args = append(args, *filter.EmployeeUID)
	}
	if filter.DateFrom != nil {
		query += ` AND pr.permission_date >= ` + nextPlaceholder(args)
		args = append(args, filter.DateFrom.Format("2006-01-02"))
	}
	if filter.DateTo != nil {
		query += ` AND pr.permission_date <= ` + nextPlaceholder(args)
		args = append(args, filter.DateTo.Format("2006-01-02"))
	}
	if filter.DepartmentUIDs != nil {
		if len(filter.DepartmentUIDs) == 0 {
			query += ` AND 1=0`
		} else {
			startIndex := len(args) + 1
			placeholders := buildPlaceholders(len(filter.DepartmentUIDs), startIndex)
			query += ` AND EXISTS (
				SELECT 1 FROM employees e
				WHERE e.uid = pr.employee_uid AND e.department_uid IN (` + placeholders + `)
			)`
			for _, dept := range filter.DepartmentUIDs {
				args = append(args, dept)
			}
		}
	}
	return query, args
}

func (r *PermissionRequestRepository) FindExpiredPending(ctx context.Context, q ports.Querier, asOf time.Time) ([]*domain.PermissionRequest, error) {
	// We rely on the use case to actually decide expiry per type (it knows the employee's shift).
	// The repository simply returns all pending requests whose permission_date <= asOf's date.
	query := `
		SELECT ` + qualifyColumns(permissionRequestColumns, "pr") + `
		FROM permission_requests pr
		JOIN approval_requests ar ON pr.approval_request_uid = ar.uid
		WHERE ar.status = 'pending'
		  AND pr.permission_date <= $1`
	return r.query(ctx, q, query, asOf.Format("2006-01-02"))
}

// ---- helpers ----

func qualifyColumns(columns, alias string) string {
	parts := strings.Split(columns, ", ")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = alias + "." + p
	}
	return strings.Join(out, ", ")
}

func (r *PermissionRequestRepository) query(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.PermissionRequest, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("permission_request_repository.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]*domain.PermissionRequest, 0)
	for rows.Next() {
		req, err := r.scanRowSet(rows)
		if err != nil {
			slog.Error("permission_request_repository.query.scan", "error", err)
			return nil, err
		}
		out = append(out, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PermissionRequestRepository) scanRow(row *sql.Row) (*domain.PermissionRequest, error) {
	req := &domain.PermissionRequest{}
	var typeStr string
	var permissionDate, submittedAt, createdAt, updatedAt domain.Time
	var decidedAt domain.NullTime
	var startTime, endTime, reason sql.NullString
	err := row.Scan(
		&req.ID, &req.UID, &req.EmployeeUID, &typeStr,
		&permissionDate, &startTime, &endTime, &reason,
		&submittedAt, &decidedAt, &req.ApprovalRequestUID,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	hydratePermissionRequest(req, typeStr, permissionDate, submittedAt, createdAt, updatedAt, decidedAt, startTime, endTime, reason)
	return req, nil
}

func (r *PermissionRequestRepository) scanRowSet(rows *sql.Rows) (*domain.PermissionRequest, error) {
	req := &domain.PermissionRequest{}
	var typeStr string
	var permissionDate, submittedAt, createdAt, updatedAt domain.Time
	var decidedAt domain.NullTime
	var startTime, endTime, reason sql.NullString
	err := rows.Scan(
		&req.ID, &req.UID, &req.EmployeeUID, &typeStr,
		&permissionDate, &startTime, &endTime, &reason,
		&submittedAt, &decidedAt, &req.ApprovalRequestUID,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	hydratePermissionRequest(req, typeStr, permissionDate, submittedAt, createdAt, updatedAt, decidedAt, startTime, endTime, reason)
	return req, nil
}

func hydratePermissionRequest(req *domain.PermissionRequest, typeStr string, permissionDate, submittedAt, createdAt, updatedAt domain.Time, decidedAt domain.NullTime, startTime, endTime, reason sql.NullString) {
	req.Type = domain.PermissionType(typeStr)
	req.PermissionDate = permissionDate.Time
	req.SubmittedAt = submittedAt.Time
	req.CreatedAt = createdAt.Time
	req.UpdatedAt = updatedAt.Time
	if decidedAt.Valid {
		req.DecidedAt = &decidedAt.Time
	}
	if startTime.Valid {
		req.StartTime = startTime.String
	}
	if endTime.Valid {
		req.EndTime = endTime.String
	}
	if reason.Valid {
		v := reason.String
		req.Reason = &v
	}
}
