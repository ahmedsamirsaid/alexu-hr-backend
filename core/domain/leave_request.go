package domain

import "time"

type LeaveRequest struct {
	ID                 int64
	UID                string
	EmployeeUID        string
	LeaveTypeUID       string
	SubLeaveTypeUID    *string
	OtherSubLeaveName  *string
	StartDate          time.Time
	EndDate            time.Time
	Days               int
	Notes              *string
	StudyDestination   *string
	Assignment         *string
	AssignmentCountry  *string
	SpouseWorkCountry  *string
	SubmittedAt        time.Time
	DecidedAt          *time.Time
	ApprovalRequestUID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewLeaveRequest(
	employeeUID, leaveTypeUID string,
	subLeaveTypeUID *string,
	otherSubLeaveName *string,
	startDate, endDate time.Time,
	days int,
	notes *string,
	studyDestination *string,
	assignment *string,
	assignmentCountry *string,
	spouseWorkCountry *string,
	approvalRequestUID string,
) *LeaveRequest {
	return &LeaveRequest{
		UID:                GenerateUID("lrq"),
		EmployeeUID:        employeeUID,
		LeaveTypeUID:       leaveTypeUID,
		SubLeaveTypeUID:    subLeaveTypeUID,
		OtherSubLeaveName:  otherSubLeaveName,
		StartDate:          startDate,
		EndDate:            endDate,
		Days:               days,
		Notes:              notes,
		StudyDestination:   studyDestination,
		Assignment:         assignment,
		AssignmentCountry:  assignmentCountry,
		SpouseWorkCountry:  spouseWorkCountry,
		SubmittedAt:        time.Now(),
		ApprovalRequestUID: approvalRequestUID,
	}
}

func (lr *LeaveRequest) SetDecided() {
	now := time.Now()
	lr.DecidedAt = &now
}
