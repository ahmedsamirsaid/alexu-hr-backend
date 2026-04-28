package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewCreateUserUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	auditor audit.Auditor,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		db:       db,
		userRepo: userRepo,
		auditor:  auditor,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
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

	// Determine employee link description for the action sentence
	employeeLink := ""
	if input.EmployeeUID != nil && *input.EmployeeUID != "" {
		employeeLink = " linked to employee " + *input.EmployeeUID
	}

	// Audit log after successful creation
	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityUser, user.UID).
		WithMeta("action", "New user account created for phone "+input.Phone+employeeLink).
		WithMeta("phone", input.Phone).
		WithMeta("employee_uid", input.EmployeeUID).
		WithMeta("new_state", map[string]interface{}{
			"uid":          user.UID,
			"phone":        input.Phone,
			"employee_uid": input.EmployeeUID,
			"is_active":    user.IsActive,
		}).
		Save(ctx)

	return &CreateUserOutput{
		UID: user.UID,
	}, nil
}
