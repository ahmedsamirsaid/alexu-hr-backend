package domain

import "time"

type AttendanceEditHistory struct {
	ID                  int64
	UID                 string
	AttendanceRecordUID string
	FieldChanged        string
	OldValue            *string
	NewValue            *string
	Reason              *string
	EditedByUID         string
	CreatedAt           time.Time
}

func NewAttendanceEditHistory(attendanceRecordUID, fieldChanged, editedByUID string, oldValue, newValue, reason *string) *AttendanceEditHistory {
	return &AttendanceEditHistory{
		UID:                 GenerateUID("aeh"),
		AttendanceRecordUID: attendanceRecordUID,
		FieldChanged:        fieldChanged,
		OldValue:            oldValue,
		NewValue:            newValue,
		Reason:              reason,
		EditedByUID:         editedByUID,
	}
}
