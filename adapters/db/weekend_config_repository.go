package db

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type WeekendConfigRepository struct{}

func NewWeekendConfigRepository() *WeekendConfigRepository {
	return &WeekendConfigRepository{}
}

func (r *WeekendConfigRepository) List(ctx context.Context, q ports.Querier) ([]*domain.WeekendConfig, error) {
	query := `
		SELECT id, uid, day_of_week, created_at, updated_at
		FROM weekend_config
		ORDER BY day_of_week`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("weekend_config_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var configs []*domain.WeekendConfig
	for rows.Next() {
		var wc domain.WeekendConfig
		var createdAt, updatedAt domain.Time
		err := rows.Scan(&wc.ID, &wc.UID, &wc.DayOfWeek, &createdAt, &updatedAt)
		if err != nil {
			slog.Error("weekend_config_repository.List.scan_row", "error", err)
			return nil, err
		}
		wc.CreatedAt = createdAt.Time
		wc.UpdatedAt = updatedAt.Time
		configs = append(configs, &wc)
	}

	if err := rows.Err(); err != nil {
		slog.Error("weekend_config_repository.List.rows_iteration", "error", err)
		return nil, err
	}
	return configs, nil
}

func (r *WeekendConfigRepository) GetWeekendDays(ctx context.Context, q ports.Querier) ([]int, error) {
	query := `SELECT day_of_week FROM weekend_config ORDER BY day_of_week`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("weekend_config_repository.GetWeekendDays.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var days []int
	for rows.Next() {
		var day int
		if err := rows.Scan(&day); err != nil {
			slog.Error("weekend_config_repository.GetWeekendDays.scan_row", "error", err)
			return nil, err
		}
		days = append(days, day)
	}

	if err := rows.Err(); err != nil {
		slog.Error("weekend_config_repository.GetWeekendDays.rows_iteration", "error", err)
		return nil, err
	}
	return days, nil
}
