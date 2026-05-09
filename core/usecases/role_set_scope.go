package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type SetRoleScopeInput struct {
	RoleUID   string
	ScopeType string
}

type SetRoleScopeUseCase struct {
	db       ports.DB
	roleRepo ports.RoleRepository
	auditor  audit.Auditor
}

func NewSetRoleScopeUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *SetRoleScopeUseCase {
	return &SetRoleScopeUseCase{
		db:       db,
		roleRepo: roleRepo,
		auditor:  auditor,
	}
}

func (uc *SetRoleScopeUseCase) Execute(ctx context.Context, input SetRoleScopeInput) error {
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

	normalizedScopeType := domain.NormalizeRoleScopeType(input.ScopeType)
	if normalizedScopeType == "" {
		return ErrInvalidRoleScopeType
	}

	// Capture old scope type for audit
	oldScopeType := role.ScopeType

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionParams := map[string]interface{}{
		"Actor":    actorName,
		"Role":     role.Name,
		"OldScope": oldScopeType,
		"NewScope": normalizedScopeType,
	}

	// Audit log will fire after successful scope update
	defer uc.auditor.From(ctx).
		Did(audit.ActionUpdate).
		On(audit.EntityRole, input.RoleUID).
		WithMeta("action_key", "audit.sentence.set_role_scope").
		WithMeta("action_params", actionParams).
		WithMeta("role_name", role.Name).
		WithMeta("old_scope_type", oldScopeType).
		WithMeta("new_scope_type", normalizedScopeType).
		Save(ctx)

	role.ScopeType = normalizedScopeType
	return uc.roleRepo.Update(ctx, uc.db, role)
}
