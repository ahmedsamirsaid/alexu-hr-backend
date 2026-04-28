package usecases

import (
	"context"
	"fmt"

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

	currentPerms, err := uc.permRepo.GetPermissionsForRole(ctx, uc.db, role.ID)
	if err != nil {
		return err
	}

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

	currentPermUIDs := make(map[string]bool)
	oldPermCodes := make([]string, 0, len(currentPerms))
	for _, p := range currentPerms {
		currentPermUIDs[p.UID] = true
		oldPermCodes = append(oldPermCodes, p.Code)
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

	newPermCodes := make([]string, 0, len(newPermUIDs))
	for uid := range newPermUIDs {
		newPermCodes = append(newPermCodes, uidToCode[uid])
	}

	actionSentence := fmt.Sprintf(
		"Permissions for role '%s' were updated: %d added, %d removed",
		role.Name, len(addedPerms), len(removedPerms),
	)

	defer uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityRole, input.RoleUID).
		WithMeta("action", actionSentence).
		WithMeta("role_name", role.Name).
		WithMeta("added_permissions", addedPerms).
		WithMeta("removed_permissions", removedPerms).
		WithMeta("old_permissions", oldPermCodes).
		WithMeta("new_permissions", newPermCodes).
		Save(ctx)

	return uc.roleRepo.SetRolePermissions(ctx, uc.db, role.ID, permIDs)
}
