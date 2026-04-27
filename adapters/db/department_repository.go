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

type DepartmentRepository struct{}

func NewDepartmentRepository() *DepartmentRepository {
	return &DepartmentRepository{}
}

func (r *DepartmentRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Department, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, is_active, default_shift_uid, created_at, updated_at
		FROM departments
		WHERE id = ?`

	return r.scanDepartment(q.QueryRowContext(ctx, query, id))
}

func (r *DepartmentRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Department, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, is_active, default_shift_uid, created_at, updated_at
		FROM departments
		WHERE uid = ?`

	return r.scanDepartment(q.QueryRowContext(ctx, query, uid))
}

func (r *DepartmentRepository) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.Department, error) {
	if len(uids) == 0 {
		return []*domain.Department{}, nil
	}

	// Build IN clause with placeholders
	query := `
		SELECT id, uid, code, name_en, name_ar, is_active, default_shift_uid, created_at, updated_at
		FROM departments
		WHERE uid IN (`

	args := make([]any, len(uids))
	for i, uid := range uids {
		if i > 0 {
			query += ", "
		}
		query += "?"
		args[i] = uid
	}
	query += ")"

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("department_repository.GetByUIDs.query", "error", err, "count", len(uids))
		return nil, err
	}
	defer rows.Close()

	var departments []*domain.Department
	for rows.Next() {
		dept, err := r.scanDepartmentRow(rows)
		if err != nil {
			slog.Error("department_repository.GetByUIDs.scan", "error", err)
			return nil, err
		}
		departments = append(departments, dept)
	}

	if err := rows.Err(); err != nil {
		slog.Error("department_repository.GetByUIDs.rows_err", "error", err)
		return nil, err
	}

	return departments, nil
}

func (r *DepartmentRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.Department, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, is_active, default_shift_uid, created_at, updated_at
		FROM departments
		WHERE code = ?`

	return r.scanDepartment(q.QueryRowContext(ctx, query, code))
}

func (r *DepartmentRepository) Create(ctx context.Context, q ports.Querier, department *domain.Department) error {
	query := `
		INSERT INTO departments (uid, code, name_en, name_ar, is_active, default_shift_uid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	department.CreatedAt = now
	department.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		department.UID, department.Code, department.NameEN, department.NameAR,
		department.IsActive, department.DefaultShiftUID, department.CreatedAt, department.UpdatedAt)
	if err != nil {
		slog.Error("department_repository.Create.exec_query", "error", err, "uid", department.UID, "code", department.Code)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("department_repository.Create.last_insert_id", "error", err, "uid", department.UID)
		return err
	}
	department.ID = id

	return nil
}

func (r *DepartmentRepository) Update(ctx context.Context, q ports.Querier, department *domain.Department) error {
	query := `
		UPDATE departments
		SET code = ?, name_en = ?, name_ar = ?, is_active = ?, default_shift_uid = ?, updated_at = ?
		WHERE id = ?`

	department.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		department.Code, department.NameEN, department.NameAR,
		department.IsActive, department.DefaultShiftUID, department.UpdatedAt, department.ID)
	if err != nil {
		slog.Error("department_repository.Update.exec_query", "error", err, "id", department.ID, "uid", department.UID)
	}
	return err
}

func (r *DepartmentRepository) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.Department, error) {
	query := `
		SELECT id, uid, code, name_en, name_ar, is_active, default_shift_uid, created_at, updated_at
		FROM departments`

	if activeOnly {
		query += ` WHERE is_active = 1`
	}

	query += ` ORDER BY name_en ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("department_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var departments []*domain.Department
	for rows.Next() {
		d, err := r.scanDepartmentRow(rows)
		if err != nil {
			slog.Error("department_repository.List.scan_row", "error", err)
			return nil, err
		}
		departments = append(departments, d)
	}

	if err := rows.Err(); err != nil {
		slog.Error("department_repository.List.rows_iteration", "error", err)
		return nil, err
	}
	return departments, nil
}

func (r *DepartmentRepository) scanDepartment(row *sql.Row) (*domain.Department, error) {
	var d domain.Department
	var createdAt, updatedAt domain.Time
	var isActive int
	err := row.Scan(
		&d.ID, &d.UID, &d.Code, &d.NameEN, &d.NameAR,
		&isActive, &d.DefaultShiftUID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("department_repository.scanDepartment.scan_row", "error", err)
		return nil, err
	}
	d.IsActive = isActive == 1
	d.CreatedAt = createdAt.Time
	d.UpdatedAt = updatedAt.Time
	return &d, nil
}

func (r *DepartmentRepository) scanDepartmentRow(rows *sql.Rows) (*domain.Department, error) {
	var d domain.Department
	var createdAt, updatedAt domain.Time
	var isActive int
	err := rows.Scan(
		&d.ID, &d.UID, &d.Code, &d.NameEN, &d.NameAR,
		&isActive, &d.DefaultShiftUID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	d.IsActive = isActive == 1
	d.CreatedAt = createdAt.Time
	d.UpdatedAt = updatedAt.Time
	return &d, nil
}
