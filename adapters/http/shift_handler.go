package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type ShiftHandler struct {
	listUC   *usecases.ListShiftsUseCase
	getUC    *usecases.GetShiftUseCase
	createUC *usecases.CreateShiftUseCase
	updateUC *usecases.UpdateShiftUseCase
}

func NewShiftHandler(
	listUC *usecases.ListShiftsUseCase,
	getUC *usecases.GetShiftUseCase,
	createUC *usecases.CreateShiftUseCase,
	updateUC *usecases.UpdateShiftUseCase,
) *ShiftHandler {
	return &ShiftHandler{listUC: listUC, getUC: getUC, createUC: createUC, updateUC: updateUC}
}

type ShiftResponse struct {
	UID          string `json:"uid"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	GraceMinutes int    `json:"graceMinutes"`
}

type ListShiftsResponse struct {
	Shifts []ShiftResponse `json:"shifts"`
}

type ShiftRequest struct {
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	GraceMinutes int    `json:"graceMinutes"`
}

func (h *ShiftHandler) List(w http.ResponseWriter, r *http.Request) {
	output, err := h.listUC.Execute(r.Context())
	if err != nil {
		slog.Error("shift_handler.List.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	shifts := make([]ShiftResponse, 0, len(output.Shifts))
	for _, shift := range output.Shifts {
		shifts = append(shifts, ShiftResponse{UID: shift.UID, StartTime: shift.StartTime, EndTime: shift.EndTime, GraceMinutes: shift.GraceMinutes})
	}

	writeJSON(w, http.StatusOK, ListShiftsResponse{Shifts: shifts})
}

func (h *ShiftHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.getUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrShiftNotFound) {
			statusCode = http.StatusNotFound
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ShiftResponse{UID: output.Shift.UID, StartTime: output.Shift.StartTime, EndTime: output.Shift.EndTime, GraceMinutes: output.Shift.GraceMinutes})
}

func (h *ShiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	output, err := h.createUC.Execute(r.Context(), usecases.CreateShiftInput{StartTime: req.StartTime, EndTime: req.EndTime, GraceMinutes: req.GraceMinutes})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ShiftResponse{UID: output.Shift.UID, StartTime: output.Shift.StartTime, EndTime: output.Shift.EndTime, GraceMinutes: output.Shift.GraceMinutes})
}

func (h *ShiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req ShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.StartTime == "" || req.EndTime == "" {
		writeError(w, http.StatusBadRequest, "startTime and endTime are required")
		return
	}

	output, err := h.updateUC.Execute(r.Context(), usecases.UpdateShiftInput{UID: uid, StartTime: &req.StartTime, EndTime: &req.EndTime, GraceMinutes: &req.GraceMinutes})
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, usecases.ErrShiftNotFound) {
			statusCode = http.StatusNotFound
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ShiftResponse{UID: output.Shift.UID, StartTime: output.Shift.StartTime, EndTime: output.Shift.EndTime, GraceMinutes: output.Shift.GraceMinutes})
}
