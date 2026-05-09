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

type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

func (r *RoleRepository) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Role, error) {
	query := `
		SELECT id, uid, name, description, scope_type, is_system, created_at, updated_at
		FROM roles
		WHERE id = ?`

	return r.scanRole(q.QueryRowContext(ctx, query, id))
}

func (r *RoleRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Role, error) {
	query := `
		SELECT id, uid, name, description, scope_type, is_system, created_at, updated_at
		FROM roles
		WHERE uid = ?`

	return r.scanRole(q.QueryRowContext(ctx, query, uid))
}

func (r *RoleRepository) GetByName(ctx context.Context, q ports.Querier, name string) (*domain.Role, error) {
	query := `
		SELECT id, uid, name, description, scope_type, is_system, created_at, updated_at
		FROM roles
		WHERE name = ?`

	return r.scanRole(q.QueryRowContext(ctx, query, name))
}

func (r *RoleRepository) Create(ctx context.Context, q ports.Querier, role *domain.Role) error {
	query := `
		INSERT INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		role.UID, role.Name, role.Description, role.ScopeType, role.IsSystem,
		role.CreatedAt, role.UpdatedAt)
	if err != nil {
		slog.Error("role_repository.Create.exec_query", "error", err, "uid", role.UID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("role_repository.Create.last_insert_id", "error", err, "uid", role.UID)
		return err
	}
	role.ID = id

	return nil
}

func (r *RoleRepository) Update(ctx context.Context, q ports.Querier, role *domain.Role) error {
	query := `
		UPDATE roles
		SET name = ?, description = ?, scope_type = ?, updated_at = ?
		WHERE id = ?`

	role.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		role.Name, role.Description, role.ScopeType, role.UpdatedAt, role.ID)
	if err != nil {
		slog.Error("role_repository.Update.exec_query", "error", err, "uid", role.UID)
	}

	return err
}

func (r *RoleRepository) List(ctx context.Context, q ports.Querier) ([]*domain.Role, error) {
	query := `
		SELECT id, uid, name, description, scope_type, is_system, created_at, updated_at
		FROM roles
		ORDER BY name ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("role_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role, err := r.scanRoleRow(rows)
		if err != nil {
			slog.Error("role_repository.List.scan_row", "error", err)
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		slog.Error("role_repository.List.rows_iteration", "error", err)
		return nil, err
	}

	return roles, nil
}

func (r *RoleRepository) Delete(ctx context.Context, q ports.Querier, id int64) error {
	query := `DELETE FROM roles WHERE id = ? AND is_system = 0`
	_, err := q.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error("role_repository.Delete.exec_query", "error", err, "id", id)
	}
	return err
}

func (r *RoleRepository) GetRolesForUser(ctx context.Context, q ports.Querier, userID int64) ([]*domain.Role, error) {
	query := `
		SELECT r.id, r.uid, r.name, r.description, r.scope_type, r.is_system, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY r.name ASC`

	rows, err := q.QueryContext(ctx, query, userID)
	if err != nil {
		slog.Error("role_repository.GetRolesForUser.query", "error", err, "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role, err := r.scanRoleRow(rows)
		if err != nil {
			slog.Error("role_repository.GetRolesForUser.scan_row", "error", err, "user_id", userID)
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		slog.Error("role_repository.GetRolesForUser.rows_iteration", "error", err, "user_id", userID)
		return nil, err
	}

	return roles, nil
}

func (r *RoleRepository) AssignRoleToUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	query := `INSERT OR IGNORE INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, ?)`
	_, err := q.ExecContext(ctx, query, userID, roleID, time.Now())
	if err != nil {
		slog.Error("role_repository.AssignRoleToUser.exec_query", "error", err, "user_id", userID, "role_id", roleID)
	}
	return err
}

func (r *RoleRepository) RemoveRoleFromUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	query := `DELETE FROM user_roles WHERE user_id = ? AND role_id = ?`
	_, err := q.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		slog.Error("role_repository.RemoveRoleFromUser.exec_query", "error", err, "user_id", userID, "role_id", roleID)
	}
	return err
}

func (r *RoleRepository) AssignPermissionToRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	query := `INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES (?, ?, ?)`
	_, err := q.ExecContext(ctx, query, roleID, permissionID, time.Now())
	if err != nil {
		slog.Error("role_repository.AssignPermissionToRole.exec_query", "error", err, "role_id", roleID, "permission_id", permissionID)
	}
	return err
}

func (r *RoleRepository) RemovePermissionFromRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?`
	_, err := q.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		slog.Error("role_repository.RemovePermissionFromRole.exec_query", "error", err, "role_id", roleID, "permission_id", permissionID)
	}
	return err
}

func (r *RoleRepository) SetRolePermissions(ctx context.Context, q ports.Querier, roleID int64, permissionIDs []int64) error {
	// Delete existing permissions
	_, err := q.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID)
	if err != nil {
		slog.Error("role_repository.SetRolePermissions.delete", "error", err, "role_id", roleID)
		return err
	}

	// Insert new permissions
	for _, permID := range permissionIDs {
		_, err := q.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES (?, ?, ?)`,
			roleID, permID, time.Now())
		if err != nil {
			slog.Error("role_repository.SetRolePermissions.insert", "error", err, "role_id", roleID, "permission_id", permID)
			return err
		}
	}

	return nil
}

func (r *RoleRepository) AssignRoleToUserWithDepartment(ctx context.Context, q ports.Querier, userID, roleID int64, departmentUID *string) error {
	// First remove any existing assignment of this role to this user (to avoid duplicates)
	_, err := q.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ? AND role_id = ?`, userID, roleID)
	if err != nil {
		slog.Error("role_repository.AssignRoleToUserWithDepartment.delete", "error", err, "user_id", userID, "role_id", roleID)
		return err
	}

	query := `INSERT INTO user_roles (user_id, role_id, department_uid, created_at) VALUES (?, ?, ?, ?)`
	_, err = q.ExecContext(ctx, query, userID, roleID, departmentUID, time.Now())
	if err != nil {
		slog.Error("role_repository.AssignRoleToUserWithDepartment.insert", "error", err, "user_id", userID, "role_id", roleID)
	}
	return err
}

func (r *RoleRepository) GetUsersByRoleAndDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID *string) ([]*domain.User, error) {
	// Find users with the given role who either:
	// 1. Have a global role assignment (department_uid IS NULL), OR
	// 2. Have a role scoped to the specific department (if departmentUID provided)
	query := `
		SELECT DISTINCT u.id, u.uid, u.phone, u.password_hash, u.employee_uid, u.is_active,
		       COALESCE(u.preferred_language,''), u.created_at, u.updated_at
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.uid = ? AND (ur.department_uid IS NULL`

	args := []any{roleUID}

	if departmentUID != nil {
		query += ` OR ur.department_uid = ?`
		args = append(args, *departmentUID)
	}
	query += `)`

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("role_repository.GetUsersByRoleAndDepartment.query", "error", err, "role_uid", roleUID)
		return nil, err
	}
	defer rows.Close()

	userRepo := NewUserRepository()
	var users []*domain.User
	for rows.Next() {
		user, err := userRepo.scanUserRow(rows)
		if err != nil {
			slog.Error("role_repository.GetUsersByRoleAndDepartment.scan_row", "error", err, "role_uid", roleUID)
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		slog.Error("role_repository.GetUsersByRoleAndDepartment.rows_iteration", "error", err, "role_uid", roleUID)
		return nil, err
	}

	return users, nil
}

func (r *RoleRepository) IsUserAuthorizedApprover(ctx context.Context, q ports.Querier, userID int64, roleUID string, departmentUID string) (bool, error) {
	// User is authorized if they have the role with:
	// 1. Global scope (department_uid IS NULL), OR
	// 2. Scope matching the requester's department
	query := `
		SELECT COUNT(*)
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r.uid = ? AND (ur.department_uid IS NULL OR ur.department_uid = ?)`

	var count int
	err := q.QueryRowContext(ctx, query, userID, roleUID, departmentUID).Scan(&count)
	if err != nil {
		slog.Error("role_repository.IsUserAuthorizedApprover.scan", "error", err, "user_id", userID, "role_uid", roleUID)
		return false, err
	}
	return count > 0, nil
}

func (r *RoleRepository) scanRole(row *sql.Row) (*domain.Role, error) {
	var role domain.Role
	var scopeType sql.NullString
	var createdAt, updatedAt domain.Time
	var isSystem int
	err := row.Scan(
		&role.ID, &role.UID, &role.Name, &role.Description, &scopeType, &isSystem,
		&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("role_repository.scanRole.scan_row", "error", err)
		return nil, err
	}
	role.IsSystem = isSystem == 1
	role.ScopeType = domain.NormalizeRoleScopeType(scopeType.String)
	if role.ScopeType == "" {
		role.ScopeType = domain.RoleScopeGlobal
	}
	role.CreatedAt = createdAt.Time
	role.UpdatedAt = updatedAt.Time
	return &role, nil
}

func (r *RoleRepository) scanRoleRow(rows *sql.Rows) (*domain.Role, error) {
	var role domain.Role
	var scopeType sql.NullString
	var createdAt, updatedAt domain.Time
	var isSystem int
	err := rows.Scan(
		&role.ID, &role.UID, &role.Name, &role.Description, &scopeType, &isSystem,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	role.IsSystem = isSystem == 1
	role.ScopeType = domain.NormalizeRoleScopeType(scopeType.String)
	if role.ScopeType == "" {
		role.ScopeType = domain.RoleScopeGlobal
	}
	role.CreatedAt = createdAt.Time
	role.UpdatedAt = updatedAt.Time
	return &role, nil
}

func (r *RoleRepository) RemoveRoleFromUserForDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID string) error {
	query := `
		DELETE FROM user_roles
		WHERE role_id = (SELECT id FROM roles WHERE uid = ?)
		AND department_uid = ?`
	_, err := q.ExecContext(ctx, query, roleUID, departmentUID)
	if err != nil {
		slog.Error("role_repository.RemoveRoleFromUserForDepartment.exec_query", "error", err, "role_uid", roleUID, "department_uid", departmentUID)
	}
	return err
}

func (r *RoleRepository) GetDepartmentManager(ctx context.Context, q ports.Querier, departmentUID string) (*domain.User, error) {
	query := `
		SELECT u.id, u.uid, u.phone, u.password_hash, u.employee_uid, u.is_active,
		       COALESCE(u.preferred_language,''), u.created_at, u.updated_at
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		JOIN roles r ON ur.role_id = r.id
		WHERE r.uid = 'role_department_manager' AND ur.department_uid = ?
		LIMIT 1`

	userRepo := NewUserRepository()
	row := q.QueryRowContext(ctx, query, departmentUID)
	return userRepo.scanUser(row)
}

func (r *RoleRepository) GetManagedDepartmentUIDs(ctx context.Context, q ports.Querier, userID int64) ([]string, error) {
	query := `
		SELECT DISTINCT ur.department_uid
		FROM user_roles ur
		WHERE ur.user_id = ?
		  AND ur.department_uid IS NOT NULL
		ORDER BY ur.department_uid ASC`

	rows, err := q.QueryContext(ctx, query, userID)
	if err != nil {
		slog.Error("role_repository.GetManagedDepartmentUIDs.query", "error", err, "user_id", userID)
		return nil, err
	}
	defer rows.Close()

	managedDepartmentUIDs := make([]string, 0)
	for rows.Next() {
		var departmentUID string
		if err := rows.Scan(&departmentUID); err != nil {
			slog.Error("role_repository.GetManagedDepartmentUIDs.scan_row", "error", err, "user_id", userID)
			return nil, err
		}
		managedDepartmentUIDs = append(managedDepartmentUIDs, departmentUID)
	}

	if err := rows.Err(); err != nil {
		slog.Error("role_repository.GetManagedDepartmentUIDs.rows_iteration", "error", err, "user_id", userID)
		return nil, err
	}

	return managedDepartmentUIDs, nil
}
