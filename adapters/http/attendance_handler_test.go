package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type stubMonthlyAttendanceStatsUseCase struct {
	output *usecases.GetMonthlyAttendanceStatsOutput
	err    error
	input  *usecases.GetMonthlyAttendanceStatsInput
}

func (s *stubMonthlyAttendanceStatsUseCase) Execute(ctx context.Context, input usecases.GetMonthlyAttendanceStatsInput) (*usecases.GetMonthlyAttendanceStatsOutput, error) {
	s.input = &input
	return s.output, s.err
}

type mockDepartmentRepoForAttendance struct {
	department *domain.Department
}

func (m *mockDepartmentRepoForAttendance) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Department, error) {
	return nil, nil
}

func (m *mockDepartmentRepoForAttendance) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Department, error) {
	if m.department != nil && m.department.UID == uid {
		return m.department, nil
	}
	return nil, nil
}

func (m *mockDepartmentRepoForAttendance) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.Department, error) {
	return nil, nil
}

func (m *mockDepartmentRepoForAttendance) Create(ctx context.Context, q ports.Querier, department *domain.Department) error {
	return nil
}

func (m *mockDepartmentRepoForAttendance) Update(ctx context.Context, q ports.Querier, department *domain.Department) error {
	return nil
}

func (m *mockDepartmentRepoForAttendance) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.Department, error) {
	return nil, nil
}

type mockAttendanceRecordRepo struct {
	countReturn     int
	listReturn      []*ports.AttendanceRecordWithEmployee
	dailyListReturn []*ports.DailyAttendanceGroup
	dateRangeReturn []*domain.AttendanceRecord
	gotDeptUID      string
	gotEmployeeUID  string
	gotFilter       ports.DepartmentAttendanceLogsFilter
	gotListParams   ports.ListParams
}

func (m *mockAttendanceRecordRepo) Create(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) (bool, error) {
	return false, nil
}

func (m *mockAttendanceRecordRepo) ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepo) ListByDateRange(ctx context.Context, q ports.Querier, startDate, endDate time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return m.dateRangeReturn, nil
}

func (m *mockAttendanceRecordRepo) ListByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	m.gotDeptUID = departmentUID
	m.gotFilter = filter
	m.gotListParams = params
	return m.listReturn, nil
}

func (m *mockAttendanceRecordRepo) CountByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	m.gotDeptUID = departmentUID
	m.gotFilter = filter
	return m.countReturn, nil
}

func (m *mockAttendanceRecordRepo) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	m.gotEmployeeUID = employeeUID
	m.gotFilter = filter
	m.gotListParams = params
	return m.listReturn, nil
}

func (m *mockAttendanceRecordRepo) CountByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	m.gotEmployeeUID = employeeUID
	m.gotFilter = filter
	return m.countReturn, nil
}

func (m *mockAttendanceRecordRepo) ListDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	m.gotDeptUID = departmentUID
	m.gotFilter = filter
	m.gotListParams = params
	return m.dailyListReturn, nil
}

func (m *mockAttendanceRecordRepo) CountDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	m.gotDeptUID = departmentUID
	m.gotFilter = filter
	return m.countReturn, nil
}

func (m *mockAttendanceRecordRepo) ListDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	m.gotEmployeeUID = employeeUID
	m.gotFilter = filter
	m.gotListParams = params
	return m.dailyListReturn, nil
}

func (m *mockAttendanceRecordRepo) CountDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	m.gotEmployeeUID = employeeUID
	m.gotFilter = filter
	return m.countReturn, nil
}

func (m *mockAttendanceRecordRepo) ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q ports.Querier, deviceUserID string) (*string, error) {
	return nil, nil
}

type mockEmployeeRepoForAttendance struct {
	employee *domain.Employee
}

func (m *mockEmployeeRepoForAttendance) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForAttendance) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	if m.employee != nil && m.employee.UID == uid {
		return m.employee, nil
	}
	return nil, nil
}

func (m *mockEmployeeRepoForAttendance) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForAttendance) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForAttendance) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForAttendance) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForAttendance) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForAttendance) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForAttendance) Count(ctx context.Context, q ports.Querier) (int, error) {
	return 0, nil
}

type mockShiftRepoForAttendance struct {
	shift *domain.Shift
}

func (m *mockShiftRepoForAttendance) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Shift, error) {
	if m.shift != nil && m.shift.UID == uid {
		return m.shift, nil
	}
	return nil, nil
}

func (m *mockShiftRepoForAttendance) List(ctx context.Context, q ports.Querier) ([]*domain.Shift, error) {
	if m.shift == nil {
		return nil, nil
	}
	return []*domain.Shift{m.shift}, nil
}

func (m *mockShiftRepoForAttendance) Upsert(ctx context.Context, q ports.Querier, shift *domain.Shift) error {
	return nil
}

func TestAttendanceHandlerListDepartmentLogs(t *testing.T) {
	recordRepo := &mockAttendanceRecordRepo{
		countReturn: 2,
		listReturn: []*ports.AttendanceRecordWithEmployee{
			{
				Record: &domain.AttendanceRecord{
					UID:          "atr_1",
					EmployeeUID:  "emp_1",
					DeviceUID:    "dev_1",
					DeviceUserID: "1001",
					PunchedAt:    time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC),
					PunchType:    domain.AttendancePunchTypeCheckIn,
				},
				EmployeeName: "Alice",
				DepartmentUID: func() *string {
					v := "dept_1"
					return &v
				}(),
				DeviceName: "Front Gate",
			},
		},
	}

	listUC := usecases.NewListDepartmentAttendanceLogsUseCase(
		&mockDB{},
		&mockDepartmentRepoForAttendance{department: &domain.Department{UID: "dept_1"}},
		recordRepo,
	)

	handler := NewAttendanceHandler(listUC, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/departments/dept_1/logs?page=2&pageSize=1&sortBy=employeeName&sortOrder=asc&employeeUid=emp_1&deviceUid=dev_1&punchType=check_in&startDate=2026-04-01&endDate=2026-04-30", nil)
	req.SetPathValue("departmentUid", "dept_1")
	rr := httptest.NewRecorder()

	handler.ListDepartmentLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	if recordRepo.gotDeptUID != "dept_1" {
		t.Fatalf("department UID = %q, want %q", recordRepo.gotDeptUID, "dept_1")
	}
	if recordRepo.gotListParams.Page != 2 {
		t.Fatalf("page = %d, want 2", recordRepo.gotListParams.Page)
	}
	if recordRepo.gotListParams.PageSize != 1 {
		t.Fatalf("pageSize = %d, want 1", recordRepo.gotListParams.PageSize)
	}
	if recordRepo.gotListParams.SortBy != "employeeName" {
		t.Fatalf("sortBy = %q, want %q", recordRepo.gotListParams.SortBy, "employeeName")
	}
	if recordRepo.gotListParams.SortOrder != ports.SortOrderAsc {
		t.Fatalf("sortOrder = %q, want %q", recordRepo.gotListParams.SortOrder, ports.SortOrderAsc)
	}
	if recordRepo.gotFilter.EmployeeUID == nil || *recordRepo.gotFilter.EmployeeUID != "emp_1" {
		t.Fatalf("employeeUid filter = %v, want emp_1", recordRepo.gotFilter.EmployeeUID)
	}
	if recordRepo.gotFilter.DeviceUID == nil || *recordRepo.gotFilter.DeviceUID != "dev_1" {
		t.Fatalf("deviceUid filter = %v, want dev_1", recordRepo.gotFilter.DeviceUID)
	}
	if recordRepo.gotFilter.PunchType == nil || *recordRepo.gotFilter.PunchType != domain.AttendancePunchTypeCheckIn {
		t.Fatalf("punchType filter = %v, want %q", recordRepo.gotFilter.PunchType, domain.AttendancePunchTypeCheckIn)
	}
	if recordRepo.gotFilter.StartDate == nil || recordRepo.gotFilter.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("startDate filter = %v, want 2026-04-01", recordRepo.gotFilter.StartDate)
	}
	if recordRepo.gotFilter.EndDate == nil || recordRepo.gotFilter.EndDate.Format("2006-01-02 15:04:05") != "2026-04-30 23:59:59" {
		t.Fatalf("endDate filter = %v, want end of 2026-04-30", recordRepo.gotFilter.EndDate)
	}

	var response struct {
		DepartmentUID string `json:"departmentUid"`
		Total         int    `json:"total"`
		Page          int    `json:"page"`
		PageSize      int    `json:"pageSize"`
		TotalPages    int    `json:"totalPages"`
		Logs          []struct {
			UID          string `json:"uid"`
			EmployeeUID  string `json:"employeeUid"`
			EmployeeName string `json:"employeeName"`
			DeviceName   string `json:"deviceName"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.DepartmentUID != "dept_1" || response.Total != 2 || response.Page != 2 || response.PageSize != 1 || response.TotalPages != 2 {
		t.Fatalf("unexpected response metadata: %+v", response)
	}
	if len(response.Logs) != 1 || response.Logs[0].UID != "atr_1" {
		t.Fatalf("unexpected logs response: %+v", response.Logs)
	}
	if response.Logs[0].DeviceName != "Front Gate" {
		t.Fatalf("deviceName = %q, want %q", response.Logs[0].DeviceName, "Front Gate")
	}
}

func TestAttendanceHandlerListDepartmentLogsWithEmployeeNameEquals(t *testing.T) {
	recordRepo := &mockAttendanceRecordRepo{}

	listUC := usecases.NewListDepartmentAttendanceLogsUseCase(
		&mockDB{},
		&mockDepartmentRepoForAttendance{department: &domain.Department{UID: "dept_1"}},
		recordRepo,
	)

	handler := NewAttendanceHandler(listUC, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/departments/dept_1/logs?employeeName=Alice&employeeNameMode=equals", nil)
	req.SetPathValue("departmentUid", "dept_1")
	rr := httptest.NewRecorder()

	handler.ListDepartmentLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if recordRepo.gotFilter.EmployeeName == nil || *recordRepo.gotFilter.EmployeeName != "Alice" {
		t.Fatalf("employeeName filter = %v, want Alice", recordRepo.gotFilter.EmployeeName)
	}
	if recordRepo.gotFilter.EmployeeNameMode == nil || *recordRepo.gotFilter.EmployeeNameMode != "equals" {
		t.Fatalf("employeeNameMode filter = %v, want equals", recordRepo.gotFilter.EmployeeNameMode)
	}
}

func TestAttendanceHandlerListDepartmentLogsRejectsInvalidSortBy(t *testing.T) {
	handler := NewAttendanceHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/departments/dept_1/logs?sortBy=createdAt", nil)
	req.SetPathValue("departmentUid", "dept_1")
	rr := httptest.NewRecorder()

	handler.ListDepartmentLogs(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAttendanceHandlerListDepartmentLogsRejectsInvalidEmployeeNameMode(t *testing.T) {
	handler := NewAttendanceHandler(nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/departments/dept_1/logs?employeeName=Ali&employeeNameMode=startsWith", nil)
	req.SetPathValue("departmentUid", "dept_1")
	rr := httptest.NewRecorder()

	handler.ListDepartmentLogs(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAttendanceHandlerListEmployeeLogs(t *testing.T) {
	recordRepo := &mockAttendanceRecordRepo{
		countReturn: 2,
		listReturn: []*ports.AttendanceRecordWithEmployee{
			{
				Record: &domain.AttendanceRecord{
					UID:          "atr_1",
					EmployeeUID:  "emp_1",
					DeviceUID:    "dev_1",
					DeviceUserID: "1001",
					PunchedAt:    time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC),
					PunchType:    domain.AttendancePunchTypeCheckIn,
				},
				EmployeeName: "Alice",
				DepartmentUID: func() *string {
					v := "dept_1"
					return &v
				}(),
				DeviceName: "Front Gate",
			},
		},
	}

	listUC := usecases.NewListEmployeeAttendanceLogsUseCase(
		&mockDB{},
		&mockEmployeeRepoForAttendance{employee: &domain.Employee{UID: "emp_1", Name: "Alice"}},
		recordRepo,
	)

	handler := NewAttendanceHandler(nil, listUC, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/logs?page=2&pageSize=1&sortBy=employeeName&sortOrder=asc&deviceUid=dev_1&punchType=check_in&startDate=2026-04-01&endDate=2026-04-30", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.ListEmployeeLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	if recordRepo.gotEmployeeUID != "emp_1" {
		t.Fatalf("employee UID = %q, want %q", recordRepo.gotEmployeeUID, "emp_1")
	}
	if recordRepo.gotListParams.Page != 2 {
		t.Fatalf("page = %d, want 2", recordRepo.gotListParams.Page)
	}
	if recordRepo.gotListParams.PageSize != 1 {
		t.Fatalf("pageSize = %d, want 1", recordRepo.gotListParams.PageSize)
	}
	if recordRepo.gotListParams.SortBy != "employeeName" {
		t.Fatalf("sortBy = %q, want %q", recordRepo.gotListParams.SortBy, "employeeName")
	}
	if recordRepo.gotListParams.SortOrder != ports.SortOrderAsc {
		t.Fatalf("sortOrder = %q, want %q", recordRepo.gotListParams.SortOrder, ports.SortOrderAsc)
	}
	if recordRepo.gotFilter.DeviceUID == nil || *recordRepo.gotFilter.DeviceUID != "dev_1" {
		t.Fatalf("deviceUid filter = %v, want dev_1", recordRepo.gotFilter.DeviceUID)
	}
	if recordRepo.gotFilter.PunchType == nil || *recordRepo.gotFilter.PunchType != domain.AttendancePunchTypeCheckIn {
		t.Fatalf("punchType filter = %v, want %q", recordRepo.gotFilter.PunchType, domain.AttendancePunchTypeCheckIn)
	}
	if recordRepo.gotFilter.StartDate == nil || recordRepo.gotFilter.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("startDate filter = %v, want 2026-04-01", recordRepo.gotFilter.StartDate)
	}
	if recordRepo.gotFilter.EndDate == nil || recordRepo.gotFilter.EndDate.Format("2006-01-02 15:04:05") != "2026-04-30 23:59:59" {
		t.Fatalf("endDate filter = %v, want end of 2026-04-30", recordRepo.gotFilter.EndDate)
	}

	var response struct {
		EmployeeUID string `json:"employeeUid"`
		Total       int    `json:"total"`
		Page        int    `json:"page"`
		PageSize    int    `json:"pageSize"`
		TotalPages  int    `json:"totalPages"`
		Logs        []struct {
			UID          string `json:"uid"`
			EmployeeUID  string `json:"employeeUid"`
			EmployeeName string `json:"employeeName"`
			DeviceName   string `json:"deviceName"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.EmployeeUID != "emp_1" || response.Total != 2 || response.Page != 2 || response.PageSize != 1 || response.TotalPages != 2 {
		t.Fatalf("unexpected response metadata: %+v", response)
	}
	if len(response.Logs) != 1 || response.Logs[0].UID != "atr_1" {
		t.Fatalf("unexpected logs response: %+v", response.Logs)
	}
	if response.Logs[0].DeviceName != "Front Gate" {
		t.Fatalf("deviceName = %q, want %q", response.Logs[0].DeviceName, "Front Gate")
	}
}

func TestAttendanceHandlerListDailyDepartmentLogs(t *testing.T) {
	recordRepo := &mockAttendanceRecordRepo{
		countReturn: 1,
		dailyListReturn: []*ports.DailyAttendanceGroup{
			{
				Date:         time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				EmployeeUID:  "emp_1",
				EmployeeName: "Alice",
				DepartmentUID: func() *string {
					v := "dept_1"
					return &v
				}(),
				CheckIn: func() *time.Time {
					v := time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC)
					return &v
				}(),
				CheckOut: func() *time.Time {
					v := time.Date(2026, 4, 10, 17, 0, 0, 0, time.UTC)
					return &v
				}(),
				CheckInDevice: func() *string {
					v := "Front Gate"
					return &v
				}(),
				CheckOutDevice: func() *string {
					v := "Back Gate"
					return &v
				}(),
			},
		},
	}

	listUC := usecases.NewListDailyDepartmentAttendanceLogsUseCase(
		&mockDB{},
		&mockDepartmentRepoForAttendance{department: &domain.Department{UID: "dept_1", DefaultShiftUID: func() *string { v := "shf_general"; return &v }()}},
		recordRepo,
		&mockEmployeeRepoForAttendance{employee: &domain.Employee{UID: "emp_1", Name: "Alice", DepartmentUID: func() *string { v := "dept_1"; return &v }()}},
		&mockShiftRepoForAttendance{shift: &domain.Shift{UID: "shf_general", StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15}},
		nil,
	)

	handler := NewAttendanceHandler(nil, nil, listUC, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/departments/dept_1/daily-logs?page=1&pageSize=10&sortBy=date&sortOrder=desc", nil)
	req.SetPathValue("departmentUid", "dept_1")
	rr := httptest.NewRecorder()

	handler.ListDailyDepartmentLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response struct {
		DepartmentUID string `json:"departmentUid"`
		Total         int    `json:"total"`
		Page          int    `json:"page"`
		PageSize      int    `json:"pageSize"`
		TotalPages    int    `json:"totalPages"`
		Records       []struct {
			Date           string  `json:"date"`
			EmployeeUID    string  `json:"employeeUid"`
			CheckIn        *string `json:"checkIn"`
			CheckOut       *string `json:"checkOut"`
			CheckInDevice  *string `json:"checkInDevice"`
			CheckOutDevice *string `json:"checkOutDevice"`
		} `json:"records"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.DepartmentUID != "dept_1" || response.Total != 1 || response.TotalPages != 1 {
		t.Fatalf("unexpected response metadata: %+v", response)
	}
	if len(response.Records) != 1 || response.Records[0].Date != "2026-04-10" || response.Records[0].EmployeeUID != "emp_1" {
		t.Fatalf("unexpected records: %+v", response.Records)
	}
	if response.Records[0].CheckIn == nil || *response.Records[0].CheckIn != "08:30:00" {
		t.Fatalf("checkIn = %v, want 08:30:00", response.Records[0].CheckIn)
	}
	if response.Records[0].CheckOut == nil || *response.Records[0].CheckOut != "17:00:00" {
		t.Fatalf("checkOut = %v, want 17:00:00", response.Records[0].CheckOut)
	}
	if response.Records[0].CheckInDevice == nil || *response.Records[0].CheckInDevice != "Front Gate" {
		t.Fatalf("checkInDevice = %v, want Front Gate", response.Records[0].CheckInDevice)
	}
	if response.Records[0].CheckOutDevice == nil || *response.Records[0].CheckOutDevice != "Back Gate" {
		t.Fatalf("checkOutDevice = %v, want Back Gate", response.Records[0].CheckOutDevice)
	}
}

func TestAttendanceHandlerListDailyEmployeeLogs(t *testing.T) {
	recordRepo := &mockAttendanceRecordRepo{
		countReturn: 1,
		dailyListReturn: []*ports.DailyAttendanceGroup{
			{
				Date:         time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
				EmployeeUID:  "emp_1",
				EmployeeName: "Alice",
				DepartmentUID: func() *string {
					v := "dept_1"
					return &v
				}(),
				CheckIn: func() *time.Time {
					v := time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC)
					return &v
				}(),
				CheckOut: func() *time.Time {
					v := time.Date(2026, 4, 10, 17, 0, 0, 0, time.UTC)
					return &v
				}(),
				CheckInDevice: func() *string {
					v := "Front Gate"
					return &v
				}(),
				CheckOutDevice: func() *string {
					v := "Back Gate"
					return &v
				}(),
			},
		},
	}

	listUC := usecases.NewListDailyEmployeeAttendanceLogsUseCase(
		&mockDB{},
		&mockEmployeeRepoForAttendance{employee: &domain.Employee{UID: "emp_1", Name: "Alice", DepartmentUID: func() *string { v := "dept_1"; return &v }()}},
		recordRepo,
		&mockDepartmentRepoForAttendance{department: &domain.Department{UID: "dept_1", DefaultShiftUID: func() *string { v := "shf_general"; return &v }()}},
		&mockShiftRepoForAttendance{shift: &domain.Shift{UID: "shf_general", StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15}},
		nil,
	)

	handler := NewAttendanceHandler(nil, nil, nil, listUC, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/daily-logs?page=1&pageSize=10&sortBy=date&sortOrder=desc", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.ListDailyEmployeeLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response struct {
		EmployeeUID string `json:"employeeUid"`
		Total       int    `json:"total"`
		Records     []struct {
			Date           string  `json:"date"`
			EmployeeUID    string  `json:"employeeUid"`
			CheckIn        *string `json:"checkIn"`
			CheckOut       *string `json:"checkOut"`
			CheckInDevice  *string `json:"checkInDevice"`
			CheckOutDevice *string `json:"checkOutDevice"`
		} `json:"records"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if response.EmployeeUID != "emp_1" || response.Total != 1 {
		t.Fatalf("unexpected response metadata: %+v", response)
	}
	if len(response.Records) != 1 || response.Records[0].Date != "2026-04-10" || response.Records[0].EmployeeUID != "emp_1" {
		t.Fatalf("unexpected records: %+v", response.Records)
	}
	if response.Records[0].CheckIn == nil || *response.Records[0].CheckIn != "08:30:00" {
		t.Fatalf("checkIn = %v, want 08:30:00", response.Records[0].CheckIn)
	}
	if response.Records[0].CheckOut == nil || *response.Records[0].CheckOut != "17:00:00" {
		t.Fatalf("checkOut = %v, want 17:00:00", response.Records[0].CheckOut)
	}
	if response.Records[0].CheckInDevice == nil || *response.Records[0].CheckInDevice != "Front Gate" {
		t.Fatalf("checkInDevice = %v, want Front Gate", response.Records[0].CheckInDevice)
	}
	if response.Records[0].CheckOutDevice == nil || *response.Records[0].CheckOutDevice != "Back Gate" {
		t.Fatalf("checkOutDevice = %v, want Back Gate", response.Records[0].CheckOutDevice)
	}
}

func TestAttendanceHandlerGetMonthlyStats(t *testing.T) {
	statsUC := &stubMonthlyAttendanceStatsUseCase{
		output: &usecases.GetMonthlyAttendanceStatsOutput{
			Month:                "2026-04",
			TotalWorkedHours:     26.25,
			MissingCheckOutCount: 1,
			AverageCheckInTime: func() *time.Time {
				v := time.Date(2026, 4, 1, 8, 26, 15, 0, time.UTC)
				return &v
			}(),
			WorkingHoursByDay: []usecases.WorkingHoursByDay{
				{Date: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), WorkedHours: 18},
				{Date: time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC), WorkedHours: 0},
			},
		},
	}

	handler := &AttendanceHandler{getMonthlyStatsUC: statsUC}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/logs/stats/monthly", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()
	handler.GetMonthlyStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var response struct {
		EmployeeUID          string  `json:"employeeUid"`
		Month                string  `json:"month"`
		TotalWorkedHours     float64 `json:"totalWorkedHours"`
		AverageCheckInTime   *string `json:"averageCheckInTime"`
		MissingCheckOutCount int     `json:"missingCheckOutCount"`
		WorkingHoursByDay    []struct {
			Date        string  `json:"date"`
			WorkedHours float64 `json:"workedHours"`
		} `json:"workingHoursByDay"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if statsUC.input == nil || statsUC.input.EmployeeUID != "emp_1" {
		t.Fatalf("employeeUid input = %+v, want emp_1", statsUC.input)
	}
	if statsUC.input.StartDate != nil || statsUC.input.EndDate != nil || statsUC.input.PeriodLabel != nil {
		t.Fatalf("expected default period input, got %+v", statsUC.input)
	}
	if response.EmployeeUID != "emp_1" || response.Month != "2026-04" || response.TotalWorkedHours != 26.25 || response.MissingCheckOutCount != 1 {
		t.Fatalf("unexpected response metadata: %+v", response)
	}
	if response.AverageCheckInTime == nil || *response.AverageCheckInTime != "08:26:15" {
		t.Fatalf("AverageCheckInTime = %v, want 08:26:15", response.AverageCheckInTime)
	}
	if len(response.WorkingHoursByDay) != 2 || response.WorkingHoursByDay[0].Date != "2026-04-01" || response.WorkingHoursByDay[0].WorkedHours != 18 {
		t.Fatalf("unexpected WorkingHoursByDay: %+v", response.WorkingHoursByDay)
	}
}

func TestAttendanceHandlerGetMonthlyStats_WithMonthFilter(t *testing.T) {
	statsUC := &stubMonthlyAttendanceStatsUseCase{
		output: &usecases.GetMonthlyAttendanceStatsOutput{Month: "2026-04"},
	}

	handler := &AttendanceHandler{getMonthlyStatsUC: statsUC}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/logs/stats/monthly?month=2026-04", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()
	handler.GetMonthlyStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if statsUC.input == nil {
		t.Fatal("expected input to be passed to use case")
	}
	if statsUC.input.StartDate == nil || statsUC.input.StartDate.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("startDate input = %v, want 2026-04-01", statsUC.input.StartDate)
	}
	if statsUC.input.EndDate == nil || statsUC.input.EndDate.Format("2006-01") != "2026-04" {
		t.Fatalf("endDate input = %v, want month 2026-04", statsUC.input.EndDate)
	}
	if statsUC.input.PeriodLabel == nil || *statsUC.input.PeriodLabel != "2026-04" {
		t.Fatalf("periodLabel input = %v, want 2026-04", statsUC.input.PeriodLabel)
	}
}

func TestAttendanceHandlerGetMonthlyStats_WithRangeFilter(t *testing.T) {
	statsUC := &stubMonthlyAttendanceStatsUseCase{
		output: &usecases.GetMonthlyAttendanceStatsOutput{Month: "2026-04-10 to 2026-04-15"},
	}

	handler := &AttendanceHandler{getMonthlyStatsUC: statsUC}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/logs/stats/monthly?startDate=2026-04-10&endDate=2026-04-15", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()
	handler.GetMonthlyStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if statsUC.input == nil {
		t.Fatal("expected input to be passed to use case")
	}
	if statsUC.input.StartDate == nil || statsUC.input.StartDate.Format("2006-01-02 15:04:05") != "2026-04-10 00:00:00" {
		t.Fatalf("startDate input = %v, want 2026-04-10 00:00:00", statsUC.input.StartDate)
	}
	if statsUC.input.EndDate == nil || statsUC.input.EndDate.Format("2006-01-02 15:04:05") != "2026-04-15 23:59:59" {
		t.Fatalf("endDate input = %v, want 2026-04-15 23:59:59", statsUC.input.EndDate)
	}
	if statsUC.input.PeriodLabel == nil || *statsUC.input.PeriodLabel != "2026-04-10 to 2026-04-15" {
		t.Fatalf("periodLabel input = %v, want range label", statsUC.input.PeriodLabel)
	}
}

func TestAttendanceHandlerGetMonthlyStats_WithYearFilter(t *testing.T) {
	statsUC := &stubMonthlyAttendanceStatsUseCase{
		output: &usecases.GetMonthlyAttendanceStatsOutput{Month: "2026"},
	}

	handler := &AttendanceHandler{getMonthlyStatsUC: statsUC}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/logs/stats/monthly?year=2026", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()
	handler.GetMonthlyStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if statsUC.input == nil {
		t.Fatal("expected input to be passed to use case")
	}
	if statsUC.input.StartDate == nil || statsUC.input.StartDate.Format("2006-01-02 15:04:05") != "2026-01-01 00:00:00" {
		t.Fatalf("startDate input = %v, want 2026-01-01 00:00:00", statsUC.input.StartDate)
	}
	if statsUC.input.PeriodLabel == nil || *statsUC.input.PeriodLabel != "2026" {
		t.Fatalf("periodLabel input = %v, want 2026", statsUC.input.PeriodLabel)
	}
}

func TestAttendanceHandlerGetMonthlyStats_RejectsMixedFilters(t *testing.T) {
	handler := &AttendanceHandler{getMonthlyStatsUC: &stubMonthlyAttendanceStatsUseCase{}}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/employees/emp_1/logs/stats/monthly?month=2026-04&startDate=2026-04-01&endDate=2026-04-10", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()
	handler.GetMonthlyStats(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
