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
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(
	getTrailUC *usecases.GetAuditTrailUseCase,
	getActorEventsUC *usecases.GetActorAuditEventsUseCase,
	listAllLogsUC *usecases.ListAllAuditLogsUseCase,
) *AuditHandler {
	return &AuditHandler{
		getTrailUC:       getTrailUC,
		getActorEventsUC: getActorEventsUC,
		listAllLogsUC:    listAllLogsUC,
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
