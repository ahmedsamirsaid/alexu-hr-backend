package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

// LeaveTypeHandler handles leave type HTTP requests (admin).
type LeaveTypeHandler struct {
	listLeaveTypesUC   *usecases.ListLeaveTypesUseCase
	toggleLeaveTypeUC  *usecases.ToggleLeaveTypeUseCase
}

func NewLeaveTypeHandler(
	listLeaveTypesUC *usecases.ListLeaveTypesUseCase,
	toggleLeaveTypeUC *usecases.ToggleLeaveTypeUseCase,
) *LeaveTypeHandler {
	return &LeaveTypeHandler{
		listLeaveTypesUC:  listLeaveTypesUC,
		toggleLeaveTypeUC: toggleLeaveTypeUC,
	}
}

// Leave Type responses

type LeaveTypeResponse struct {
	UID                   string `json:"uid"`
	Code                  string `json:"code"`
	NameEN                string `json:"nameEn"`
	NameAR                string `json:"nameAr"`
	DefaultBalance        int    `json:"defaultBalance"`
	MaxConsecutive        *int   `json:"maxConsecutive,omitempty"`
	RecordingDeadlineDays *int   `json:"recordingDeadlineDays,omitempty"`
	AdvanceNoticeDays     *int   `json:"advanceNoticeDays,omitempty"`
	IsActive              bool   `json:"isActive"`
	CreatedAt             string `json:"createdAt"`
	UpdatedAt             string `json:"updatedAt"`
}

type ListLeaveTypesResponse struct {
	LeaveTypes []LeaveTypeResponse `json:"leaveTypes"`
}

// Leave Type requests

type ToggleLeaveTypeRequest struct {
	IsActive bool `json:"isActive"`
}

// ListLeaveTypes handles GET /api/v1/admin/leave-types
func (h *LeaveTypeHandler) ListLeaveTypes(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	output, err := h.listLeaveTypesUC.Execute(r.Context(), activeOnly)
	if err != nil {
		slog.Error("leave_type_handler.ListLeaveTypes.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	leaveTypes := make([]LeaveTypeResponse, 0, len(output.LeaveTypes))
	for _, lt := range output.LeaveTypes {
		leaveTypes = append(leaveTypes, LeaveTypeResponse{
			UID:                   lt.UID,
			Code:                  lt.Code,
			NameEN:                lt.NameEN,
			NameAR:                lt.NameAR,
			DefaultBalance:        lt.DefaultBalance,
			MaxConsecutive:        lt.MaxConsecutive,
			RecordingDeadlineDays: lt.RecordingDeadlineDays,
			AdvanceNoticeDays:     lt.AdvanceNoticeDays,
			IsActive:              lt.IsActive,
			CreatedAt:             lt.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:             lt.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, ListLeaveTypesResponse{LeaveTypes: leaveTypes})
}

// ToggleLeaveType handles PATCH /api/v1/admin/leave-types/{uid}/active
func (h *LeaveTypeHandler) ToggleLeaveType(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req ToggleLeaveTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("leave_type_handler.ToggleLeaveType.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.ToggleLeaveTypeInput{
		UID:      uid,
		IsActive: req.IsActive,
	}
	output, err := h.toggleLeaveTypeUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrLeaveTypeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("leave_type_handler.ToggleLeaveType.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, LeaveTypeResponse{
		UID:                   output.LeaveType.UID,
		Code:                  output.LeaveType.Code,
		NameEN:                output.LeaveType.NameEN,
		NameAR:                output.LeaveType.NameAR,
		DefaultBalance:        output.LeaveType.DefaultBalance,
		MaxConsecutive:        output.LeaveType.MaxConsecutive,
		RecordingDeadlineDays: output.LeaveType.RecordingDeadlineDays,
		AdvanceNoticeDays:     output.LeaveType.AdvanceNoticeDays,
		IsActive:              output.LeaveType.IsActive,
		CreatedAt:             output.LeaveType.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             output.LeaveType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
