package usecases

import (
	"context"
	"errors"
	"strings"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
)

var ErrNoEmployeeLinked = errors.New("no_employee_linked")

type UpdateOwnEmployeeProfileInput struct {
	UserID          int64
	Name            string
	Mobile          string
	TelephoneNumber *string
	Email           *string
}

type UpdateOwnEmployeeProfileOutput struct {
	UID             string
	Name            string
	Mobile          string
	TelephoneNumber *string
	Email           *string
}

type UpdateOwnEmployeeProfileUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	userRepo     ports.UserRepository
	auditor      audit.Auditor
}

func NewUpdateOwnEmployeeProfileUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	auditor audit.Auditor,
) *UpdateOwnEmployeeProfileUseCase {
	return &UpdateOwnEmployeeProfileUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		auditor:      auditor,
	}
}

func (uc *UpdateOwnEmployeeProfileUseCase) Execute(ctx context.Context, input UpdateOwnEmployeeProfileInput) (*UpdateOwnEmployeeProfileOutput, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	mobile := strings.TrimSpace(input.Mobile)
	if mobile == "" {
		return nil, errors.New("mobile is required")
	}

	user, err := uc.userRepo.GetByID(ctx, uc.db, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.EmployeeUID == nil || *user.EmployeeUID == "" {
		return nil, ErrNoEmployeeLinked
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	if mobile != employee.Mobile {
		existingMobiles, err := uc.employeeRepo.ExistingMobiles(ctx, uc.db, []string{mobile})
		if err != nil {
			return nil, err
		}
		if len(existingMobiles) > 0 {
			return nil, ErrEmployeeMobileAlreadyExists
		}

		existingUser, err := uc.userRepo.GetByPhone(ctx, uc.db, mobile)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrPhoneAlreadyExists
		}
	}

	oldName := employee.Name
	oldMobile := employee.Mobile
	oldTelephoneNumber := stringPtrValue(employee.TelephoneNumber)
	oldEmail := stringPtrValue(employee.Email)

	employee.Name = name
	employee.Mobile = mobile
	employee.TelephoneNumber = trimStringPtr(input.TelephoneNumber)
	employee.Email = trimStringPtr(input.Email)
	user.Phone = mobile

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := uc.employeeRepo.Update(ctx, tx, employee); err != nil {
		return nil, err
	}
	if err := uc.userRepo.Update(ctx, tx, user); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	defer uc.auditor.From(ctx).
		Did(audit.ActionUpdate).
		On(audit.EntityEmployee, employee.UID).
		WithMeta("old_name", oldName).
		WithMeta("new_name", employee.Name).
		WithMeta("old_mobile", oldMobile).
		WithMeta("new_mobile", employee.Mobile).
		WithMeta("old_telephone_number", oldTelephoneNumber).
		WithMeta("new_telephone_number", stringPtrValue(employee.TelephoneNumber)).
		WithMeta("old_email", oldEmail).
		WithMeta("new_email", stringPtrValue(employee.Email)).
		WithMeta("synced_user_phone", true).
		Save(ctx)

	return &UpdateOwnEmployeeProfileOutput{
		UID:             employee.UID,
		Name:            employee.Name,
		Mobile:          employee.Mobile,
		TelephoneNumber: employee.TelephoneNumber,
		Email:           employee.Email,
	}, nil
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
