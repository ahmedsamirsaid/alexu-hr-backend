package usecases

import "errors"

var (
	ErrEmployeeProfileChangeRequestNotFound = errors.New("employee profile change request not found")
	ErrEmployeeProfileChangeAlreadyPending  = errors.New("employee profile change request already pending")
	ErrNoProfileChangesRequested            = errors.New("no employee profile changes requested")
)
