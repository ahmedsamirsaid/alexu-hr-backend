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

type ApprovalFlowStepRepository struct{}

func NewApprovalFlowStepRepository() *ApprovalFlowStepRepository {
	return &ApprovalFlowStepRepository{}
}

func (r *ApprovalFlowStepRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalFlowStep, error) {
	query := `
		SELECT id, uid, approval_flow_uid, step_order, role_uid, created_at, updated_at
		FROM approval_flow_steps
		WHERE id = ?`

	return r.scanApprovalFlowStep(q.QueryRowContext(ctx, query, id))
}

func (r *ApprovalFlowStepRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalFlowStep, error) {
	query := `
		SELECT id, uid, approval_flow_uid, step_order, role_uid, created_at, updated_at
		FROM approval_flow_steps
		WHERE uid = ?`

	return r.scanApprovalFlowStep(q.QueryRowContext(ctx, query, uid))
}

func (r *ApprovalFlowStepRepository) Create(ctx context.Context, q ports.Querier, step *domain.ApprovalFlowStep) error {
	query := `
		INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	step.CreatedAt = now
	step.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		step.UID, step.ApprovalFlowUID, step.StepOrder, step.RoleUID,
		step.CreatedAt, step.UpdatedAt)
	if err != nil {
		slog.Error("approval_flow_step_repository.Create.exec_query", "error", err, "uid", step.UID, "flow_uid", step.ApprovalFlowUID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("approval_flow_step_repository.Create.last_insert_id", "error", err, "uid", step.UID)
		return err
	}
	step.ID = id

	return nil
}

func (r *ApprovalFlowStepRepository) Update(ctx context.Context, q ports.Querier, step *domain.ApprovalFlowStep) error {
	query := `
		UPDATE approval_flow_steps
		SET step_order = ?, role_uid = ?, updated_at = ?
		WHERE id = ?`

	step.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		step.StepOrder, step.RoleUID, step.UpdatedAt, step.ID)
	if err != nil {
		slog.Error("approval_flow_step_repository.Update.exec_query", "error", err, "id", step.ID, "uid", step.UID)
	}
	return err
}

func (r *ApprovalFlowStepRepository) Delete(ctx context.Context, q ports.Querier, uid string) error {
	query := `DELETE FROM approval_flow_steps WHERE uid = ?`
	_, err := q.ExecContext(ctx, query, uid)
	if err != nil {
		slog.Error("approval_flow_step_repository.Delete.exec_query", "error", err, "uid", uid)
	}
	return err
}

func (r *ApprovalFlowStepRepository) ListByFlow(ctx context.Context, q ports.Querier, approvalFlowUID string) ([]*domain.ApprovalFlowStep, error) {
	query := `
		SELECT id, uid, approval_flow_uid, step_order, role_uid, created_at, updated_at
		FROM approval_flow_steps
		WHERE approval_flow_uid = ?
		ORDER BY step_order ASC`

	rows, err := q.QueryContext(ctx, query, approvalFlowUID)
	if err != nil {
		slog.Error("approval_flow_step_repository.ListByFlow.query", "error", err, "flow_uid", approvalFlowUID)
		return nil, err
	}
	defer rows.Close()

	var steps []*domain.ApprovalFlowStep
	for rows.Next() {
		s, err := r.scanApprovalFlowStepRow(rows)
		if err != nil {
			slog.Error("approval_flow_step_repository.ListByFlow.scan_row", "error", err, "flow_uid", approvalFlowUID)
			return nil, err
		}
		steps = append(steps, s)
	}

	if err := rows.Err(); err != nil {
		slog.Error("approval_flow_step_repository.ListByFlow.rows_iteration", "error", err, "flow_uid", approvalFlowUID)
		return nil, err
	}
	return steps, nil
}

func (r *ApprovalFlowStepRepository) CountByFlow(ctx context.Context, q ports.Querier, approvalFlowUID string) (int, error) {
	query := `SELECT COUNT(*) FROM approval_flow_steps WHERE approval_flow_uid = ?`
	var count int
	err := q.QueryRowContext(ctx, query, approvalFlowUID).Scan(&count)
	if err != nil {
		slog.Error("approval_flow_step_repository.CountByFlow.scan_row", "error", err, "flow_uid", approvalFlowUID)
	}
	return count, err
}

func (r *ApprovalFlowStepRepository) GetByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (*domain.ApprovalFlowStep, error) {
	query := `
		SELECT id, uid, approval_flow_uid, step_order, role_uid, created_at, updated_at
		FROM approval_flow_steps
		WHERE approval_flow_uid = ? AND step_order = ?`

	return r.scanApprovalFlowStep(q.QueryRowContext(ctx, query, approvalFlowUID, stepOrder))
}

func (r *ApprovalFlowStepRepository) HasPendingRequestsAtStep(ctx context.Context, q ports.Querier, stepUID string) (bool, error) {
	// Get the step to find its flow and step_order
	step, err := r.GetByUID(ctx, q, stepUID)
	if err != nil {
		return false, err
	}
	if step == nil {
		return false, nil
	}

	// Check if any pending approval requests are at this step
	query := `
		SELECT COUNT(*)
		FROM approval_requests
		WHERE approval_flow_uid = ? AND current_step = ? AND status = 'pending'`

	var count int
	err = q.QueryRowContext(ctx, query, step.ApprovalFlowUID, step.StepOrder).Scan(&count)
	if err != nil {
		slog.Error("approval_flow_step_repository.HasPendingRequestsAtStep.scan_row", "error", err, "step_uid", stepUID)
		return false, err
	}
	return count > 0, nil
}

func (r *ApprovalFlowStepRepository) scanApprovalFlowStep(row *sql.Row) (*domain.ApprovalFlowStep, error) {
	var s domain.ApprovalFlowStep
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&s.ID, &s.UID, &s.ApprovalFlowUID, &s.StepOrder, &s.RoleUID,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("approval_flow_step_repository.scanApprovalFlowStep.scan_row", "error", err)
		return nil, err
	}
	s.CreatedAt = createdAt.Time
	s.UpdatedAt = updatedAt.Time
	return &s, nil
}

func (r *ApprovalFlowStepRepository) scanApprovalFlowStepRow(rows *sql.Rows) (*domain.ApprovalFlowStep, error) {
	var s domain.ApprovalFlowStep
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&s.ID, &s.UID, &s.ApprovalFlowUID, &s.StepOrder, &s.RoleUID,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	s.CreatedAt = createdAt.Time
	s.UpdatedAt = updatedAt.Time
	return &s, nil
}
