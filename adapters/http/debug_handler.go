package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type DebugHandler struct {
	sendTestPushUC *usecases.SendTestPushNotificationUseCase
}

func NewDebugHandler(sendTestPushUC *usecases.SendTestPushNotificationUseCase) *DebugHandler {
	return &DebugHandler{sendTestPushUC: sendTestPushUC}
}

type sendTestPushRequest struct {
	UserUID string            `json:"userUid"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data"`
}

func (h *DebugHandler) SendTestPush(w http.ResponseWriter, r *http.Request) {
	var req sendTestPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("debug_handler.SendTestPush.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.UserUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "User UID is required")
		return
	}

	output, err := h.sendTestPushUC.Execute(r.Context(), usecases.SendTestPushNotificationInput{
		UserUID: req.UserUID,
		Title:   req.Title,
		Body:    req.Body,
		Data:    ports.NotificationData(req.Data),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrTestPushUserNotFound):
			writeJSONError(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			slog.Error("debug_handler.SendTestPush.execute_usecase", "error", err, "user_uid", req.UserUID)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to send test push")
		}
		return
	}

	writeJSON(w, http.StatusOK, output)
}
