package usecases

import (
	"context"

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
}

func NewSetRoleScopeUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
) *SetRoleScopeUseCase {
	return &SetRoleScopeUseCase{
		db:       db,
		roleRepo: roleRepo,
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

	role.ScopeType = normalizedScopeType
	return uc.roleRepo.Update(ctx, uc.db, role)
}
