package domain

import "time"

type WorkHoursConfig struct {
	ID                int64
	WorkDayStart      string
	WorkDayEnd        string
	LateGraceMinutes  int
	EarlyGraceMinutes int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewDefaultWorkHoursConfig() *WorkHoursConfig {
	return &WorkHoursConfig{
		ID:                1,
		WorkDayStart:      "09:00",
		WorkDayEnd:        "17:00",
		LateGraceMinutes:  15,
		EarlyGraceMinutes: 15,
	}
}
