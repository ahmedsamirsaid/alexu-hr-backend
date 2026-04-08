package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type ListUsersInput struct {
	Page     int
	PageSize int
}

type ListUsersOutput struct {
	Users      []UserListItem `json:"users"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
}

type UserListItem struct {
	UID          string   `json:"uid"`
	Phone        string   `json:"phone"`
	EmployeeUID  *string  `json:"employeeUid"`
	EmployeeName *string  `json:"employeeName,omitempty"`
	IsActive     bool     `json:"isActive"`
	Roles        []string `json:"roles"`
}

type ListUsersUseCase struct {
	db       ports.DB
	userRepo ports.UserRepository
	roleRepo ports.RoleRepository
	empRepo  ports.EmployeeRepository
}

func NewListUsersUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	empRepo ports.EmployeeRepository,
) *ListUsersUseCase {
	return &ListUsersUseCase{
		db:       db,
		userRepo: userRepo,
		roleRepo: roleRepo,
		empRepo:  empRepo,
	}
}

func (uc *ListUsersUseCase) Execute(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}

	offset := (input.Page - 1) * input.PageSize

	users, err := uc.userRepo.List(ctx, uc.db, input.PageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.userRepo.Count(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	items := make([]UserListItem, len(users))
	for i, user := range users {
		roles, err := uc.roleRepo.GetRolesForUser(ctx, uc.db, user.ID)
		if err != nil {
			return nil, err
		}

		roleNames := make([]string, len(roles))
		for j, role := range roles {
			roleNames[j] = role.Name
		}

		var employeeName *string
		if user.EmployeeUID != nil {
			emp, err := uc.empRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
			if err == nil && emp != nil {
				employeeName = &emp.Name
			}
		}

		items[i] = UserListItem{
			UID:          user.UID,
			Phone:        user.Phone,
			EmployeeUID:  user.EmployeeUID,
			EmployeeName: employeeName,
			IsActive:     user.IsActive,
			Roles:        roleNames,
		}
	}

	totalPages := (total + input.PageSize - 1) / input.PageSize

	return &ListUsersOutput{
		Users:      items,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
