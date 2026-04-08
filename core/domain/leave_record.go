package domain

import "time"

type LeaveRecord struct {
	ID              int64
	UID             string
	EmployeeID      int64
	LeaveTypeID     int64
	StartDate       time.Time
	EndDate         time.Time
	Days            int
	RecordedAt      time.Time
	RecordedBy      *int64
	Notes           *string
	LeaveRequestUID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewLeaveRecord(employeeID, leaveTypeID int64, startDate, endDate time.Time, days int, recordedBy *int64, notes *string, leaveRequestUID *string) *LeaveRecord {
	return &LeaveRecord{
		UID:             GenerateUID("leave"),
		EmployeeID:      employeeID,
		LeaveTypeID:     leaveTypeID,
		StartDate:       startDate,
		EndDate:         endDate,
		Days:            days,
		RecordedAt:      time.Now(),
		RecordedBy:      recordedBy,
		Notes:           notes,
		LeaveRequestUID: leaveRequestUID,
	}
}
