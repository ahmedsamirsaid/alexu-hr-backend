package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type ListPermissionsOutput struct {
	Permissions []PermissionItem `json:"permissions"`
}

type ListPermissionsUseCase struct {
	db       ports.DB
	permRepo ports.PermissionRepository
}

func NewListPermissionsUseCase(
	db ports.DB,
	permRepo ports.PermissionRepository,
) *ListPermissionsUseCase {
	return &ListPermissionsUseCase{
		db:       db,
		permRepo: permRepo,
	}
}

func (uc *ListPermissionsUseCase) Execute(ctx context.Context) (*ListPermissionsOutput, error) {
	perms, err := uc.permRepo.List(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	items := make([]PermissionItem, len(perms))
	for i, p := range perms {
		items[i] = PermissionItem{
			UID:         p.UID,
			Code:        p.Code,
			Description: p.Description,
		}
	}

	return &ListPermissionsOutput{
		Permissions: items,
	}, nil
}
