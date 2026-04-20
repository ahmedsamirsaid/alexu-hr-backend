package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CreateRoleInput struct {
	Name        string
	Description string
	ScopeType   string
}

type CreateRoleOutput struct {
	UID string
}

type CreateRoleUseCase struct {
	db       ports.DB
	roleRepo ports.RoleRepository
}

func NewCreateRoleUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
) *CreateRoleUseCase {
	return &CreateRoleUseCase{
		db:       db,
		roleRepo: roleRepo,
	}
}

func (uc *CreateRoleUseCase) Execute(ctx context.Context, input CreateRoleInput) (*CreateRoleOutput, error) {
	// Check if name already exists
	existing, err := uc.roleRepo.GetByName(ctx, uc.db, input.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRoleNameExists
	}

	normalizedScopeType := domain.NormalizeRoleScopeType(input.ScopeType)
	if normalizedScopeType == "" {
		return nil, ErrInvalidRoleScopeType
	}

	role := domain.NewRole(input.Name, input.Description, normalizedScopeType)

	if err := uc.roleRepo.Create(ctx, uc.db, role); err != nil {
		return nil, err
	}

	return &CreateRoleOutput{
		UID: role.UID,
	}, nil
}
