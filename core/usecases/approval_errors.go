package usecases

import "errors"

var (
	// Approval Flow errors
	ErrApprovalFlowNotFound      = errors.New("approval flow not found")
	ErrApprovalFlowCodeExists    = errors.New("approval flow with this code already exists")
	ErrApprovalFlowInactive      = errors.New("approval flow is inactive")
	ErrApprovalFlowHasNoSteps    = errors.New("approval flow has no steps configured")

	// Approval Flow Step errors
	ErrApprovalFlowStepNotFound    = errors.New("approval flow step not found")
	ErrDuplicateStepOrder          = errors.New("step with this order already exists")
	ErrStepHasPendingRequests      = errors.New("cannot delete step with pending requests")

	// Leave Request errors
	ErrLeaveRequestNotFound        = errors.New("leave request not found")
	ErrLeaveTypeNoApprovalRequired = errors.New("leave type does not require approval")
	ErrNoDepartmentAssigned        = errors.New("employee has no department assigned")
	ErrOverlappingRequest          = errors.New("overlapping leave request exists")

	// Approval Request errors
	ErrApprovalRequestNotFound  = errors.New("approval request not found")
	ErrRequestNotPending        = errors.New("request is not pending")
	ErrNotAuthorizedApprover    = errors.New("not authorized to approve this request")
	ErrCannotCancelApproved     = errors.New("cannot cancel approved request")
	ErrNotRequestOwner          = errors.New("not the owner of this request")
)
