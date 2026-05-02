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
	Events []*domain.AuditLog
}

// GetAuditTrailUseCase handles retrieving the complete audit history for an entity.
type GetAuditTrailUseCase struct {
	db              ports.DB
	auditLogRepo    ports.AuditLogRepository
	employeeRepo    ports.EmployeeRepository
}

// NewGetAuditTrailUseCase creates a new get audit trail use case.
func NewGetAuditTrailUseCase(
	db ports.DB,
	auditLogRepo ports.AuditLogRepository,
	employeeRepo ports.EmployeeRepository,
) *GetAuditTrailUseCase {
	return &GetAuditTrailUseCase{
		db:           db,
		auditLogRepo: auditLogRepo,
		employeeRepo: employeeRepo,
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

	return &GetAuditTrailOutput{
		Events: events,
	}, nil
}
