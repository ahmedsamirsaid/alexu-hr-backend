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

type EmployeeProfileChangeRequestRepository struct{}

func NewEmployeeProfileChangeRequestRepository() *EmployeeProfileChangeRequestRepository {
	return &EmployeeProfileChangeRequestRepository{}
}

func (r *EmployeeProfileChangeRequestRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.EmployeeProfileChangeRequest, error) {
	query := `
		SELECT id, uid, employee_uid, approval_request_uid, submitted_by_employee_uid,
		       current_financial_grade, current_id_card_valid_until, current_marital_status,
		       requested_financial_grade, requested_id_card_valid_until, requested_marital_status,
		       comments, created_at, updated_at
		FROM employee_profile_change_requests
		WHERE uid = $1`
	return r.scanRequest(q.QueryRowContext(ctx, query, uid))
}

func (r *EmployeeProfileChangeRequestRepository) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.EmployeeProfileChangeRequest, error) {
	query := `
		SELECT id, uid, employee_uid, approval_request_uid, submitted_by_employee_uid,
		       current_financial_grade, current_id_card_valid_until, current_marital_status,
		       requested_financial_grade, requested_id_card_valid_until, requested_marital_status,
		       comments, created_at, updated_at
		FROM employee_profile_change_requests
		WHERE approval_request_uid = $1`
	return r.scanRequest(q.QueryRowContext(ctx, query, approvalRequestUID))
}

func (r *EmployeeProfileChangeRequestRepository) Create(ctx context.Context, q ports.Querier, request *domain.EmployeeProfileChangeRequest) error {
	query := `
		INSERT INTO employee_profile_change_requests (
			uid, employee_uid, approval_request_uid, submitted_by_employee_uid,
			current_financial_grade, current_id_card_valid_until, current_marital_status,
			requested_financial_grade, requested_id_card_valid_until, requested_marital_status,
			comments, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now

	var id int64
	err := q.QueryRowContext(
		ctx,
		query,
		request.UID,
		request.EmployeeUID,
		request.ApprovalRequestUID,
		request.SubmittedByEmployeeUID,
		request.CurrentFinancialGrade,
		nullableDate(request.CurrentIDCardValidUntil),
		request.CurrentMaritalStatus,
		request.RequestedFinancialGrade,
		nullableDate(request.RequestedIDCardValidUntil),
		request.RequestedMaritalStatus,
		request.Comments,
		request.CreatedAt,
		request.UpdatedAt,
	).Scan(&id)
	if err != nil {
		slog.Error("employee_profile_change_request_repository.Create.query_row", "error", err, "uid", request.UID)
		return err
	}
	request.ID = id
	return nil
}

func (r *EmployeeProfileChangeRequestRepository) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.EmployeeProfileChangeRequest, error) {
	query := `
		SELECT id, uid, employee_uid, approval_request_uid, submitted_by_employee_uid,
		       current_financial_grade, current_id_card_valid_until, current_marital_status,
		       requested_financial_grade, requested_id_card_valid_until, requested_marital_status,
		       comments, created_at, updated_at
		FROM employee_profile_change_requests
		WHERE employee_uid = $1
		ORDER BY created_at DESC`
	return r.queryRequests(ctx, q, query, employeeUID)
}

func (r *EmployeeProfileChangeRequestRepository) ListPending(ctx context.Context, q ports.Querier) ([]*domain.EmployeeProfileChangeRequest, error) {
	query := `
		SELECT epcr.id, epcr.uid, epcr.employee_uid, epcr.approval_request_uid, epcr.submitted_by_employee_uid,
		       epcr.current_financial_grade, epcr.current_id_card_valid_until, epcr.current_marital_status,
		       epcr.requested_financial_grade, epcr.requested_id_card_valid_until, epcr.requested_marital_status,
		       epcr.comments, epcr.created_at, epcr.updated_at
		FROM employee_profile_change_requests epcr
		JOIN approval_requests ar ON ar.uid = epcr.approval_request_uid
		WHERE ar.status = 'pending'
		ORDER BY epcr.created_at ASC`
	return r.queryRequests(ctx, q, query)
}

func (r *EmployeeProfileChangeRequestRepository) HasPendingForEmployee(ctx context.Context, q ports.Querier, employeeUID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM employee_profile_change_requests epcr
		JOIN approval_requests ar ON ar.uid = epcr.approval_request_uid
		WHERE epcr.employee_uid = $1 AND ar.status = 'pending'`

	var count int
	if err := q.QueryRowContext(ctx, query, employeeUID).Scan(&count); err != nil {
		slog.Error("employee_profile_change_request_repository.HasPendingForEmployee.scan", "error", err, "employee_uid", employeeUID)
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeProfileChangeRequestRepository) queryRequests(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.EmployeeProfileChangeRequest, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("employee_profile_change_request_repository.queryRequests.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*domain.EmployeeProfileChangeRequest
	for rows.Next() {
		request, err := r.scanRequestRow(rows)
		if err != nil {
			slog.Error("employee_profile_change_request_repository.queryRequests.scan_row", "error", err)
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		slog.Error("employee_profile_change_request_repository.queryRequests.rows_iteration", "error", err)
		return nil, err
	}
	return requests, nil
}

func (r *EmployeeProfileChangeRequestRepository) scanRequest(row *sql.Row) (*domain.EmployeeProfileChangeRequest, error) {
	var request domain.EmployeeProfileChangeRequest
	var currentIDCardValidUntil sql.NullString
	var requestedIDCardValidUntil sql.NullString
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&request.ID,
		&request.UID,
		&request.EmployeeUID,
		&request.ApprovalRequestUID,
		&request.SubmittedByEmployeeUID,
		&request.CurrentFinancialGrade,
		&currentIDCardValidUntil,
		&request.CurrentMaritalStatus,
		&request.RequestedFinancialGrade,
		&requestedIDCardValidUntil,
		&request.RequestedMaritalStatus,
		&request.Comments,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	request.CurrentIDCardValidUntil = nullableDatePtr(currentIDCardValidUntil)
	request.RequestedIDCardValidUntil = nullableDatePtr(requestedIDCardValidUntil)
	request.CreatedAt = createdAt.Time
	request.UpdatedAt = updatedAt.Time
	return &request, nil
}

func (r *EmployeeProfileChangeRequestRepository) scanRequestRow(rows *sql.Rows) (*domain.EmployeeProfileChangeRequest, error) {
	var request domain.EmployeeProfileChangeRequest
	var currentIDCardValidUntil sql.NullString
	var requestedIDCardValidUntil sql.NullString
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&request.ID,
		&request.UID,
		&request.EmployeeUID,
		&request.ApprovalRequestUID,
		&request.SubmittedByEmployeeUID,
		&request.CurrentFinancialGrade,
		&currentIDCardValidUntil,
		&request.CurrentMaritalStatus,
		&request.RequestedFinancialGrade,
		&requestedIDCardValidUntil,
		&request.RequestedMaritalStatus,
		&request.Comments,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	request.CurrentIDCardValidUntil = nullableDatePtr(currentIDCardValidUntil)
	request.RequestedIDCardValidUntil = nullableDatePtr(requestedIDCardValidUntil)
	request.CreatedAt = createdAt.Time
	request.UpdatedAt = updatedAt.Time
	return &request, nil
}
