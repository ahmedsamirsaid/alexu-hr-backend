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

type WorkHoursConfigRepository struct{}

func NewWorkHoursConfigRepository() *WorkHoursConfigRepository {
	return &WorkHoursConfigRepository{}
}

func (r *WorkHoursConfigRepository) Get(ctx context.Context, q ports.Querier) (*domain.WorkHoursConfig, error) {
	query := `
		SELECT id, work_day_start, work_day_end, late_grace_minutes, early_grace_minutes, created_at, updated_at
		FROM work_hours_configs
		WHERE id = 1`

	row := q.QueryRowContext(ctx, query)

	var cfg domain.WorkHoursConfig
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&cfg.ID,
		&cfg.WorkDayStart,
		&cfg.WorkDayEnd,
		&cfg.LateGraceMinutes,
		&cfg.EarlyGraceMinutes,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("work_hours_config_repository.Get.scan", "error", err)
		return nil, err
	}

	cfg.CreatedAt = createdAt.Time
	cfg.UpdatedAt = updatedAt.Time
	return &cfg, nil
}

func (r *WorkHoursConfigRepository) Upsert(ctx context.Context, q ports.Querier, cfg *domain.WorkHoursConfig) error {
	query := `
		INSERT INTO work_hours_configs (
			id, work_day_start, work_day_end, late_grace_minutes, early_grace_minutes, created_at, updated_at
		)
		VALUES (1, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			work_day_start = excluded.work_day_start,
			work_day_end = excluded.work_day_end,
			late_grace_minutes = excluded.late_grace_minutes,
			early_grace_minutes = excluded.early_grace_minutes,
			updated_at = excluded.updated_at`

	now := time.Now()
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = now
	}
	cfg.UpdatedAt = now

	_, err := q.ExecContext(ctx, query,
		cfg.WorkDayStart,
		cfg.WorkDayEnd,
		cfg.LateGraceMinutes,
		cfg.EarlyGraceMinutes,
		cfg.CreatedAt,
		cfg.UpdatedAt,
	)
	if err != nil {
		slog.Error("work_hours_config_repository.Upsert.exec", "error", err)
	}

	return err
}
