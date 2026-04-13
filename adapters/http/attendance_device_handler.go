package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

type AttendanceDeviceHandler struct {
	registerUC *usecases.RegisterAttendanceDeviceUseCase
	listUC     *usecases.ListAttendanceDevicesUseCase
	getUC      *usecases.GetAttendanceDeviceUseCase
	updateUC   *usecases.UpdateAttendanceDeviceUseCase
	deleteUC   *usecases.DeleteAttendanceDeviceUseCase
	activateUC *usecases.ActivateAttendanceDeviceUseCase
	statsUC    *usecases.GetAttendanceDeviceStatsUseCase
	checkUC    *usecases.CheckAttendanceDeviceConnectionUseCase
	checkAllUC *usecases.CheckAllAttendanceDevicesConnectionUseCase
}

func NewAttendanceDeviceHandler(
	registerUC *usecases.RegisterAttendanceDeviceUseCase,
	listUC *usecases.ListAttendanceDevicesUseCase,
	getUC *usecases.GetAttendanceDeviceUseCase,
	updateUC *usecases.UpdateAttendanceDeviceUseCase,
	deleteUC *usecases.DeleteAttendanceDeviceUseCase,
	activateUC *usecases.ActivateAttendanceDeviceUseCase,
	statsUC *usecases.GetAttendanceDeviceStatsUseCase,
	checkUC *usecases.CheckAttendanceDeviceConnectionUseCase,
	checkAllUC *usecases.CheckAllAttendanceDevicesConnectionUseCase,
) *AttendanceDeviceHandler {
	return &AttendanceDeviceHandler{
		registerUC: registerUC,
		listUC:     listUC,
		getUC:      getUC,
		updateUC:   updateUC,
		deleteUC:   deleteUC,
		activateUC: activateUC,
		statsUC:    statsUC,
		checkUC:    checkUC,
		checkAllUC: checkAllUC,
	}
}

type updateAttendanceDeviceRequest struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type registerAttendanceDeviceRequest struct {
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	SerialNumber string `json:"serialNumber"`
}

func (h *AttendanceDeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerAttendanceDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("attendance_device_handler.Register.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	output, err := h.registerUC.Execute(r.Context(), usecases.RegisterAttendanceDeviceInput{
		IP:           req.IP,
		Port:         req.Port,
		Name:         req.Name,
		Location:     req.Location,
		SerialNumber: req.SerialNumber,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrDeviceNameRequired):
			writeJSONError(w, http.StatusBadRequest, "validation_error", "Name is required")
		case errors.Is(err, usecases.ErrDeviceIPRequired):
			writeJSONError(w, http.StatusBadRequest, "validation_error", "IP is required")
		case errors.Is(err, usecases.ErrDeviceInvalidIP):
			writeJSONError(w, http.StatusBadRequest, "invalid_ip", err.Error())
		case errors.Is(err, usecases.ErrDeviceInvalidPort):
			writeJSONError(w, http.StatusBadRequest, "invalid_port", err.Error())
		case errors.Is(err, usecases.ErrDeviceSerialNumberRequired):
			writeJSONError(w, http.StatusBadRequest, "validation_error", "Serial number is required")
		case errors.Is(err, usecases.ErrDeviceSerialNumberExists):
			writeJSONError(w, http.StatusConflict, "serial_number_exists", err.Error())
		case errors.Is(err, usecases.ErrDeviceAddressExists):
			writeJSONError(w, http.StatusConflict, "device_address_exists", err.Error())
		default:
			slog.Error("attendance_device_handler.Register.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to register device")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

func (h *AttendanceDeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20
	search := r.URL.Query().Get("search")

	var status *domain.AttendanceDeviceStatus
	if rawStatus := r.URL.Query().Get("status"); rawStatus != "" && rawStatus != "all" {
		parsedStatus := domain.AttendanceDeviceStatus(strings.ToLower(rawStatus))
		switch parsedStatus {
		case domain.AttendanceDeviceStatusOnline, domain.AttendanceDeviceStatusOffline, domain.AttendanceDeviceStatusDeactivated:
			status = &parsedStatus
		default:
			writeJSONError(w, http.StatusBadRequest, "validation_error", "Status must be one of: all, online, offline, deactivated")
			return
		}
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	output, err := h.listUC.Execute(r.Context(), usecases.ListAttendanceDevicesInput{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Search:   search,
	})
	if err != nil {
		slog.Error("attendance_device_handler.List.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list devices")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *AttendanceDeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Device UID is required")
		return
	}

	err := h.deleteUC.Execute(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrAttendanceDeviceNotFound):
			writeJSONError(w, http.StatusNotFound, "not_found", "Device not found")
		default:
			slog.Error("attendance_device_handler.Delete.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to deactivate device")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Device deactivated"})
}

func (h *AttendanceDeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Device UID is required")
		return
	}

	output, err := h.getUC.Execute(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrAttendanceDeviceNotFound):
			writeJSONError(w, http.StatusNotFound, "not_found", "Device not found")
		default:
			slog.Error("attendance_device_handler.Get.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch device")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *AttendanceDeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Device UID is required")
		return
	}

	var req updateAttendanceDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("attendance_device_handler.Update.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	output, err := h.updateUC.Execute(r.Context(), usecases.UpdateAttendanceDeviceInput{
		UID:      uid,
		IP:       req.IP,
		Port:     req.Port,
		Name:     req.Name,
		Location: req.Location,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrAttendanceDeviceNotFound):
			writeJSONError(w, http.StatusNotFound, "not_found", "Device not found")
		case errors.Is(err, usecases.ErrDeviceNameRequired):
			writeJSONError(w, http.StatusBadRequest, "validation_error", "Name is required")
		case errors.Is(err, usecases.ErrDeviceIPRequired):
			writeJSONError(w, http.StatusBadRequest, "validation_error", "IP is required")
		case errors.Is(err, usecases.ErrDeviceInvalidIP):
			writeJSONError(w, http.StatusBadRequest, "invalid_ip", err.Error())
		case errors.Is(err, usecases.ErrDeviceInvalidPort):
			writeJSONError(w, http.StatusBadRequest, "invalid_port", err.Error())
		case errors.Is(err, usecases.ErrDeviceAddressExists):
			writeJSONError(w, http.StatusConflict, "device_address_exists", err.Error())
		default:
			slog.Error("attendance_device_handler.Update.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update device")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *AttendanceDeviceHandler) Activate(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Device UID is required")
		return
	}

	err := h.activateUC.Execute(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrAttendanceDeviceNotFound):
			writeJSONError(w, http.StatusNotFound, "not_found", "Device not found")
		case errors.Is(err, usecases.ErrAttendanceDeviceAlreadyActive):
			writeJSONError(w, http.StatusConflict, "device_already_active", "Device is already active")
		default:
			slog.Error("attendance_device_handler.Activate.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to activate device")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Device activated"})
}

func (h *AttendanceDeviceHandler) Stats(w http.ResponseWriter, r *http.Request) {
	output, err := h.statsUC.Execute(r.Context())
	if err != nil {
		slog.Error("attendance_device_handler.Stats.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch device stats")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *AttendanceDeviceHandler) CheckConnection(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	output, err := h.checkUC.Execute(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrCheckDeviceUIDRequired):
			writeJSONError(w, http.StatusBadRequest, "validation_error", "Device UID is required")
		case errors.Is(err, usecases.ErrAttendanceDeviceNotFound):
			writeJSONError(w, http.StatusNotFound, "not_found", "Device not found")
		case errors.Is(err, usecases.ErrAttendanceDeviceDeactivated):
			writeJSONError(w, http.StatusConflict, "device_deactivated", "Device is deactivated")
		default:
			slog.Error("attendance_device_handler.CheckConnection.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to check connection")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (h *AttendanceDeviceHandler) CheckAllConnections(w http.ResponseWriter, r *http.Request) {
	output, err := h.checkAllUC.Execute(r.Context())
	if err != nil {
		slog.Error("attendance_device_handler.CheckAllConnections.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to check all device connections")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
