package usecases

import "errors"

var (
	ErrPermissionRequestNotFound       = errors.New("permission request not found")
	ErrPermissionTypeInvalid           = errors.New("invalid permission type")
	ErrPermissionDeadlineMissed        = errors.New("permission request submission deadline has passed")
	ErrPermissionWindowRequired        = errors.New("start time and end time are required for this permission type")
	ErrPermissionWindowNotAllowed      = errors.New("start time and end time are not allowed for this permission type")
	ErrPermissionWindowOutsideShift    = errors.New("permission window must be inside the employee's shift")
	ErrPermissionWindowInvalid         = errors.New("permission end time must be after start time")
	ErrPermissionOverlap               = errors.New("permission overlaps with another permission on the same day")
	ErrPermissionWeeklyCapExceeded     = errors.New("weekly cap reached for this permission category")
	ErrPermissionDateInPast            = errors.New("permission date cannot be in the past")
	ErrPermissionNotEditable           = errors.New("permission can only be edited while pending")
	ErrPermissionNotCancellable        = errors.New("permission cannot be cancelled in its current state")
	ErrPermissionAlreadyExpired        = errors.New("permission deadline has expired")
)
