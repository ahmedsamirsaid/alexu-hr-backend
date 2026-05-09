package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// GetActorAuditEventsInput defines the input for retrieving audit events by actor.
type GetActorAuditEventsInput struct {
	ActorUID string
	Page     int
	PageSize int
}

// GetActorAuditEventsOutput contains the paginated audit events for an actor.
type GetActorAuditEventsOutput struct {
	Events      []LocalizedAuditEventDTO
	Page        int
	PageSize    int
	TotalPages  int
	HasNextPage bool
}

// GetActorAuditEventsUseCase handles retrieving audit events for a specific actor with pagination.
type GetActorAuditEventsUseCase struct {
	db           ports.DB
	auditLogRepo ports.AuditLogRepository
	i18n         ports.I18nService
}

// NewGetActorAuditEventsUseCase creates a new get actor audit events use case.
func NewGetActorAuditEventsUseCase(
	db ports.DB,
	auditLogRepo ports.AuditLogRepository,
	i18n ports.I18nService,
) *GetActorAuditEventsUseCase {
	return &GetActorAuditEventsUseCase{
		db:           db,
		auditLogRepo: auditLogRepo,
		i18n:         i18n,
	}
}

// Execute retrieves paginated audit events for an actor.
// Returns events ordered by occurred_at descending (newest first).
// Returns an empty list if no audit events exist for the actor.
// Validates pagination parameters: page size must be between 1 and 100.
func (uc *GetActorAuditEventsUseCase) Execute(ctx context.Context, input GetActorAuditEventsInput) (*GetActorAuditEventsOutput, error) {
	// Validate pagination parameters
	if input.PageSize < 1 || input.PageSize > 100 {
		return nil, fmt.Errorf("page size must be between 1 and 100, got %d", input.PageSize)
	}
	if input.Page < 1 {
		input.Page = 1
	}

	// Build list params
	listParams := ports.ListParams{
		Page:      input.Page,
		PageSize:  input.PageSize,
		SortBy:    "occurred_at",
		SortOrder: ports.SortOrderDesc,
	}

	// Retrieve events
	events, err := uc.auditLogRepo.ListByActor(ctx, uc.db, input.ActorUID, listParams)
	if err != nil {
		return nil, err
	}

	// Return empty list if no events found (not an error)
	if events == nil {
		events = []*domain.AuditLog{}
	}

	// Convert to localized DTOs
	localizedEvents := make([]LocalizedAuditEventDTO, len(events))
	for i, event := range events {
		localizedEvents[i] = LocalizedAuditEventDTO{
			UID:             event.UID,
			ActorUID:        event.ActorUID,
			Action:          event.Action,
			ActionLabel:     uc.i18n.T(ctx, "audit.action."+event.Action),
			EntityType:      event.EntityType,
			EntityTypeLabel: uc.i18n.T(ctx, "audit.entity."+event.EntityType),
			EntityUID:       event.EntityUID,
			Meta:            event.Meta,
			OccurredAt:      event.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:       event.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Calculate pagination metadata
	hasNextPage := len(events) == input.PageSize

	return &GetActorAuditEventsOutput{
		Events:      localizedEvents,
		Page:        input.Page,
		PageSize:    input.PageSize,
		HasNextPage: hasNextPage,
	}, nil
}
