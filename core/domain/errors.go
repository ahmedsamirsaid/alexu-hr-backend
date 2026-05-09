package domain

// LocalizedError is an error that carries a translation key and optional parameters
// for i18n-aware error message formatting.
//
// It implements error.Is() by comparing Code fields, so sentinel variables work
// correctly with errors.Is() even when params differ.
type LocalizedError struct {
	// Code is the machine-readable error code (also returned by Error()).
	Code string
	// Key is the i18n translation key, e.g. "error.leave.insufficient_balance".
	Key string
	// Params holds template parameters for parameterized translation strings.
	Params map[string]interface{}
}

// Error implements the error interface; returns the machine-readable code.
func (e *LocalizedError) Error() string {
	return e.Code
}

// Is allows errors.Is() to match by Code, enabling sentinel comparisons.
func (e *LocalizedError) Is(target error) bool {
	t, ok := target.(*LocalizedError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewLocalizedError creates a LocalizedError with the given code, translation key, and params.
func NewLocalizedError(code, key string, params map[string]interface{}) *LocalizedError {
	return &LocalizedError{Code: code, Key: key, Params: params}
}

// ─── Common localized error sentinels ───────────────────────────────────────

var (
	// Employee errors
	ErrEmployeeNotFound = NewLocalizedError("employee_not_found", "error.employee.not_found", nil)
	ErrEmployeeAlreadyExists = NewLocalizedError("employee_already_exists", "error.employee.already_exists", nil)

	// Leave errors
	ErrInsufficientBalance    = NewLocalizedError("insufficient_leave_balance", "error.leave.insufficient_balance", nil)
	ErrInvalidDateRange       = NewLocalizedError("invalid_date_range", "error.leave.invalid_date_range", nil)
	ErrExceedsConsecutiveDays = NewLocalizedError("exceeds_consecutive_days", "error.leave.exceeds_consecutive_days", nil)
	ErrRecordingDeadlinePassed = NewLocalizedError("recording_deadline_passed", "error.leave.recording_deadline_passed", nil)
	ErrLeaveNotFound          = NewLocalizedError("leave_not_found", "error.leave.not_found", nil)
	ErrLeaveOverlapping       = NewLocalizedError("leave_overlapping", "error.leave.overlapping_dates", nil)

	// Auth errors
	ErrInvalidCredentials = NewLocalizedError("invalid_credentials", "error.auth.invalid_credentials", nil)
	ErrTokenExpired       = NewLocalizedError("token_expired", "error.auth.token_expired", nil)
	ErrUnauthorized       = NewLocalizedError("unauthorized", "error.auth.unauthorized", nil)
	ErrInvalidOTP         = NewLocalizedError("invalid_otp", "error.auth.invalid_otp", nil)
	ErrOTPExpired         = NewLocalizedError("otp_expired", "error.auth.otp_expired", nil)

	// Approval errors
	ErrApprovalNotFound     = NewLocalizedError("approval_not_found", "error.approval.not_found", nil)
	ErrNotAuthorizedApprover = NewLocalizedError("not_authorized_approver", "error.approval.not_authorized", nil)

	// Department errors
	ErrDepartmentNotFound = NewLocalizedError("department_not_found", "error.department.not_found", nil)
)
