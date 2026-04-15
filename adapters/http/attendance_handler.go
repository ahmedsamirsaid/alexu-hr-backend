package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type monthlyAttendanceStatsExecutor interface {
	Execute(ctx context.Context, input usecases.GetMonthlyAttendanceStatsInput) (*usecases.GetMonthlyAttendanceStatsOutput, error)
}

type AttendanceHandler struct {
	listDepartmentLogsUC      *usecases.ListDepartmentAttendanceLogsUseCase
	listEmployeeLogsUC        *usecases.ListEmployeeAttendanceLogsUseCase
	listDailyDepartmentLogsUC *usecases.ListDailyDepartmentAttendanceLogsUseCase
	listDailyEmployeeLogsUC   *usecases.ListDailyEmployeeAttendanceLogsUseCase
	getMonthlyStatsUC         monthlyAttendanceStatsExecutor
	getDailySummaryUC         *usecases.GetDailyAttendanceSummaryUseCase
	getWorkHoursUC            *usecases.GetWorkHoursConfigUseCase
	setWorkHoursUC            *usecases.SetWorkHoursConfigUseCase
}

func NewAttendanceHandler(
	listDepartmentLogsUC *usecases.ListDepartmentAttendanceLogsUseCase,
	listEmployeeLogsUC *usecases.ListEmployeeAttendanceLogsUseCase,
	listDailyDepartmentLogsUC *usecases.ListDailyDepartmentAttendanceLogsUseCase,
	listDailyEmployeeLogsUC *usecases.ListDailyEmployeeAttendanceLogsUseCase,
	getMonthlyStatsUC *usecases.GetMonthlyAttendanceStatsUseCase,
	getDailySummaryUC *usecases.GetDailyAttendanceSummaryUseCase,
	getWorkHoursUC *usecases.GetWorkHoursConfigUseCase,
	setWorkHoursUC *usecases.SetWorkHoursConfigUseCase,
) *AttendanceHandler {
	return &AttendanceHandler{
		listDepartmentLogsUC:      listDepartmentLogsUC,
		listEmployeeLogsUC:        listEmployeeLogsUC,
		listDailyDepartmentLogsUC: listDailyDepartmentLogsUC,
		listDailyEmployeeLogsUC:   listDailyEmployeeLogsUC,
		getMonthlyStatsUC:         getMonthlyStatsUC,
		getDailySummaryUC:         getDailySummaryUC,
		getWorkHoursUC:            getWorkHoursUC,
		setWorkHoursUC:            setWorkHoursUC,
	}
}

type workHoursRequest struct {
	WorkDayStart      string `json:"workDayStart"`
	WorkDayEnd        string `json:"workDayEnd"`
	LateGraceMinutes  int    `json:"lateGraceMinutes"`
	EarlyGraceMinutes int    `json:"earlyGraceMinutes"`
}

type workHoursResponse struct {
	WorkDayStart      string `json:"workDayStart"`
	WorkDayEnd        string `json:"workDayEnd"`
	LateGraceMinutes  int    `json:"lateGraceMinutes"`
	EarlyGraceMinutes int    `json:"earlyGraceMinutes"`
}

type dailySummaryItemResponse struct {
	EmployeeUID     string  `json:"employeeUid"`
	EmployeeName    string  `json:"employeeName,omitempty"`
	CheckIn         *string `json:"checkIn,omitempty"`
	CheckOut        *string `json:"checkOut,omitempty"`
	WorkedHours     float64 `json:"workedHours"`
	LateArrival     bool    `json:"lateArrival"`
	EarlyDeparture  bool    `json:"earlyDeparture"`
	MissingCheckIn  bool    `json:"missingCheckIn"`
	MissingCheckOut bool    `json:"missingCheckOut"`
}

type attendanceLogItemResponse struct {
	UID           string  `json:"uid"`
	EmployeeUID   string  `json:"employeeUid"`
	EmployeeName  string  `json:"employeeName"`
	DepartmentUID *string `json:"departmentUid,omitempty"`
	DeviceUID     string  `json:"deviceUid"`
	DeviceName    string  `json:"deviceName"`
	DeviceUserID  string  `json:"deviceUserId"`
	PunchedAt     string  `json:"punchedAt"`
	PunchType     string  `json:"punchType"`
	RawPayload    *string `json:"rawPayload,omitempty"`
}

type listDepartmentLogsResponse struct {
	DepartmentUID string                      `json:"departmentUid"`
	Logs          []attendanceLogItemResponse `json:"logs"`
	Total         int                         `json:"total"`
	Page          int                         `json:"page"`
	PageSize      int                         `json:"pageSize"`
	TotalPages    int                         `json:"totalPages"`
}

type listEmployeeLogsResponse struct {
	EmployeeUID string                      `json:"employeeUid"`
	Logs        []attendanceLogItemResponse `json:"logs"`
	Total       int                         `json:"total"`
	Page        int                         `json:"page"`
	PageSize    int                         `json:"pageSize"`
	TotalPages  int                         `json:"totalPages"`
}

type dailyAttendanceLogItemResponse struct {
	Date            string  `json:"date"`
	EmployeeUID     string  `json:"employeeUid"`
	EmployeeName    string  `json:"employeeName"`
	DepartmentUID   *string `json:"departmentUid,omitempty"`
	CheckIn         *string `json:"checkIn,omitempty"`
	CheckOut        *string `json:"checkOut,omitempty"`
	CheckInDevice   *string `json:"checkInDevice,omitempty"`
	CheckOutDevice  *string `json:"checkOutDevice,omitempty"`
	WorkedHours     float64 `json:"workedHours"`
	LateArrival     bool    `json:"lateArrival"`
	EarlyDeparture  bool    `json:"earlyDeparture"`
	MissingCheckIn  bool    `json:"missingCheckIn"`
	MissingCheckOut bool    `json:"missingCheckOut"`
}

type listDailyDepartmentLogsResponse struct {
	DepartmentUID string                           `json:"departmentUid"`
	Records       []dailyAttendanceLogItemResponse `json:"records"`
	Total         int                              `json:"total"`
	Page          int                              `json:"page"`
	PageSize      int                              `json:"pageSize"`
	TotalPages    int                              `json:"totalPages"`
}

type listDailyEmployeeLogsResponse struct {
	EmployeeUID string                           `json:"employeeUid"`
	Records     []dailyAttendanceLogItemResponse `json:"records"`
	Total       int                              `json:"total"`
	Page        int                              `json:"page"`
	PageSize    int                              `json:"pageSize"`
	TotalPages  int                              `json:"totalPages"`
}

type workingHoursByDayResponse struct {
	Date        string  `json:"date"`
	WorkedHours float64 `json:"workedHours"`
}

type monthlyAttendanceStatsResponse struct {
	EmployeeUID          string                      `json:"employeeUid"`
	Month                string                      `json:"month"`
	TotalWorkedHours     float64                     `json:"totalWorkedHours"`
	AverageCheckInTime   *string                     `json:"averageCheckInTime,omitempty"`
	MissingCheckOutCount int                         `json:"missingCheckOutCount"`
	WorkingHoursByDay    []workingHoursByDayResponse `json:"workingHoursByDay"`
}

func (h *AttendanceHandler) ListDepartmentLogs(w http.ResponseWriter, r *http.Request) {
	departmentUID := r.PathValue("departmentUid")
	if departmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	query, err := parseAttendanceLogsListQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := usecases.ListDepartmentAttendanceLogsInput{
		DepartmentUID: departmentUID,
		ListParams:    query.Params,
	}

	if err := applyAttendanceLogsFilters(query.Filters, &input.EmployeeUID, &input.EmployeeName, &input.EmployeeNameMode, &input.DeviceUID, &input.PunchType, &input.StartDate, &input.EndDate); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.listDepartmentLogsUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	logs := make([]attendanceLogItemResponse, 0, len(output.Logs))
	for _, item := range output.Logs {
		logs = append(logs, attendanceLogItemResponse{
			UID:           item.UID,
			EmployeeUID:   item.EmployeeUID,
			EmployeeName:  item.EmployeeName,
			DepartmentUID: item.DepartmentUID,
			DeviceUID:     item.DeviceUID,
			DeviceName:    item.DeviceName,
			DeviceUserID:  item.DeviceUserID,
			PunchedAt:     item.PunchedAt.Format(time.RFC3339),
			PunchType:     string(item.PunchType),
			RawPayload:    item.RawPayload,
		})
	}

	writeJSON(w, http.StatusOK, listDepartmentLogsResponse{
		DepartmentUID: output.DepartmentUID,
		Logs:          logs,
		Total:         output.Total,
		Page:          output.Page,
		PageSize:      output.PageSize,
		TotalPages:    output.TotalPages,
	})
}

func (h *AttendanceHandler) ListEmployeeLogs(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	query, err := parseAttendanceLogsListQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := usecases.ListEmployeeAttendanceLogsInput{
		EmployeeUID: employeeUID,
		ListParams:  query.Params,
	}

	var ignoredEmployeeUID *string
	if err := applyAttendanceLogsFilters(query.Filters, &ignoredEmployeeUID, &input.EmployeeName, &input.EmployeeNameMode, &input.DeviceUID, &input.PunchType, &input.StartDate, &input.EndDate); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.listEmployeeLogsUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	logs := make([]attendanceLogItemResponse, 0, len(output.Logs))
	for _, item := range output.Logs {
		logs = append(logs, attendanceLogItemResponse{
			UID:           item.UID,
			EmployeeUID:   item.EmployeeUID,
			EmployeeName:  item.EmployeeName,
			DepartmentUID: item.DepartmentUID,
			DeviceUID:     item.DeviceUID,
			DeviceName:    item.DeviceName,
			DeviceUserID:  item.DeviceUserID,
			PunchedAt:     item.PunchedAt.Format(time.RFC3339),
			PunchType:     string(item.PunchType),
			RawPayload:    item.RawPayload,
		})
	}

	writeJSON(w, http.StatusOK, listEmployeeLogsResponse{
		EmployeeUID: output.EmployeeUID,
		Logs:        logs,
		Total:       output.Total,
		Page:        output.Page,
		PageSize:    output.PageSize,
		TotalPages:  output.TotalPages,
	})
}

func (h *AttendanceHandler) ListDailyDepartmentLogs(w http.ResponseWriter, r *http.Request) {
	departmentUID := r.PathValue("departmentUid")
	if departmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	query, err := parseDailyAttendanceLogsListQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := usecases.ListDailyDepartmentAttendanceLogsInput{
		DepartmentUID: departmentUID,
		ListParams:    query.Params,
	}
	if err := applyAttendanceLogsFilters(query.Filters, &input.EmployeeUID, &input.EmployeeName, &input.EmployeeNameMode, &input.DeviceUID, &input.PunchType, &input.StartDate, &input.EndDate); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.listDailyDepartmentLogsUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, listDailyDepartmentLogsResponse{
		DepartmentUID: output.DepartmentUID,
		Records:       buildDailyAttendanceLogResponses(output.Records),
		Total:         output.Total,
		Page:          output.Page,
		PageSize:      output.PageSize,
		TotalPages:    output.TotalPages,
	})
}

func (h *AttendanceHandler) ListDailyEmployeeLogs(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	query, err := parseDailyAttendanceLogsListQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := usecases.ListDailyEmployeeAttendanceLogsInput{
		EmployeeUID: employeeUID,
		ListParams:  query.Params,
	}
	var ignoredEmployeeUID *string
	if err := applyAttendanceLogsFilters(query.Filters, &ignoredEmployeeUID, &input.EmployeeName, &input.EmployeeNameMode, &input.DeviceUID, &input.PunchType, &input.StartDate, &input.EndDate); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.listDailyEmployeeLogsUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, listDailyEmployeeLogsResponse{
		EmployeeUID: output.EmployeeUID,
		Records:     buildDailyAttendanceLogResponses(output.Records),
		Total:       output.Total,
		Page:        output.Page,
		PageSize:    output.PageSize,
		TotalPages:  output.TotalPages,
	})
}

func (h *AttendanceHandler) GetDailySummary(w http.ResponseWriter, r *http.Request) {
	date := time.Now()
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			return
		}
		date = parsed
	}

	var employeeUID *string
	if uid := r.URL.Query().Get("employeeUid"); uid != "" {
		employeeUID = &uid
	}

	output, err := h.getDailySummaryUC.Execute(r.Context(), usecases.GetDailyAttendanceSummaryInput{
		Date:        date,
		EmployeeUID: employeeUID,
	})
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	summaries := make([]dailySummaryItemResponse, 0, len(output.Summaries))
	for _, item := range output.Summaries {
		var checkIn *string
		var checkOut *string
		if item.CheckIn != nil {
			v := item.CheckIn.Format(time.RFC3339)
			checkIn = &v
		}
		if item.CheckOut != nil {
			v := item.CheckOut.Format(time.RFC3339)
			checkOut = &v
		}

		summaries = append(summaries, dailySummaryItemResponse{
			EmployeeUID:     item.EmployeeUID,
			EmployeeName:    item.EmployeeName,
			CheckIn:         checkIn,
			CheckOut:        checkOut,
			WorkedHours:     item.WorkedHours,
			LateArrival:     item.LateArrival,
			EarlyDeparture:  item.EarlyDeparture,
			MissingCheckIn:  item.MissingCheckIn,
			MissingCheckOut: item.MissingCheckOut,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"date":      output.Date.Format("2006-01-02"),
		"summaries": summaries,
	})
}

func (h *AttendanceHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	output, err := h.getMonthlyStatsUC.Execute(r.Context(), usecases.GetMonthlyAttendanceStatsInput{
		EmployeeUID: employeeUID,
	})
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	var averageCheckInTime *string
	if output.AverageCheckInTime != nil {
		v := output.AverageCheckInTime.Format("15:04:05")
		averageCheckInTime = &v
	}

	workingHoursByDay := make([]workingHoursByDayResponse, 0, len(output.WorkingHoursByDay))
	for _, item := range output.WorkingHoursByDay {
		workingHoursByDay = append(workingHoursByDay, workingHoursByDayResponse{
			Date:        item.Date.Format("2006-01-02"),
			WorkedHours: item.WorkedHours,
		})
	}

	writeJSON(w, http.StatusOK, monthlyAttendanceStatsResponse{
		EmployeeUID:          employeeUID,
		Month:                output.Month,
		TotalWorkedHours:     output.TotalWorkedHours,
		AverageCheckInTime:   averageCheckInTime,
		MissingCheckOutCount: output.MissingCheckOutCount,
		WorkingHoursByDay:    workingHoursByDay,
	})
}

func (h *AttendanceHandler) GetWorkHours(w http.ResponseWriter, r *http.Request) {
	output, err := h.getWorkHoursUC.Execute(r.Context())
	if err != nil {
		slog.Error("attendance_handler.GetWorkHours.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get work hours config")
		return
	}

	writeJSON(w, http.StatusOK, workHoursResponse{
		WorkDayStart:      output.Config.WorkDayStart,
		WorkDayEnd:        output.Config.WorkDayEnd,
		LateGraceMinutes:  output.Config.LateGraceMinutes,
		EarlyGraceMinutes: output.Config.EarlyGraceMinutes,
	})
}

func (h *AttendanceHandler) SetWorkHours(w http.ResponseWriter, r *http.Request) {
	var req workHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cfg, err := h.setWorkHoursUC.Execute(r.Context(), usecases.SetWorkHoursConfigInput{
		WorkDayStart:      req.WorkDayStart,
		WorkDayEnd:        req.WorkDayEnd,
		LateGraceMinutes:  req.LateGraceMinutes,
		EarlyGraceMinutes: req.EarlyGraceMinutes,
	})
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, workHoursResponse{
		WorkDayStart:      cfg.WorkDayStart,
		WorkDayEnd:        cfg.WorkDayEnd,
		LateGraceMinutes:  cfg.LateGraceMinutes,
		EarlyGraceMinutes: cfg.EarlyGraceMinutes,
	})
}

func (h *AttendanceHandler) writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecases.ErrDepartmentNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usecases.ErrEmployeeNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usecases.ErrInvalidWorkDayStart),
		errors.Is(err, usecases.ErrInvalidWorkDayEnd),
		errors.Is(err, usecases.ErrInvalidWorkHoursRange),
		errors.Is(err, usecases.ErrInvalidGraceMinutes):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		slog.Error("attendance_handler.writeUseCaseError", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseAttendanceListDateTime(value string, endOfDay bool) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		return parsed.Add(24*time.Hour - time.Nanosecond), nil
	}
	return parsed, nil
}

func parseAttendanceLogsListQuery(r *http.Request) (listQuery, error) {
	return parseListQuery(r, listQueryOptions{
		DefaultPageSize:  20,
		MaxPageSize:      100,
		DefaultSortBy:    "punchedAt",
		DefaultSortOrder: ports.SortOrderDesc,
		AllowedSortBy: map[string]struct{}{
			"punchedAt":    {},
			"employeeName": {},
			"punchType":    {},
		},
		AllowedFilters: map[string]struct{}{
			"employeeUid":      {},
			"employeeName":     {},
			"employeeNameMode": {},
			"deviceUid":        {},
			"punchType":        {},
			"startDate":        {},
			"endDate":          {},
		},
	})
}

func parseDailyAttendanceLogsListQuery(r *http.Request) (listQuery, error) {
	return parseListQuery(r, listQueryOptions{
		DefaultPageSize:  20,
		MaxPageSize:      100,
		DefaultSortBy:    "date",
		DefaultSortOrder: ports.SortOrderDesc,
		AllowedSortBy: map[string]struct{}{
			"date":         {},
			"employeeName": {},
		},
		AllowedFilters: map[string]struct{}{
			"employeeUid":      {},
			"employeeName":     {},
			"employeeNameMode": {},
			"deviceUid":        {},
			"punchType":        {},
			"startDate":        {},
			"endDate":          {},
		},
	})
}

func applyAttendanceLogsFilters(filters map[string]string, employeeUID, employeeName, employeeNameMode, deviceUID **string, punchType **domain.AttendancePunchType, startDate, endDate **time.Time) error {
	if employeeUIDValue, ok := filters["employeeUid"]; ok && employeeUID != nil {
		*employeeUID = &employeeUIDValue
	}
	if employeeNameValue, ok := filters["employeeName"]; ok {
		*employeeName = &employeeNameValue
	}
	if employeeNameModeValue, ok := filters["employeeNameMode"]; ok {
		switch employeeNameModeValue {
		case "equals", "contains":
			*employeeNameMode = &employeeNameModeValue
		default:
			return errors.New("invalid employeeNameMode parameter")
		}
	}
	if deviceUIDValue, ok := filters["deviceUid"]; ok {
		*deviceUID = &deviceUIDValue
	}
	if punchTypeValue, ok := filters["punchType"]; ok {
		value := domain.AttendancePunchType(punchTypeValue)
		switch value {
		case domain.AttendancePunchTypeCheckIn,
			domain.AttendancePunchTypeCheckOut,
			domain.AttendancePunchTypeBreakStart,
			domain.AttendancePunchTypeBreakEnd,
			domain.AttendancePunchTypeUnknown:
			*punchType = &value
		default:
			return errors.New("invalid punchType parameter")
		}
	}
	if startDateStr, ok := filters["startDate"]; ok {
		value, err := parseAttendanceListDateTime(startDateStr, false)
		if err != nil {
			return errors.New("invalid startDate parameter")
		}
		*startDate = &value
	}
	if endDateStr, ok := filters["endDate"]; ok {
		value, err := parseAttendanceListDateTime(endDateStr, true)
		if err != nil {
			return errors.New("invalid endDate parameter")
		}
		*endDate = &value
	}
	return nil
}

func buildDailyAttendanceLogResponses(items []usecases.DailyAttendanceLogItem) []dailyAttendanceLogItemResponse {
	records := make([]dailyAttendanceLogItemResponse, 0, len(items))
	for _, item := range items {
		var checkIn *string
		var checkOut *string
		if item.CheckIn != nil {
			v := item.CheckIn.Format("15:04:05")
			checkIn = &v
		}
		if item.CheckOut != nil {
			v := item.CheckOut.Format("15:04:05")
			checkOut = &v
		}

		records = append(records, dailyAttendanceLogItemResponse{
			Date:            item.Date.Format("2006-01-02"),
			EmployeeUID:     item.EmployeeUID,
			EmployeeName:    item.EmployeeName,
			DepartmentUID:   item.DepartmentUID,
			CheckIn:         checkIn,
			CheckOut:        checkOut,
			CheckInDevice:   item.CheckInDevice,
			CheckOutDevice:  item.CheckOutDevice,
			WorkedHours:     item.WorkedHours,
			LateArrival:     item.LateArrival,
			EarlyDeparture:  item.EarlyDeparture,
			MissingCheckIn:  item.MissingCheckIn,
			MissingCheckOut: item.MissingCheckOut,
		})
	}
	return records
}
