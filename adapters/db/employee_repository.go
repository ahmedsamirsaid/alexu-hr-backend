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

type EmployeeRepository struct{}

func NewEmployeeRepository() *EmployeeRepository {
	return &EmployeeRepository{}
}

func (r *EmployeeRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	query := `
		SELECT id, uid, name, mobile, government_id, university_id, email,
		       hire_date, status, department_uid, shift_uid, created_at, updated_at
		FROM employees
		WHERE id = ?`

	return r.scanEmployee(q.QueryRowContext(ctx, query, id))
}

func (r *EmployeeRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	query := `
		SELECT id, uid, name, mobile, government_id, university_id, email,
		       hire_date, status, department_uid, shift_uid, created_at, updated_at
		FROM employees
		WHERE uid = ?`

	return r.scanEmployee(q.QueryRowContext(ctx, query, uid))
}

func (r *EmployeeRepository) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	query := `
		INSERT INTO employees (uid, name, mobile, government_id, university_id, email,
		                       hire_date, status, department_uid, shift_uid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	employee.CreatedAt = now
	employee.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		employee.UID, employee.Name, employee.Mobile,
		employee.GovernmentID, employee.UniversityID, employee.Email,
		employee.HireDate, employee.Status, employee.DepartmentUID, employee.ShiftUID, employee.CreatedAt, employee.UpdatedAt)
	if err != nil {
		slog.Error("employee_repository.Create.exec_query", "error", err, "uid", employee.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("employee_repository.Create.last_insert_id", "error", err, "uid", employee.UID)
		return err
	}
	employee.ID = id

	return nil
}

func (r *EmployeeRepository) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	query := `
		UPDATE employees
		SET name = ?, mobile = ?, government_id = ?, university_id = ?, email = ?,
		    hire_date = ?, status = ?, department_uid = ?, shift_uid = ?, updated_at = ?
		WHERE id = ?`

	employee.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		employee.Name, employee.Mobile, employee.GovernmentID, employee.UniversityID,
		employee.Email, employee.HireDate, employee.Status, employee.DepartmentUID, employee.ShiftUID, employee.UpdatedAt, employee.ID)
	if err != nil {
		slog.Error("employee_repository.Update.exec_query", "error", err, "uid", employee.UID)
	}

	return err
}

func (r *EmployeeRepository) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	query := `
		SELECT id, uid, name, mobile, government_id, university_id, email,
		       hire_date, status, department_uid, shift_uid, created_at, updated_at
		FROM employees
		WHERE 1=1`

	var args []any

	if filter != nil {
		if filter.Status != nil {
			query += ` AND status = ?`
			args = append(args, *filter.Status)
		}
		if filter.HireDateFrom != nil {
			query += ` AND hire_date >= ?`
			args = append(args, *filter.HireDateFrom)
		}
		if filter.HireDateTo != nil {
			query += ` AND hire_date <= ?`
			args = append(args, *filter.HireDateTo)
		}
	}

	query += ` ORDER BY name ASC`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("employee_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		e, err := r.scanEmployeeRow(rows)
		if err != nil {
			slog.Error("employee_repository.List.scan_row", "error", err)
			return nil, err
		}
		employees = append(employees, e)
	}

	if err := rows.Err(); err != nil {
		slog.Error("employee_repository.List.rows_iteration", "error", err)
		return nil, err
	}

	return employees, nil
}

func (r *EmployeeRepository) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	if len(governmentIDs) == 0 {
		return nil, nil
	}
	return r.existingValues(ctx, q, "government_id", governmentIDs)
}

func (r *EmployeeRepository) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	if len(mobiles) == 0 {
		return nil, nil
	}
	return r.existingValues(ctx, q, "mobile", mobiles)
}

func (r *EmployeeRepository) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	if len(universityIDs) == 0 {
		return nil, nil
	}
	return r.existingValues(ctx, q, "university_id", universityIDs)
}

func (r *EmployeeRepository) existingValues(ctx context.Context, q ports.Querier, column string, values []string) ([]string, error) {
	placeholders := make([]byte, 0, len(values)*2)
	args := make([]any, len(values))
	for i, v := range values {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args[i] = v
	}

	query := `SELECT ` + column + ` FROM employees WHERE ` + column + ` IN (` + string(placeholders) + `)`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("employee_repository.existingValues.query", "error", err, "column", column)
		return nil, err
	}
	defer rows.Close()

	var existing []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			slog.Error("employee_repository.existingValues.scan_row", "error", err, "column", column)
			return nil, err
		}
		existing = append(existing, val)
	}

	if err := rows.Err(); err != nil {
		slog.Error("employee_repository.existingValues.rows_iteration", "error", err, "column", column)
		return nil, err
	}

	return existing, nil
}

func (r *EmployeeRepository) scanEmployee(row *sql.Row) (*domain.Employee, error) {
	var e domain.Employee
	var hireDate, createdAt, updatedAt domain.Time
	err := row.Scan(
		&e.ID, &e.UID, &e.Name, &e.Mobile, &e.GovernmentID, &e.UniversityID,
		&e.Email, &hireDate, &e.Status, &e.DepartmentUID, &e.ShiftUID, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("employee_repository.scanEmployee.scan_row", "error", err)
		return nil, err
	}
	e.HireDate = hireDate.Time
	e.CreatedAt = createdAt.Time
	e.UpdatedAt = updatedAt.Time
	return &e, nil
}

func (r *EmployeeRepository) scanEmployeeRow(rows *sql.Rows) (*domain.Employee, error) {
	var e domain.Employee
	var hireDate, createdAt, updatedAt domain.Time
	err := rows.Scan(
		&e.ID, &e.UID, &e.Name, &e.Mobile, &e.GovernmentID, &e.UniversityID,
		&e.Email, &hireDate, &e.Status, &e.DepartmentUID, &e.ShiftUID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	e.HireDate = hireDate.Time
	e.CreatedAt = createdAt.Time
	e.UpdatedAt = updatedAt.Time
	return &e, nil
}

func (r *EmployeeRepository) Count(ctx context.Context, q ports.Querier) (int, error) {
	query := `SELECT COUNT(*) FROM employees`

	var count int
	err := q.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		slog.Error("employee_repository.Count.scan", "error", err)
		return 0, err
	}
	return count, nil
}
