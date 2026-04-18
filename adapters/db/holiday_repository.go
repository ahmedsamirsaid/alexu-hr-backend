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

type HolidayDefinitionRepository struct{}

func NewHolidayDefinitionRepository() *HolidayDefinitionRepository {
	return &HolidayDefinitionRepository{}
}

func (r *HolidayDefinitionRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, date, is_manual, created_at, updated_at
		FROM holiday_definitions
		WHERE id = ?`

	return r.scanHolidayDefinition(q.QueryRowContext(ctx, query, id))
}

func (r *HolidayDefinitionRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, date, is_manual, created_at, updated_at
		FROM holiday_definitions
		WHERE code = ?`

	return r.scanHolidayDefinition(q.QueryRowContext(ctx, query, code))
}

func (r *HolidayDefinitionRepository) List(ctx context.Context, q ports.Querier) ([]*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, date, is_manual, created_at, updated_at
		FROM holiday_definitions
		ORDER BY id`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("holiday_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var defs []*domain.HolidayDefinition
	for rows.Next() {
		def, err := r.scanHolidayDefinitionRow(rows)
		if err != nil {
			slog.Error("holiday_repository.List.scan_row", "error", err)
			return nil, err
		}
		defs = append(defs, def)
	}

	if err := rows.Err(); err != nil {
		slog.Error("holiday_repository.List.rows_iteration", "error", err)
		return nil, err
	}
	return defs, nil
}

func (r *HolidayDefinitionRepository) GetByDate(ctx context.Context, q ports.Querier, date time.Time) ([]*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, date, is_manual, created_at, updated_at
		FROM holiday_definitions
		WHERE date = ?
		ORDER BY id`

	return r.queryHolidayDefinitions(ctx, q, query, date)
}

func (r *HolidayDefinitionRepository) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, date, is_manual, created_at, updated_at
		FROM holiday_definitions
		WHERE date >= ? AND date <= ?
		ORDER BY date ASC, id ASC`

	return r.queryHolidayDefinitions(ctx, q, query, start, end)
}

func (r *HolidayDefinitionRepository) Create(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	query := `
		INSERT INTO holiday_definitions (uid, code, name_en, name_ar, date, is_manual, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	def.CreatedAt = now
	def.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		def.UID, def.Code, def.NameEN, def.NameAR,
		def.Date, def.IsManual, def.CreatedAt, def.UpdatedAt)
	if err != nil {
		slog.Error("holiday_repository.Create.exec_query", "error", err, "uid", def.UID, "code", def.Code)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("holiday_repository.Create.last_insert_id", "error", err, "uid", def.UID)
		return err
	}
	def.ID = id

	return nil
}

func (r *HolidayDefinitionRepository) queryHolidayDefinitions(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.HolidayDefinition, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("holiday_repository.queryHolidayDefinitions.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var defs []*domain.HolidayDefinition
	for rows.Next() {
		def, err := r.scanHolidayDefinitionRow(rows)
		if err != nil {
			slog.Error("holiday_repository.queryHolidayDefinitions.scan_row", "error", err)
			return nil, err
		}
		defs = append(defs, def)
	}

	if err := rows.Err(); err != nil {
		slog.Error("holiday_repository.queryHolidayDefinitions.rows_iteration", "error", err)
		return nil, err
	}

	return defs, nil
}

func (r *HolidayDefinitionRepository) scanHolidayDefinition(row *sql.Row) (*domain.HolidayDefinition, error) {
	var def domain.HolidayDefinition
	var dateValue, createdAt, updatedAt domain.Time
	err := row.Scan(
		&def.ID, &def.UID, &def.Code, &def.NameEN, &def.NameAR,
		&dateValue, &def.IsManual, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("holiday_repository.scanHolidayDefinition.scan_row", "error", err)
		return nil, err
	}
	def.Date = dateValue.Time
	def.CreatedAt = createdAt.Time
	def.UpdatedAt = updatedAt.Time
	return &def, nil
}

func (r *HolidayDefinitionRepository) scanHolidayDefinitionRow(rows *sql.Rows) (*domain.HolidayDefinition, error) {
	var def domain.HolidayDefinition
	var dateValue, createdAt, updatedAt domain.Time
	err := rows.Scan(
		&def.ID, &def.UID, &def.Code, &def.NameEN, &def.NameAR,
		&dateValue, &def.IsManual, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	def.Date = dateValue.Time
	def.CreatedAt = createdAt.Time
	def.UpdatedAt = updatedAt.Time
	return &def, nil
}
