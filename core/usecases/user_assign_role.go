package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type AssignRoleInput struct {
	UserUID string
	RoleUID string
}

type AssignRoleUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
	roleRepo ports.RoleRepository
}

func NewAssignRoleUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		db:       db,
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (uc *AssignRoleUseCase) Execute(ctx context.Context, input AssignRoleInput) error {
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	role, err := uc.roleRepo.GetByUID(ctx, uc.db, input.RoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	return uc.roleRepo.AssignRoleToUser(ctx, uc.db, user.ID, role.ID)
}

type RemoveRoleInput struct {
	UserUID string
	RoleUID string
}

type RemoveRoleUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
	roleRepo ports.RoleRepository
}

func NewRemoveRoleUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
) *RemoveRoleUseCase {
	return &RemoveRoleUseCase{
		db:       db,
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (uc *RemoveRoleUseCase) Execute(ctx context.Context, input RemoveRoleInput) error {
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	role, err := uc.roleRepo.GetByUID(ctx, uc.db, input.RoleUID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	return uc.roleRepo.RemoveRoleFromUser(ctx, uc.db, user.ID, role.ID)
}
