package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.User, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.User, error)
	GetByPhone(ctx context.Context, q Querier, phone string) (*domain.User, error)
	Create(ctx context.Context, q Querier, user *domain.User) error
	Update(ctx context.Context, q Querier, user *domain.User) error
	List(ctx context.Context, q Querier, limit, offset int) ([]*domain.User, error)
	Count(ctx context.Context, q Querier) (int, error)

	// ExistingPhones returns the subset of phones that already exist in the users table.
	ExistingPhones(ctx context.Context, q Querier, phones []string) ([]string, error)

	// GetByEmployeeUID returns the user linked to the given employee UID.
	GetByEmployeeUID(ctx context.Context, q Querier, employeeUID string) (*domain.User, error)
}

type RoleRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.Role, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.Role, error)
	GetByName(ctx context.Context, q Querier, name string) (*domain.Role, error)
	Create(ctx context.Context, q Querier, role *domain.Role) error
	Update(ctx context.Context, q Querier, role *domain.Role) error
	List(ctx context.Context, q Querier) ([]*domain.Role, error)
	Delete(ctx context.Context, q Querier, id int64) error
	GetRolesForUser(ctx context.Context, q Querier, userID int64) ([]*domain.Role, error)
	AssignRoleToUser(ctx context.Context, q Querier, userID, roleID int64) error
	RemoveRoleFromUser(ctx context.Context, q Querier, userID, roleID int64) error
	AssignPermissionToRole(ctx context.Context, q Querier, roleID, permissionID int64) error
	RemovePermissionFromRole(ctx context.Context, q Querier, roleID, permissionID int64) error
	SetRolePermissions(ctx context.Context, q Querier, roleID int64, permissionIDs []int64) error

	// AssignRoleToUserWithDepartment assigns a role to a user with optional department scope
	AssignRoleToUserWithDepartment(ctx context.Context, q Querier, userID, roleID int64, departmentUID *string) error

	// GetUsersByRoleAndDepartment returns users with the given role, scoped to department
	// If departmentUID is nil, returns users with global role assignment
	// If departmentUID is set, returns users with matching department scope OR global role
	GetUsersByRoleAndDepartment(ctx context.Context, q Querier, roleUID string, departmentUID *string) ([]*domain.User, error)

	// IsUserAuthorizedApprover checks if user can approve for the given role and department
	// Returns true if user has the role globally OR scoped to the specific department
	IsUserAuthorizedApprover(ctx context.Context, q Querier, userID int64, roleUID string, departmentUID string) (bool, error)

	// RemoveRoleFromUserForDepartment removes a role assignment scoped to a specific department
	RemoveRoleFromUserForDepartment(ctx context.Context, q Querier, roleUID string, departmentUID string) error

	// GetDepartmentManager returns the user who has the Department Manager role scoped to the given department
	GetDepartmentManager(ctx context.Context, q Querier, departmentUID string) (*domain.User, error)

	// GetManagedDepartmentUIDs returns department UIDs from any department-scoped role assignments
	GetManagedDepartmentUIDs(ctx context.Context, q Querier, userID int64) ([]string, error)

	// GetRoleNamesByEmployeeUIDs returns a map of employee_uid → primary role name for a batch
	// of employees. The "primary" role is the highest-priority non-employee role found.
	GetRoleNamesByEmployeeUIDs(ctx context.Context, q Querier, employeeUIDs []string) (map[string]string, error)
}

type PermissionRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.Permission, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.Permission, error)
	List(ctx context.Context, q Querier) ([]*domain.Permission, error)
	GetPermissionsForRole(ctx context.Context, q Querier, roleID int64) ([]*domain.Permission, error)
	GetPermissionsForRoles(ctx context.Context, q Querier, roleIDs []int64) (map[int64][]*domain.Permission, error)
}

type OTPRepository interface {
	Create(ctx context.Context, q Querier, otp *domain.OTPCode) error
	GetLatestByPhone(ctx context.Context, q Querier, phone string) (*domain.OTPCode, error)
	MarkUsed(ctx context.Context, q Querier, id int64) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, q Querier, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, q Querier, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, q Querier, id int64) error
	RevokeAllForUser(ctx context.Context, q Querier, userID int64) error
}
