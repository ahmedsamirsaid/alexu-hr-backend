package usecases

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type mockMonthlyStatsDB struct{}

func (m *mockMonthlyStatsDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	return nil, nil
}

func (m *mockMonthlyStatsDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockMonthlyStatsDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockMonthlyStatsDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

type mockAttendanceRecordRepoForMonthlyStats struct {
	records        []*domain.AttendanceRecord
	gotEmployeeUID *string
}

func (m *mockAttendanceRecordRepoForMonthlyStats) Create(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) (bool, error) {
	return false, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.AttendanceRecord, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) Update(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) error {
	return nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ListByDateRange(ctx context.Context, q ports.Querier, startDate, endDate time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	m.gotEmployeeUID = employeeUID
	return m.records, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ListByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) CountByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) CountByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ListDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) CountDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ListDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return nil, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) CountDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (m *mockAttendanceRecordRepoForMonthlyStats) ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q ports.Querier, deviceUserID string) (*string, error) {
	return nil, nil
}

func TestGetMonthlyAttendanceStatsUseCaseExecute(t *testing.T) {
	repo := &mockAttendanceRecordRepoForMonthlyStats{
		records: []*domain.AttendanceRecord{
			{EmployeeUID: "emp_1", PunchedAt: time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC), PunchType: domain.AttendancePunchTypeCheckIn},
			{EmployeeUID: "emp_1", PunchedAt: time.Date(2026, 4, 1, 17, 0, 0, 0, time.UTC), PunchType: domain.AttendancePunchTypeCheckOut},
			{EmployeeUID: "emp_1", PunchedAt: time.Date(2026, 4, 2, 8, 30, 0, 0, time.UTC), PunchType: domain.AttendancePunchTypeCheckIn},
			{EmployeeUID: "emp_1", PunchedAt: time.Date(2026, 4, 3, 8, 15, 0, 0, time.UTC), PunchType: domain.AttendancePunchTypeUnknown},
			{EmployeeUID: "emp_1", PunchedAt: time.Date(2026, 4, 3, 16, 30, 0, 0, time.UTC), PunchType: domain.AttendancePunchTypeCheckOut},
		},
	}

	uc := NewGetMonthlyAttendanceStatsUseCase(&mockMonthlyStatsDB{}, repo)
	uc.now = func() time.Time {
		return time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	}

	output, err := uc.Execute(context.Background(), GetMonthlyAttendanceStatsInput{
		EmployeeUID: "emp_1",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if repo.gotEmployeeUID == nil || *repo.gotEmployeeUID != "emp_1" {
		t.Fatalf("employeeUID filter = %v, want emp_1", repo.gotEmployeeUID)
	}

	if output.Month != "2026-04" {
		t.Fatalf("Month = %q, want %q", output.Month, "2026-04")
	}
	if output.TotalWorkedHours != 17.25 {
		t.Fatalf("TotalWorkedHours = %v, want 17.25", output.TotalWorkedHours)
	}
	if output.AverageCheckInTime == nil || output.AverageCheckInTime.Format("15:04:05") != "08:15:00" {
		t.Fatalf("AverageCheckInTime = %v, want 08:15:00", output.AverageCheckInTime)
	}
	if output.MissingCheckOutCount != 1 {
		t.Fatalf("MissingCheckOutCount = %d, want 1", output.MissingCheckOutCount)
	}
	if len(output.WorkingHoursByDay) != 3 {
		t.Fatalf("len(WorkingHoursByDay) = %d, want 3", len(output.WorkingHoursByDay))
	}
	if output.WorkingHoursByDay[0].Date.Format("2006-01-02") != "2026-04-01" || output.WorkingHoursByDay[0].WorkedHours != 9 {
		t.Fatalf("day 1 = %+v, want date 2026-04-01 with 9 hours", output.WorkingHoursByDay[0])
	}
	if output.WorkingHoursByDay[1].Date.Format("2006-01-02") != "2026-04-02" || output.WorkingHoursByDay[1].WorkedHours != 0 {
		t.Fatalf("day 2 = %+v, want date 2026-04-02 with 0 hours", output.WorkingHoursByDay[1])
	}
	if output.WorkingHoursByDay[2].Date.Format("2006-01-02") != "2026-04-03" || output.WorkingHoursByDay[2].WorkedHours != 8.25 {
		t.Fatalf("day 3 = %+v, want date 2026-04-03 with 8.25 hours", output.WorkingHoursByDay[2])
	}
}
