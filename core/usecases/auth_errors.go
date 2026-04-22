package usecases

import "errors"

var (
	ErrInvalidOTP             = errors.New("invalid_otp")
	ErrOTPExpired             = errors.New("otp_expired")
	ErrInvalidCredentials     = errors.New("invalid_credentials")
	ErrPasswordNotSet         = errors.New("password_not_set")
	ErrInvalidRefreshToken    = errors.New("invalid_refresh_token")
	ErrUserNotFound           = errors.New("user_not_found")
	ErrUserInactive           = errors.New("user_inactive")
	ErrWebAccessDenied        = errors.New("web_access_denied")
	ErrPhoneAlreadyExists     = errors.New("phone_already_exists")
	ErrRoleNotFound           = errors.New("role_not_found")
	ErrRoleNameExists         = errors.New("role_name_exists")
	ErrInvalidRoleScopeType   = errors.New("invalid_role_scope_type")
	ErrRoleScopeRequired      = errors.New("role_scope_required")
	ErrRoleScopeConflict      = errors.New("role_scope_conflict")
	ErrCannotDeleteSystemRole = errors.New("cannot_delete_system_role")
	ErrCannotModifySystemRole = errors.New("cannot_modify_system_role")
)
