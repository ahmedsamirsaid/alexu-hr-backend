package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	query := `
		SELECT id, uid, phone, password_hash, employee_uid, is_active, created_at, updated_at
		FROM users
		WHERE id = ?`

	return r.scanUser(q.QueryRowContext(ctx, query, id))
}

func (r *UserRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	query := `
		SELECT id, uid, phone, password_hash, employee_uid, is_active, created_at, updated_at
		FROM users
		WHERE uid = ?`

	return r.scanUser(q.QueryRowContext(ctx, query, uid))
}

func (r *UserRepository) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	query := `
		SELECT id, uid, phone, password_hash, employee_uid, is_active, created_at, updated_at
		FROM users
		WHERE phone = ?`

	return r.scanUser(q.QueryRowContext(ctx, query, phone))
}

func (r *UserRepository) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	query := `
		INSERT INTO users (uid, phone, password_hash, employee_uid, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		user.UID, user.Phone, user.PasswordHash, user.EmployeeUID,
		user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		slog.Error("user_repository.Create.exec_query", "error", err, "uid", user.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("user_repository.Create.last_insert_id", "error", err, "uid", user.UID)
		return err
	}
	user.ID = id

	return nil
}

func (r *UserRepository) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
	query := `
		UPDATE users
		SET phone = ?, password_hash = ?, employee_uid = ?, is_active = ?, updated_at = ?
		WHERE id = ?`

	user.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		user.Phone, user.PasswordHash, user.EmployeeUID,
		user.IsActive, user.UpdatedAt, user.ID)
	if err != nil {
		slog.Error("user_repository.Update.exec_query", "error", err, "uid", user.UID)
	}

	return err
}

func (r *UserRepository) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, uid, phone, password_hash, employee_uid, is_active, created_at, updated_at
		FROM users
		ORDER BY id DESC
		LIMIT ? OFFSET ?`

	rows, err := q.QueryContext(ctx, query, limit, offset)
	if err != nil {
		slog.Error("user_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := r.scanUserRow(rows)
		if err != nil {
			slog.Error("user_repository.List.scan_row", "error", err)
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		slog.Error("user_repository.List.rows_iteration", "error", err)
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Count(ctx context.Context, q ports.Querier) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := q.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		slog.Error("user_repository.Count.scan", "error", err)
	}
	return count, err
}

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var createdAt, updatedAt domain.Time
	var isActive int
	err := row.Scan(
		&u.ID, &u.UID, &u.Phone, &u.PasswordHash, &u.EmployeeUID,
		&isActive, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user_repository.scanUser.scan_row", "error", err)
		return nil, err
	}
	u.IsActive = isActive == 1
	u.CreatedAt = createdAt.Time
	u.UpdatedAt = updatedAt.Time
	return &u, nil
}

func (r *UserRepository) scanUserRow(rows *sql.Rows) (*domain.User, error) {
	var u domain.User
	var createdAt, updatedAt domain.Time
	var isActive int
	err := rows.Scan(
		&u.ID, &u.UID, &u.Phone, &u.PasswordHash, &u.EmployeeUID,
		&isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	u.IsActive = isActive == 1
	u.CreatedAt = createdAt.Time
	u.UpdatedAt = updatedAt.Time
	return &u, nil
}

func (r *UserRepository) ExistingPhones(ctx context.Context, q ports.Querier, phones []string) ([]string, error) {
	if len(phones) == 0 {
		return nil, nil
	}

	// Build query with placeholders
	placeholders := make([]string, len(phones))
	args := make([]any, len(phones))
	for i, phone := range phones {
		placeholders[i] = "?"
		args[i] = phone
	}

	query := `SELECT phone FROM users WHERE phone IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("user_repository.ExistingPhones.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var existing []string
	for rows.Next() {
		var phone string
		if err := rows.Scan(&phone); err != nil {
			slog.Error("user_repository.ExistingPhones.scan_row", "error", err)
			return nil, err
		}
		existing = append(existing, phone)
	}

	if err := rows.Err(); err != nil {
		slog.Error("user_repository.ExistingPhones.rows_iteration", "error", err)
		return nil, err
	}

	return existing, nil
}

func (r *UserRepository) GetByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) (*domain.User, error) {
	query := `
		SELECT id, uid, phone, password_hash, employee_uid, is_active, created_at, updated_at
		FROM users
		WHERE employee_uid = ?`

	return r.scanUser(q.QueryRowContext(ctx, query, employeeUID))
}
