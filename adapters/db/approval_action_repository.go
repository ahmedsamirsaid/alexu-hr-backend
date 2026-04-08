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

type ApprovalActionRepository struct{}

func NewApprovalActionRepository() *ApprovalActionRepository {
	return &ApprovalActionRepository{}
}

func (r *ApprovalActionRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalAction, error) {
	query := `
		SELECT id, uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at
		FROM approval_actions
		WHERE id = ?`

	return r.scanApprovalAction(q.QueryRowContext(ctx, query, id))
}

func (r *ApprovalActionRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalAction, error) {
	query := `
		SELECT id, uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at
		FROM approval_actions
		WHERE uid = ?`

	return r.scanApprovalAction(q.QueryRowContext(ctx, query, uid))
}

func (r *ApprovalActionRepository) Create(ctx context.Context, q ports.Querier, action *domain.ApprovalAction) error {
	query := `
		INSERT INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	action.CreatedAt = now
	if action.ActedAt.IsZero() {
		action.ActedAt = now
	}

	result, err := q.ExecContext(ctx, query,
		action.UID, action.ApprovalRequestUID, action.Action,
		action.StepOrder, action.ActorUID, action.Comments,
		action.ActedAt, action.CreatedAt)
	if err != nil {
		slog.Error("approval_action_repository.Create.exec_query", "error", err, "uid", action.UID, "request_uid", action.ApprovalRequestUID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("approval_action_repository.Create.last_insert_id", "error", err, "uid", action.UID)
		return err
	}
	action.ID = id

	return nil
}

func (r *ApprovalActionRepository) ListByRequest(ctx context.Context, q ports.Querier, approvalRequestUID string) ([]*domain.ApprovalAction, error) {
	query := `
		SELECT id, uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at
		FROM approval_actions
		WHERE approval_request_uid = ?
		ORDER BY acted_at ASC`

	rows, err := q.QueryContext(ctx, query, approvalRequestUID)
	if err != nil {
		slog.Error("approval_action_repository.ListByRequest.query", "error", err, "request_uid", approvalRequestUID)
		return nil, err
	}
	defer rows.Close()

	var actions []*domain.ApprovalAction
	for rows.Next() {
		a, err := r.scanApprovalActionRow(rows)
		if err != nil {
			slog.Error("approval_action_repository.ListByRequest.scan_row", "error", err, "request_uid", approvalRequestUID)
			return nil, err
		}
		actions = append(actions, a)
	}

	if err := rows.Err(); err != nil {
		slog.Error("approval_action_repository.ListByRequest.rows_iteration", "error", err, "request_uid", approvalRequestUID)
		return nil, err
	}
	return actions, nil
}

func (r *ApprovalActionRepository) scanApprovalAction(row *sql.Row) (*domain.ApprovalAction, error) {
	var a domain.ApprovalAction
	var actedAt, createdAt domain.Time
	err := row.Scan(
		&a.ID, &a.UID, &a.ApprovalRequestUID, &a.Action,
		&a.StepOrder, &a.ActorUID, &a.Comments,
		&actedAt, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("approval_action_repository.scanApprovalAction.scan_row", "error", err)
		return nil, err
	}
	a.ActedAt = actedAt.Time
	a.CreatedAt = createdAt.Time
	return &a, nil
}

func (r *ApprovalActionRepository) scanApprovalActionRow(rows *sql.Rows) (*domain.ApprovalAction, error) {
	var a domain.ApprovalAction
	var actedAt, createdAt domain.Time
	err := rows.Scan(
		&a.ID, &a.UID, &a.ApprovalRequestUID, &a.Action,
		&a.StepOrder, &a.ActorUID, &a.Comments,
		&actedAt, &createdAt)
	if err != nil {
		return nil, err
	}
	a.ActedAt = actedAt.Time
	a.CreatedAt = createdAt.Time
	return &a, nil
}
