package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

// mockEmployeeRepoForDashboard implements EmployeeRepository for dashboard tests
type mockEmployeeRepoForDashboard struct {
	count     int
	employees []*domain.Employee
	err       error
}

func (m *mockEmployeeRepoForDashboard) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForDashboard) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForDashboard) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForDashboard) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForDashboard) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	if m.employees != nil {
		return m.employees, m.err
	}
	return nil, m.err
}

func (m *mockEmployeeRepoForDashboard) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForDashboard) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForDashboard) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForDashboard) Count(ctx context.Context, q ports.Querier) (int, error) {
	return m.count, m.err
}

// mockLeaveRecordRepoForDashboard implements LeaveRecordRepository for dashboard tests
type mockLeaveRecordRepoForDashboard struct {
	leavesToday int
	err         error
}

type mockLeaveRequestRepoForDashboard struct {
	count int
	err   error
}

type mockAttendanceRepoForDashboard struct {
	records []*domain.AttendanceRecord
	err     error
}

func (m *mockAttendanceRepoForDashboard) ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return m.records, m.err
}

func (m *mockLeaveRequestRepoForDashboard) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForDashboard) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForDashboard) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForDashboard) Create(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	return nil
}

func (m *mockLeaveRequestRepoForDashboard) Update(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	return nil
}

func (m *mockLeaveRequestRepoForDashboard) ListByEmployee(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForDashboard) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeUID string, limit, offset int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForDashboard) CountByEmployee(ctx context.Context, q ports.Querier, employeeUID string) (int, error) {
	return 0, nil
}

func (m *mockLeaveRequestRepoForDashboard) HasOverlapping(ctx context.Context, q ports.Querier, employeeUID string, startDate, endDate time.Time, excludeUID *string) (bool, error) {
	return false, nil
}

func (m *mockLeaveRequestRepoForDashboard) List(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter, limit, offset int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForDashboard) Count(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter) (int, error) {
	return m.count, m.err
}

func (m *mockLeaveRequestRepoForDashboard) FindExpiredPending(ctx context.Context, q ports.Querier, graceDays int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) Create(ctx context.Context, q ports.Querier, record *domain.LeaveRecord) error {
	return nil
}

func (m *mockLeaveRecordRepoForDashboard) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeID int64, limit, offset int) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) CountByEmployee(ctx context.Context, q ports.Querier, employeeID int64) (int, error) {
	return 0, nil
}

func (m *mockLeaveRecordRepoForDashboard) ListByEmployeeAndDateRange(ctx context.Context, q ports.Querier, employeeID int64, start, end time.Time) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) ListByEmployeeAndType(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64) ([]*domain.LeaveRecord, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) ListAllPaginated(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter, limit, offset int) ([]*ports.LeaveRecordWithEmployee, error) {
	return nil, nil
}

func (m *mockLeaveRecordRepoForDashboard) CountAll(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter) (int, error) {
	return 0, nil
}

func (m *mockLeaveRecordRepoForDashboard) CountOnLeaveToday(ctx context.Context, q ports.Querier, date time.Time) (int, error) {
	return m.leavesToday, m.err
}

func (m *mockLeaveRecordRepoForDashboard) HasLeaveOnDate(ctx context.Context, q ports.Querier, employeeID int64, date time.Time) (bool, error) {
	return false, nil
}

func TestGetDashboardStatsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name                    string
		employeeCount           int
		leavesToday             int
		checkedInRecords        []*domain.AttendanceRecord
		pendingRequests         int
		expectedTotalEmployees  int
		expectedCheckedInToday  int
		expectedLeavesToday     int
		expectedPendingRequests int
	}{
		{
			name:          "returns correct stats",
			employeeCount: 150,
			leavesToday:   5,
			checkedInRecords: []*domain.AttendanceRecord{
				{EmployeeUID: "emp_1", PunchType: domain.AttendancePunchTypeCheckIn},
				{EmployeeUID: "emp_1", PunchType: domain.AttendancePunchTypeCheckOut},
				{EmployeeUID: "emp_2", PunchType: domain.AttendancePunchTypeUnknown},
			},
			pendingRequests:         7,
			expectedTotalEmployees:  150,
			expectedCheckedInToday:  2,
			expectedLeavesToday:     5,
			expectedPendingRequests: 7,
		},
		{
			name:                    "zero employees",
			employeeCount:           0,
			leavesToday:             0,
			pendingRequests:         0,
			expectedTotalEmployees:  0,
			expectedCheckedInToday:  0,
			expectedLeavesToday:     0,
			expectedPendingRequests: 0,
		},
		{
			name:          "many employees on leave",
			employeeCount: 500,
			leavesToday:   100,
			checkedInRecords: []*domain.AttendanceRecord{
				{EmployeeUID: "emp_10", PunchType: domain.AttendancePunchTypeCheckIn},
				{EmployeeUID: "emp_11", PunchType: domain.AttendancePunchTypeCheckIn},
				{EmployeeUID: "emp_12", PunchType: domain.AttendancePunchTypeUnknown},
				{EmployeeUID: "emp_12", PunchType: domain.AttendancePunchTypeCheckIn},
			},
			pendingRequests:         12,
			expectedTotalEmployees:  500,
			expectedCheckedInToday:  3,
			expectedLeavesToday:     100,
			expectedPendingRequests: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockDB{tx: &mockTx{}}
			employeeRepo := &mockEmployeeRepoForDashboard{count: tt.employeeCount}
			leaveRecordRepo := &mockLeaveRecordRepoForDashboard{leavesToday: tt.leavesToday}
			leaveRequestRepo := &mockLeaveRequestRepoForDashboard{count: tt.pendingRequests}
			attendanceRepo := &mockAttendanceRepoForDashboard{records: tt.checkedInRecords}

			uc := usecases.NewGetDashboardStatsUseCase(db, employeeRepo, leaveRecordRepo, leaveRequestRepo, attendanceRepo)

			output, err := uc.Execute(context.Background(), usecases.GetDashboardStatsInput{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if output.TotalEmployees != tt.expectedTotalEmployees {
				t.Errorf("TotalEmployees = %d, want %d", output.TotalEmployees, tt.expectedTotalEmployees)
			}

			if output.CheckedInToday != tt.expectedCheckedInToday {
				t.Errorf("CheckedInToday = %d, want %d", output.CheckedInToday, tt.expectedCheckedInToday)
			}

			if output.LeavesToday != tt.expectedLeavesToday {
				t.Errorf("LeavesToday = %d, want %d", output.LeavesToday, tt.expectedLeavesToday)
			}

			if output.PendingRequests != tt.expectedPendingRequests {
				t.Errorf("PendingRequests = %d, want %d", output.PendingRequests, tt.expectedPendingRequests)
			}
		})
	}
}

func TestGetDashboardStatsUseCase_CountsPendingRequests(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepoForDashboard{count: 100}
	leaveRecordRepo := &mockLeaveRecordRepoForDashboard{leavesToday: 10}
	leaveRequestRepo := &mockLeaveRequestRepoForDashboard{count: 4}
	attendanceRepo := &mockAttendanceRepoForDashboard{}

	uc := usecases.NewGetDashboardStatsUseCase(db, employeeRepo, leaveRecordRepo, leaveRequestRepo, attendanceRepo)

	output, err := uc.Execute(context.Background(), usecases.GetDashboardStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.PendingRequests != 4 {
		t.Errorf("PendingRequests = %d, want %d", output.PendingRequests, 4)
	}
}

func TestGetDashboardStatsUseCase_ExecuteScopedCountsCheckedInForManagedDepartments(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	departmentA := "dep_a"
	departmentB := "dep_b"

	employeeRepo := &mockEmployeeRepoForDashboard{
		employees: []*domain.Employee{
			{ID: 1, UID: "emp_a1", DepartmentUID: &departmentA},
			{ID: 2, UID: "emp_a2", DepartmentUID: &departmentA},
			{ID: 3, UID: "emp_b1", DepartmentUID: &departmentB},
		},
	}
	leaveRecordRepo := &mockLeaveRecordRepoForDashboard{}
	leaveRequestRepo := &mockLeaveRequestRepoForDashboard{}
	attendanceRepo := &mockAttendanceRepoForDashboard{
		records: []*domain.AttendanceRecord{
			{EmployeeUID: "emp_a1", PunchType: domain.AttendancePunchTypeCheckIn},
			{EmployeeUID: "emp_a2", PunchType: domain.AttendancePunchTypeCheckOut},
			{EmployeeUID: "emp_b1", PunchType: domain.AttendancePunchTypeCheckIn},
		},
	}

	uc := usecases.NewGetDashboardStatsUseCase(db, employeeRepo, leaveRecordRepo, leaveRequestRepo, attendanceRepo)

	output, err := uc.Execute(context.Background(), usecases.GetDashboardStatsInput{
		ManagedDepartmentUIDs: []string{departmentA},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.TotalEmployees != 2 {
		t.Errorf("TotalEmployees = %d, want %d", output.TotalEmployees, 2)
	}

	if output.CheckedInToday != 1 {
		t.Errorf("CheckedInToday = %d, want %d", output.CheckedInToday, 1)
	}
}
