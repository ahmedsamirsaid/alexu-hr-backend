package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type attendanceLogHistoryExecutor interface {
	Execute(ctx context.Context, input usecases.GetAttendanceLogHistoryInput) (*usecases.GetAttendanceLogHistoryOutput, error)
}

type monthlyAttendanceStatsExecutor interface {
	Execute(ctx context.Context, input usecases.GetMonthlyAttendanceStatsInput) (*usecases.GetMonthlyAttendanceStatsOutput, error)
}

type departmentAttendanceReportExecutor interface {
	Execute(ctx context.Context, input usecases.GetDepartmentAttendanceReportInput) (*usecases.DepartmentAttendanceReportOutput, error)
}

type exportDepartmentAttendanceReportExecutor interface {
	Execute(ctx context.Context, input usecases.GetDepartmentAttendanceReportInput) (*usecases.ExportDepartmentAttendanceReportOutput, error)
}

type exportEmployeeAttendanceReportExecutor interface {
	Execute(ctx context.Context, input usecases.ExportEmployeeAttendanceReportInput) (*usecases.ExportEmployeeAttendanceReportOutput, error)
}

type AttendanceReportDependencies struct {
	GetDepartmentReportUC    departmentAttendanceReportExecutor
	ExportDepartmentReportUC exportDepartmentAttendanceReportExecutor
	ExportEmployeeReportUC   exportEmployeeAttendanceReportExecutor
}

type AttendanceHandler struct {
	listDepartmentLogsUC      *usecases.ListDepartmentAttendanceLogsUseCase
	listEmployeeLogsUC        *usecases.ListEmployeeAttendanceLogsUseCase
	listDailyDepartmentLogsUC *usecases.ListDailyDepartmentAttendanceLogsUseCase
	listDailyEmployeeLogsUC   *usecases.ListDailyEmployeeAttendanceLogsUseCase
	createLogUC               *usecases.CreateAttendanceLogUseCase
	updateLogUC               *usecases.UpdateAttendanceLogUseCase
	getLogHistoryUC           attendanceLogHistoryExecutor
	getMonthlyStatsUC         monthlyAttendanceStatsExecutor
	getDailySummaryUC         *usecases.GetDailyAttendanceSummaryUseCase
	getDepartmentReportUC     departmentAttendanceReportExecutor
	exportDepartmentReportUC  exportDepartmentAttendanceReportExecutor
	exportEmployeeReportUC    exportEmployeeAttendanceReportExecutor
}

func NewAttendanceHandler(
	listDepartmentLogsUC *usecases.ListDepartmentAttendanceLogsUseCase,
	listEmployeeLogsUC *usecases.ListEmployeeAttendanceLogsUseCase,
	listDailyDepartmentLogsUC *usecases.ListDailyDepartmentAttendanceLogsUseCase,
	listDailyEmployeeLogsUC *usecases.ListDailyEmployeeAttendanceLogsUseCase,
	createLogUC *usecases.CreateAttendanceLogUseCase,
	updateLogUC *usecases.UpdateAttendanceLogUseCase,
	getLogHistoryUC *usecases.GetAttendanceLogHistoryUseCase,
	getMonthlyStatsUC *usecases.GetMonthlyAttendanceStatsUseCase,
	getDailySummaryUC *usecases.GetDailyAttendanceSummaryUseCase,
	reportDeps ...AttendanceReportDependencies,
) *AttendanceHandler {
	handler := &AttendanceHandler{
		listDepartmentLogsUC:      listDepartmentLogsUC,
		listEmployeeLogsUC:        listEmployeeLogsUC,
		listDailyDepartmentLogsUC: listDailyDepartmentLogsUC,
		listDailyEmployeeLogsUC:   listDailyEmployeeLogsUC,
		createLogUC:               createLogUC,
		updateLogUC:               updateLogUC,
		getLogHistoryUC:           getLogHistoryUC,
		getMonthlyStatsUC:         getMonthlyStatsUC,
		getDailySummaryUC:         getDailySummaryUC,
	}
	if len(reportDeps) > 0 {
		handler.getDepartmentReportUC = reportDeps[0].GetDepartmentReportUC
		handler.exportDepartmentReportUC = reportDeps[0].ExportDepartmentReportUC
		handler.exportEmployeeReportUC = reportDeps[0].ExportEmployeeReportUC
	}
	return handler
}

type createAttendanceLogRequest struct {
	EmployeeUID string  `json:"employeeUid"`
	DeviceUID   string  `json:"deviceUid"`
	PunchedAt   string  `json:"punchedAt"`
	PunchType   string  `json:"punchType"`
	Reason      *string `json:"reason"`
}

type updateAttendanceLogRequest struct {
	DeviceUID string  `json:"deviceUid"`
	PunchedAt string  `json:"punchedAt"`
	PunchType string  `json:"punchType"`
	Reason    *string `json:"reason"`
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
	Date              string                        `json:"date"`
	EmployeeUID       string                        `json:"employeeUid"`
	EmployeeName      string                        `json:"employeeName"`
	DepartmentUID     *string                       `json:"departmentUid,omitempty"`
	CheckIn           *string                       `json:"checkIn,omitempty"`
	CheckInLogUID     *string                       `json:"checkInLogUid,omitempty"`
	CheckOut          *string                       `json:"checkOut,omitempty"`
	CheckOutLogUID    *string                       `json:"checkOutLogUid,omitempty"`
	CheckInDevice     *string                       `json:"checkInDevice,omitempty"`
	CheckInDeviceUID  *string                       `json:"checkInDeviceUid,omitempty"`
	CheckOutDevice    *string                       `json:"checkOutDevice,omitempty"`
	CheckOutDeviceUID *string                       `json:"checkOutDeviceUid,omitempty"`
	HasEditHistory    bool                          `json:"hasEditHistory"`
	WorkedHours       float64                       `json:"workedHours"`
	LateArrival       bool                          `json:"lateArrival"`
	EarlyDeparture    bool                          `json:"earlyDeparture"`
	MissingCheckIn    bool                          `json:"missingCheckIn"`
	MissingCheckOut   bool                          `json:"missingCheckOut"`
	Exceptions        []attendanceExceptionResponse `json:"exceptions"`
}

type attendanceExceptionResponse struct {
	Type         string `json:"type"`
	MinutesDelta *int   `json:"minutesDelta,omitempty"`
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

type attendanceDayBreakdownResponse struct {
	Date                  string  `json:"date"`
	CheckIn               *string `json:"check_in"`
	CheckOut              *string `json:"check_out"`
	WorkedHours           float64 `json:"worked_hours"`
	IsLate                bool    `json:"is_late"`
	LateMinutes           int     `json:"late_minutes"`
	IsEarlyDeparture      bool    `json:"is_early_departure"`
	EarlyDepartureMinutes int     `json:"early_departure_minutes"`
	IsAbsent              bool    `json:"is_absent"`
	IsOnLeave             bool    `json:"is_on_leave"`
	LeaveTypeName         *string `json:"leave_type_name"`
}

type monthlyAttendanceStatsResponse struct {
	EmployeeUID             string                           `json:"employeeUid"`
	Month                   string                           `json:"month"`
	TotalWorkedHours        float64                          `json:"totalWorkedHours"`
	AverageCheckInTime      *string                          `json:"averageCheckInTime,omitempty"`
	AverageCheckOutTime     *string                          `json:"averageCheckOutTime,omitempty"`
	MissingCheckInCount     int                              `json:"missingCheckInCount"`
	MissingCheckOutCount    int                              `json:"missingCheckOutCount"`
	WorkingHoursByDay       []workingHoursByDayResponse      `json:"workingHoursByDay"`
	LateDaysCount           int                              `json:"late_days_count"`
	EarlyDepartureDaysCount int                              `json:"early_departure_days_count"`
	AbsentDaysCount         int                              `json:"absent_days_count"`
	DaysBreakdown           []attendanceDayBreakdownResponse `json:"days_breakdown"`
}

type departmentAttendanceReportEmployeeResponse struct {
	EmployeeUID         string  `json:"employeeUid"`
	EmployeeName        string  `json:"employeeName"`
	TotalWorkingDays    int     `json:"totalWorkingDays"`
	DaysPresent         int     `json:"daysPresent"`
	DaysAbsent          int     `json:"daysAbsent"`
	LateDays            int     `json:"lateDays"`
	EarlyDepartureDays  int     `json:"earlyDepartureDays"`
	MissingCheckInDays  int     `json:"missingCheckInDays"`
	MissingCheckOutDays int     `json:"missingCheckOutDays"`
	TotalWorkedHours    float64 `json:"totalWorkedHours"`
	AverageCheckInTime  *string `json:"averageCheckInTime,omitempty"`
	AverageCheckOutTime *string `json:"averageCheckOutTime,omitempty"`
}

type departmentAttendanceReportResponse struct {
	DepartmentUID          string                                       `json:"departmentUid"`
	StartDate              string                                       `json:"startDate"`
	EndDate                string                                       `json:"endDate"`
	TotalHeadcount         int                                          `json:"totalHeadcount"`
	AverageAttendanceRate  float64                                      `json:"averageAttendanceRate"`
	AveragePunctualityRate float64                                      `json:"averagePunctualityRate"`
	Employees              []departmentAttendanceReportEmployeeResponse `json:"employees"`
}

func (h *AttendanceHandler) ListDepartmentLogs(w http.ResponseWriter, r *http.Request) {
	departmentUID := r.PathValue("departmentUid")
	if departmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	claims := GetClaims(r)
	if !canAccessDepartmentAttendance(claims, departmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
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

	claims := GetClaims(r)
	if !canAccessEmployeeAttendance(claims, employeeUID, output.DepartmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
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

	claims := GetClaims(r)
	if !canAccessDepartmentAttendance(claims, departmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
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

	claims := GetClaims(r)
	if !canAccessEmployeeAttendance(claims, employeeUID, output.DepartmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
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

func (h *AttendanceHandler) GetDepartmentReport(w http.ResponseWriter, r *http.Request) {
	departmentUID := r.PathValue("departmentUid")
	if departmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	claims := GetClaims(r)
	if !canAccessDepartmentAttendance(claims, departmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

	input, ok := h.parseDepartmentReportInput(w, r, departmentUID)
	if !ok {
		return
	}
	if h.getDepartmentReportUC == nil {
		writeError(w, http.StatusInternalServerError, "department report use case is not configured")
		return
	}

	output, err := h.getDepartmentReportUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, buildDepartmentAttendanceReportResponse(output))
}

func (h *AttendanceHandler) ExportDepartmentReport(w http.ResponseWriter, r *http.Request) {
	departmentUID := r.PathValue("departmentUid")
	if departmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	claims := GetClaims(r)
	if !canAccessDepartmentAttendance(claims, departmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

	input, ok := h.parseDepartmentReportInput(w, r, departmentUID)
	if !ok {
		return
	}
	if h.exportDepartmentReportUC == nil {
		writeError(w, http.StatusInternalServerError, "department report export use case is not configured")
		return
	}

	output, err := h.exportDepartmentReportUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

func (h *AttendanceHandler) ExportEmployeeReport(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	if !h.ensureCanAccessEmployeeAttendance(w, r, employeeUID) {
		return
	}

	input, ok := h.parseEmployeeReportInput(w, r, employeeUID)
	if !ok {
		return
	}
	if h.exportEmployeeReportUC == nil {
		writeError(w, http.StatusInternalServerError, "employee report export use case is not configured")
		return
	}

	output, err := h.exportEmployeeReportUC.Execute(r.Context(), input)
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
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

func (h *AttendanceHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	if !canManuallyManageAttendanceLogs(claims) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Only IT Managers and Admins can manually edit attendance logs")
		return
	}

	req, ok := decodeCreateAttendanceLogRequest(w, r)
	if !ok {
		return
	}

	output, err := h.createLogUC.Execute(r.Context(), usecases.CreateAttendanceLogInput{
		EmployeeUID: req.EmployeeUID,
		DeviceUID:   req.DeviceUID,
		PunchedAt:   req.PunchedAt,
		PunchType:   req.PunchType,
		EditedByUID: claims.UserUID,
		Reason:      req.Reason,
	})
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, output)
}

func (h *AttendanceHandler) UpdateLog(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	if !canManuallyManageAttendanceLogs(claims) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Only IT Managers and Admins can manually edit attendance logs")
		return
	}

	req, ok := decodeUpdateAttendanceLogRequest(w, r)
	if !ok {
		return
	}

	output, err := h.updateLogUC.Execute(r.Context(), usecases.UpdateAttendanceLogInput{
		UID:         uid,
		DeviceUID:   req.DeviceUID,
		PunchedAt:   req.PunchedAt,
		PunchType:   req.PunchType,
		EditedByUID: claims.UserUID,
		Reason:      req.Reason,
	})
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *AttendanceHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	employeeUID := r.PathValue("employeeUid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "employeeUid is required")
		return
	}

	if !h.ensureCanAccessEmployeeAttendance(w, r, employeeUID) {
		return
	}

	monthValue := strings.TrimSpace(r.URL.Query().Get("month"))
	yearValue := strings.TrimSpace(r.URL.Query().Get("year"))
	startDateValue := firstNonEmptyQueryValue(r, "startDate", "start_date")
	endDateValue := firstNonEmptyQueryValue(r, "endDate", "end_date", "end_data")

	if monthValue != "" && (yearValue != "" || startDateValue != "" || endDateValue != "") {
		writeError(w, http.StatusBadRequest, "month cannot be combined with year/startDate/endDate")
		return
	}
	if yearValue != "" && (startDateValue != "" || endDateValue != "") {
		writeError(w, http.StatusBadRequest, "year cannot be combined with startDate/endDate")
		return
	}
	if (startDateValue == "") != (endDateValue == "") {
		writeError(w, http.StatusBadRequest, "startDate and endDate must be provided together")
		return
	}

	var startDate *time.Time
	var endDate *time.Time
	var periodLabel *string

	if monthValue != "" {
		monthStart, err := time.Parse("2006-01", monthValue)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid month format, expected YYYY-MM")
			return
		}

		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
		now := time.Now()
		if monthStart.Year() == now.Year() && monthStart.Month() == now.Month() {
			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
			if monthEnd.After(todayEnd) {
				monthEnd = todayEnd
			}
		}

		startDate = &monthStart
		endDate = &monthEnd
		periodLabel = &monthValue
	} else if yearValue != "" {
		year, err := strconv.Atoi(yearValue)
		if err != nil || year < 1 || year > 9999 {
			writeError(w, http.StatusBadRequest, "invalid year format, expected YYYY")
			return
		}

		yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		yearEnd := time.Date(year, time.December, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
		now := time.Now()
		if year == now.Year() {
			todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
			if yearEnd.After(todayEnd) {
				yearEnd = todayEnd
			}
		}

		startDate = &yearStart
		endDate = &yearEnd
		periodLabel = &yearValue
	} else if startDateValue != "" && endDateValue != "" {
		start, err := parseAttendanceListDateTime(startDateValue, false)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid startDate parameter")
			return
		}
		end, err := parseAttendanceListDateTime(endDateValue, true)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid endDate parameter")
			return
		}
		if start.After(end) {
			writeError(w, http.StatusBadRequest, "startDate must be before or equal to endDate")
			return
		}

		startDate = &start
		endDate = &end
		label := fmt.Sprintf("%s to %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
		periodLabel = &label
	}

	output, err := h.getMonthlyStatsUC.Execute(r.Context(), usecases.GetMonthlyAttendanceStatsInput{
		EmployeeUID: employeeUID,
		StartDate:   startDate,
		EndDate:     endDate,
		PeriodLabel: periodLabel,
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

	var averageCheckOutTime *string
	if output.AverageCheckOutTime != nil {
		v := output.AverageCheckOutTime.Format("15:04:05")
		averageCheckOutTime = &v
	}

	workingHoursByDay := make([]workingHoursByDayResponse, 0, len(output.WorkingHoursByDay))
	for _, item := range output.WorkingHoursByDay {
		workingHoursByDay = append(workingHoursByDay, workingHoursByDayResponse{
			Date:        item.Date.Format("2006-01-02"),
			WorkedHours: item.WorkedHours,
		})
	}

	daysBreakdown := buildAttendanceDayBreakdownResponses(output.DaysBreakdown)

	writeJSON(w, http.StatusOK, monthlyAttendanceStatsResponse{
		EmployeeUID:             employeeUID,
		Month:                   output.Month,
		TotalWorkedHours:        output.TotalWorkedHours,
		AverageCheckInTime:      averageCheckInTime,
		AverageCheckOutTime:     averageCheckOutTime,
		MissingCheckInCount:     output.MissingCheckInCount,
		MissingCheckOutCount:    output.MissingCheckOutCount,
		WorkingHoursByDay:       workingHoursByDay,
		LateDaysCount:           output.LateDaysCount,
		EarlyDepartureDaysCount: output.EarlyDepartureDaysCount,
		AbsentDaysCount:         output.AbsentDaysCount,
		DaysBreakdown:           daysBreakdown,
	})
}

func (h *AttendanceHandler) GetLogHistory(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.getLogHistoryUC.Execute(r.Context(), usecases.GetAttendanceLogHistoryInput{
		AttendanceLogUID: uid,
	})
	if err != nil {
		h.writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *AttendanceHandler) writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecases.ErrDepartmentNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usecases.ErrEmployeeNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usecases.ErrAttendanceDeviceNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usecases.ErrAttendanceLogNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, usecases.ErrInvalidWorkDayStart),
		errors.Is(err, usecases.ErrInvalidWorkDayEnd),
		errors.Is(err, usecases.ErrInvalidWorkHoursRange),
		errors.Is(err, usecases.ErrInvalidGraceMinutes),
		errors.Is(err, usecases.ErrInvalidReportDateRange),
		errors.Is(err, usecases.ErrAttendanceLogEmployeeRequired),
		errors.Is(err, usecases.ErrAttendanceLogDeviceRequired),
		errors.Is(err, usecases.ErrAttendanceLogDeviceUserIDReq),
		errors.Is(err, usecases.ErrAttendanceLogPunchedAtReq),
		errors.Is(err, usecases.ErrAttendanceLogPunchTypeInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, usecases.ErrAttendanceLogConflict):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, usecases.ErrAttendanceLogCheckoutBeforeCheckin):
		writeError(w, http.StatusPreconditionFailed, err.Error())
	default:
		slog.Error("attendance_handler.writeUseCaseError", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func canAccessRequestedEmployee(claims *JWTClaims, employeeUID string) bool {
	if claims == nil {
		return false
	}
	if claims.HasPermission("*") {
		return true
	}
	if claims.IsGlobalScope() {
		return true
	}
	if claims.IsDepartmentScope() {
		return true
	}
	return claims.EmployeeUID != nil && *claims.EmployeeUID == employeeUID
}

func (h *AttendanceHandler) ensureCanAccessEmployeeAttendance(w http.ResponseWriter, r *http.Request, employeeUID string) bool {
	claims := GetClaims(r)
	if !canAccessRequestedEmployee(claims, employeeUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
		return false
	}

	if claims != nil && !claims.HasPermission("*") && claims.IsDepartmentScope() {
		if h.listEmployeeLogsUC == nil {
			writeError(w, http.StatusInternalServerError, "employee access use case is not configured")
			return false
		}
		accessOutput, err := h.listEmployeeLogsUC.Execute(r.Context(), usecases.ListEmployeeAttendanceLogsInput{
			EmployeeUID: employeeUID,
			ListParams: ports.ListParams{
				Page:     1,
				PageSize: 1,
			},
		})
		if err != nil {
			h.writeUseCaseError(w, err)
			return false
		}
		if accessOutput.DepartmentUID == nil || !claims.HasDepartmentAccess(*accessOutput.DepartmentUID) {
			writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
			return false
		}
	}

	return true
}

func canAccessEmployeeAttendance(claims *JWTClaims, employeeUID string, departmentUID *string) bool {
	if claims == nil {
		return false
	}
	if claims.HasPermission("*") {
		return true
	}
	if claims.IsGlobalScope() {
		return true
	}
	if claims.IsDepartmentScope() {
		return departmentUID != nil && claims.HasDepartmentAccess(*departmentUID)
	}
	return claims.EmployeeUID != nil && *claims.EmployeeUID == employeeUID
}

func canAccessDepartmentAttendance(claims *JWTClaims, departmentUID string) bool {
	if claims == nil {
		return false
	}
	if claims.HasPermission("*") {
		return true
	}
	if claims.IsGlobalScope() {
		return true
	}
	if claims.IsDepartmentScope() {
		return claims.HasDepartmentAccess(departmentUID)
	}
	return false
}

func canManuallyManageAttendanceLogs(claims *JWTClaims) bool {
	if claims == nil {
		return false
	}
	if !claims.HasPermission("attendance:write") {
		return false
	}
	if claims.HasPermission("*") {
		return true
	}
	return claims.HasAnyRole("admin", "it_manager")
}

func decodeCreateAttendanceLogRequest(w http.ResponseWriter, r *http.Request) (*parsedCreateAttendanceLogRequest, bool) {
	var req createAttendanceLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return nil, false
	}

	punchedAt, err := time.Parse(time.RFC3339, req.PunchedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid punchedAt format, expected RFC3339")
		return nil, false
	}

	return &parsedCreateAttendanceLogRequest{
		EmployeeUID: req.EmployeeUID,
		DeviceUID:   req.DeviceUID,
		PunchedAt:   punchedAt,
		PunchType:   req.PunchType,
		Reason:      normalizeOptionalRequestString(req.Reason),
	}, true
}

func decodeUpdateAttendanceLogRequest(w http.ResponseWriter, r *http.Request) (*parsedUpdateAttendanceLogRequest, bool) {
	var req updateAttendanceLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return nil, false
	}

	punchedAt, err := time.Parse(time.RFC3339, req.PunchedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid punchedAt format, expected RFC3339")
		return nil, false
	}

	return &parsedUpdateAttendanceLogRequest{
		DeviceUID: req.DeviceUID,
		PunchedAt: punchedAt,
		PunchType: req.PunchType,
		Reason:    normalizeOptionalRequestString(req.Reason),
	}, true
}

type parsedCreateAttendanceLogRequest struct {
	EmployeeUID string
	DeviceUID   string
	PunchedAt   time.Time
	PunchType   string
	Reason      *string
}

type parsedUpdateAttendanceLogRequest struct {
	DeviceUID string
	PunchedAt time.Time
	PunchType string
	Reason    *string
}

func normalizeOptionalRequestString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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

func (h *AttendanceHandler) parseDepartmentReportInput(w http.ResponseWriter, r *http.Request, departmentUID string) (usecases.GetDepartmentAttendanceReportInput, bool) {
	startDateValue := firstNonEmptyQueryValue(r, "start_date", "startDate")
	endDateValue := firstNonEmptyQueryValue(r, "end_date", "end_data", "endDate")

	if startDateValue == "" || endDateValue == "" {
		writeError(w, http.StatusBadRequest, "start_date and end_date are required")
		return usecases.GetDepartmentAttendanceReportInput{}, false
	}

	startDate, err := time.Parse("2006-01-02", startDateValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_date format, expected YYYY-MM-DD")
		return usecases.GetDepartmentAttendanceReportInput{}, false
	}
	endDate, err := time.Parse("2006-01-02", endDateValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_date format, expected YYYY-MM-DD")
		return usecases.GetDepartmentAttendanceReportInput{}, false
	}
	if endDate.Before(startDate) {
		writeError(w, http.StatusBadRequest, "start_date must be before or equal to end_date")
		return usecases.GetDepartmentAttendanceReportInput{}, false
	}

	return usecases.GetDepartmentAttendanceReportInput{
		DepartmentUID: departmentUID,
		StartDate:     startDate,
		EndDate:       endDate,
	}, true
}

func (h *AttendanceHandler) parseEmployeeReportInput(w http.ResponseWriter, r *http.Request, employeeUID string) (usecases.ExportEmployeeAttendanceReportInput, bool) {
	startDateValue := firstNonEmptyQueryValue(r, "start_date", "startDate")
	endDateValue := firstNonEmptyQueryValue(r, "end_date", "end_data", "endDate")

	if startDateValue == "" || endDateValue == "" {
		writeError(w, http.StatusBadRequest, "start_date and end_date are required")
		return usecases.ExportEmployeeAttendanceReportInput{}, false
	}

	startDate, err := time.Parse("2006-01-02", startDateValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid start_date format, expected YYYY-MM-DD")
		return usecases.ExportEmployeeAttendanceReportInput{}, false
	}
	endDate, err := time.Parse("2006-01-02", endDateValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid end_date format, expected YYYY-MM-DD")
		return usecases.ExportEmployeeAttendanceReportInput{}, false
	}
	if endDate.Before(startDate) {
		writeError(w, http.StatusBadRequest, "start_date must be before or equal to end_date")
		return usecases.ExportEmployeeAttendanceReportInput{}, false
	}

	return usecases.ExportEmployeeAttendanceReportInput{
		EmployeeUID: employeeUID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, true
}

func firstNonEmptyQueryValue(r *http.Request, names ...string) string {
	for _, name := range names {
		value := strings.TrimSpace(r.URL.Query().Get(name))
		if value != "" {
			return value
		}
	}
	return ""
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
	if startDate != nil && endDate != nil && *startDate != nil && *endDate != nil {
		if (*startDate).After(**endDate) {
			return errors.New("startDate must be before or equal to endDate")
		}
	}
	return nil
}

func buildDailyAttendanceLogResponses(items []usecases.DailyAttendanceLogItem) []dailyAttendanceLogItemResponse {
	records := make([]dailyAttendanceLogItemResponse, 0, len(items))
	for _, item := range items {
		var checkIn *string
		var checkOut *string
		if item.CheckIn != nil {
			v := item.CheckIn.Format("15:04:05Z07:00")
			checkIn = &v
		}
		if item.CheckOut != nil {
			v := item.CheckOut.Format("15:04:05Z07:00")
			checkOut = &v
		}

		records = append(records, dailyAttendanceLogItemResponse{
			Date:              item.Date.Format("2006-01-02"),
			EmployeeUID:       item.EmployeeUID,
			EmployeeName:      item.EmployeeName,
			DepartmentUID:     item.DepartmentUID,
			CheckIn:           checkIn,
			CheckInLogUID:     item.CheckInLogUID,
			CheckOut:          checkOut,
			CheckOutLogUID:    item.CheckOutLogUID,
			CheckInDevice:     item.CheckInDevice,
			CheckInDeviceUID:  item.CheckInDeviceUID,
			CheckOutDevice:    item.CheckOutDevice,
			CheckOutDeviceUID: item.CheckOutDeviceUID,
			HasEditHistory:    item.HasEditHistory,
			WorkedHours:       item.WorkedHours,
			LateArrival:       item.LateArrival,
			EarlyDeparture:    item.EarlyDeparture,
			MissingCheckIn:    item.MissingCheckIn,
			MissingCheckOut:   item.MissingCheckOut,
			Exceptions:        buildAttendanceExceptionResponses(item),
		})
	}
	return records
}

func buildAttendanceDayBreakdownResponses(items []usecases.AttendanceDayBreakdown) []attendanceDayBreakdownResponse {
	records := make([]attendanceDayBreakdownResponse, 0, len(items))
	for _, item := range items {
		var checkIn *string
		if item.CheckIn != nil {
			value := item.CheckIn.Format("15:04:05Z07:00")
			checkIn = &value
		}

		var checkOut *string
		if item.CheckOut != nil {
			value := item.CheckOut.Format("15:04:05Z07:00")
			checkOut = &value
		}

		records = append(records, attendanceDayBreakdownResponse{
			Date:                  item.Date.Format("2006-01-02"),
			CheckIn:               checkIn,
			CheckOut:              checkOut,
			WorkedHours:           item.WorkedHours,
			IsLate:                item.IsLate,
			LateMinutes:           item.LateMinutes,
			IsEarlyDeparture:      item.IsEarlyDeparture,
			EarlyDepartureMinutes: item.EarlyDepartureMinutes,
			IsAbsent:              item.IsAbsent,
			IsOnLeave:             item.IsOnLeave,
			LeaveTypeName:         item.LeaveTypeName,
		})
	}
	return records
}

func buildDepartmentAttendanceReportResponse(output *usecases.DepartmentAttendanceReportOutput) departmentAttendanceReportResponse {
	employees := make([]departmentAttendanceReportEmployeeResponse, 0, len(output.Employees))
	for _, employee := range output.Employees {
		var averageCheckInTime *string
		if employee.AverageCheckInTime != nil {
			value := employee.AverageCheckInTime.Format("15:04:05")
			averageCheckInTime = &value
		}

		var averageCheckOutTime *string
		if employee.AverageCheckOutTime != nil {
			value := employee.AverageCheckOutTime.Format("15:04:05")
			averageCheckOutTime = &value
		}

		employees = append(employees, departmentAttendanceReportEmployeeResponse{
			EmployeeUID:         employee.EmployeeUID,
			EmployeeName:        employee.EmployeeName,
			TotalWorkingDays:    employee.TotalWorkingDays,
			DaysPresent:         employee.DaysPresent,
			DaysAbsent:          employee.DaysAbsent,
			LateDays:            employee.LateDays,
			EarlyDepartureDays:  employee.EarlyDepartureDays,
			MissingCheckInDays:  employee.MissingCheckInDays,
			MissingCheckOutDays: employee.MissingCheckOutDays,
			TotalWorkedHours:    employee.TotalWorkedHours,
			AverageCheckInTime:  averageCheckInTime,
			AverageCheckOutTime: averageCheckOutTime,
		})
	}

	return departmentAttendanceReportResponse{
		DepartmentUID:          output.DepartmentUID,
		StartDate:              output.StartDate.Format("2006-01-02"),
		EndDate:                output.EndDate.Format("2006-01-02"),
		TotalHeadcount:         output.TotalHeadcount,
		AverageAttendanceRate:  output.AverageAttendanceRate,
		AveragePunctualityRate: output.AveragePunctualityRate,
		Employees:              employees,
	}
}

func buildAttendanceExceptionResponses(item usecases.DailyAttendanceLogItem) []attendanceExceptionResponse {
	if len(item.Exceptions) == 0 {
		return []attendanceExceptionResponse{}
	}

	result := make([]attendanceExceptionResponse, 0, len(item.Exceptions))
	for _, value := range item.Exceptions {
		exception := attendanceExceptionResponse{Type: string(value)}
		switch value {
		case domain.AttendanceExceptionTypeLateArrival:
			exception.MinutesDelta = item.LateMinutes
		case domain.AttendanceExceptionTypeEarlyDeparture:
			exception.MinutesDelta = item.EarlyMinutes
		}
		result = append(result, exception)
	}

	return result
}
