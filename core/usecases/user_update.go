package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type UpdateUserInput struct {
	UserUID     string
	Phone       *string
	Password    *string
	EmployeeUID *string
	IsActive    *bool
}

type UpdateUserUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
}

func NewUpdateUserUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		db:       db,
		userRepo: userRepo,
	}
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, input UpdateUserInput) error {
	user, err := uc.userRepo.GetByUID(ctx, uc.db, input.UserUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Check phone uniqueness if changed
	if input.Phone != nil && *input.Phone != user.Phone {
		existing, err := uc.userRepo.GetByPhone(ctx, uc.db, *input.Phone)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrPhoneAlreadyExists
		}
		user.Phone = *input.Phone
	}

	// Update password if provided
	if input.Password != nil && *input.Password != "" {
		if err := SetUserPassword(user, *input.Password); err != nil {
			return err
		}
	}

	// Update employee link
	if input.EmployeeUID != nil {
		if *input.EmployeeUID == "" {
			user.EmployeeUID = nil
		} else {
			user.EmployeeUID = input.EmployeeUID
		}
	}

	// Update active status
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	return uc.userRepo.Update(ctx, uc.db, user)
}
