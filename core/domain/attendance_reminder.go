package domain

import "time"

type AttendanceReminderType string

const (
	AttendanceReminderTypeMissingCheckIn  AttendanceReminderType = "missing_check_in"
	AttendanceReminderTypeMissingCheckOut AttendanceReminderType = "missing_check_out"
)

type AttendanceReminder struct {
	ID             int64
	UID            string
	EmployeeUID    string
	AttendanceDate time.Time
	ReminderType   AttendanceReminderType
	SentAt         time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewAttendanceReminder(employeeUID string, attendanceDate time.Time, reminderType AttendanceReminderType, sentAt time.Time) *AttendanceReminder {
	return &AttendanceReminder{
		UID:            GenerateUID("arn"),
		EmployeeUID:    employeeUID,
		AttendanceDate: attendanceDate,
		ReminderType:   reminderType,
		SentAt:         sentAt,
	}
}
