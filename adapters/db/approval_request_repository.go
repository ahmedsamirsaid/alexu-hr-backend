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

type ApprovalRequestRepository struct{}

func NewApprovalRequestRepository() *ApprovalRequestRepository {
	return &ApprovalRequestRepository{}
}

func (r *ApprovalRequestRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalRequest, error) {
	query := `
		SELECT id, uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at
		FROM approval_requests
		WHERE id = ?`

	return r.scanApprovalRequest(q.QueryRowContext(ctx, query, id))
}

func (r *ApprovalRequestRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalRequest, error) {
	query := `
		SELECT id, uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at
		FROM approval_requests
		WHERE uid = ?`

	return r.scanApprovalRequest(q.QueryRowContext(ctx, query, uid))
}

func (r *ApprovalRequestRepository) Create(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	query := `
		INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		request.UID, request.ApprovalFlowUID, request.RequesterUID,
		request.CurrentStep, request.MaxStep, request.Status,
		request.CreatedAt, request.UpdatedAt)
	if err != nil {
		slog.Error("approval_request_repository.Create.exec_query", "error", err, "uid", request.UID, "requester_uid", request.RequesterUID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("approval_request_repository.Create.last_insert_id", "error", err, "uid", request.UID)
		return err
	}
	request.ID = id

	return nil
}

func (r *ApprovalRequestRepository) Update(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	query := `
		UPDATE approval_requests
		SET approval_flow_uid = ?, current_step = ?, max_step = ?, status = ?, updated_at = ?
		WHERE id = ?`

	request.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		request.ApprovalFlowUID, request.CurrentStep, request.MaxStep, request.Status, request.UpdatedAt, request.ID)
	if err != nil {
		slog.Error("approval_request_repository.Update.exec_query", "error", err, "id", request.ID, "uid", request.UID)
	}
	return err
}

func (r *ApprovalRequestRepository) ListByRequester(ctx context.Context, q ports.Querier, requesterUID string) ([]*domain.ApprovalRequest, error) {
	query := `
		SELECT id, uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at
		FROM approval_requests
		WHERE requester_uid = ?
		ORDER BY created_at DESC`

	return r.queryApprovalRequests(ctx, q, query, requesterUID)
}

func (r *ApprovalRequestRepository) ListPending(ctx context.Context, q ports.Querier) ([]*domain.ApprovalRequest, error) {
	query := `
		SELECT id, uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at
		FROM approval_requests
		WHERE status = 'pending'
		ORDER BY created_at ASC`

	return r.queryApprovalRequests(ctx, q, query)
}

func (r *ApprovalRequestRepository) ListPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) ([]*domain.ApprovalRequest, error) {
	query := `
		SELECT id, uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at
		FROM approval_requests
		WHERE approval_flow_uid = ? AND current_step = ? AND status = 'pending'
		ORDER BY created_at ASC`

	return r.queryApprovalRequests(ctx, q, query, approvalFlowUID, stepOrder)
}

func (r *ApprovalRequestRepository) CountPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM approval_requests
		WHERE approval_flow_uid = ? AND current_step = ? AND status = 'pending'`

	var count int
	err := q.QueryRowContext(ctx, query, approvalFlowUID, stepOrder).Scan(&count)
	if err != nil {
		slog.Error("approval_request_repository.CountPendingByFlowAndStep.scan_row", "error", err, "flow_uid", approvalFlowUID, "step_order", stepOrder)
	}
	return count, err
}

func (r *ApprovalRequestRepository) queryApprovalRequests(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.ApprovalRequest, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("approval_request_repository.queryApprovalRequests.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var requests []*domain.ApprovalRequest
	for rows.Next() {
		req, err := r.scanApprovalRequestRow(rows)
		if err != nil {
			slog.Error("approval_request_repository.queryApprovalRequests.scan_row", "error", err)
			return nil, err
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		slog.Error("approval_request_repository.queryApprovalRequests.rows_iteration", "error", err)
		return nil, err
	}
	return requests, nil
}

func (r *ApprovalRequestRepository) scanApprovalRequest(row *sql.Row) (*domain.ApprovalRequest, error) {
	var req domain.ApprovalRequest
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&req.ID, &req.UID, &req.ApprovalFlowUID, &req.RequesterUID,
		&req.CurrentStep, &req.MaxStep, &req.Status,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("approval_request_repository.scanApprovalRequest.scan_row", "error", err)
		return nil, err
	}
	req.CreatedAt = createdAt.Time
	req.UpdatedAt = updatedAt.Time
	return &req, nil
}

func (r *ApprovalRequestRepository) scanApprovalRequestRow(rows *sql.Rows) (*domain.ApprovalRequest, error) {
	var req domain.ApprovalRequest
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&req.ID, &req.UID, &req.ApprovalFlowUID, &req.RequesterUID,
		&req.CurrentStep, &req.MaxStep, &req.Status,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	req.CreatedAt = createdAt.Time
	req.UpdatedAt = updatedAt.Time
	return &req, nil
}
