package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewSetRolePermissionsUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	permRepo ports.PermissionRepository,
	auditor audit.Auditor,
) *SetRolePermissionsUseCase {
	return &SetRolePermissionsUseCase{
		db:       db,
		roleRepo: roleRepo,
		permRepo: permRepo,
		auditor:  auditor,
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

	// Get current permissions for audit trail
	currentPerms, err := uc.permRepo.GetPermissionsForRole(ctx, uc.db, role.ID)
	if err != nil {
		return err
	}

	// Get permission IDs from UIDs
	allPerms, err := uc.permRepo.List(ctx, uc.db)
	if err != nil {
		return err
	}

	uidToID := make(map[string]int64)
	uidToCode := make(map[string]string)
	for _, p := range allPerms {
		uidToID[p.UID] = p.ID
		uidToCode[p.UID] = p.Code
	}

	permIDs := make([]int64, 0, len(input.PermissionUIDs))
	newPermUIDs := make(map[string]bool)
	for _, uid := range input.PermissionUIDs {
		if id, ok := uidToID[uid]; ok {
			permIDs = append(permIDs, id)
			newPermUIDs[uid] = true
		}
	}

	// Calculate added and removed permissions
	currentPermUIDs := make(map[string]bool)
	for _, p := range currentPerms {
		currentPermUIDs[p.UID] = true
	}

	addedPerms := []string{}
	removedPerms := []string{}

	for uid := range newPermUIDs {
		if !currentPermUIDs[uid] {
			addedPerms = append(addedPerms, uidToCode[uid])
		}
	}

	for uid := range currentPermUIDs {
		if !newPermUIDs[uid] {
			removedPerms = append(removedPerms, uidToCode[uid])
		}
	}

	// Audit log after successful permission update
	defer uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityRole, input.RoleUID).
		WithMeta("added_permissions", addedPerms).
		WithMeta("removed_permissions", removedPerms).
		Save(ctx)

	return uc.roleRepo.SetRolePermissions(ctx, uc.db, role.ID, permIDs)
}
