package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type MeHandler struct {
	updateUserUC *usecases.UpdateUserUseCase
}

func NewMeHandler(updateUserUC *usecases.UpdateUserUseCase) *MeHandler {
	return &MeHandler{updateUserUC: updateUserUC}
}

type updateLanguageRequest struct {
	Language string `json:"language"`
}

func (h *MeHandler) UpdateLanguage(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	var req updateLanguageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Language != "ar" && req.Language != "en" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "language must be 'ar' or 'en'")
		return
	}

	err := h.updateUserUC.Execute(r.Context(), usecases.UpdateUserInput{
		UserUID:           claims.UserUID,
		PreferredLanguage: &req.Language,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			writeJSONError(w, http.StatusNotFound, "user_not_found", "User not found")
			return
		}
		slog.Error("me_handler.UpdateLanguage.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update language")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Language updated"})
}
