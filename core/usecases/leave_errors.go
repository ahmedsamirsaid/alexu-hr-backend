package usecases

import "github.com/banumusa/backend/core/domain"

// Re-export domain-level localized errors so handlers can use errors.Is().
// All usecases.ErrXxx variables now carry translation keys for i18n-aware
// error responses while remaining comparable via errors.Is().

var (
	ErrEmployeeNotFound                     = domain.ErrEmployeeNotFound
	ErrLeaveTypeNotFound                    = domain.NewLocalizedError("leave_type_not_found", "error.leave.not_found", nil)
	ErrSubLeaveTypeNotFound                 = domain.NewLocalizedError("sub_leave_type_not_found", "error.leave.not_found", nil)
	ErrSubLeaveTypeDoesNotBelongToLeaveType = domain.NewLocalizedError("sub_leave_type_mismatch", "error.leave.not_found", nil)
	ErrInsufficientBalance                  = domain.ErrInsufficientBalance
	ErrExceedsConsecutiveDays               = domain.ErrExceedsConsecutiveDays
	ErrRecordingDeadlinePassed              = domain.ErrRecordingDeadlinePassed
	ErrInvalidDateRange                     = domain.ErrInvalidDateRange
	ErrNoWorkingDays                        = domain.NewLocalizedError("no_working_days", "error.leave.invalid_date_range", nil)
	ErrOverlappingRequest                   = domain.ErrLeaveOverlapping
)
