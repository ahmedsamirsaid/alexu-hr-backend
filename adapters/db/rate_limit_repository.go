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

type RateLimitRepository struct{}

func NewRateLimitRepository() *RateLimitRepository {
	return &RateLimitRepository{}
}

func (r *RateLimitRepository) GetByScopeAndSubject(ctx context.Context, q ports.Querier, scope, subjectKey string) (*domain.RateLimitRecord, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, scope, subject_key, window_started_at, attempt_count, locked_until, updated_at
		FROM rate_limit_records
		WHERE scope = ? AND subject_key = ?`,
		scope, subjectKey,
	)

	var record domain.RateLimitRecord
	var windowStartedAt, updatedAt time.Time
	var lockedUntil sql.NullTime
	err := row.Scan(
		&record.ID,
		&record.Scope,
		&record.SubjectKey,
		&windowStartedAt,
		&record.AttemptCount,
		&lockedUntil,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("rate_limit_repository.GetByScopeAndSubject.scan", "error", err, "scope", scope)
		return nil, err
	}

	record.WindowStartedAt = windowStartedAt
	record.UpdatedAt = updatedAt
	if lockedUntil.Valid {
		record.LockedUntil = &lockedUntil.Time
	}
	return &record, nil
}

func (r *RateLimitRepository) Upsert(ctx context.Context, q ports.Querier, record *domain.RateLimitRecord) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO rate_limit_records (
			scope, subject_key, window_started_at, attempt_count, locked_until, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(scope, subject_key) DO UPDATE SET
			window_started_at = excluded.window_started_at,
			attempt_count = excluded.attempt_count,
			locked_until = excluded.locked_until,
			updated_at = excluded.updated_at`,
		record.Scope,
		record.SubjectKey,
		record.WindowStartedAt,
		record.AttemptCount,
		record.LockedUntil,
		record.UpdatedAt,
	)
	if err != nil {
		slog.Error("rate_limit_repository.Upsert.exec", "error", err, "scope", record.Scope)
	}
	return err
}

func (r *RateLimitRepository) DeleteByScopeAndSubject(ctx context.Context, q ports.Querier, scope, subjectKey string) error {
	_, err := q.ExecContext(ctx, `DELETE FROM rate_limit_records WHERE scope = ? AND subject_key = ?`, scope, subjectKey)
	if err != nil {
		slog.Error("rate_limit_repository.DeleteByScopeAndSubject.exec", "error", err, "scope", scope)
	}
	return err
}
