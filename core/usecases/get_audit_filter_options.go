package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/banumusa/backend/core/ports"
)

// FilterOption represents a filter option with key and localized label.
type FilterOption struct {
	Value string `json:"value"` // Technical value (e.g., "approve")
	Label string `json:"label"` // Localized label based on request locale
}

// GetAuditFilterOptionsInput defines the input for getting filter options.
type GetAuditFilterOptionsInput struct {
	Language string // "en" or "ar" (optional, for future use)
}

// GetAuditFilterOptionsOutput contains the available filter options.
type GetAuditFilterOptionsOutput struct {
	ActionTypes []FilterOption `json:"actionTypes"`
	EntityTypes []FilterOption `json:"entityTypes"`
}

// GetAuditFilterOptionsUseCase handles retrieving available filter options for audit logs.
type GetAuditFilterOptionsUseCase struct {
	auditRepo ports.AuditLogRepository
	db        ports.DB
	i18n      ports.I18nService
}

// NewGetAuditFilterOptionsUseCase creates a new filter options use case.
func NewGetAuditFilterOptionsUseCase(
	auditRepo ports.AuditLogRepository,
	db ports.DB,
	i18n ports.I18nService,
) *GetAuditFilterOptionsUseCase {
	return &GetAuditFilterOptionsUseCase{
		auditRepo: auditRepo,
		db:        db,
		i18n:      i18n,
	}
}

// Execute retrieves available filter options from the audit log.
func (uc *GetAuditFilterOptionsUseCase) Execute(ctx context.Context, input GetAuditFilterOptionsInput) (*GetAuditFilterOptionsOutput, error) {
	// Get distinct action types
	actionTypes, err := uc.auditRepo.GetDistinctActionTypes(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	// Get distinct entity types
	entityTypes, err := uc.getDistinctEntityTypesExcludingSystemActions(ctx)
	if err != nil {
		return nil, err
	}

	// Map action types to filter options with localized labels
	actionOptions := make([]FilterOption, 0, len(actionTypes))
	for _, action := range actionTypes {
		if isSystemAuditAction(action) {
			continue
		}

		actionOptions = append(actionOptions, FilterOption{
			Value: action,
			Label: uc.i18n.T(ctx, "audit.action."+action),
		})
	}

	// Map entity types to filter options with localized labels
	entityOptions := make([]FilterOption, 0, len(entityTypes))
	for _, entityType := range entityTypes {
		if entityType == "system" {
			continue
		}

		entityOptions = append(entityOptions, FilterOption{
			Value: entityType,
			Label: uc.i18n.T(ctx, "audit.entity."+entityType),
		})
	}

	return &GetAuditFilterOptionsOutput{
		ActionTypes: actionOptions,
		EntityTypes: entityOptions,
	}, nil
}

func (uc *GetAuditFilterOptionsUseCase) getDistinctEntityTypesExcludingSystemActions(ctx context.Context) ([]string, error) {
	systemActions := systemAuditActionList()
	placeholders := make([]string, 0, len(systemActions))
	for i := range systemActions {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}
	query := fmt.Sprintf(
		"SELECT DISTINCT entity_type FROM audit_logs WHERE action NOT IN (%s) ORDER BY entity_type ASC",
		strings.Join(placeholders, ","),
	)

	args := make([]any, 0, len(systemActions))
	for _, action := range systemActions {
		args = append(args, action)
	}

	rows, err := uc.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entityTypes := make([]string, 0)
	for rows.Next() {
		var entityType string
		if err := rows.Scan(&entityType); err != nil {
			return nil, err
		}
		entityTypes = append(entityTypes, entityType)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entityTypes, nil
}

func systemAuditActionList() []string {
	return []string{
		"deduct_balance",
		"annual_reset",
		"auto_reject",
		"system_adjustment",
		"cascade_delete",
		"auto_create_transaction",
	}
}

func isSystemAuditAction(action string) bool {
	systemActions := make(map[string]bool)
	for _, actionName := range systemAuditActionList() {
		systemActions[actionName] = true
	}

	return systemActions[action]
}
