package domain

type AttendanceExceptionType string

const (
	AttendanceExceptionTypeMissedPunchIn  AttendanceExceptionType = "missed_punch_in"
	AttendanceExceptionTypeMissedPunchOut AttendanceExceptionType = "missed_punch_out"
	AttendanceExceptionTypeLateArrival    AttendanceExceptionType = "late_arrival"
	AttendanceExceptionTypeEarlyDeparture AttendanceExceptionType = "early_departure"
	AttendanceExceptionTypeAbsence        AttendanceExceptionType = "absence"
)
