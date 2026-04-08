package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type DeviceTokenHandler struct {
	registerUC   *usecases.RegisterDeviceTokenUseCase
	unregisterUC *usecases.UnregisterDeviceTokenUseCase
}

func NewDeviceTokenHandler(
	registerUC *usecases.RegisterDeviceTokenUseCase,
	unregisterUC *usecases.UnregisterDeviceTokenUseCase,
) *DeviceTokenHandler {
	return &DeviceTokenHandler{
		registerUC:   registerUC,
		unregisterUC: unregisterUC,
	}
}

type registerDeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type registerDeviceTokenResponse struct {
	UID string `json:"uid"`
}

func (h *DeviceTokenHandler) Register(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	var req registerDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("device_token_handler.Register.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Token == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Token is required")
		return
	}

	if req.Platform == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Platform is required")
		return
	}

	output, err := h.registerUC.Execute(r.Context(), usecases.RegisterDeviceTokenInput{
		UserUID:  claims.UserUID,
		Token:    req.Token,
		Platform: req.Platform,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrInvalidPlatform):
			writeJSONError(w, http.StatusBadRequest, "invalid_platform", "Platform must be 'android' or 'ios'")
		case errors.Is(err, usecases.ErrInvalidToken):
			writeJSONError(w, http.StatusBadRequest, "invalid_token", "Token is required")
		default:
			slog.Error("device_token_handler.Register.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to register device token")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registerDeviceTokenResponse{UID: output.UID})
}

func (h *DeviceTokenHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Token UID is required")
		return
	}

	err := h.unregisterUC.Execute(r.Context(), usecases.UnregisterDeviceTokenInput{
		UID:     uid,
		UserUID: claims.UserUID,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrDeviceTokenNotFound):
			writeJSONError(w, http.StatusNotFound, "not_found", "Device token not found")
		case errors.Is(err, usecases.ErrNotTokenOwner):
			writeJSONError(w, http.StatusForbidden, "forbidden", "Not authorized to delete this token")
		default:
			slog.Error("device_token_handler.Unregister.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to unregister device token")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Device token unregistered"})
}
