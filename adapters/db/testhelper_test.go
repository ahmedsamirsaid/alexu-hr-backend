package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

// TestDB wraps a SQLite database for testing
type TestDB struct {
	*SQLiteDB
	t *testing.T
}

var testDBCounter int

// NewTestDB creates an in-memory SQLite database with the schema applied
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Use a unique name for each test to ensure isolation while using shared cache
	// Shared cache ensures migrations and queries use the same database connection
	testDBCounter++
	dsn := fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", testDBCounter)

	db, err := NewSQLiteDB(dsn)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	tdb := &TestDB{SQLiteDB: db, t: t}
	tdb.applySchema()

	return tdb
}

func (tdb *TestDB) applySchema() {
	tdb.t.Helper()

	migrationsPath := findMigrationsDir(tdb.t)

	driver, err := sqlite.WithInstance(tdb.db, &sqlite.Config{})
	if err != nil {
		tdb.t.Fatalf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite",
		driver,
	)
	if err != nil {
		tdb.t.Fatalf("failed to create migrate instance: %v", err)
	}

	// Run all migrations (schema + seeds)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		tdb.t.Fatalf("failed to run migrations: %v", err)
	}

	// Clear seed data to start fresh for tests
	tdb.clearSeedData()
}

func (tdb *TestDB) clearSeedData() {
	tdb.t.Helper()
	ctx := context.Background()

	// Clear seed data tables (order matters for foreign keys)
	tables := []string{
		"annual_reports",
		"incentive_bonus",
		"penalties_removed",
		"penalties",
		"leave_request_documents",
		"sub_leave_types",
		"leave_requests",
		"approval_actions",
		"approval_requests",
		"approval_flow_steps",
		"approval_flows",
		"user_roles",
		"role_permissions",
		"refresh_tokens",
		"otp_codes",
		"rate_limit_records",
		"leave_balance_transactions",
		"leave_records",
		"leave_balances",
		"leave_types",
		"holiday_definitions",
		"weekend_config",
		"employees",
		"users",
		"roles",
		"permissions",
	}

	for _, table := range tables {
		_, err := tdb.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			// Ignore errors for tables that might not exist yet
			continue
		}
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()

	// Try relative paths from where tests might run
	candidates := []string{
		"migrations",
		"./migrations",
		"adapters/db/migrations",
		"backend/adapters/db/migrations",
	}

	// Also try walking up from current directory
	cwd, err := os.Getwd()
	if err == nil {
		dir := cwd
		for i := range 5 {
			_ = i
			candidate := filepath.Join(dir, "adapters", "db", "migrations")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate
			}
			candidate = filepath.Join(dir, "backend", "adapters", "db", "migrations")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate
			}
			candidate = filepath.Join(dir, "migrations")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate
			}
			dir = filepath.Dir(dir)
		}
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	t.Fatalf("could not find migrations directory")
	return ""
}

// Close cleans up the test database
func (tdb *TestDB) Close() {
	if err := tdb.SQLiteDB.Close(); err != nil {
		tdb.t.Errorf("failed to close test database: %v", err)
	}
}

// SeedEmployee creates a test employee and returns it
func (tdb *TestDB) SeedEmployee(name string) *domain.Employee {
	tdb.t.Helper()

	employee := domain.NewEmployee(
		name,
		"010"+time.Now().Format("150405.000000"),
		"gov"+time.Now().Format("150405.000000"),
		"uni"+time.Now().Format("150405.000000"),
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	repo := NewEmployeeRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, tdb.SQLiteDB, employee); err != nil {
		tdb.t.Fatalf("failed to seed employee: %v", err)
	}

	return employee
}

// SeedLeaveType creates a test leave type and returns it
func (tdb *TestDB) SeedLeaveType(code, nameEN, nameAR string, defaultBalance int) *domain.LeaveType {
	tdb.t.Helper()

	maxConsec := 2
	recordingDeadline := 2
	lt := &domain.LeaveType{
		UID:                   domain.GenerateUID("ltype"),
		Code:                  code,
		NameEN:                nameEN,
		NameAR:                nameAR,
		DefaultBalance:        defaultBalance,
		MaxConsecutive:        &maxConsec,
		RecordingDeadlineDays: &recordingDeadline,
		IsActive:              true,
	}

	ctx := context.Background()
	now := time.Now()
	_, err := tdb.db.ExecContext(ctx, `
		INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive,
			recording_deadline_days, advance_notice_days, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		lt.UID, lt.Code, lt.NameEN, lt.NameAR, lt.DefaultBalance, lt.MaxConsecutive,
		lt.RecordingDeadlineDays, lt.AdvanceNoticeDays, lt.IsActive, now, now)
	if err != nil {
		tdb.t.Fatalf("failed to seed leave type: %v", err)
	}

	// Get the ID
	row := tdb.db.QueryRowContext(ctx, "SELECT id FROM leave_types WHERE uid = ?", lt.UID)
	if err := row.Scan(&lt.ID); err != nil {
		tdb.t.Fatalf("failed to get leave type ID: %v", err)
	}

	lt.CreatedAt = now
	lt.UpdatedAt = now

	return lt
}

func (tdb *TestDB) SeedSubLeaveType(leaveTypeUID, nameEN, nameAR string) *domain.SubLeaveType {
	tdb.t.Helper()

	subLeaveType := &domain.SubLeaveType{
		UID:          domain.GenerateUID("slt"),
		LeaveTypeUID: leaveTypeUID,
		NameEN:       nameEN,
		NameAR:       nameAR,
	}

	ctx := context.Background()
	now := time.Now()
	_, err := tdb.db.ExecContext(ctx, `
		INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		subLeaveType.UID, subLeaveType.LeaveTypeUID, subLeaveType.NameEN, subLeaveType.NameAR, now, now)
	if err != nil {
		tdb.t.Fatalf("failed to seed sub leave type: %v", err)
	}

	row := tdb.db.QueryRowContext(ctx, "SELECT id FROM sub_leave_types WHERE uid = ?", subLeaveType.UID)
	if err := row.Scan(&subLeaveType.ID); err != nil {
		tdb.t.Fatalf("failed to get sub leave type ID: %v", err)
	}

	subLeaveType.CreatedAt = now
	subLeaveType.UpdatedAt = now

	return subLeaveType
}

// SeedLeaveBalance creates a test leave balance and returns it
func (tdb *TestDB) SeedLeaveBalance(employeeID, leaveTypeID int64, year, totalDays, usedDays int) *domain.LeaveBalance {
	tdb.t.Helper()

	balance := &domain.LeaveBalance{
		UID:         domain.GenerateUID("lbal"),
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
		TotalDays:   totalDays,
		UsedDays:    usedDays,
	}

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, tdb.SQLiteDB, balance); err != nil {
		tdb.t.Fatalf("failed to seed leave balance: %v", err)
	}

	return balance
}

// SeedWeekendConfig seeds weekend configuration for the given days
func (tdb *TestDB) SeedWeekendConfig(days ...int) {
	tdb.t.Helper()

	ctx := context.Background()
	now := time.Now()
	for _, day := range days {
		uid := domain.GenerateUID("wknd")
		_, err := tdb.db.ExecContext(ctx, `
			INSERT INTO weekend_config (uid, day_of_week, created_at, updated_at)
			VALUES (?, ?, ?, ?)`, uid, day, now, now)
		if err != nil {
			tdb.t.Fatalf("failed to seed weekend config for day %d: %v", day, err)
		}
	}
}

// SeedHolidayDefinition creates a test holiday definition and returns it
func (tdb *TestDB) SeedHolidayDefinition(code, nameEN, nameAR string, month, day *int) *domain.HolidayDefinition {
	tdb.t.Helper()

	now := time.Now()
	holidayDate := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	if month != nil && day != nil {
		holidayDate = time.Date(now.Year(), time.Month(*month), *day, 0, 0, 0, 0, time.UTC)
	}

	def := &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     code,
		NameEN:   nameEN,
		NameAR:   nameAR,
		Date:     holidayDate,
		IsManual: month == nil || day == nil,
	}

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, tdb.SQLiteDB, def); err != nil {
		tdb.t.Fatalf("failed to seed holiday definition: %v", err)
	}

	return def
}

// MustExec executes a query and fails the test if it errors
func (tdb *TestDB) MustExec(query string, args ...any) sql.Result {
	tdb.t.Helper()

	result, err := tdb.db.Exec(query, args...)
	if err != nil {
		tdb.t.Fatalf("failed to execute query: %v", err)
	}
	return result
}

// ParseDate parses a date string in YYYY-MM-DD format
func ParseDate(t *testing.T, s string) time.Time {
	t.Helper()

	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", s, err)
	}
	return d
}

// SeedUser creates a test user and returns it
func (tdb *TestDB) SeedUser(phone string) *domain.User {
	tdb.t.Helper()

	user := domain.NewUser(phone)
	repo := NewUserRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		tdb.t.Fatalf("failed to seed user: %v", err)
	}

	return user
}

// SeedRole creates a test role and returns it
func (tdb *TestDB) SeedRole(name, description string) *domain.Role {
	tdb.t.Helper()

	role := domain.NewRole(name, description)
	repo := NewRoleRepository()
	ctx := context.Background()
	if err := repo.Create(ctx, tdb.SQLiteDB, role); err != nil {
		tdb.t.Fatalf("failed to seed role: %v", err)
	}

	return role
}

// SeedPermission creates a test permission and returns it
func (tdb *TestDB) SeedPermission(code, description string) *domain.Permission {
	tdb.t.Helper()

	perm := &domain.Permission{
		UID:         domain.GenerateUID("perm"),
		Code:        code,
		Description: description,
	}

	ctx := context.Background()
	now := time.Now()
	result, err := tdb.db.ExecContext(ctx, `
		INSERT INTO permissions (uid, code, description, created_at)
		VALUES (?, ?, ?, ?)`, perm.UID, perm.Code, perm.Description, now)
	if err != nil {
		tdb.t.Fatalf("failed to seed permission: %v", err)
	}

	id, _ := result.LastInsertId()
	perm.ID = id
	perm.CreatedAt = now

	return perm
}
