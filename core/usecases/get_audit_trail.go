package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// GetAuditTrailInput defines the input for retrieving audit trail for an entity.
type GetAuditTrailInput struct {
	EntityType string
	EntityUID  string
}

// GetAuditTrailOutput contains the audit trail for an entity.
type GetAuditTrailOutput struct {
	Events []LocalizedAuditEventDTO
}

// LocalizedAuditEventDTO represents an audit event with localized labels.
type LocalizedAuditEventDTO struct {
	UID             string  `json:"uid"`
	ActorUID        string  `json:"actorUid"`
	Action          string  `json:"action"`          // Raw key
	ActionLabel     string  `json:"actionLabel"`     // Translated
	EntityType      string  `json:"entityType"`      // Raw key
	EntityTypeLabel string  `json:"entityTypeLabel"` // Translated
	EntityUID       string  `json:"entityUid"`
	Meta            *string `json:"meta"`
	OccurredAt      string  `json:"occurredAt"`
	CreatedAt       string  `json:"createdAt"`
}

// GetAuditTrailUseCase handles retrieving the complete audit history for an entity.
type GetAuditTrailUseCase struct {
	db           ports.DB
	auditLogRepo ports.AuditLogRepository
	employeeRepo ports.EmployeeRepository
	i18n         ports.I18nService
}

// NewGetAuditTrailUseCase creates a new get audit trail use case.
func NewGetAuditTrailUseCase(
	db ports.DB,
	auditLogRepo ports.AuditLogRepository,
	employeeRepo ports.EmployeeRepository,
	i18n ports.I18nService,
) *GetAuditTrailUseCase {
	return &GetAuditTrailUseCase{
		db:           db,
		auditLogRepo: auditLogRepo,
		employeeRepo: employeeRepo,
		i18n:         i18n,
	}
}

// Execute retrieves the audit trail for an entity.
// Returns all audit events ordered by occurred_at descending (newest first).
// Returns an empty list if no audit events exist for the entity.
func (uc *GetAuditTrailUseCase) Execute(ctx context.Context, input GetAuditTrailInput) (*GetAuditTrailOutput, error) {
	events, err := uc.auditLogRepo.ListByEntity(ctx, uc.db, input.EntityType, input.EntityUID)
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

	return &GetAuditTrailOutput{
		Events: localizedEvents,
	}, nil
}
