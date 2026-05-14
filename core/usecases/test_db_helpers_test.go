package usecases

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
)

func newUsecaseTestSQLiteDB(t *testing.T) *dbadapter.SQLiteDB {
	t.Helper()

	dsn := fmt.Sprintf("file:usecase_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := dbadapter.NewSQLiteDB(dsn)
	if err != nil {
		t.Fatalf("failed to create sqlite db: %v", err)
	}

	ctx := context.Background()
	statements := []string{
		`CREATE TABLE employees (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			uid TEXT UNIQUE NOT NULL
		);`,
		`CREATE TABLE holiday_definitions (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			uid TEXT UNIQUE NOT NULL,
			code TEXT UNIQUE NOT NULL,
			name_en TEXT NOT NULL,
			name_ar TEXT NOT NULL,
			date DATE NOT NULL,
			is_manual BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE holiday_definition_departments (
			holiday_definition_id INTEGER NOT NULL,
			department_uid TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (holiday_definition_id, department_uid)
		);`,
		`CREATE TABLE attendance_exceptions (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			uid TEXT UNIQUE NOT NULL,
			employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE CASCADE,
			attendance_date DATE NOT NULL,
			exception_type TEXT NOT NULL,
			check_in TEXT,
			check_out TEXT,
			grace_minutes INTEGER,
			minutes_delta INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(employee_uid, attendance_date, exception_type)
		);`,
		`CREATE TABLE weekend_config (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			uid TEXT UNIQUE NOT NULL,
			day_of_week INTEGER NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("failed to execute schema statement: %v", err)
		}
	}

	return db
}

func seedWeekendDays(t *testing.T, db *dbadapter.SQLiteDB, days ...int) {
	t.Helper()

	ctx := context.Background()
	for _, day := range days {
		uid := fmt.Sprintf("weekend_%d", day)
		if _, err := db.ExecContext(ctx, `INSERT INTO weekend_config (uid, day_of_week) VALUES ($1, $2)`, uid, day); err != nil {
			t.Fatalf("failed to seed weekend day %d: %v", day, err)
		}
	}
}
