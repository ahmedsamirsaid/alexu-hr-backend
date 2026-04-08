package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type SetRolePermissionsInput struct {
	RoleUID        string
	PermissionUIDs []string
}

type SetRolePermissionsUseCase struct {
	db       ports.DB
	roleRepo ports.RoleRepository
	permRepo ports.PermissionRepository
}

func NewSetRolePermissionsUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	permRepo ports.PermissionRepository,
) *SetRolePermissionsUseCase {
	return &SetRolePermissionsUseCase{
		db:       db,
		roleRepo: roleRepo,
		permRepo: permRepo,
	}
}

func (uc *SetRolePermissionsUseCase) Execute(ctx context.Context, input SetRolePermissionsInput) error {
	role, err := uc.roleRepo.GetByUID(ctx, uc.db, input.RoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	if role.IsSystem {
		return ErrCannotModifySystemRole
	}

	// Get permission IDs from UIDs
	allPerms, err := uc.permRepo.List(ctx, uc.db)
	if err != nil {
		return err
	}

	uidToID := make(map[string]int64)
	for _, p := range allPerms {
		uidToID[p.UID] = p.ID
	}

	permIDs := make([]int64, 0, len(input.PermissionUIDs))
	for _, uid := range input.PermissionUIDs {
		if id, ok := uidToID[uid]; ok {
			permIDs = append(permIDs, id)
		}
	}

	return uc.roleRepo.SetRolePermissions(ctx, uc.db, role.ID, permIDs)
}
