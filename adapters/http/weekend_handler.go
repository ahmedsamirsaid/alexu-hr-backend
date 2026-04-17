package http

import (
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type WeekendHandler struct {
	listUC *usecases.ListWeekendDaysUseCase
}

func NewWeekendHandler(listUC *usecases.ListWeekendDaysUseCase) *WeekendHandler {
	return &WeekendHandler{listUC: listUC}
}

type weekendDaysResponse struct {
	WeekendDays []int `json:"weekendDays"`
}

func (h *WeekendHandler) ListWeekendDays(w http.ResponseWriter, r *http.Request) {
	output, err := h.listUC.Execute(r.Context())
	if err != nil {
		slog.Error("weekend_handler.ListWeekendDays.execute", "error", err)
		writeError(w, http.StatusInternalServerError, "Unable to load weekend configuration.")
		return
	}

	writeJSON(w, http.StatusOK, weekendDaysResponse{WeekendDays: output.WeekendDays})
}
