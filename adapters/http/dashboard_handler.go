package http

import (
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

// DashboardStatsResponse represents the dashboard stats API response.
type DashboardStatsResponse struct {
	TotalEmployees  int `json:"totalEmployees"`
	LeavesToday     int `json:"leavesToday"`
	PendingRequests int `json:"pendingRequests"`
}

// DashboardHandler handles dashboard HTTP requests.
type DashboardHandler struct {
	getStatsUC *usecases.GetDashboardStatsUseCase
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(getStatsUC *usecases.GetDashboardStatsUseCase) *DashboardHandler {
	return &DashboardHandler{
		getStatsUC: getStatsUC,
	}
}

// GetStats handles GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	output, err := h.getStatsUC.Execute(r.Context())
	if err != nil {
		slog.Error("dashboard_handler.GetStats.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, DashboardStatsResponse{
		TotalEmployees:  output.TotalEmployees,
		LeavesToday:     output.LeavesToday,
		PendingRequests: output.PendingRequests,
	})
}
