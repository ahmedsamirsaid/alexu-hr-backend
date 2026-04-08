package domain

import "time"

type Permission struct {
	ID          int64
	UID         string
	Code        string
	Description string
	CreatedAt   time.Time
}
