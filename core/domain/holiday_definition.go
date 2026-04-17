package domain

import "time"

type HolidayDefinition struct {
	ID        int64
	UID       string
	Code      string
	NameEN    string
	NameAR    string
	Date      time.Time
	IsManual  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (hd *HolidayDefinition) IsFixed() bool {
	return !hd.Date.IsZero()
}
