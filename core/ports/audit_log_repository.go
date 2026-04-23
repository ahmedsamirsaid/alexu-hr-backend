package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

// AuditLogRepository persists and queries audit log entries.
type AuditLogRepository interface {
	// Create inserts a new audit log entry.
	Create(ctx context.Context, q Querier, log *domain.AuditLog) error

	// ListByEntity returns all audit entries for a given entity, newest first.
	ListByEntity(ctx context.Context, q Querier, entityType, entityUID string) ([]*domain.AuditLog, error)

	// ListByActor returns paginated audit entries for a given actor, newest first.
	ListByActor(ctx context.Context, q Querier, actorUID string, p ListParams) ([]*domain.AuditLog, error)
}
