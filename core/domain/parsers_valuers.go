package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// SQLite stores dates as TEXT strings. These types handle conversion between
// Go's time.Time and SQLite's text representation.

// Time is a wrapper around time.Time that implements sql.Scanner and driver.Valuer
// for proper SQLite datetime handling.
type Time struct {
	time.Time
}

// Supported time formats for parsing SQLite datetime strings
var timeFormats = []string{
	"2006-01-02 15:04:05.999999999 -0700 MST", // Go's time.Time.String() format
	"2006-01-02 15:04:05.999999999 +0000 UTC", // Go's time.Time.String() format (UTC)
	"2006-01-02 15:04:05.999999999-07:00",     // SQLite format with timezone
	"2006-01-02T15:04:05.999999999-07:00",     // RFC3339 with nanoseconds
	"2006-01-02T15:04:05.999999999Z07:00",     // RFC3339 variant
	"2006-01-02T15:04:05Z07:00",               // RFC3339
	"2006-01-02T15:04:05",                     // ISO8601 without timezone
	"2006-01-02 15:04:05",                     // SQLite format without timezone
	"2006-01-02",                              // Date only
	time.RFC3339Nano,
	time.RFC3339,
}

// Scan implements the sql.Scanner interface for reading from database
func (t *Time) Scan(value any) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		return t.parseString(v)
	case []byte:
		return t.parseString(string(v))
	case int64:
		// Unix timestamp
		t.Time = time.Unix(v, 0)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Time: %v", value, value)
	}
}

func (t *Time) parseString(s string) error {
	if s == "" {
		t.Time = time.Time{}
		return nil
	}

	// Strip monotonic clock reading if present (appears as " m=+..." or " m=-...")
	if idx := strings.Index(s, " m="); idx != -1 {
		s = s[:idx]
	}

	var parseErr error
	for _, format := range timeFormats {
		parsed, err := time.Parse(format, s)
		if err == nil {
			t.Time = parsed
			return nil
		}
		parseErr = err
	}
	return fmt.Errorf("cannot parse time string %q: %v", s, parseErr)
}

// Value implements the driver.Valuer interface for writing to database
func (t Time) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	// Use SQLite-compatible format
	return t.Time.Format("2006-01-02 15:04:05.999999999-07:00"), nil
}

// Date is a wrapper for date-only values (no time component)
type Date struct {
	time.Time
}

// Scan implements the sql.Scanner interface for reading from database
func (d *Date) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		return d.parseString(v)
	case []byte:
		return d.parseString(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into Date: %v", value, value)
	}
}

func (d *Date) parseString(s string) error {
	if s == "" {
		d.Time = time.Time{}
		return nil
	}

	// Strip monotonic clock reading if present (appears as " m=+..." or " m=-...")
	if idx := strings.Index(s, " m="); idx != -1 {
		s = s[:idx]
	}

	// Try date-only format first
	parsed, err := time.Parse("2006-01-02", s)
	if err == nil {
		d.Time = parsed
		return nil
	}

	// Fall back to datetime formats
	for _, format := range timeFormats {
		parsed, err := time.Parse(format, s)
		if err == nil {
			d.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("cannot parse date string %q", s)
}

// Value implements the driver.Valuer interface for writing to database
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	// Store as date only
	return d.Time.Format("2006-01-02"), nil
}

// NullTime is like Time but allows NULL values
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface
func (nt *NullTime) Scan(value any) error {
	if value == nil {
		nt.Time = time.Time{}
		nt.Valid = false
		return nil
	}

	t := &Time{}
	if err := t.Scan(value); err != nil {
		return err
	}
	nt.Time = t.Time
	nt.Valid = true
	return nil
}

// Value implements the driver.Valuer interface
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return Time{nt.Time}.Value()
}
