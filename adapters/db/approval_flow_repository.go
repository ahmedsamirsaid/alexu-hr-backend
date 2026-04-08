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

type ApprovalFlowRepository struct{}

func NewApprovalFlowRepository() *ApprovalFlowRepository {
	return &ApprovalFlowRepository{}
}

func (r *ApprovalFlowRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalFlow, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, description, is_active, created_at, updated_at
		FROM approval_flows
		WHERE id = ?`

	return r.scanApprovalFlow(q.QueryRowContext(ctx, query, id))
}

func (r *ApprovalFlowRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalFlow, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, description, is_active, created_at, updated_at
		FROM approval_flows
		WHERE uid = ?`

	return r.scanApprovalFlow(q.QueryRowContext(ctx, query, uid))
}

func (r *ApprovalFlowRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.ApprovalFlow, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, description, is_active, created_at, updated_at
		FROM approval_flows
		WHERE code = ?`

	return r.scanApprovalFlow(q.QueryRowContext(ctx, query, code))
}

func (r *ApprovalFlowRepository) Create(ctx context.Context, q ports.Querier, flow *domain.ApprovalFlow) error {
	query := `
		INSERT INTO approval_flows (uid, code, name_en, name_ar, description, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	flow.CreatedAt = now
	flow.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		flow.UID, flow.Code, flow.NameEN, flow.NameAR, flow.Description,
		flow.IsActive, flow.CreatedAt, flow.UpdatedAt)
	if err != nil {
		slog.Error("approval_flow_repository.Create.exec_query", "error", err, "uid", flow.UID, "code", flow.Code)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("approval_flow_repository.Create.last_insert_id", "error", err, "uid", flow.UID)
		return err
	}
	flow.ID = id

	return nil
}

func (r *ApprovalFlowRepository) Update(ctx context.Context, q ports.Querier, flow *domain.ApprovalFlow) error {
	query := `
		UPDATE approval_flows
		SET code = ?, name_en = ?, name_ar = ?, description = ?, is_active = ?, updated_at = ?
		WHERE id = ?`

	flow.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		flow.Code, flow.NameEN, flow.NameAR, flow.Description,
		flow.IsActive, flow.UpdatedAt, flow.ID)
	if err != nil {
		slog.Error("approval_flow_repository.Update.exec_query", "error", err, "id", flow.ID, "uid", flow.UID)
	}
	return err
}

func (r *ApprovalFlowRepository) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.ApprovalFlow, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, description, is_active, created_at, updated_at
		FROM approval_flows`

	if activeOnly {
		query += ` WHERE is_active = 1`
	}

	query += ` ORDER BY code ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("approval_flow_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var flows []*domain.ApprovalFlow
	for rows.Next() {
		f, err := r.scanApprovalFlowRow(rows)
		if err != nil {
			slog.Error("approval_flow_repository.List.scan_row", "error", err)
			return nil, err
		}
		flows = append(flows, f)
	}

	if err := rows.Err(); err != nil {
		slog.Error("approval_flow_repository.List.rows_iteration", "error", err)
		return nil, err
	}
	return flows, nil
}

func (r *ApprovalFlowRepository) scanApprovalFlow(row *sql.Row) (*domain.ApprovalFlow, error) {
	var f domain.ApprovalFlow
	var createdAt, updatedAt domain.Time
	var isActive int
	err := row.Scan(
		&f.ID, &f.UID, &f.Code, &f.NameEN, &f.NameAR, &f.Description,
		&isActive, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("approval_flow_repository.scanApprovalFlow.scan_row", "error", err)
		return nil, err
	}
	f.IsActive = isActive == 1
	f.CreatedAt = createdAt.Time
	f.UpdatedAt = updatedAt.Time
	return &f, nil
}

func (r *ApprovalFlowRepository) scanApprovalFlowRow(rows *sql.Rows) (*domain.ApprovalFlow, error) {
	var f domain.ApprovalFlow
	var createdAt, updatedAt domain.Time
	var isActive int
	err := rows.Scan(
		&f.ID, &f.UID, &f.Code, &f.NameEN, &f.NameAR, &f.Description,
		&isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	f.IsActive = isActive == 1
	f.CreatedAt = createdAt.Time
	f.UpdatedAt = updatedAt.Time
	return &f, nil
}
