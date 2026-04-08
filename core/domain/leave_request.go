package domain

import "time"

type LeaveRequest struct {
	ID                 int64
	UID                string
	EmployeeUID        string
	LeaveTypeUID       string
	StartDate          time.Time
	EndDate            time.Time
	Days               int
	Notes              *string
	SubmittedAt        time.Time
	DecidedAt          *time.Time
	ApprovalRequestUID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewLeaveRequest(employeeUID, leaveTypeUID string, startDate, endDate time.Time, days int, notes *string, approvalRequestUID string) *LeaveRequest {
	return &LeaveRequest{
		UID:                GenerateUID("lrq"),
		EmployeeUID:        employeeUID,
		LeaveTypeUID:       leaveTypeUID,
		StartDate:          startDate,
		EndDate:            endDate,
		Days:               days,
		Notes:              notes,
		SubmittedAt:        time.Now(),
		ApprovalRequestUID: approvalRequestUID,
	}
}

func (lr *LeaveRequest) SetDecided() {
	now := time.Now()
	lr.DecidedAt = &now
}
