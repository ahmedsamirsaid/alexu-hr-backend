package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type IncentiveBonusRepository struct{}

func NewIncentiveBonusRepository() *IncentiveBonusRepository {
	return &IncentiveBonusRepository{}
}

func (r *IncentiveBonusRepository) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.IncentiveBonus, error) {
	query := `
		SELECT id, employee_uid, bonus_date, decision_number, decision_date, decision_image_url, created_at, updated_at
		FROM incentive_bonus
		WHERE employee_uid = ?
		ORDER BY bonus_date DESC, id DESC`

	rows, err := q.QueryContext(ctx, query, employeeUID)
	if err != nil {
		slog.Error("incentive_bonus_repository.ListByEmployeeUID.query", "error", err, "employee_uid", employeeUID)
		return nil, err
	}
	defer rows.Close()

	bonuses := make([]*domain.IncentiveBonus, 0)
	for rows.Next() {
		var bonus domain.IncentiveBonus
		if err := rows.Scan(
			&bonus.ID,
			&bonus.EmployeeUID,
			&bonus.BonusDate,
			&bonus.DecisionNumber,
			&bonus.DecisionDate,
			&bonus.DecisionImageURL,
			&bonus.CreatedAt,
			&bonus.UpdatedAt,
		); err != nil {
			slog.Error("incentive_bonus_repository.ListByEmployeeUID.scan", "error", err, "employee_uid", employeeUID)
			return nil, err
		}
		bonuses = append(bonuses, &bonus)
	}

	if err := rows.Err(); err != nil {
		slog.Error("incentive_bonus_repository.ListByEmployeeUID.rows", "error", err, "employee_uid", employeeUID)
		return nil, err
	}

	return bonuses, nil
}

func (r *IncentiveBonusRepository) Create(ctx context.Context, q ports.Querier, bonus *domain.IncentiveBonus) error {
	query := `
		INSERT INTO incentive_bonus (
			employee_uid, bonus_date, decision_number, decision_date, decision_image_url, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	bonus.CreatedAt = now
	bonus.UpdatedAt = now

	result, err := q.ExecContext(
		ctx,
		query,
		bonus.EmployeeUID,
		bonus.BonusDate,
		bonus.DecisionNumber,
		bonus.DecisionDate,
		bonus.DecisionImageURL,
		bonus.CreatedAt,
		bonus.UpdatedAt,
	)
	if err != nil {
		slog.Error("incentive_bonus_repository.Create.exec_query", "error", err, "employee_uid", bonus.EmployeeUID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("incentive_bonus_repository.Create.last_insert_id", "error", err, "employee_uid", bonus.EmployeeUID)
		return err
	}
	bonus.ID = id

	return nil
}

func (r *IncentiveBonusRepository) DeleteByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) error {
	query := `DELETE FROM incentive_bonus WHERE employee_uid = ?`
	if _, err := q.ExecContext(ctx, query, employeeUID); err != nil {
		slog.Error("incentive_bonus_repository.DeleteByEmployeeUID.exec_query", "error", err, "employee_uid", employeeUID)
		return err
	}
	return nil
}
