package ports

import (
	"context"
	"time"

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

	// CountAll returns the total count of audit logs matching the filters.
	CountAll(ctx context.Context, q Querier, filters AuditLogFilters) (int, error)

	// GetDistinctActionTypes returns all unique action types in the audit log.
	GetDistinctActionTypes(ctx context.Context, q Querier) ([]string, error)

	// GetDistinctEntityTypes returns all unique entity types in the audit log.
	GetDistinctEntityTypes(ctx context.Context, q Querier) ([]string, error)
}

// AuditLogFilters contains optional filters for listing audit logs.
type AuditLogFilters struct {
	EntityType        string     // Filter by entity type (e.g., "employee", "leave_request")
	EntityTypes       []string   // Filter by multiple entity types
	ActorUID          string     // Filter by actor UID
	ActorName         string     // Filter by actor name (requires JOIN)
	ActionTypes       []string   // Filter by multiple action types
	SearchText        string     // Free text search in entity_uid, action, or meta
	StartDate         *time.Time // Filter by date range start
	EndDate           *time.Time // Filter by date range end
	HideSystemActions bool       // Exclude system-generated actions
}
