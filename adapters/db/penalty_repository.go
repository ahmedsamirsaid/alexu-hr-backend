package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type PenaltyRepository struct{}

func NewPenaltyRepository() *PenaltyRepository {
	return &PenaltyRepository{}
}

func (r *PenaltyRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Penalty, error) {
	query := `
		SELECT id, employee_uid, penalty_type, penalty_reason, penalty_decision_number, penalty_decision_date, penalty_decision_file_url, created_at, updated_at
		FROM penalties
		WHERE id = $1`

	row := q.QueryRowContext(ctx, query, id)

	var penalty domain.Penalty
	if err := row.Scan(
		&penalty.ID,
		&penalty.EmployeeUID,
		&penalty.PenaltyType,
		&penalty.PenaltyReason,
		&penalty.PenaltyDecisionNumber,
		&penalty.PenaltyDecisionDate,
		&penalty.PenaltyDecisionFileURL,
		&penalty.CreatedAt,
		&penalty.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("penalty_repository.GetByID.scan", "error", err, "penalty_id", id)
		return nil, err
	}

	return &penalty, nil
}

func (r *PenaltyRepository) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.Penalty, error) {
	query := `
		SELECT id, employee_uid, penalty_type, penalty_reason, penalty_decision_number, penalty_decision_date, penalty_decision_file_url, created_at, updated_at
		FROM penalties
		WHERE employee_uid = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM penalties_removed pr
			WHERE pr.penalty_id = penalties.id
		  )
		ORDER BY penalty_decision_date DESC, id DESC`

	rows, err := q.QueryContext(ctx, query, employeeUID)
	if err != nil {
		slog.Error("penalty_repository.ListByEmployeeUID.query", "error", err, "employee_uid", employeeUID)
		return nil, err
	}
	defer rows.Close()

	penalties := make([]*domain.Penalty, 0)
	for rows.Next() {
		var penalty domain.Penalty
		if err := rows.Scan(
			&penalty.ID,
			&penalty.EmployeeUID,
			&penalty.PenaltyType,
			&penalty.PenaltyReason,
			&penalty.PenaltyDecisionNumber,
			&penalty.PenaltyDecisionDate,
			&penalty.PenaltyDecisionFileURL,
			&penalty.CreatedAt,
			&penalty.UpdatedAt,
		); err != nil {
			slog.Error("penalty_repository.ListByEmployeeUID.scan", "error", err, "employee_uid", employeeUID)
			return nil, err
		}
		penalties = append(penalties, &penalty)
	}

	if err := rows.Err(); err != nil {
		slog.Error("penalty_repository.ListByEmployeeUID.rows", "error", err, "employee_uid", employeeUID)
		return nil, err
	}

	return penalties, nil
}

func (r *PenaltyRepository) Create(ctx context.Context, q ports.Querier, penalty *domain.Penalty) error {
	query := `
		INSERT INTO penalties (
			employee_uid, penalty_type, penalty_reason, penalty_decision_number, penalty_decision_date, penalty_decision_file_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	penalty.CreatedAt = now
	penalty.UpdatedAt = now

	var id int64
	err := q.QueryRowContext(
		ctx,
		query,
		penalty.EmployeeUID,
		penalty.PenaltyType,
		penalty.PenaltyReason,
		penalty.PenaltyDecisionNumber,
		penalty.PenaltyDecisionDate,
		penalty.PenaltyDecisionFileURL,
		penalty.CreatedAt,
		penalty.UpdatedAt,
	).Scan(&id)
	if err != nil {
		slog.Error("penalty_repository.Create.query_row", "error", err, "employee_uid", penalty.EmployeeUID)
		return err
	}
	penalty.ID = id

	return nil
}

func (r *PenaltyRepository) IsRemoved(ctx context.Context, q ports.Querier, penaltyID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM penalties_removed WHERE penalty_id = $1)`

	var exists bool
	if err := q.QueryRowContext(ctx, query, penaltyID).Scan(&exists); err != nil {
		slog.Error("penalty_repository.IsRemoved.scan", "error", err, "penalty_id", penaltyID)
		return false, err
	}

	return exists, nil
}

func (r *PenaltyRepository) CreateRemoval(ctx context.Context, q ports.Querier, removal *domain.PenaltyRemoval) error {
	query := `
		INSERT INTO penalties_removed (
			penalty_id, penalty_removal_type, penalty_removal_number, penalty_removal_date, notes,
			penalty_withdrawal_decision_file_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	removal.CreatedAt = now
	removal.UpdatedAt = now

	var id int64
	err := q.QueryRowContext(
		ctx,
		query,
		removal.PenaltyID,
		removal.PenaltyRemovalType,
		removal.PenaltyRemovalNumber,
		removal.PenaltyRemovalDate,
		removal.Notes,
		removal.PenaltyWithdrawalDecisionFileURL,
		removal.CreatedAt,
		removal.UpdatedAt,
	).Scan(&id)
	if err != nil {
		slog.Error("penalty_repository.CreateRemoval.query_row", "error", err, "penalty_id", removal.PenaltyID)
		return err
	}
	removal.ID = id

	return nil
}

func (r *PenaltyRepository) DeleteByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) error {
	query := `DELETE FROM penalties WHERE employee_uid = $1`
	if _, err := q.ExecContext(ctx, query, employeeUID); err != nil {
		slog.Error("penalty_repository.DeleteByEmployeeUID.exec_query", "error", err, "employee_uid", employeeUID)
		return err
	}
	return nil
}
