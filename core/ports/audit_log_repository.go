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

	// ListAll returns paginated audit entries with optional filters, newest first.
	ListAll(ctx context.Context, q Querier, filters AuditLogFilters, p ListParams) ([]*domain.AuditLog, error)
}

// AuditLogFilters contains optional filters for listing audit logs.
type AuditLogFilters struct {
	EntityType string // Filter by entity type (e.g., "employee", "leave_request")
	ActorUID   string // Filter by actor UID
	SearchText string // Free text search in entity_uid, action, or meta
}
