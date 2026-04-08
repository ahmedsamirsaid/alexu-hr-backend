package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type PermissionRepository struct{}

func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{}
}

func (r *PermissionRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Permission, error) {
	query := `
		SELECT id, uid, code, description, created_at
		FROM permissions
		WHERE id = ?`

	return r.scanPermission(q.QueryRowContext(ctx, query, id))
}

func (r *PermissionRepository) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.Permission, error) {
	query := `
		SELECT id, uid, code, description, created_at
		FROM permissions
		WHERE code = ?`

	return r.scanPermission(q.QueryRowContext(ctx, query, code))
}

func (r *PermissionRepository) List(ctx context.Context, q ports.Querier) ([]*domain.Permission, error) {
	query := `
		SELECT id, uid, code, description, created_at
		FROM permissions
		ORDER BY code ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("permission_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var perms []*domain.Permission
	for rows.Next() {
		p, err := r.scanPermissionRow(rows)
		if err != nil {
			slog.Error("permission_repository.List.scan_row", "error", err)
			return nil, err
		}
		perms = append(perms, p)
	}

	if err := rows.Err(); err != nil {
		slog.Error("permission_repository.List.rows_iteration", "error", err)
		return nil, err
	}
	return perms, nil
}

func (r *PermissionRepository) GetPermissionsForRole(ctx context.Context, q ports.Querier, roleID int64) ([]*domain.Permission, error) {
	result, err := r.GetPermissionsForRoles(ctx, q, []int64{roleID})
	if err != nil {
		return nil, err
	}
	return result[roleID], nil
}

func (r *PermissionRepository) GetPermissionsForRoles(ctx context.Context, q ports.Querier, roleIDs []int64) (map[int64][]*domain.Permission, error) {
	result := make(map[int64][]*domain.Permission)
	if len(roleIDs) == 0 {
		return result, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]byte, 0, len(roleIDs)*2-1)
	args := make([]any, len(roleIDs))
	for i, id := range roleIDs {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args[i] = id
		result[id] = []*domain.Permission{} // Initialize empty slice
	}

	query := `
		SELECT rp.role_id, p.id, p.uid, p.code, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id IN (` + string(placeholders) + `)
		ORDER BY p.code ASC`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("permission_repository.GetPermissionsForRoles.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleID int64
		var p domain.Permission
		var createdAt domain.Time
		err := rows.Scan(&roleID, &p.ID, &p.UID, &p.Code, &p.Description, &createdAt)
		if err != nil {
			slog.Error("permission_repository.GetPermissionsForRoles.scan_row", "error", err)
			return nil, err
		}
		p.CreatedAt = createdAt.Time
		result[roleID] = append(result[roleID], &p)
	}

	if err := rows.Err(); err != nil {
		slog.Error("permission_repository.GetPermissionsForRoles.rows_iteration", "error", err)
		return nil, err
	}
	return result, nil
}

func (r *PermissionRepository) scanPermission(row *sql.Row) (*domain.Permission, error) {
	var p domain.Permission
	var createdAt domain.Time
	err := row.Scan(&p.ID, &p.UID, &p.Code, &p.Description, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("permission_repository.scanPermission.scan_row", "error", err)
		return nil, err
	}
	p.CreatedAt = createdAt.Time
	return &p, nil
}

func (r *PermissionRepository) scanPermissionRow(rows *sql.Rows) (*domain.Permission, error) {
	var p domain.Permission
	var createdAt domain.Time
	err := rows.Scan(&p.ID, &p.UID, &p.Code, &p.Description, &createdAt)
	if err != nil {
		return nil, err
	}
	p.CreatedAt = createdAt.Time
	return &p, nil
}
