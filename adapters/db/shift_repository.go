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

type ShiftRepository struct{}

func NewShiftRepository() *ShiftRepository {
	return &ShiftRepository{}
}

func (r *ShiftRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Shift, error) {
	query := `
		SELECT id, uid, start_time, end_time, grace_minutes, created_at, updated_at
		FROM shifts
		WHERE uid = $1`

	return r.scanShift(q.QueryRowContext(ctx, query, uid))
}

func (r *ShiftRepository) List(ctx context.Context, q ports.Querier) ([]*domain.Shift, error) {
	query := `
		SELECT id, uid, start_time, end_time, grace_minutes, created_at, updated_at
		FROM shifts
		ORDER BY start_time ASC, end_time ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("shift_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var shifts []*domain.Shift
	for rows.Next() {
		shift, err := r.scanShiftRow(rows)
		if err != nil {
			slog.Error("shift_repository.List.scan_row", "error", err)
			return nil, err
		}
		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		slog.Error("shift_repository.List.rows_iteration", "error", err)
		return nil, err
	}

	return shifts, nil
}

func (r *ShiftRepository) Upsert(ctx context.Context, q ports.Querier, shift *domain.Shift) error {
	query := `
		INSERT INTO shifts (id, uid, start_time, end_time, grace_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(uid) DO UPDATE SET
			start_time = excluded.start_time,
			end_time = excluded.end_time,
			grace_minutes = excluded.grace_minutes,
			updated_at = excluded.updated_at`

	now := time.Now()
	if shift.CreatedAt.IsZero() {
		shift.CreatedAt = now
	}
	shift.UpdatedAt = now

	var id any = shift.ID
	if shift.ID == 0 {
		id = nil
	}

	_, err := q.ExecContext(ctx, query,
		id,
		shift.UID,
		shift.StartTime,
		shift.EndTime,
		shift.GraceMinutes,
		shift.CreatedAt,
		shift.UpdatedAt,
	)
	if err != nil {
		slog.Error("shift_repository.Upsert.exec", "error", err)
	}

	return err
}

func (r *ShiftRepository) scanShift(row *sql.Row) (*domain.Shift, error) {
	var shift domain.Shift
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&shift.ID,
		&shift.UID,
		&shift.StartTime,
		&shift.EndTime,
		&shift.GraceMinutes,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("shift_repository.scanShift.scan", "error", err)
		return nil, err
	}
	shift.CreatedAt = createdAt.Time
	shift.UpdatedAt = updatedAt.Time
	return &shift, nil
}

func (r *ShiftRepository) scanShiftRow(rows *sql.Rows) (*domain.Shift, error) {
	var shift domain.Shift
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&shift.ID,
		&shift.UID,
		&shift.StartTime,
		&shift.EndTime,
		&shift.GraceMinutes,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	shift.CreatedAt = createdAt.Time
	shift.UpdatedAt = updatedAt.Time
	return &shift, nil
}
