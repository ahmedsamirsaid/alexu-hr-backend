package usecases

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type VerifyOTPInput struct {
	Phone string
	OTP   string
}

type VerifyOTPOutput struct {
	AccessToken  string
	RefreshToken string
	User         *UserOutput
}

type UserOutput struct {
	UID                   string   `json:"uid"`
	Phone                 string   `json:"phone"`
	EmployeeUID           *string  `json:"employeeUid"`
	EmployeeName          *string  `json:"employeeName,omitempty"`
	Roles                 []string `json:"roles"`
	Permissions           []string `json:"permissions"`
	ManagedDepartmentUIDs []string `json:"managedDepartmentUids"`
}

type VerifyOTPUseCase struct {
	db               ports.DB
	userRepo         ports.UserRepository
	roleRepo         ports.RoleRepository
	permissionRepo   ports.PermissionRepository
	employeeRepo     ports.EmployeeRepository
	otpRepo          ports.OTPRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtService       JWTService
	devOTPBypass     bool
	devBypassOTP     string
	refreshTokenDays int
}

type JWTService interface {
	GenerateAccessToken(user *domain.User) (string, error)
}

func NewVerifyOTPUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	permissionRepo ports.PermissionRepository,
	employeeRepo ports.EmployeeRepository,
	otpRepo ports.OTPRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtService JWTService,
	devOTPBypass bool,
	devBypassOTP string,
	refreshTokenDays int,
) *VerifyOTPUseCase {
	return &VerifyOTPUseCase{
		db:               db,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		permissionRepo:   permissionRepo,
		employeeRepo:     employeeRepo,
		otpRepo:          otpRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		devOTPBypass:     devOTPBypass,
		devBypassOTP:     devBypassOTP,
		refreshTokenDays: refreshTokenDays,
	}
}

func (uc *VerifyOTPUseCase) Execute(ctx context.Context, input VerifyOTPInput) (*VerifyOTPOutput, error) {
	// Check for dev bypass
	if uc.devOTPBypass && input.OTP == uc.devBypassOTP {
		slog.Info("auth_verify_otp.Execute.otp_bypass_used", "phone", input.Phone)
		return uc.handleSuccessfulAuth(ctx, input.Phone)
	}

	// Validate OTP
	otp, err := uc.otpRepo.GetLatestByPhone(ctx, uc.db, input.Phone)
	if err != nil {
		return nil, err
	}
	if otp == nil {
		return nil, ErrInvalidOTP
	}

	if otp.IsExpired() {
		return nil, ErrOTPExpired
	}

	if !otp.IsValid(input.OTP) {
		return nil, ErrInvalidOTP
	}

	// Mark OTP as used
	if err := uc.otpRepo.MarkUsed(ctx, uc.db, otp.ID); err != nil {
		return nil, err
	}

	return uc.handleSuccessfulAuth(ctx, input.Phone)
}

func (uc *VerifyOTPUseCase) handleSuccessfulAuth(ctx context.Context, phone string) (*VerifyOTPOutput, error) {
	// Get or create user
	user, err := uc.userRepo.GetByPhone(ctx, uc.db, phone)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
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

	// Generate tokens
	accessToken, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken := domain.NewRefreshToken(user.ID, uc.refreshTokenDays)
	if err := uc.refreshTokenRepo.Create(ctx, uc.db, &refreshToken.RefreshToken); err != nil {
		return nil, err
	}

	return &VerifyOTPOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.RawToken,
		User:         uc.toUserOutput(ctx, user),
	}, nil
}

func (uc *VerifyOTPUseCase) toUserOutput(ctx context.Context, user *domain.User) *UserOutput {
	roleNames := make([]string, len(user.Roles))
	permSet := make(map[string]bool)

	for i, role := range user.Roles {
		roleNames[i] = role.Name
		if role.HasAllPermissions() {
			// Admin has all permissions - we'll populate this on the frontend
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
