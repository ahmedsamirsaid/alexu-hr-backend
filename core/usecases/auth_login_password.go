package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LoginPasswordInput struct {
	Phone    string
	Password string
}

type LoginPasswordOutput struct {
	AccessToken  string
	RefreshToken string
	User         *UserOutput
}

type LoginPasswordUseCase struct {
	db               ports.DB
	userRepo         ports.UserRepository
	roleRepo         ports.RoleRepository
	permissionRepo   ports.PermissionRepository
	employeeRepo     ports.EmployeeRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtService       JWTService
	devOTPBypass     bool
	devBypassOTP     string
	refreshTokenDays int
	auditor          audit.Auditor
}

func NewLoginPasswordUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	permissionRepo ports.PermissionRepository,
	employeeRepo ports.EmployeeRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtService JWTService,
	devOTPBypass bool,
	devBypassOTP string,
	refreshTokenDays int,
	auditor audit.Auditor,
) *LoginPasswordUseCase {
	return &LoginPasswordUseCase{
		db:               db,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		permissionRepo:   permissionRepo,
		employeeRepo:     employeeRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		devOTPBypass:     devOTPBypass,
		devBypassOTP:     devBypassOTP,
		refreshTokenDays: refreshTokenDays,
		auditor:          auditor,
	}
}

func (uc *LoginPasswordUseCase) Execute(ctx context.Context, input LoginPasswordInput) (*LoginPasswordOutput, error) {
	user, err := uc.userRepo.GetByPhone(ctx, uc.db, input.Phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Check for dev bypass - allows using dev OTP as password
	isDevBypass := uc.devOTPBypass && input.Password == uc.devBypassOTP
	if !isDevBypass {
		if user.PasswordHash == nil {
			return nil, ErrPasswordNotSet
		}

		if !CheckPassword(user, input.Password) {
			return nil, ErrInvalidCredentials
		}
	}

	// Load roles
	roles, err := uc.roleRepo.GetRolesForUser(ctx, uc.db, user.ID)
	if err != nil {
		return nil, err
	}

	// Bulk fetch permissions for all roles
	roleIDs := make([]int64, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}
	permsByRole, err := uc.permissionRepo.GetPermissionsForRoles(ctx, uc.db, roleIDs)
	if err != nil {
		return nil, err
	}

	// Assign permissions to roles
	user.Roles = make([]domain.Role, len(roles))
	for i, r := range roles {
		perms := permsByRole[r.ID]
		r.Permissions = make([]domain.Permission, len(perms))
		for j, p := range perms {
			r.Permissions[j] = *p
		}
		user.Roles[i] = *r
	}

	managedDepartmentUIDs, err := uc.roleRepo.GetManagedDepartmentUIDs(ctx, uc.db, user.ID)
	if err != nil {
		return nil, err
	}
	user.ManagedDepartmentUIDs = managedDepartmentUIDs

	// Check web portal access (user must have a role other than "Employee")
	if !user.HasWebPortalAccess() {
		return nil, ErrWebAccessDenied
	}

	// Generate tokens
	accessToken, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken := domain.NewRefreshToken(user.ID, uc.refreshTokenDays)
	if err := uc.refreshTokenRepo.Create(ctx, uc.db, &refreshToken.RefreshToken); err != nil {
		return nil, err
	}

	// Audit log for login (skip admin user)
	if user.Phone != "+201000000000" {
		actorUID := user.UID // Use user UID as actor if no employee UID
		if user.EmployeeUID != nil {
			actorUID = *user.EmployeeUID
		}
		actionParams := map[string]interface{}{
			"Phone": user.Phone,
		}
		uc.auditor.Actor(actorUID).
			Did(audit.ActionLogin).
			On(audit.EntityUser, user.UID).
			WithMeta("action_key", "audit.sentence.user_login").
			WithMeta("action_params", actionParams).
			WithMeta("phone", user.Phone).
			WithMeta("login_method", "password").
			Save(ctx)
	}

	return &LoginPasswordOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.RawToken,
		User:         uc.toUserOutput(ctx, user),
	}, nil
}

func (uc *LoginPasswordUseCase) toUserOutput(ctx context.Context, user *domain.User) *UserOutput {
	roleNames := make([]string, len(user.Roles))
	permSet := make(map[string]bool)

	for i, role := range user.Roles {
		roleNames[i] = role.Name
		if role.HasAllPermissions() {
			permSet["*"] = true
		} else {
			for _, perm := range role.Permissions {
				permSet[perm.Code] = true
			}
		}
	}

	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	output := &UserOutput{
		UID:                   user.UID,
		Phone:                 user.Phone,
		EmployeeUID:           user.EmployeeUID,
		AccessScope:           determineAccessScope(user.Roles),
		Roles:                 roleNames,
		Permissions:           permissions,
		ManagedDepartmentUIDs: append([]string(nil), user.ManagedDepartmentUIDs...),
	}

	// Fetch employee name if linked
	if user.EmployeeUID != nil {
		employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
		if err == nil && employee != nil {
			output.EmployeeName = &employee.Name
		}
	}

	return output
}
