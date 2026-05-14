package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type RateLimitConfig struct {
	OTPRequestLimit  int
	OTPRequestWindow time.Duration
	OTPVerifyLimit   int
	OTPVerifyWindow  time.Duration
	OTPRequestLock   time.Duration
	OTPVerifyLock    time.Duration
	LoginLock        time.Duration
	LoginLimit       int
	LoginWindow      time.Duration
}

type RateLimitResult struct {
	Allowed    bool
	RetryAfter time.Duration
	Locked     bool
}

type RateLimitService struct {
	db     ports.DB
	repo   ports.RateLimitRepository
	config RateLimitConfig
	now    func() time.Time
}

func NewRateLimitService(db ports.DB, repo ports.RateLimitRepository, config RateLimitConfig) *RateLimitService {
	return &RateLimitService{
		db:     db,
		repo:   repo,
		config: config,
		now:    time.Now,
	}
}

func NormalizePhoneForRateLimit(phone string) string {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for i, r := range trimmed {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		if r == '+' && i == 0 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (s *RateLimitService) AllowOTPRequest(ctx context.Context, subject string) (*RateLimitResult, error) {
	return s.incrementWindow(ctx, domain.RateLimitScopeOTPRequest, subject, s.config.OTPRequestLimit, s.config.OTPRequestWindow, s.config.OTPRequestLock)
}

func (s *RateLimitService) CheckOTPVerify(ctx context.Context, subject string) (*RateLimitResult, error) {
	return s.checkWindow(ctx, domain.RateLimitScopeOTPVerify, subject, s.config.OTPVerifyWindow)
}

func (s *RateLimitService) RegisterOTPVerifyFailure(ctx context.Context, subject string) (*RateLimitResult, error) {
	return s.incrementWindow(ctx, domain.RateLimitScopeOTPVerify, subject, s.config.OTPVerifyLimit, s.config.OTPVerifyWindow, s.config.OTPVerifyLock)
}

func (s *RateLimitService) ResetOTPVerify(ctx context.Context, subject string) error {
	return s.resetScope(ctx, domain.RateLimitScopeOTPVerify, subject)
}

func (s *RateLimitService) CheckLogin(ctx context.Context, subject string) (*RateLimitResult, error) {
	return s.checkWindow(ctx, domain.RateLimitScopeLogin, subject, s.config.LoginWindow)
}

func (s *RateLimitService) RegisterLoginFailure(ctx context.Context, subject string) (*RateLimitResult, error) {
	return s.incrementWindow(ctx, domain.RateLimitScopeLogin, subject, s.config.LoginLimit, s.config.LoginWindow, s.config.LoginLock)
}

func (s *RateLimitService) ResetLogin(ctx context.Context, subject string) error {
	return s.resetScope(ctx, domain.RateLimitScopeLogin, subject)
}

func (s *RateLimitService) resetScope(ctx context.Context, scope, subject string) error {
	subjectKey := NormalizeRateLimitSubject(subject)
	if subjectKey == "" {
		return nil
	}
	return s.repo.DeleteByScopeAndSubject(ctx, s.db, scope, subjectKey)
}

func (s *RateLimitService) checkWindow(ctx context.Context, scope, subject string, window time.Duration) (*RateLimitResult, error) {
	subjectKey := NormalizeRateLimitSubject(subject)
	if subjectKey == "" {
		return &RateLimitResult{Allowed: true}, nil
	}

	record, err := s.repo.GetByScopeAndSubject(ctx, s.db, scope, subjectKey)
	if err != nil {
		return nil, err
	}
	return s.evaluateRecord(record, window), nil
}

func (s *RateLimitService) evaluateRecord(record *domain.RateLimitRecord, window time.Duration) *RateLimitResult {
	now := s.now()
	if record == nil {
		return &RateLimitResult{Allowed: true}
	}

	if record.LockedUntil != nil && record.LockedUntil.After(now) {
		return &RateLimitResult{
			Allowed:    false,
			Locked:     true,
			RetryAfter: record.LockedUntil.Sub(now),
		}
	}

	windowEnd := record.WindowStartedAt.Add(window)
	if !windowEnd.After(now) {
		return &RateLimitResult{Allowed: true}
	}

	return &RateLimitResult{Allowed: true}
}

func (s *RateLimitService) incrementWindow(ctx context.Context, scope, subject string, limit int, window, lockDuration time.Duration) (*RateLimitResult, error) {
	subjectKey := NormalizeRateLimitSubject(subject)
	if subjectKey == "" {
		return &RateLimitResult{Allowed: true}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	record, err := s.repo.GetByScopeAndSubject(ctx, tx, scope, subjectKey)
	if err != nil {
		return nil, err
	}

	now := s.now()
	if record == nil {
		record = &domain.RateLimitRecord{
			Scope:           scope,
			SubjectKey:      subjectKey,
			WindowStartedAt: now,
			AttemptCount:    1,
			UpdatedAt:       now,
		}
		if err := s.repo.Upsert(ctx, tx, record); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &RateLimitResult{Allowed: true}, nil
	}

	if record.LockedUntil != nil && record.LockedUntil.After(now) {
		return &RateLimitResult{
			Allowed:    false,
			Locked:     true,
			RetryAfter: record.LockedUntil.Sub(now),
		}, nil
	}

	if !record.WindowStartedAt.Add(window).After(now) {
		record.WindowStartedAt = now
		record.AttemptCount = 0
		record.LockedUntil = nil
	}

	record.AttemptCount++
	record.UpdatedAt = now

	if record.AttemptCount > limit {
		retryAfter := record.WindowStartedAt.Add(window).Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		if lockDuration > 0 {
			lockedUntil := now.Add(lockDuration)
			record.LockedUntil = &lockedUntil
			retryAfter = lockDuration
		}
		if err := s.repo.Upsert(ctx, tx, record); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &RateLimitResult{
			Allowed:    false,
			Locked:     lockDuration > 0,
			RetryAfter: retryAfter,
		}, nil
	}

	if err := s.repo.Upsert(ctx, tx, record); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &RateLimitResult{Allowed: true}, nil
}

func IsCredentialRateLimitError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrPasswordNotSet)
}

func NormalizeRateLimitSubject(subject string) string {
	return strings.TrimSpace(subject)
}
