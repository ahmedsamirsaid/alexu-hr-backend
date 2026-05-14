package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type RateLimitRepository interface {
	GetByScopeAndSubject(ctx context.Context, q Querier, scope, subjectKey string) (*domain.RateLimitRecord, error)
	Upsert(ctx context.Context, q Querier, record *domain.RateLimitRecord) error
	DeleteByScopeAndSubject(ctx context.Context, q Querier, scope, subjectKey string) error
}
