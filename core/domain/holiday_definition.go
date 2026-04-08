package domain

import "time"

type HolidayDefinition struct {
	ID           int64
	UID          string
	Code         string
	NameEN       string
	NameAR       string
	DefaultMonth *int
	DefaultDay   *int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (hd *HolidayDefinition) IsFixed() bool {
	return hd.DefaultMonth != nil && hd.DefaultDay != nil
}
