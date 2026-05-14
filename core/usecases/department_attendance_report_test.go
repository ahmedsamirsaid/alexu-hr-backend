package usecases

import (
	"context"
	"math"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type reportDepartmentRepo struct {
	department *domain.Department
}

func (r *reportDepartmentRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Department, error) {
	return nil, nil
}

func (r *reportDepartmentRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Department, error) {
	if r.department != nil && r.department.UID == uid {
		return r.department, nil
	}
	return nil, nil
}

func (r *reportDepartmentRepo) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.Department, error) {
	return nil, nil
}

func (r *reportDepartmentRepo) Create(ctx context.Context, q ports.Querier, department *domain.Department) error {
	return nil
}

func (r *reportDepartmentRepo) Update(ctx context.Context, q ports.Querier, department *domain.Department) error {
	return nil
}

func (r *reportDepartmentRepo) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.Department, error) {
	return nil, nil
}

type reportAttendanceRecordRepo struct {
	groups []*ports.DailyAttendanceGroup
}

func (r *reportAttendanceRecordRepo) Create(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) (bool, error) {
	return false, nil
}

func (r *reportAttendanceRecordRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.AttendanceRecord, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) Update(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) error {
	return nil
}

func (r *reportAttendanceRecordRepo) ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) ListByDateRange(ctx context.Context, q ports.Querier, startDate, endDate time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) ListByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) CountByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (r *reportAttendanceRecordRepo) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) CountByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (r *reportAttendanceRecordRepo) ListDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return r.groups, nil
}

func (r *reportAttendanceRecordRepo) CountDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (r *reportAttendanceRecordRepo) ListDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) CountDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}

func (r *reportAttendanceRecordRepo) ListDaily(ctx context.Context, q ports.Querier, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return nil, nil
}

func (r *reportAttendanceRecordRepo) ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q ports.Querier, deviceUserID string) (*string, error) {
	return nil, nil
}

type reportWeekendRepo struct {
	days []int
}

func (r *reportWeekendRepo) List(ctx context.Context, q ports.Querier) ([]*domain.WeekendConfig, error) {
	return nil, nil
}

func (r *reportWeekendRepo) GetWeekendDays(ctx context.Context, q ports.Querier) ([]int, error) {
	return r.days, nil
}

type reportHolidayRepo struct {
	holidays []*domain.HolidayDefinition
}

func (r *reportHolidayRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayDefinition, error) {
	return nil, nil
}

func (r *reportHolidayRepo) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.HolidayDefinition, error) {
	return nil, nil
}

func (r *reportHolidayRepo) GetByDate(ctx context.Context, q ports.Querier, date time.Time) ([]*domain.HolidayDefinition, error) {
	return r.holidays, nil
}

func (r *reportHolidayRepo) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	result := make([]*domain.HolidayDefinition, 0)
	for _, holiday := range r.holidays {
		if !holiday.Date.Before(start) && !holiday.Date.After(end) {
			result = append(result, holiday)
		}
	}
	return result, nil
}

func (r *reportHolidayRepo) List(ctx context.Context, q ports.Querier) ([]*domain.HolidayDefinition, error) {
	return r.holidays, nil
}

func (r *reportHolidayRepo) Create(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	return nil
}

func (r *reportHolidayRepo) Update(ctx context.Context, q ports.Querier, def *domain.HolidayDefinition) error {
	return nil
}

func (r *reportHolidayRepo) Delete(ctx context.Context, q ports.Querier, id int64) error {
	return nil
}

func (r *reportHolidayRepo) ListAllByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	return r.holidays, nil
}

func (r *reportHolidayRepo) ListByDateRangeForDepartment(ctx context.Context, q ports.Querier, start, end time.Time, departmentUID string) ([]*domain.HolidayDefinition, error) {
	return r.holidays, nil
}

func TestGetDepartmentAttendanceReportUseCase(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.FixedZone("Africa/Cairo", 2*60*60)
	defer func() {
		time.Local = previousLocal
	}()

	db := newDepartmentReportTestDB(t)
	defer db.Close()

	ctx := context.Background()
	departmentUID := "dep_engineering"
	seedReportEmployee(t, db, "emp_alice", "Alice", departmentUID, "2026-04-01")
	seedReportEmployee(t, db, "emp_bob", "Bob", departmentUID, "2026-04-08")

	seedReportException(t, db, "emp_alice", "2026-04-09", domain.AttendanceExceptionTypeAbsence)
	seedReportException(t, db, "emp_alice", "2026-04-06", domain.AttendanceExceptionTypeMissedPunchIn)
	seedReportException(t, db, "emp_alice", "2026-04-08", domain.AttendanceExceptionTypeLateArrival)
	seedReportException(t, db, "emp_bob", "2026-04-09", domain.AttendanceExceptionTypeAbsence)

	deptUID := departmentUID
	aliceCheckIn := time.Date(2026, 4, 5, 7, 20, 0, 0, time.UTC)
	aliceCheckOut := time.Date(2026, 4, 5, 15, 0, 0, 0, time.UTC)
	aliceCheckOutOnly := time.Date(2026, 4, 6, 15, 10, 0, 0, time.UTC)
	bobCheckIn := time.Date(2026, 4, 8, 6, 55, 0, 0, time.UTC)
	bobCheckOut := time.Date(2026, 4, 8, 14, 30, 0, 0, time.UTC)

	recordRepo := &reportAttendanceRecordRepo{
		groups: []*ports.DailyAttendanceGroup{
			{
				Date:          time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
				EmployeeUID:   "emp_alice",
				EmployeeName:  "Alice",
				DepartmentUID: &deptUID,
				CheckIn:       &aliceCheckIn,
				CheckOut:      &aliceCheckOut,
			},
			{
				Date:          time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
				EmployeeUID:   "emp_alice",
				EmployeeName:  "Alice",
				DepartmentUID: &deptUID,
				CheckOut:      &aliceCheckOutOnly,
			},
			{
				Date:          time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
				EmployeeUID:   "emp_bob",
				EmployeeName:  "Bob",
				DepartmentUID: &deptUID,
				CheckIn:       &bobCheckIn,
				CheckOut:      &bobCheckOut,
			},
		},
	}

	uc := NewGetDepartmentAttendanceReportUseCase(
		db,
		&reportDepartmentRepo{department: &domain.Department{UID: departmentUID, IsActive: true}},
		recordRepo,
		nil,
		nil,
		&reportWeekendRepo{days: []int{5, 6}},
		&reportHolidayRepo{holidays: []*domain.HolidayDefinition{
			{Date: time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)},
		}},
	)

	output, err := uc.Execute(ctx, GetDepartmentAttendanceReportInput{
		DepartmentUID: departmentUID,
		StartDate:     time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
		EndDate:       time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.TotalHeadcount != 2 {
		t.Fatalf("TotalHeadcount = %d, want 2", output.TotalHeadcount)
	}
	if !almostEqual(output.AverageAttendanceRate, 50) {
		t.Fatalf("AverageAttendanceRate = %v, want 50", output.AverageAttendanceRate)
	}
	if !almostEqual(output.AveragePunctualityRate, 66.6666666667) {
		t.Fatalf("AveragePunctualityRate = %v, want about 66.67", output.AveragePunctualityRate)
	}
	if len(output.Employees) != 2 {
		t.Fatalf("len(Employees) = %d, want 2", len(output.Employees))
	}

	alice := output.Employees[0]
	if alice.EmployeeUID != "emp_alice" {
		t.Fatalf("first employee = %s, want emp_alice", alice.EmployeeUID)
	}
	if alice.TotalWorkingDays != 4 || alice.DaysPresent != 2 || alice.DaysAbsent != 1 {
		t.Fatalf("alice working/present/absent = %d/%d/%d, want 4/2/1", alice.TotalWorkingDays, alice.DaysPresent, alice.DaysAbsent)
	}
	if alice.LateDays != 2 || alice.MissingCheckInDays != 1 || alice.MissingCheckOutDays != 0 {
		t.Fatalf("alice late/missingIn/missingOut = %d/%d/%d, want 2/1/0", alice.LateDays, alice.MissingCheckInDays, alice.MissingCheckOutDays)
	}
	if !almostEqual(alice.TotalWorkedHours, 7.6666666667) {
		t.Fatalf("alice TotalWorkedHours = %v, want about 7.67", alice.TotalWorkedHours)
	}
	if alice.AverageCheckInTime == nil || alice.AverageCheckInTime.Format("15:04:05") != "09:20:00" {
		t.Fatalf("alice AverageCheckInTime = %v, want 09:20:00", alice.AverageCheckInTime)
	}
	if alice.AverageCheckOutTime == nil || alice.AverageCheckOutTime.Format("15:04:05") != "17:05:00" {
		t.Fatalf("alice AverageCheckOutTime = %v, want 17:05:00", alice.AverageCheckOutTime)
	}

	bob := output.Employees[1]
	if bob.TotalWorkingDays != 2 || bob.DaysPresent != 1 || bob.DaysAbsent != 1 {
		t.Fatalf("bob working/present/absent = %d/%d/%d, want 2/1/1", bob.TotalWorkingDays, bob.DaysPresent, bob.DaysAbsent)
	}
	if bob.EarlyDepartureDays != 1 || bob.LateDays != 0 {
		t.Fatalf("bob early/late = %d/%d, want 1/0", bob.EarlyDepartureDays, bob.LateDays)
	}
}

func newDepartmentReportTestDB(t *testing.T) *dbadapter.SQLiteDB {
	t.Helper()

	db, err := dbadapter.NewSQLiteDB("file:department_report_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	ctx := context.Background()
	statements := []string{
		`CREATE TABLE employees (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			uid TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			mobile TEXT NOT NULL,
			government_id TEXT NOT NULL,
			university_id TEXT NOT NULL,
			email TEXT,
			hire_date DATE NOT NULL,
			status TEXT NOT NULL,
			department_uid TEXT,
			shift_uid TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE attendance_exceptions (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			uid TEXT UNIQUE NOT NULL,
			employee_uid TEXT NOT NULL,
			attendance_date DATE NOT NULL,
			exception_type TEXT NOT NULL,
			check_in TEXT,
			check_out TEXT,
			grace_minutes INTEGER,
			minutes_delta INTEGER,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			UNIQUE(employee_uid, attendance_date, exception_type)
		);`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("failed to execute schema statement: %v", err)
		}
	}

	return db
}

func seedReportEmployee(t *testing.T, db ports.DB, uid, name, departmentUID, hireDate string) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO employees (
			uid, name, mobile, government_id, university_id, hire_date, status, department_uid
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uid, name, "01000000000", uid+"_gov", uid+"_uni", hireDate, domain.EmployeeStatusActive, departmentUID)
	if err != nil {
		t.Fatalf("failed to seed employee %s: %v", uid, err)
	}
}

func seedReportException(t *testing.T, db ports.DB, employeeUID, date string, exceptionType domain.AttendanceExceptionType) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type)
		VALUES (?, ?, ?, ?)
	`, employeeUID+"_"+date+"_"+string(exceptionType), employeeUID, date, exceptionType)
	if err != nil {
		t.Fatalf("failed to seed attendance exception: %v", err)
	}
}

func almostEqual(got, want float64) bool {
	return math.Abs(got-want) < 0.0001
}
