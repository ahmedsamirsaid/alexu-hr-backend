package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/banumusa/backend/core/usecases"
)

type HolidayHandler struct {
	listUC        *usecases.ListHolidaysUseCase
	listWeekendUC *usecases.ListWeekendDaysUseCase
	createUC      *usecases.CreateManualHolidayUseCase
}

func NewHolidayHandler(
	listUC *usecases.ListHolidaysUseCase,
	listWeekendUC *usecases.ListWeekendDaysUseCase,
	createUC *usecases.CreateManualHolidayUseCase,
) *HolidayHandler {
	return &HolidayHandler{listUC: listUC, listWeekendUC: listWeekendUC, createUC: createUC}
}

type holidayItemResponse struct {
	UID       string `json:"uid"`
	Date      string `json:"date"`
	Name      string `json:"name"`
	LocalName string `json:"localName"`
	IsManual  bool   `json:"isManual"`
}

type listHolidaysResponse struct {
	Holidays    []holidayItemResponse `json:"holidays"`
	WeekendDays []int                 `json:"weekendDays"`
}

type createHolidayRequest struct {
	Date   string  `json:"date"`
	NameEN string  `json:"nameEn"`
	NameAR *string `json:"nameAr,omitempty"`
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

	items := make([]holidayItemResponse, 0, len(output.Holidays))
	for _, holiday := range output.Holidays {
		items = append(items, holidayItemResponse{
			UID:       holiday.UID,
			Date:      holiday.Date.Format("2006-01-02"),
			Name:      holiday.NameEN,
			LocalName: holiday.NameAR,
			IsManual:  holiday.IsManual,
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
		Date:   dateValue,
		NameEN: req.NameEN,
		NameAR: req.NameAR,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrHolidayDateAlreadyExists):
			writeError(w, http.StatusConflict, "A holiday already exists on this date.")
		case errors.Is(err, usecases.ErrHolidayDateFallsOnWeekend):
			writeError(w, http.StatusConflict, "Manual holidays cannot be created on weekends.")
		default:
			slog.Error("holiday_handler.CreateHoliday.execute", "error", err)
			writeError(w, http.StatusInternalServerError, "Unable to create holiday.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, holidayItemResponse{
		UID:       output.UID,
		Date:      output.Date.Format("2006-01-02"),
		Name:      output.NameEN,
		LocalName: output.NameAR,
		IsManual:  output.IsManual,
	})
}
