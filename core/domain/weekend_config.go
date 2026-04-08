package domain

import "time"

type WeekendConfig struct {
	ID        int64
	UID       string
	DayOfWeek int // 0=Sunday, 1=Monday, ..., 6=Saturday
	CreatedAt time.Time
	UpdatedAt time.Time
}
