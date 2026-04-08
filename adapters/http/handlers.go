package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/banumusa/backend/core/usecases"
)

type RecordLeaveRequest struct {
	EmployeeUID  string  `json:"employeeUid"`
	LeaveTypeUID string  `json:"leaveTypeUid"`
	StartDate    string  `json:"startDate"`
	EndDate      string  `json:"endDate"`
	Notes        *string `json:"notes,omitempty"`
}

type LeaveRecordResponse struct {
	UID       string  `json:"uid"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Days      int     `json:"days"`
	Notes     *string `json:"notes,omitempty"`
}

type RecordLeaveResponse struct {
	Records []LeaveRecordResponse `json:"records"`
}

type BalanceResponse struct {
	LeaveTypeUID     string `json:"leaveTypeUid"`
	LeaveTypeCode    string `json:"leaveTypeCode"`
	LeaveTypeNameEN  string `json:"leaveTypeNameEn"`
	LeaveTypeNameAR  string `json:"leaveTypeNameAr"`
	Year             int    `json:"year"`
	InitialBalance   int    `json:"initialBalance"`
	UsedBalance      int    `json:"usedBalance"`
	RemainingBalance int    `json:"remainingBalance"`
}

type GetBalanceResponse struct {
	Balances []BalanceResponse `json:"balances"`
}

type LeaveRecordListItem struct {
	UID             string  `json:"uid"`
	LeaveTypeUID    string  `json:"leaveTypeUid"`
	LeaveTypeCode   string  `json:"leaveTypeCode"`
	LeaveTypeNameEN string  `json:"leaveTypeNameEn"`
	LeaveTypeNameAR string  `json:"leaveTypeNameAr"`
	StartDate       string  `json:"startDate"`
	EndDate         string  `json:"endDate"`
	Days            int     `json:"days"`
	RecordedAt      string  `json:"recordedAt"`
	Notes           *string `json:"notes,omitempty"`
}

type ListLeaveRecordsResponse struct {
	Records  []LeaveRecordListItem `json:"records"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type AllLeaveRecordListItem struct {
	UID             string  `json:"uid"`
	EmployeeUID     string  `json:"employeeUid"`
	EmployeeName    string  `json:"employeeName"`
	LeaveTypeUID    string  `json:"leaveTypeUid"`
	LeaveTypeCode   string  `json:"leaveTypeCode"`
	LeaveTypeNameEN string  `json:"leaveTypeNameEn"`
	LeaveTypeNameAR string  `json:"leaveTypeNameAr"`
	StartDate       string  `json:"startDate"`
	EndDate         string  `json:"endDate"`
	Days            int     `json:"days"`
	RecordedAt      string  `json:"recordedAt"`
	Notes           *string `json:"notes,omitempty"`
}

type ListAllLeaveRecordsResponse struct {
	Records  []AllLeaveRecordListItem `json:"records"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LeaveHandler struct {
	recordLeaveUC         *usecases.RecordLeaveUseCase
	getBalanceUC          *usecases.GetBalanceUseCase
	listLeaveRecordsUC    *usecases.ListLeaveRecordsUseCase
	listAllLeaveRecordsUC *usecases.ListAllLeaveRecordsUseCase
}

func NewLeaveHandler(
	recordLeaveUC *usecases.RecordLeaveUseCase,
	getBalanceUC *usecases.GetBalanceUseCase,
	listLeaveRecordsUC *usecases.ListLeaveRecordsUseCase,
	listAllLeaveRecordsUC *usecases.ListAllLeaveRecordsUseCase,
) *LeaveHandler {
	return &LeaveHandler{
		recordLeaveUC:         recordLeaveUC,
		getBalanceUC:          getBalanceUC,
		listLeaveRecordsUC:    listLeaveRecordsUC,
		listAllLeaveRecordsUC: listAllLeaveRecordsUC,
	}
}

func (h *LeaveHandler) RecordLeave(w http.ResponseWriter, r *http.Request) {
	var req RecordLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startDate format, expected YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid endDate format, expected YYYY-MM-DD")
		return
	}

	input := usecases.RecordLeaveInput{
		EmployeeUID:  req.EmployeeUID,
		LeaveTypeUID: req.LeaveTypeUID,
		StartDate:    startDate,
		EndDate:      endDate,
		Notes:        req.Notes,
	}

	output, err := h.recordLeaveUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrLeaveTypeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrInsufficientBalance):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrExceedsConsecutiveDays):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrRecordingDeadlinePassed):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidDateRange):
			statusCode = http.StatusBadRequest
		}
		writeError(w, statusCode, err.Error())
		return
	}

	var records []LeaveRecordResponse
	for _, rec := range output.Records {
		records = append(records, LeaveRecordResponse{
			UID:       rec.UID,
			StartDate: rec.StartDate.Format("2006-01-02"),
			EndDate:   rec.EndDate.Format("2006-01-02"),
			Days:      rec.Days,
			Notes:     rec.Notes,
		})
	}

	writeJSON(w, http.StatusCreated, RecordLeaveResponse{Records: records})
}

func (h *LeaveHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid year parameter")
			return
		}
	}

	input := usecases.GetBalanceInput{
		EmployeeUID: employeeUID,
		Year:        year,
	}

	output, err := h.getBalanceUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		}
		writeError(w, statusCode, err.Error())
		return
	}

	var balances []BalanceResponse
	for _, b := range output.Balances {
		balances = append(balances, BalanceResponse{
			LeaveTypeUID:     b.LeaveTypeUID,
			LeaveTypeCode:    b.LeaveTypeCode,
			LeaveTypeNameEN:  b.LeaveTypeNameEN,
			LeaveTypeNameAR:  b.LeaveTypeNameAR,
			Year:             b.Year,
			InitialBalance:   b.InitialBalance,
			UsedBalance:      b.UsedBalance,
			RemainingBalance: b.RemainingBalance,
		})
	}

	writeJSON(w, http.StatusOK, GetBalanceResponse{Balances: balances})
}

func (h *LeaveHandler) ListLeaveRecords(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	input := usecases.ListLeaveRecordsInput{
		EmployeeUID: employeeUID,
	}

	if leaveTypeUID := r.URL.Query().Get("leaveTypeUid"); leaveTypeUID != "" {
		input.LeaveTypeUID = &leaveTypeUID
	}

	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid startDate format, expected YYYY-MM-DD")
			return
		}
		input.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid endDate format, expected YYYY-MM-DD")
			return
		}
		input.EndDate = &endDate
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid page parameter")
			return
		}
		input.Page = page
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid pageSize parameter")
			return
		}
		input.PageSize = pageSize
	}

	output, err := h.listLeaveRecordsUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrLeaveTypeNotFound):
			statusCode = http.StatusNotFound
		}
		writeError(w, statusCode, err.Error())
		return
	}

	var records []LeaveRecordListItem
	for _, rec := range output.Records {
		records = append(records, LeaveRecordListItem{
			UID:             rec.UID,
			LeaveTypeUID:    rec.LeaveTypeUID,
			LeaveTypeCode:   rec.LeaveTypeCode,
			LeaveTypeNameEN: rec.LeaveTypeNameEN,
			LeaveTypeNameAR: rec.LeaveTypeNameAR,
			StartDate:       rec.StartDate.Format("2006-01-02"),
			EndDate:         rec.EndDate.Format("2006-01-02"),
			Days:            rec.Days,
			RecordedAt:      rec.RecordedAt.Format(time.RFC3339),
			Notes:           rec.Notes,
		})
	}

	writeJSON(w, http.StatusOK, ListLeaveRecordsResponse{
		Records:  records,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
	})
}

func (h *LeaveHandler) ListAllLeaveRecords(w http.ResponseWriter, r *http.Request) {
	input := usecases.ListAllLeaveRecordsInput{}

	if search := r.URL.Query().Get("search"); search != "" {
		input.Search = search
	}

	if leaveTypeUID := r.URL.Query().Get("leaveTypeUid"); leaveTypeUID != "" {
		input.LeaveTypeUID = &leaveTypeUID
	}

	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid startDate format, expected YYYY-MM-DD")
			return
		}
		input.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid endDate format, expected YYYY-MM-DD")
			return
		}
		input.EndDate = &endDate
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid page parameter")
			return
		}
		input.Page = page
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid pageSize parameter")
			return
		}
		input.PageSize = pageSize
	}

	output, err := h.listAllLeaveRecordsUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrLeaveTypeNotFound) {
			statusCode = http.StatusNotFound
		}
		writeError(w, statusCode, err.Error())
		return
	}

	var records []AllLeaveRecordListItem
	for _, rec := range output.Records {
		records = append(records, AllLeaveRecordListItem{
			UID:             rec.UID,
			EmployeeUID:     rec.EmployeeUID,
			EmployeeName:    rec.EmployeeName,
			LeaveTypeUID:    rec.LeaveTypeUID,
			LeaveTypeCode:   rec.LeaveTypeCode,
			LeaveTypeNameEN: rec.LeaveTypeNameEN,
			LeaveTypeNameAR: rec.LeaveTypeNameAR,
			StartDate:       rec.StartDate.Format("2006-01-02"),
			EndDate:         rec.EndDate.Format("2006-01-02"),
			Days:            rec.Days,
			RecordedAt:      rec.RecordedAt.Format(time.RFC3339),
			Notes:           rec.Notes,
		})
	}

	writeJSON(w, http.StatusOK, ListAllLeaveRecordsResponse{
		Records:  records,
		Total:    output.Total,
		Page:     output.Page,
		PageSize: output.PageSize,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
