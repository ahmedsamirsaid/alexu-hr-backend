package usecases

import "errors"

var (
	ErrInvalidWorkDayStart    = errors.New("invalid work day start time, expected HH:MM")
	ErrInvalidWorkDayEnd      = errors.New("invalid work day end time, expected HH:MM")
	ErrInvalidWorkHoursRange  = errors.New("work day end must be after work day start")
	ErrInvalidGraceMinutes    = errors.New("grace minutes cannot be negative")
	ErrInvalidReportDateRange = errors.New("attendance report date range is invalid")
)
