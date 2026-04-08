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
		SELECT id, uid, code, name_en, name_ar, default_month, default_day, created_at, updated_at
		FROM holiday_definitions
		WHERE id = ?`

	return r.scanHolidayDefinition(q.QueryRowContext(ctx, query, id))
}

func (r *HolidayDefinitionRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_month, default_day, created_at, updated_at
		FROM holiday_definitions
		WHERE code = ?`

	return r.scanHolidayDefinition(q.QueryRowContext(ctx, query, code))
}

func (r *HolidayDefinitionRepository) List(ctx context.Context, q ports.Querier) ([]*domain.HolidayDefinition, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, default_month, default_day, created_at, updated_at
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

func (r *HolidayDefinitionRepository) Create(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	query := `
		INSERT INTO holiday_definitions (uid, code, name_en, name_ar, default_month, default_day, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	def.CreatedAt = now
	def.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		def.UID, def.Code, def.NameEN, def.NameAR,
		def.DefaultMonth, def.DefaultDay, def.CreatedAt, def.UpdatedAt)
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

func (r *HolidayDefinitionRepository) scanHolidayDefinition(row *sql.Row) (*domain.HolidayDefinition, error) {
	var def domain.HolidayDefinition
	var createdAt, updatedAt domain.Time
	err := row.Scan(
		&def.ID, &def.UID, &def.Code, &def.NameEN, &def.NameAR,
		&def.DefaultMonth, &def.DefaultDay, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("holiday_repository.scanHolidayDefinition.scan_row", "error", err)
		return nil, err
	}
	def.CreatedAt = createdAt.Time
	def.UpdatedAt = updatedAt.Time
	return &def, nil
}

func (r *HolidayDefinitionRepository) scanHolidayDefinitionRow(rows *sql.Rows) (*domain.HolidayDefinition, error) {
	var def domain.HolidayDefinition
	var createdAt, updatedAt domain.Time
	err := rows.Scan(
		&def.ID, &def.UID, &def.Code, &def.NameEN, &def.NameAR,
		&def.DefaultMonth, &def.DefaultDay, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	def.CreatedAt = createdAt.Time
	def.UpdatedAt = updatedAt.Time
	return &def, nil
}

type HolidayInstanceRepository struct{}

func NewHolidayInstanceRepository() *HolidayInstanceRepository {
	return &HolidayInstanceRepository{}
}

func (r *HolidayInstanceRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayInstance, error) {
	query := `
		SELECT id, uid, definition_id, year, actual_date, observed_date, is_confirmed, notes, created_at, updated_at
		FROM holiday_instances
		WHERE id = ?`

	return r.scanHolidayInstance(q.QueryRowContext(ctx, query, id))
}

func (r *HolidayInstanceRepository) GetByDefinitionAndYear(ctx context.Context, q ports.Querier, definitionID int64, year int) (*domain.HolidayInstance, error) {
	query := `
		SELECT id, uid, definition_id, year, actual_date, observed_date, is_confirmed, notes, created_at, updated_at
		FROM holiday_instances
		WHERE definition_id = ? AND year = ?`

	return r.scanHolidayInstance(q.QueryRowContext(ctx, query, definitionID, year))
}

func (r *HolidayInstanceRepository) ListByYear(ctx context.Context, q ports.Querier, year int) ([]*domain.HolidayInstance, error) {
	query := `
		SELECT id, uid, definition_id, year, actual_date, observed_date, is_confirmed, notes, created_at, updated_at
		FROM holiday_instances
		WHERE year = ?
		ORDER BY observed_date`

	return r.queryHolidayInstances(ctx, q, query, year)
}

func (r *HolidayInstanceRepository) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayInstance, error) {
	query := `
		SELECT id, uid, definition_id, year, actual_date, observed_date, is_confirmed, notes, created_at, updated_at
		FROM holiday_instances
		WHERE observed_date >= ? AND observed_date <= ?
		ORDER BY observed_date`

	return r.queryHolidayInstances(ctx, q, query, start, end)
}

func (r *HolidayInstanceRepository) Create(ctx context.Context, q ports.Querier, instance *domain.HolidayInstance) error {
	query := `
		INSERT INTO holiday_instances (uid, definition_id, year, actual_date, observed_date, is_confirmed, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	instance.CreatedAt = now
	instance.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		instance.UID, instance.DefinitionID, instance.Year,
		instance.ActualDate, instance.ObservedDate, instance.IsConfirmed,
		instance.Notes, instance.CreatedAt, instance.UpdatedAt)
	if err != nil {
		slog.Error("holiday_repository.Create.exec_query", "error", err, "uid", instance.UID, "year", instance.Year)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("holiday_repository.Create.last_insert_id", "error", err, "uid", instance.UID)
		return err
	}
	instance.ID = id

	return nil
}

func (r *HolidayInstanceRepository) Update(ctx context.Context, q ports.Querier, instance *domain.HolidayInstance) error {
	query := `
		UPDATE holiday_instances
		SET actual_date = ?, observed_date = ?, is_confirmed = ?, notes = ?, updated_at = ?
		WHERE id = ?`

	instance.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		instance.ActualDate, instance.ObservedDate, instance.IsConfirmed,
		instance.Notes, instance.UpdatedAt, instance.ID)
	if err != nil {
		slog.Error("holiday_repository.Update.exec_query", "error", err, "id", instance.ID, "uid", instance.UID)
	}
	return err
}

func (r *HolidayInstanceRepository) queryHolidayInstances(ctx context.Context, q ports.Querier, query string, args ...any) ([]*domain.HolidayInstance, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("holiday_repository.queryHolidayInstances.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.HolidayInstance
	for rows.Next() {
		inst, err := r.scanHolidayInstanceRow(rows)
		if err != nil {
			slog.Error("holiday_repository.queryHolidayInstances.scan_row", "error", err)
			return nil, err
		}
		instances = append(instances, inst)
	}

	if err := rows.Err(); err != nil {
		slog.Error("holiday_repository.queryHolidayInstances.rows_iteration", "error", err)
		return nil, err
	}
	return instances, nil
}

func (r *HolidayInstanceRepository) scanHolidayInstance(row *sql.Row) (*domain.HolidayInstance, error) {
	var inst domain.HolidayInstance
	var actualDate, observedDate, createdAt, updatedAt domain.Time
	err := row.Scan(
		&inst.ID, &inst.UID, &inst.DefinitionID, &inst.Year,
		&actualDate, &observedDate, &inst.IsConfirmed,
		&inst.Notes, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("holiday_repository.scanHolidayInstance.scan_row", "error", err)
		return nil, err
	}
	inst.ActualDate = actualDate.Time
	inst.ObservedDate = observedDate.Time
	inst.CreatedAt = createdAt.Time
	inst.UpdatedAt = updatedAt.Time
	return &inst, nil
}

func (r *HolidayInstanceRepository) scanHolidayInstanceRow(rows *sql.Rows) (*domain.HolidayInstance, error) {
	var inst domain.HolidayInstance
	var actualDate, observedDate, createdAt, updatedAt domain.Time
	err := rows.Scan(
		&inst.ID, &inst.UID, &inst.DefinitionID, &inst.Year,
		&actualDate, &observedDate, &inst.IsConfirmed,
		&inst.Notes, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	inst.ActualDate = actualDate.Time
	inst.ObservedDate = observedDate.Time
	inst.CreatedAt = createdAt.Time
	inst.UpdatedAt = updatedAt.Time
	return &inst, nil
}
