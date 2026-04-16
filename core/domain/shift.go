package domain

import "time"

type Shift struct {
	ID           int64
	UID          string
	StartTime    string
	EndTime      string
	GraceMinutes int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewDefaultShift() *Shift {
	return &Shift{
		ID:           1,
		UID:          "shf_general",
		StartTime:    "09:00",
		EndTime:      "17:00",
		GraceMinutes: 15,
	}
}
