package usecases

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
	"github.com/banumusa/backend/core/domain"
)

func TestSyncDailyAbsencesUseCase_Execute_SyncsTrueAbsencesOnly(t *testing.T) {
	ctx := context.Background()
	db := newAbsenceSyncTestSQLiteDB(t)
	defer db.Close()

	leaveRepo := dbadapter.NewLeaveRecordRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	holidayRepo := dbadapter.NewHolidayDefinitionRepository()

	targetDate := normalizeDateOnly(time.Now().UTC().AddDate(0, 0, -1))
	nonTargetWeekend := (int(targetDate.Weekday()) + 1) % 7
	seedAbsenceSyncWeekendDays(t, db, nonTargetWeekend)

	absentUID := "emp_absent"
	presentUID := "emp_present"
	leaveUID := "emp_leave"
	futureHireUID := "emp_future"

	seedAbsenceSyncEmployee(t, db, 1, absentUID, "Absent", targetDate.AddDate(0, 0, -20), domain.EmployeeStatusActive)
	seedAbsenceSyncEmployee(t, db, 2, presentUID, "Present", targetDate.AddDate(0, 0, -20), domain.EmployeeStatusActive)
	seedAbsenceSyncEmployee(t, db, 3, leaveUID, "On Leave", targetDate.AddDate(0, 0, -20), domain.EmployeeStatusActive)
	seedAbsenceSyncEmployee(t, db, 4, futureHireUID, "Future Hire", targetDate.AddDate(0, 0, 2), domain.EmployeeStatusActive)

	seedAbsenceSyncAttendance(t, db, presentUID, targetDate)
	seedAbsenceSyncLeave(t, db, 3, targetDate)

	seedAbsenceExceptionForDate(t, db, presentUID, targetDate)
	seedAbsenceExceptionForDate(t, db, leaveUID, targetDate)

	uc := NewSyncDailyAbsencesUseCase(db, leaveRepo, weekendRepo, holidayRepo, "UTC")
	output, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}

	if output.SkippedNonWorking {
		t.Fatal("expected working day, got skipped non-working")
	}
	if output.SyncedAbsenceCount != 1 {
		t.Fatalf("expected synced absences=1 got %d", output.SyncedAbsenceCount)
	}
	if output.SkippedPresentCount != 1 {
		t.Fatalf("expected skipped present=1 got %d", output.SkippedPresentCount)
	}
	if output.SkippedLeaveCount != 1 {
		t.Fatalf("expected skipped leave=1 got %d", output.SkippedLeaveCount)
	}
	if output.SkippedHireDateCount != 1 {
		t.Fatalf("expected skipped hire-date=1 got %d", output.SkippedHireDateCount)
	}

	rows, err := db.QueryContext(ctx,
		`SELECT employee_uid FROM attendance_exceptions WHERE attendance_date = ? AND exception_type = ? ORDER BY employee_uid ASC`,
		targetDate.Format("2006-01-02"),
		domain.AttendanceExceptionTypeAbsence,
	)
	if err != nil {
		t.Fatalf("query absences failed: %v", err)
	}
	defer rows.Close()

	actual := make([]string, 0)
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			t.Fatalf("scan absence row failed: %v", err)
		}
		actual = append(actual, uid)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	if len(actual) != 1 || actual[0] != absentUID {
		t.Fatalf("expected only %q as absent, got %v", absentUID, actual)
	}
}

func TestSyncDailyAbsencesUseCase_Execute_SkipsWeekendAndClearsAbsences(t *testing.T) {
	ctx := context.Background()
	db := newAbsenceSyncTestSQLiteDB(t)
	defer db.Close()

	leaveRepo := dbadapter.NewLeaveRecordRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	holidayRepo := dbadapter.NewHolidayDefinitionRepository()

	targetDate := normalizeDateOnly(time.Now().UTC().AddDate(0, 0, -1))
	seedAbsenceSyncWeekendDays(t, db, int(targetDate.Weekday()))

	seedAbsenceSyncEmployee(t, db, 1, "emp_weekend", "Weekend", targetDate.AddDate(0, 0, -10), domain.EmployeeStatusActive)
	seedAbsenceExceptionForDate(t, db, "emp_weekend", targetDate)

	uc := NewSyncDailyAbsencesUseCase(db, leaveRepo, weekendRepo, holidayRepo, "UTC")
	output, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}

	if !output.SkippedNonWorking {
		t.Fatal("expected skipped non-working on weekend")
	}
	if output.SyncedAbsenceCount != 0 {
		t.Fatalf("expected synced absences=0 got %d", output.SyncedAbsenceCount)
	}
	if output.ClearedAbsenceCount != 1 {
		t.Fatalf("expected cleared absences=1 got %d", output.ClearedAbsenceCount)
	}
}

func TestSyncDailyAbsencesUseCase_Execute_SkipsHolidayAndClearsAbsences(t *testing.T) {
	ctx := context.Background()
	db := newAbsenceSyncTestSQLiteDB(t)
	defer db.Close()

	leaveRepo := dbadapter.NewLeaveRecordRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	holidayRepo := dbadapter.NewHolidayDefinitionRepository()

	targetDate := normalizeDateOnly(time.Now().UTC().AddDate(0, 0, -1))
	nonTargetWeekend := (int(targetDate.Weekday()) + 1) % 7
	seedAbsenceSyncWeekendDays(t, db, nonTargetWeekend)

	seedAbsenceSyncEmployee(t, db, 1, "emp_holiday", "Holiday", targetDate.AddDate(0, 0, -10), domain.EmployeeStatusActive)
	seedAbsenceExceptionForDate(t, db, "emp_holiday", targetDate)

	err := holidayRepo.Create(ctx, db, &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     buildHolidayCode("test holiday", targetDate),
		NameEN:   "Test Holiday",
		NameAR:   "عطلة",
		Date:     targetDate,
		IsManual: false,
	})
	if err != nil {
		t.Fatalf("seed holiday failed: %v", err)
	}

	uc := NewSyncDailyAbsencesUseCase(db, leaveRepo, weekendRepo, holidayRepo, "UTC")
	output, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}

	if !output.SkippedNonWorking {
		t.Fatal("expected skipped non-working on holiday")
	}
	if output.ClearedAbsenceCount != 1 {
		t.Fatalf("expected cleared absences=1 got %d", output.ClearedAbsenceCount)
	}
}

func newAbsenceSyncTestSQLiteDB(t *testing.T) *dbadapter.SQLiteDB {
	t.Helper()

	dsn := fmt.Sprintf("file:absence_sync_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := dbadapter.NewSQLiteDB(dsn)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}

	ctx := context.Background()
	statements := []string{
		`CREATE TABLE employees (
			id INTEGER PRIMARY KEY,
			uid TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			department_uid TEXT,
			hire_date TEXT NOT NULL,
			status TEXT NOT NULL
		);`,
		`CREATE TABLE attendance_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_uid TEXT NOT NULL,
			punched_at TEXT NOT NULL
		);`,
		`CREATE TABLE leave_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL
		);`,
		`CREATE TABLE holiday_definitions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uid TEXT UNIQUE NOT NULL,
			code TEXT UNIQUE NOT NULL,
			name_en TEXT NOT NULL,
			name_ar TEXT NOT NULL,
			date TEXT NOT NULL,
			is_manual INTEGER NOT NULL DEFAULT 1,
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		);`,
		`CREATE TABLE weekend_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uid TEXT UNIQUE NOT NULL,
			day_of_week INTEGER NOT NULL UNIQUE,
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		);`,
		`CREATE TABLE attendance_exceptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uid TEXT UNIQUE NOT NULL,
			employee_uid TEXT NOT NULL,
			attendance_date TEXT NOT NULL,
			exception_type TEXT NOT NULL,
			check_in TEXT,
			check_out TEXT,
			grace_minutes INTEGER,
			minutes_delta INTEGER,
			created_at TEXT,
			updated_at TEXT,
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

func seedAbsenceSyncEmployee(t *testing.T, db *dbadapter.SQLiteDB, id int64, uid, name string, hireDate time.Time, status domain.EmployeeStatus) {
	t.Helper()

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO employees (id, uid, name, department_uid, hire_date, status) VALUES (?, ?, ?, ?, ?, ?)`,
		id,
		uid,
		name,
		"dept-1",
		hireDate.Format("2006-01-02"),
		status,
	)
	if err != nil {
		t.Fatalf("seed employee failed: %v", err)
	}
}

func seedAbsenceSyncAttendance(t *testing.T, db *dbadapter.SQLiteDB, employeeUID string, day time.Time) {
	t.Helper()

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO attendance_records (employee_uid, punched_at) VALUES (?, ?)`,
		employeeUID,
		day.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("seed attendance failed: %v", err)
	}
}

func seedAbsenceSyncLeave(t *testing.T, db *dbadapter.SQLiteDB, employeeID int64, day time.Time) {
	t.Helper()

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO leave_records (employee_id, start_date, end_date) VALUES (?, ?, ?)`,
		employeeID,
		day.Format("2006-01-02"),
		day.Format("2006-01-02"),
	)
	if err != nil {
		t.Fatalf("seed leave failed: %v", err)
	}
}

func seedAbsenceSyncWeekendDays(t *testing.T, db *dbadapter.SQLiteDB, days ...int) {
	t.Helper()

	for _, day := range days {
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO weekend_config (uid, day_of_week) VALUES (?, ?)`,
			fmt.Sprintf("weekend_%d", day),
			day,
		)
		if err != nil {
			t.Fatalf("seed weekend day failed: %v", err)
		}
	}
}

func seedAbsenceExceptionForDate(t *testing.T, db *dbadapter.SQLiteDB, employeeUID string, day time.Time) {
	t.Helper()

	now := time.Now().Format(time.RFC3339Nano)
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		domain.GenerateUID("aex"),
		employeeUID,
		day.Format("2006-01-02"),
		domain.AttendanceExceptionTypeAbsence,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("seed absence exception failed: %v", err)
	}
}
