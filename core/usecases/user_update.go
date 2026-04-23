package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
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
	auditor  audit.Auditor
}

func NewUpdateUserUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	auditor audit.Auditor,
) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		db:       db,
		userRepo: userRepo,
		auditor:  auditor,
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

	// Capture old values for audit metadata
	oldPhone := user.Phone
	oldEmployeeUID := ""
	if user.EmployeeUID != nil {
		oldEmployeeUID = *user.EmployeeUID
	}
	oldIsActive := user.IsActive
	passwordChanged := false

	// Prepare audit metadata
	auditBuilder := uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityUser, input.UserUID)

	// Check phone uniqueness if changed
	if input.Phone != nil && *input.Phone != user.Phone {
		existing, err := uc.userRepo.GetByPhone(ctx, uc.db, *input.Phone)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrPhoneAlreadyExists
		}
		auditBuilder.WithMeta("old_phone", oldPhone).WithMeta("new_phone", *input.Phone)
		user.Phone = *input.Phone
	}

	// Update password if provided
	if input.Password != nil && *input.Password != "" {
		if err := SetUserPassword(user, *input.Password); err != nil {
			return err
		}
		passwordChanged = true
	}

	// Update employee link
	if input.EmployeeUID != nil {
		newEmployeeUID := ""
		if *input.EmployeeUID == "" {
			user.EmployeeUID = nil
		} else {
			user.EmployeeUID = input.EmployeeUID
			newEmployeeUID = *input.EmployeeUID
		}
		auditBuilder.WithMeta("old_employee_uid", oldEmployeeUID).WithMeta("new_employee_uid", newEmployeeUID)
	}

	// Update active status
	if input.IsActive != nil {
		auditBuilder.WithMeta("old_is_active", oldIsActive).WithMeta("new_is_active", *input.IsActive)
		user.IsActive = *input.IsActive
	}

	// DO NOT log passwords - use password_changed flag instead
	if passwordChanged {
		auditBuilder.WithMeta("password_changed", true)
	}

	// Audit log after successful update
	defer auditBuilder.Save(ctx)

	return uc.userRepo.Update(ctx, uc.db, user)
}
