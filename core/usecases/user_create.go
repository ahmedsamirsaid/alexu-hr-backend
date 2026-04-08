package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CreateUserInput struct {
	Phone       string
	Password    *string
	EmployeeUID *string
}

type CreateUserOutput struct {
	UID string
}

type CreateUserUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
}

func NewCreateUserUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		db:       db,
		userRepo: userRepo,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	// Check if phone already exists
	existing, err := uc.userRepo.GetByPhone(ctx, uc.db, input.Phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPhoneAlreadyExists
	}

	user := domain.NewUser(input.Phone)
	user.EmployeeUID = input.EmployeeUID

	if input.Password != nil && *input.Password != "" {
		if err := SetUserPassword(user, *input.Password); err != nil {
			return nil, err
		}
	}

	if err := uc.userRepo.Create(ctx, uc.db, user); err != nil {
		return nil, err
	}

	return &CreateUserOutput{
		UID: user.UID,
	}, nil
}
