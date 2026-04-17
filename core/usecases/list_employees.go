package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// ListEmployeesInput defines the filters for listing employees.
type ListEmployeesInput struct {
	Status                *domain.EmployeeStatus
	HireDateFrom          *time.Time
	HireDateTo            *time.Time
	Role                  *string // Filter by role name
	ManagedDepartmentUIDs []string
}

// EmployeeUserInfo contains the linked user account information.
type EmployeeUserInfo struct {
	UserUID  string
	Phone    string
	IsActive bool
	Roles    []string
}

// EmployeeListItem represents an employee in the list response.
type EmployeeListItem struct {
	UID           string
	Name          string
	Mobile        string
	GovernmentID  string
	UniversityID  string
	Email         *string
	HireDate      time.Time
	Status        domain.EmployeeStatus
	DepartmentUID *string // Department UID (nil if not assigned)
	ShiftUID      *string
	User          *EmployeeUserInfo // Linked user account info (nil if no user linked)
}

// ListEmployeesOutput contains the list of employees.
type ListEmployeesOutput struct {
	Employees []EmployeeListItem
}

// ListEmployeesUseCase handles listing employees.
type ListEmployeesUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	userRepo     ports.UserRepository
	roleRepo     ports.RoleRepository
}

// NewListEmployeesUseCase creates a new list employees use case.
func NewListEmployeesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
) *ListEmployeesUseCase {
	return &ListEmployeesUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
	}
}

// Execute lists employees with optional filters.
func (uc *ListEmployeesUseCase) Execute(ctx context.Context, input ListEmployeesInput) (*ListEmployeesOutput, error) {
	filter := &ports.EmployeeListFilter{
		Status:       input.Status,
		HireDateFrom: input.HireDateFrom,
		HireDateTo:   input.HireDateTo,
	}

	employees, err := uc.employeeRepo.List(ctx, uc.db, filter)
	if err != nil {
		return nil, err
	}

	allowedDepartments := make(map[string]struct{}, len(input.ManagedDepartmentUIDs))
	for _, departmentUID := range input.ManagedDepartmentUIDs {
		allowedDepartments[departmentUID] = struct{}{}
	}
	limitByDepartment := len(allowedDepartments) > 0

	items := make([]EmployeeListItem, 0, len(employees))
	for _, emp := range employees {
		if limitByDepartment {
			if emp.DepartmentUID == nil {
				continue
			}
			if _, ok := allowedDepartments[*emp.DepartmentUID]; !ok {
				continue
			}
		}

		item := EmployeeListItem{
			UID:           emp.UID,
			Name:          emp.Name,
			Mobile:        emp.Mobile,
			GovernmentID:  emp.GovernmentID,
			UniversityID:  emp.UniversityID,
			Email:         emp.Email,
			HireDate:      emp.HireDate,
			Status:        emp.Status,
			DepartmentUID: emp.DepartmentUID,
			ShiftUID:      emp.ShiftUID,
		}

		// Fetch linked user details
		user, err := uc.userRepo.GetByEmployeeUID(ctx, uc.db, emp.UID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			// Fetch user roles
			roles, err := uc.roleRepo.GetRolesForUser(ctx, uc.db, user.ID)
			if err != nil {
				return nil, err
			}
			roleNames := make([]string, 0, len(roles))
			for _, role := range roles {
				roleNames = append(roleNames, role.Name)
			}

			item.User = &EmployeeUserInfo{
				UserUID:  user.UID,
				Phone:    user.Phone,
				IsActive: user.IsActive,
				Roles:    roleNames,
			}
		}

		// Apply role filter if specified
		if input.Role != nil {
			if item.User == nil {
				continue // Skip employees without linked user when filtering by role
			}
			hasRole := false
			for _, r := range item.User.Roles {
				if r == *input.Role {
					hasRole = true
					break
				}
			}
			if !hasRole {
				continue // Skip employees without the specified role
			}
		}

		items = append(items, item)
	}

	return &ListEmployeesOutput{Employees: items}, nil
}
