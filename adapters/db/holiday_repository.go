package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type HolidayDefinitionRepository struct{}

func NewHolidayDefinitionRepository() *HolidayDefinitionRepository {
	return &HolidayDefinitionRepository{}
}

func (r *HolidayDefinitionRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayDefinition, error) {
	query := r.selectHolidayDefinitionsQuery(`WHERE h.id = $1`, `ORDER BY h.id`)

	return r.scanHolidayDefinition(q.QueryRowContext(ctx, query, id))
}

func (r *HolidayDefinitionRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.HolidayDefinition, error) {
	query := r.selectHolidayDefinitionsQuery(`WHERE h.code = $1`, `ORDER BY h.id`)

	return r.scanHolidayDefinition(q.QueryRowContext(ctx, query, code))
}

func (r *HolidayDefinitionRepository) List(ctx context.Context, q ports.Querier) ([]*domain.HolidayDefinition, error) {
	query := r.selectHolidayDefinitionsQuery("", `ORDER BY h.id`)

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
	query := r.selectHolidayDefinitionsQuery(`WHERE h.date = $1`, `ORDER BY h.id`)

	return r.queryHolidayDefinitions(ctx, q, query, date)
}

func (r *HolidayDefinitionRepository) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	query := r.selectHolidayDefinitionsQuery(`WHERE h.date >= $1 AND h.date <= $2 AND NOT EXISTS (
		SELECT 1 FROM holiday_definition_departments hd
		WHERE hd.holiday_definition_id = h.id
	)`, `ORDER BY h.date ASC, h.id ASC`)

	return r.queryHolidayDefinitions(ctx, q, query, start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func (r *HolidayDefinitionRepository) ListAllByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	query := r.selectHolidayDefinitionsQuery(`WHERE h.date >= $1 AND h.date <= $2`, `ORDER BY h.date ASC, h.id ASC`)

	return r.queryHolidayDefinitions(ctx, q, query, start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func (r *HolidayDefinitionRepository) ListByDateRangeForDepartment(ctx context.Context, q ports.Querier, start, end time.Time, departmentUID string) ([]*domain.HolidayDefinition, error) {
	query := r.selectHolidayDefinitionsQuery(`WHERE h.date >= $1 AND h.date <= $2 AND (
		NOT EXISTS (
			SELECT 1 FROM holiday_definition_departments hd
			WHERE hd.holiday_definition_id = h.id
		) OR EXISTS (
			SELECT 1 FROM holiday_definition_departments hd
			WHERE hd.holiday_definition_id = h.id AND hd.department_uid = $3
		)
	)`, `ORDER BY h.date ASC, h.id ASC`)

	return r.queryHolidayDefinitions(ctx, q, query, start.Format("2006-01-02"), end.Format("2006-01-02"), departmentUID)
}

func (r *HolidayDefinitionRepository) Create(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	query := `
		INSERT INTO holiday_definitions (uid, code, name_en, name_ar, date, is_manual, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	def.CreatedAt = now
	def.UpdatedAt = now

	err := q.QueryRowContext(ctx, query,
		def.UID, def.Code, def.NameEN, def.NameAR,
		def.Date, def.IsManual, def.CreatedAt, def.UpdatedAt).Scan(&def.ID)
	if err != nil {
		slog.Error("holiday_repository.Create.exec_query", "error", err, "uid", def.UID, "code", def.Code)
		return err
	}

	if err := r.replaceHolidayDepartments(ctx, q, def.ID, def.DepartmentUIDs); err != nil {
		return err
	}

	return nil
}

func (r *HolidayDefinitionRepository) Update(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	query := `
		UPDATE holiday_definitions
		SET code = $1, name_en = $2, name_ar = $3, date = $4, is_manual = $5, updated_at = $6
		WHERE id = $7`

	def.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		def.Code, def.NameEN, def.NameAR, def.Date, def.IsManual, def.UpdatedAt, def.ID)
	if err != nil {
		slog.Error("holiday_repository.Update.exec_query", "error", err, "uid", def.UID)
		return err
	}

	return r.replaceHolidayDepartments(ctx, q, def.ID, def.DepartmentUIDs)
}

func (r *HolidayDefinitionRepository) Delete(ctx context.Context, q ports.Querier, id int64) error {
	_, err := q.ExecContext(ctx, `DELETE FROM holiday_definitions WHERE id = $1`, id)
	if err != nil {
		slog.Error("holiday_repository.Delete.exec_query", "error", err, "id", id)
	}
	return err
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
	var departmentUIDs sql.NullString
	err := row.Scan(
		&def.ID, &def.UID, &def.Code, &def.NameEN, &def.NameAR,
		&dateValue, &def.IsManual, &createdAt, &updatedAt, &departmentUIDs)
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
	def.DepartmentUIDs = parseDepartmentUIDs(departmentUIDs)
	return &def, nil
}

func (r *HolidayDefinitionRepository) scanHolidayDefinitionRow(rows *sql.Rows) (*domain.HolidayDefinition, error) {
	var def domain.HolidayDefinition
	var dateValue, createdAt, updatedAt domain.Time
	var departmentUIDs sql.NullString
	err := rows.Scan(
		&def.ID, &def.UID, &def.Code, &def.NameEN, &def.NameAR,
		&dateValue, &def.IsManual, &createdAt, &updatedAt, &departmentUIDs)
	if err != nil {
		return nil, err
	}
	def.Date = dateValue.Time
	def.CreatedAt = createdAt.Time
	def.UpdatedAt = updatedAt.Time
	def.DepartmentUIDs = parseDepartmentUIDs(departmentUIDs)
	return &def, nil
}

func (r *HolidayDefinitionRepository) selectHolidayDefinitionsQuery(whereClause, orderClause string) string {
	query := `
		SELECT h.id, h.uid, h.code, h.name_en, h.name_ar, h.date, h.is_manual, h.created_at, h.updated_at,
			STRING_AGG(hd.department_uid, ',') AS department_uids
		FROM holiday_definitions h
		LEFT JOIN holiday_definition_departments hd ON hd.holiday_definition_id = h.id`
	if whereClause != "" {
		query += "\n\t" + whereClause
	}
	query += "\n\tGROUP BY h.id"
	if orderClause != "" {
		query += "\n\t" + orderClause
	}
	return query
}

func (r *HolidayDefinitionRepository) replaceHolidayDepartments(ctx context.Context, q ports.Querier, definitionID int64, departmentUIDs []string) error {
	if _, err := q.ExecContext(ctx, `DELETE FROM holiday_definition_departments WHERE holiday_definition_id = $1`, definitionID); err != nil {
		slog.Error("holiday_repository.replaceHolidayDepartments.delete", "error", err, "definition_id", definitionID)
		return err
	}

	for _, departmentUID := range normalizeDepartmentUIDs(departmentUIDs) {
		if _, err := q.ExecContext(ctx,
			`INSERT INTO holiday_definition_departments (holiday_definition_id, department_uid, created_at) VALUES ($1, $2, $3)`,
			definitionID, departmentUID, time.Now()); err != nil {
			slog.Error("holiday_repository.replaceHolidayDepartments.insert", "error", err, "definition_id", definitionID, "department_uid", departmentUID)
			return err
		}
	}

	return nil
}

func parseDepartmentUIDs(value sql.NullString) []string {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}

	parts := strings.Split(value.String, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}

	sort.Strings(result)
	return result
}

func normalizeDepartmentUIDs(departmentUIDs []string) []string {
	if len(departmentUIDs) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(departmentUIDs))
	result := make([]string, 0, len(departmentUIDs))
	for _, departmentUID := range departmentUIDs {
		departmentUID = strings.TrimSpace(departmentUID)
		if departmentUID == "" {
			continue
		}
		if _, ok := seen[departmentUID]; ok {
			continue
		}
		seen[departmentUID] = struct{}{}
		result = append(result, departmentUID)
	}

	sort.Strings(result)
	if len(result) == 0 {
		return nil
	}
	return result
}
