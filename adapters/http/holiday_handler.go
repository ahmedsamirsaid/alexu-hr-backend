package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type HolidayHandler struct {
	db            ports.DB
	listUC        *usecases.ListHolidaysUseCase
	listWeekendUC *usecases.ListWeekendDaysUseCase
	createUC      *usecases.CreateManualHolidayUseCase
	updateUC      *usecases.UpdateHolidayUseCase
	deleteUC      *usecases.DeleteHolidayUseCase
	employeeRepo  ports.EmployeeRepository
}

func NewHolidayHandler(
	db ports.DB,
	listUC *usecases.ListHolidaysUseCase,
	listWeekendUC *usecases.ListWeekendDaysUseCase,
	createUC *usecases.CreateManualHolidayUseCase,
	updateUC *usecases.UpdateHolidayUseCase,
	deleteUC *usecases.DeleteHolidayUseCase,
	employeeRepo ports.EmployeeRepository,
) *HolidayHandler {
	return &HolidayHandler{db: db, listUC: listUC, listWeekendUC: listWeekendUC, createUC: createUC, updateUC: updateUC, deleteUC: deleteUC, employeeRepo: employeeRepo}
}

type holidayItemResponse struct {
	UID            string   `json:"uid"`
	Date           string   `json:"date"`
	Name           string   `json:"name"`
	LocalName      string   `json:"localName"`
	DepartmentUIDs []string `json:"departmentUids,omitempty"`
	IsManual       bool     `json:"isManual"`
}

type listHolidaysResponse struct {
	Holidays    []holidayItemResponse `json:"holidays"`
	WeekendDays []int                 `json:"weekendDays"`
}

type createHolidayRequest struct {
	Date           string   `json:"date"`
	NameEN         string   `json:"nameEn"`
	NameAR         *string  `json:"nameAr,omitempty"`
	DepartmentUIDs []string `json:"departmentUids,omitempty"`
}

type updateHolidayRequest struct {
	Date           string   `json:"date"`
	NameEN         string   `json:"nameEn"`
	NameAR         *string  `json:"nameAr,omitempty"`
	DepartmentUIDs []string `json:"departmentUids,omitempty"`
}

func (h *HolidayHandler) ListHolidays(w http.ResponseWriter, r *http.Request) {
	var input usecases.ListHolidaysInput

	if from := r.URL.Query().Get("from"); from != "" {
		parsed, err := time.Parse("2006-01-02", from)
		if err != nil {
			writeError(w, http.StatusBadRequest, "The 'from' date must use YYYY-MM-DD format.")
			return
		}
		input.StartDate = &parsed
	}

	if to := r.URL.Query().Get("to"); to != "" {
		parsed, err := time.Parse("2006-01-02", to)
		if err != nil {
			writeError(w, http.StatusBadRequest, "The 'to' date must use YYYY-MM-DD format.")
			return
		}
		input.EndDate = &parsed
	}

	output, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("holiday_handler.ListHolidays.execute", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	allowedDepartmentUIDs := h.allowedHolidayDepartmentUIDs(r)
	items := make([]holidayItemResponse, 0, len(output.Holidays))
	for _, holiday := range output.Holidays {
		if !holidayVisibleToDepartmentUIDs(holiday.DepartmentUIDs, allowedDepartmentUIDs) {
			continue
		}
		items = append(items, holidayItemResponse{
			UID:            holiday.UID,
			Date:           holiday.Date.Format("2006-01-02"),
			Name:           holiday.NameEN,
			LocalName:      holiday.NameAR,
			DepartmentUIDs: append([]string(nil), holiday.DepartmentUIDs...),
			IsManual:       holiday.IsManual,
		})
	}

	weekendOutput, err := h.listWeekendUC.Execute(r.Context())
	if err != nil {
		slog.Error("holiday_handler.ListHolidays.weekend_days", "error", err)
		writeError(w, http.StatusInternalServerError, "Unable to load weekend configuration.")
		return
	}

	writeJSON(w, http.StatusOK, listHolidaysResponse{Holidays: items, WeekendDays: weekendOutput.WeekendDays})
}

func (h *HolidayHandler) CreateHoliday(w http.ResponseWriter, r *http.Request) {
	var req createHolidayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "The request body is invalid.")
		return
	}

	dateValue, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "The date must use YYYY-MM-DD format.")
		return
	}

	output, err := h.createUC.Execute(r.Context(), usecases.CreateManualHolidayInput{
		Date:           dateValue,
		NameEN:         req.NameEN,
		NameAR:         req.NameAR,
		DepartmentUIDs: req.DepartmentUIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrHolidayDateAlreadyExists):
			writeError(w, http.StatusConflict, "A holiday already exists on this date.")
		default:
			slog.Error("holiday_handler.CreateHoliday.execute", "error", err)
			writeError(w, http.StatusInternalServerError, "Unable to create holiday.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, holidayItemResponse{
		UID:            output.UID,
		Date:           output.Date.Format("2006-01-02"),
		Name:           output.NameEN,
		LocalName:      output.NameAR,
		DepartmentUIDs: append([]string(nil), output.DepartmentUIDs...),
		IsManual:       output.IsManual,
	})
}

func (h *HolidayHandler) UpdateHoliday(w http.ResponseWriter, r *http.Request) {
	if h.updateUC == nil {
		writeError(w, http.StatusNotImplemented, "update holiday is unavailable")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req updateHolidayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "The request body is invalid.")
		return
	}

	dateValue, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "The date must use YYYY-MM-DD format.")
		return
	}

	definition, err := h.updateUC.Execute(r.Context(), usecases.UpdateHolidayInput{
		UID:            uid,
		Date:           dateValue,
		NameEN:         req.NameEN,
		NameAR:         req.NameAR,
		DepartmentUIDs: req.DepartmentUIDs,
		IsManual:       true,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrHolidayNotFound):
			writeError(w, http.StatusNotFound, "Holiday not found.")
		case errors.Is(err, usecases.ErrHolidayDateInPast):
			writeError(w, http.StatusConflict, "Past holidays cannot be edited.")
		case errors.Is(err, usecases.ErrHolidayDateAlreadyExists):
			writeError(w, http.StatusConflict, "A holiday already exists on this date.")
		default:
			slog.Error("holiday_handler.UpdateHoliday.execute", "error", err)
			writeError(w, http.StatusInternalServerError, "Unable to update holiday.")
		}
		return
	}

	writeJSON(w, http.StatusOK, holidayItemResponse{
		UID:            definition.UID,
		Date:           definition.Date.Format("2006-01-02"),
		Name:           definition.NameEN,
		LocalName:      definition.NameAR,
		DepartmentUIDs: append([]string(nil), definition.DepartmentUIDs...),
		IsManual:       definition.IsManual,
	})
}

func (h *HolidayHandler) DeleteHoliday(w http.ResponseWriter, r *http.Request) {
	if h.deleteUC == nil {
		writeError(w, http.StatusNotImplemented, "delete holiday is unavailable")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	if err := h.deleteUC.Execute(r.Context(), usecases.DeleteHolidayInput{UID: uid}); err != nil {
		switch {
		case errors.Is(err, usecases.ErrHolidayNotFound):
			writeError(w, http.StatusNotFound, "Holiday not found.")
		case errors.Is(err, usecases.ErrHolidayDateInPast):
			writeError(w, http.StatusConflict, "Past holidays cannot be deleted.")
		default:
			slog.Error("holiday_handler.DeleteHoliday.execute", "error", err)
			writeError(w, http.StatusInternalServerError, "Unable to delete holiday.")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HolidayHandler) allowedHolidayDepartmentUIDs(r *http.Request) []string {
	claims := GetClaims(r)
	if claims == nil {
		return nil
	}

	if claims.IsGlobalScope() || claims.HasPermission("*") || claims.HasPermission("holidays:write") {
		return nil
	}

	if claims.IsDepartmentScope() && len(claims.ManagedDepartmentUIDs) > 0 {
		return append([]string(nil), claims.ManagedDepartmentUIDs...)
	}

	if claims.EmployeeUID != nil && h.employeeRepo != nil {
		employee, err := h.employeeRepo.GetByUID(r.Context(), h.db, *claims.EmployeeUID)
		if err != nil || employee == nil || employee.DepartmentUID == nil {
			return []string{}
		}
		return []string{*employee.DepartmentUID}
	}

	return []string{}
}

func holidayVisibleToDepartmentUIDs(departmentUIDs, allowedDepartmentUIDs []string) bool {
	if allowedDepartmentUIDs == nil {
		return true
	}
	if len(departmentUIDs) == 0 {
		return true
	}
	if len(allowedDepartmentUIDs) == 0 {
		return false
	}
	allowedSet := make(map[string]struct{}, len(allowedDepartmentUIDs))
	for _, departmentUID := range allowedDepartmentUIDs {
		allowedSet[strings.TrimSpace(departmentUID)] = struct{}{}
	}
	for _, departmentUID := range departmentUIDs {
		if _, ok := allowedSet[strings.TrimSpace(departmentUID)]; ok {
			return true
		}
	}
	return false
}
