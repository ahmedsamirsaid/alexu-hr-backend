package usecases

import (
	"fmt"
	"time"
)

// parseDateOrTimestamp parses a date string in either YYYY-MM-DD or RFC3339 format.
// This handles both date-only columns and timestamp columns returned from the database.
func parseDateOrTimestamp(value string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date or timestamp format: %q", value)
}
