package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/banumusa/backend/core/usecases"
)

// AuditHandler handles audit log HTTP requests.
type AuditHandler struct {
	getTrailUC       *usecases.GetAuditTrailUseCase
	getActorEventsUC *usecases.GetActorAuditEventsUseCase
	listAllLogsUC    *usecases.ListAllAuditLogsUseCase
	listEnrichedUC   *usecases.ListEnrichedAuditLogsUseCase
	filterOptionsUC  *usecases.GetAuditFilterOptionsUseCase
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(
	getTrailUC *usecases.GetAuditTrailUseCase,
	getActorEventsUC *usecases.GetActorAuditEventsUseCase,
	listAllLogsUC *usecases.ListAllAuditLogsUseCase,
	listEnrichedUC *usecases.ListEnrichedAuditLogsUseCase,
	filterOptionsUC *usecases.GetAuditFilterOptionsUseCase,
) *AuditHandler {
	return &AuditHandler{
		getTrailUC:       getTrailUC,
		getActorEventsUC: getActorEventsUC,
		listAllLogsUC:    listAllLogsUC,
		listEnrichedUC:   listEnrichedUC,
		filterOptionsUC:  filterOptionsUC,
	}
}

// AuditEventResponse represents an audit event in the API response.
type AuditEventResponse struct {
	UID        string  `json:"uid"`
	ActorUID   string  `json:"actorUid"`
	Action     string  `json:"action"`
	EntityType string  `json:"entityType"`
	EntityUID  string  `json:"entityUid"`
	Meta       *string `json:"meta,omitempty"`
	OccurredAt string  `json:"occurredAt"`
	CreatedAt  string  `json:"createdAt"`
}

// AuditTrailResponse represents the audit trail API response.
type AuditTrailResponse struct {
	Events []AuditEventResponse `json:"events"`
}

// ActorAuditEventsResponse represents the actor audit events API response.
type ActorAuditEventsResponse struct {
	Events      []AuditEventResponse `json:"events"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"pageSize"`
	HasNextPage bool                 `json:"hasNextPage"`
}

// GetAuditTrail handles GET /api/v1/audit/entity/{entityType}/{entityUID}
func (h *AuditHandler) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("entityType")
	entityUID := r.PathValue("entityUID")

	if entityType == "" {
		writeError(w, http.StatusBadRequest, "entityType is required")
		return
	}
	if entityUID == "" {
		writeError(w, http.StatusBadRequest, "entityUID is required")
		return
	}

	input := usecases.GetAuditTrailInput{
		EntityType: entityType,
		EntityUID:  entityUID,
	}

	output, err := h.getTrailUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("audit_handler.GetAuditTrail.execute_usecase",
			"error", err,
			"entity_type", entityType,
			"entity_uid", entityUID,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve audit trail")
		return
	}

	// Transform domain models to DTOs
	events := make([]AuditEventResponse, 0, len(output.Events))
	for _, event := range output.Events {
		events = append(events, AuditEventResponse{
			UID:        event.UID,
			ActorUID:   event.ActorUID,
			Action:     event.Action,
			EntityType: event.EntityType,
			EntityUID:  event.EntityUID,
			Meta:       event.Meta,
			OccurredAt: event.OccurredAt.Format(time.RFC3339),
			CreatedAt:  event.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, AuditTrailResponse{Events: events})
}

// GetActorAuditEvents handles GET /api/v1/audit/actor/{actorUID}
func (h *AuditHandler) GetActorAuditEvents(w http.ResponseWriter, r *http.Request) {
	actorUID := r.PathValue("actorUID")
	if actorUID == "" {
		writeError(w, http.StatusBadRequest, "actorUID is required")
		return
	}

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			writeError(w, http.StatusBadRequest, "invalid page parameter")
			return
		}
	}

	pageSize := 20 // default page size
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			writeError(w, http.StatusBadRequest, "page_size must be between 1 and 100")
			return
		}
	}

	input := usecases.GetActorAuditEventsInput{
		ActorUID: actorUID,
		Page:     page,
		PageSize: pageSize,
	}

	output, err := h.getActorEventsUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("audit_handler.GetActorAuditEvents.execute_usecase",
			"error", err,
			"actor_uid", actorUID,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve actor audit events")
		return
	}

	// Transform domain models to DTOs
	events := make([]AuditEventResponse, 0, len(output.Events))
	for _, event := range output.Events {
		events = append(events, AuditEventResponse{
			UID:        event.UID,
			ActorUID:   event.ActorUID,
			Action:     event.Action,
			EntityType: event.EntityType,
			EntityUID:  event.EntityUID,
			Meta:       event.Meta,
			OccurredAt: event.OccurredAt.Format(time.RFC3339),
			CreatedAt:  event.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, ActorAuditEventsResponse{
		Events:      events,
		Page:        output.Page,
		PageSize:    output.PageSize,
		HasNextPage: output.HasNextPage,
	})
}

// ListAllAuditLogs handles GET /api/v1/audit/logs
func (h *AuditHandler) ListAllAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	entityType := r.URL.Query().Get("entity_type")
	actorUID := r.URL.Query().Get("actor_uid")
	searchText := r.URL.Query().Get("search")

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			writeError(w, http.StatusBadRequest, "invalid page parameter")
			return
		}
	}

	pageSize := 50 // default page size
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			writeError(w, http.StatusBadRequest, "page_size must be between 1 and 100")
			return
		}
	}

	input := usecases.ListAllAuditLogsInput{
		EntityType: entityType,
		ActorUID:   actorUID,
		SearchText: searchText,
		Page:       page,
		PageSize:   pageSize,
	}

	output, err := h.listAllLogsUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("audit_handler.ListAllAuditLogs.execute_usecase",
			"error", err,
			"entity_type", entityType,
			"actor_uid", actorUID,
			"search_text", searchText,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve audit logs")
		return
	}

	// Transform DTOs to response format
	events := make([]AuditEventResponse, 0, len(output.Events))
	for _, event := range output.Events {
		events = append(events, AuditEventResponse{
			UID:        event.UID,
			ActorUID:   event.ActorUID,
			Action:     event.Action,
			EntityType: event.EntityType,
			EntityUID:  event.EntityUID,
			Meta:       event.Meta,
			OccurredAt: event.OccurredAt,
			CreatedAt:  event.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, ActorAuditEventsResponse{
		Events:      events,
		Page:        output.Page,
		PageSize:    output.PageSize,
		HasNextPage: output.HasNextPage,
	})
}

// GetFilterOptions handles GET /api/v1/audit/filter-options
func (h *AuditHandler) GetFilterOptions(w http.ResponseWriter, r *http.Request) {
	// Parse language query parameter (optional, defaults to "en")
	language := r.URL.Query().Get("lang")
	if language == "" {
		language = "en"
	}

	input := usecases.GetAuditFilterOptionsInput{
		Language: language,
	}

	output, err := h.filterOptionsUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("audit_handler.GetFilterOptions.execute_usecase",
			"error", err,
			"language", language,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve filter options")
		return
	}

	writeJSON(w, http.StatusOK, output)
}

// ListEnrichedAuditLogs handles GET /api/v1/audit/logs/enriched
func (h *AuditHandler) ListEnrichedAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	entityType := r.URL.Query().Get("entity_type")
	actorUID := r.URL.Query().Get("actor_uid")
	actorName := r.URL.Query().Get("actor_name")
	searchText := r.URL.Query().Get("search")

	// Parse action_types (comma-separated)
	var actionTypes []string
	if actionTypesStr := r.URL.Query().Get("action_types"); actionTypesStr != "" {
		actionTypes = parseCommaSeparated(actionTypesStr)
	}

	// Parse entity_types (comma-separated)
	var entityTypes []string
	if entityTypesStr := r.URL.Query().Get("entity_types"); entityTypesStr != "" {
		entityTypes = parseCommaSeparated(entityTypesStr)
	}

	// Parse date range
	var startDate, endDate *time.Time
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "invalid start_date format, use RFC3339")
			return
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "invalid end_date format, use RFC3339")
			return
		}
	}

	// Validate date range
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		writeError(w, http.StatusBadRequest, "start_date must be before or equal to end_date")
		return
	}

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			writeError(w, http.StatusBadRequest, "invalid page parameter")
			return
		}
	}

	pageSize := 50 // default page size
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 200 {
			writeError(w, http.StatusBadRequest, "page_size must be between 1 and 200")
			return
		}
	}

	input := usecases.ListEnrichedAuditLogsInput{
		EntityType:  entityType,
		ActorUID:    actorUID,
		ActorName:   actorName,
		SearchText:  searchText,
		ActionTypes: actionTypes,
		EntityTypes: entityTypes,
		StartDate:   startDate,
		EndDate:     endDate,
		Page:        page,
		PageSize:    pageSize,
	}

	output, err := h.listEnrichedUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("audit_handler.ListEnrichedAuditLogs.execute_usecase",
			"error", err,
			"entity_type", entityType,
			"actor_uid", actorUID,
			"actor_name", actorName,
		)
		writeError(w, http.StatusInternalServerError, "failed to retrieve enriched audit logs")
		return
	}

	writeJSON(w, http.StatusOK, output)
}

// parseCommaSeparated splits a comma-separated string into a slice of trimmed strings.
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := []string{}
	for _, part := range splitByComma(s) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitByComma splits a string by comma.
func splitByComma(s string) []string {
	result := []string{}
	current := ""
	for _, ch := range s {
		if ch == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// trimSpace removes leading and trailing whitespace.
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
