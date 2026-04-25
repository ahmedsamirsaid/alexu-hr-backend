package domain

import "time"

type SubLeaveType struct {
	ID           int64
	UID          string
	LeaveTypeUID string
	NameEN       string
	NameAR       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
