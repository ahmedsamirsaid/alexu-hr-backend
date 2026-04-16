package domain

import "time"

type AttendancePunchType string

const (
	AttendancePunchTypeCheckIn    AttendancePunchType = "check_in"
	AttendancePunchTypeCheckOut   AttendancePunchType = "check_out"
	AttendancePunchTypeBreakStart AttendancePunchType = "break_start"
	AttendancePunchTypeBreakEnd   AttendancePunchType = "break_end"
	AttendancePunchTypeUnknown    AttendancePunchType = "unknown"
)

type AttendanceRecord struct {
	ID           int64
	UID          string
	EmployeeUID  string
	DeviceUID    string
	DeviceUserID string
	PunchedAt    time.Time
	PunchType    AttendancePunchType
	RawPayload   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewAttendanceRecord(employeeUID, deviceUID, deviceUserID string, punchedAt time.Time, punchType AttendancePunchType, rawPayload *string) *AttendanceRecord {
	return &AttendanceRecord{
		UID:          GenerateUID("atr"),
		EmployeeUID:  employeeUID,
		DeviceUID:    deviceUID,
		DeviceUserID: deviceUserID,
		PunchedAt:    punchedAt,
		PunchType:    punchType,
		RawPayload:   rawPayload,
	}
}
