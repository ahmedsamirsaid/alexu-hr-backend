package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

// ListAllAuditLogsUseCase retrieves all audit logs with optional filters.
type ListAllAuditLogsUseCase struct {
	repo ports.AuditLogRepository
	db   ports.DB
	i18n ports.I18nService
}

// NewListAllAuditLogsUseCase creates a new use case instance.
func NewListAllAuditLogsUseCase(repo ports.AuditLogRepository, db ports.DB, i18n ports.I18nService) *ListAllAuditLogsUseCase {
	return &ListAllAuditLogsUseCase{
		repo: repo,
		db:   db,
		i18n: i18n,
	}
}

// ListAllAuditLogsInput contains the input parameters.
type ListAllAuditLogsInput struct {
	EntityType string // Optional: filter by entity type
	ActorUID   string // Optional: filter by actor UID
	SearchText string // Optional: free text search
	Page       int
	PageSize   int
}

// ListAllAuditLogsOutput contains the result.
type ListAllAuditLogsOutput struct {
	Events      []LocalizedAuditEventDTO
	Page        int
	PageSize    int
	HasNextPage bool
}

// AuditEventDTO represents an audit event in the output.
type AuditEventDTO struct {
	UID        string
	ActorUID   string
	Action     string
	EntityType string
	EntityUID  string
	Meta       *string
	OccurredAt string
	CreatedAt  string
}

// Execute runs the use case.
func (uc *ListAllAuditLogsUseCase) Execute(ctx context.Context, input ListAllAuditLogsInput) (*ListAllAuditLogsOutput, error) {
	filters := ports.AuditLogFilters{
		EntityType: input.EntityType,
		ActorUID:   input.ActorUID,
		SearchText: input.SearchText,
	}

	params := ports.ListParams{
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	logs, err := uc.repo.ListAll(ctx, uc.db, filters, params)
	if err != nil {
		return nil, err
	}

	// Convert to localized DTOs
	events := make([]LocalizedAuditEventDTO, 0, len(logs))
	for _, log := range logs {
		events = append(events, LocalizedAuditEventDTO{
			UID:             log.UID,
			ActorUID:        log.ActorUID,
			Action:          log.Action,
			ActionLabel:     uc.i18n.T(ctx, "audit.action."+log.Action),
			EntityType:      log.EntityType,
			EntityTypeLabel: uc.i18n.T(ctx, "audit.entity."+log.EntityType),
			EntityUID:       log.EntityUID,
			Meta:            log.Meta,
			OccurredAt:      log.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:       log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Check if there's a next page by fetching one more record
	hasNextPage := len(logs) == input.PageSize

	return &ListAllAuditLogsOutput{
		Events:      events,
		Page:        input.Page,
		PageSize:    input.PageSize,
		HasNextPage: hasNextPage,
	}, nil
}
