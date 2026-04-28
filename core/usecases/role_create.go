package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewCreateRoleUseCase(
	db ports.DB,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *CreateRoleUseCase {
	return &CreateRoleUseCase{
		db:       db,
		roleRepo: roleRepo,
		auditor:  auditor,
	}
}

func (uc *CreateRoleUseCase) Execute(ctx context.Context, input CreateRoleInput) (*CreateRoleOutput, error) {
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

	actionSentence := fmt.Sprintf("Role '%s' (%s scope) was created", role.Name, role.ScopeType)

	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityRole, role.UID).
		WithMeta("action", actionSentence).
		WithMeta("name", role.Name).
		WithMeta("scope_type", role.ScopeType).
		WithMeta("description", role.Description).
		WithMeta("new_state", map[string]interface{}{
			"uid":        role.UID,
			"name":       role.Name,
			"scope_type": role.ScopeType,
		}).
		Save(ctx)

	if err := uc.roleRepo.Create(ctx, uc.db, role); err != nil {
		return nil, err
	}

	return &CreateRoleOutput{
		UID: role.UID,
	}, nil
}
