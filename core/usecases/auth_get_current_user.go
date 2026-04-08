package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetCurrentUserInput struct {
	UserID int64
}

type GetCurrentUserOutput struct {
	UID         string           `json:"uid"`
	Phone       string           `json:"phone"`
	EmployeeUID *string          `json:"employeeUid"`
	IsActive    bool             `json:"isActive"`
	Roles       []RoleOutput     `json:"roles"`
	Permissions []string         `json:"permissions"`
	Employee    *EmployeeBasic   `json:"employee,omitempty"`
}

type RoleOutput struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type EmployeeBasic struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type GetCurrentUserUseCase struct {
	db             ports.DB
	userRepo       ports.UserRepository
	roleRepo       ports.RoleRepository
	permissionRepo ports.PermissionRepository
	employeeRepo   ports.EmployeeRepository
}

func NewGetCurrentUserUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	permissionRepo ports.PermissionRepository,
	employeeRepo ports.EmployeeRepository,
) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{
		db:             db,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		employeeRepo:   employeeRepo,
	}
}

func (uc *GetCurrentUserUseCase) Execute(ctx context.Context, input GetCurrentUserInput) (*GetCurrentUserOutput, error) {
	user, err := uc.userRepo.GetByID(ctx, uc.db, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Load roles
	roles, err := uc.roleRepo.GetRolesForUser(ctx, uc.db, user.ID)
	if err != nil {
		return nil, err
	}

	// Bulk fetch permissions for all roles
	roleIDs := make([]int64, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}
	permsByRole, err := uc.permissionRepo.GetPermissionsForRoles(ctx, uc.db, roleIDs)
	if err != nil {
		return nil, err
	}

	// Assign permissions to roles
	user.Roles = make([]domain.Role, len(roles))
	for i, r := range roles {
		perms := permsByRole[r.ID]
		r.Permissions = make([]domain.Permission, len(perms))
		for j, p := range perms {
			r.Permissions[j] = *p
		}
		user.Roles[i] = *r
	}

	output := &GetCurrentUserOutput{
		UID:         user.UID,
		Phone:       user.Phone,
		EmployeeUID: user.EmployeeUID,
		IsActive:    user.IsActive,
		Roles:       make([]RoleOutput, len(roles)),
		Permissions: uc.collectPermissions(user),
	}

	for i, role := range roles {
		output.Roles[i] = RoleOutput{
			UID:  role.UID,
			Name: role.Name,
		}
	}

	// Load linked employee if exists
	if user.EmployeeUID != nil {
		employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
		if err != nil {
			return nil, err
		}
		if employee != nil {
			output.Employee = &EmployeeBasic{
				UID:  employee.UID,
				Name: employee.Name,
			}
		}
	}

	return output, nil
}

func (uc *GetCurrentUserUseCase) collectPermissions(user *domain.User) []string {
	permSet := make(map[string]bool)

	for _, role := range user.Roles {
		if role.IsSystem {
			return []string{"*"}
		}
		for _, perm := range role.Permissions {
			permSet[perm.Code] = true
		}
	}

	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}
	return permissions
}
