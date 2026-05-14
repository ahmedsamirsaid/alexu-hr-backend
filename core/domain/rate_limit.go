package domain

import "time"

const (
	RateLimitScopeOTPRequest = "otp_request"
	RateLimitScopeOTPVerify  = "otp_verify"
	RateLimitScopeLogin      = "login"
)

type RateLimitRecord struct {
	ID              int64
	Scope           string
	SubjectKey      string
	WindowStartedAt time.Time
	AttemptCount    int
	LockedUntil     *time.Time
	UpdatedAt       time.Time
}
