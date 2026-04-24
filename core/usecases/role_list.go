package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type ListRolesOutput struct {
	Roles []RoleListItem `json:"roles"`
}

type RoleListItem struct {
	UID         string           `json:"uid"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ScopeType   string           `json:"scopeType"`
	IsSystem    bool             `json:"isSystem"`
	Permissions []PermissionItem `json:"permissions"`
}

type PermissionItem struct {
	UID         string `json:"uid"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type ListRolesUseCase struct {
	db             ports.DB
	roleRepo       ports.RoleRepository
	permissionRepo ports.PermissionRepository
}

func NewListRolesUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	permissionRepo ports.PermissionRepository,
) *ListRolesUseCase {
	return &ListRolesUseCase{
		db:             db,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

func (uc *ListRolesUseCase) Execute(ctx context.Context) (*ListRolesOutput, error) {
	roles, err := uc.roleRepo.List(ctx, uc.db)
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

	items := make([]RoleListItem, len(roles))
	for i, role := range roles {
		rolePerms := permsByRole[role.ID]
		perms := make([]PermissionItem, len(rolePerms))
		for j, perm := range rolePerms {
			perms[j] = PermissionItem{
				UID:         perm.UID,
				Code:        perm.Code,
				Description: perm.Description,
			}
		}

		items[i] = RoleListItem{
			UID:         role.UID,
			Name:        role.Name,
			Description: role.Description,
			ScopeType:   role.ScopeType,
			IsSystem:    role.IsSystem,
			Permissions: perms,
		}
	}

	return &ListRolesOutput{
		Roles: items,
	}, nil
}
