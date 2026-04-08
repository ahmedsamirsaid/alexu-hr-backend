package domain

import "time"

type LeaveBalance struct {
	ID          int64
	UID         string
	EmployeeID  int64
	LeaveTypeID int64
	Year        int
	TotalDays   int
	UsedDays    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewLeaveBalance(employeeID, leaveTypeID int64, year, totalDays int) *LeaveBalance {
	return &LeaveBalance{
		UID:         GenerateUID("lbal"),
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
		TotalDays:   totalDays,
		UsedDays:    0,
	}
}

func (lb *LeaveBalance) RemainingDays() int {
	return lb.TotalDays - lb.UsedDays
}

func (lb *LeaveBalance) CanDeduct(days int) bool {
	return lb.RemainingDays() >= days
}
