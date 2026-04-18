package usecases

import "errors"

var (
	ErrHolidayDateAlreadyExists  = errors.New("holiday_date_already_exists")
	ErrHolidayDateFallsOnWeekend = errors.New("holiday_date_falls_on_weekend")
)
