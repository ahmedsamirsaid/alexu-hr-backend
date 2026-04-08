package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// GetEmployeeInput defines the input for getting an employee.
type GetEmployeeInput struct {
	UID string
}

// GetEmployeeOutput contains the employee details.
type GetEmployeeOutput struct {
	UID          string
	Name         string
	Mobile       string
	GovernmentID string
	UniversityID string
	Email        *string
	HireDate     time.Time
	Status       domain.EmployeeStatus
}

// GetEmployeeUseCase handles retrieving a single employee by UID.
type GetEmployeeUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
}

// NewGetEmployeeUseCase creates a new get employee use case.
func NewGetEmployeeUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
) *GetEmployeeUseCase {
	return &GetEmployeeUseCase{
		db:           db,
		employeeRepo: employeeRepo,
	}
}

// Execute retrieves an employee by UID.
func (uc *GetEmployeeUseCase) Execute(ctx context.Context, input GetEmployeeInput) (*GetEmployeeOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	return &GetEmployeeOutput{
		UID:          employee.UID,
		Name:         employee.Name,
		Mobile:       employee.Mobile,
		GovernmentID: employee.GovernmentID,
		UniversityID: employee.UniversityID,
		Email:        employee.Email,
		HireDate:     employee.HireDate,
		Status:       employee.Status,
	}, nil
}
