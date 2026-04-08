package domain

import "time"

type HolidayInstance struct {
	ID           int64
	UID          string
	DefinitionID int64
	Year         int
	ActualDate   time.Time
	ObservedDate time.Time
	IsConfirmed  bool
	Notes        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
