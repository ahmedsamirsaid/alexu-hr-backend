package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestHolidayDefinitionRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	def := &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     "REVOLUTION_JAN_25",
		NameEN:   "January 25 Revolution",
		NameAR:   "ثورة 25 يناير",
		Date:     time.Date(2026, time.January, 25, 0, 0, 0, 0, time.UTC),
		IsManual: true,
	}

	err := repo.Create(ctx, tdb.SQLiteDB, def)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if def.ID == 0 {
		t.Error("Create() did not set ID")
	}
	if def.CreatedAt.IsZero() {
		t.Error("Create() did not set CreatedAt")
	}
}

func TestHolidayDefinitionRepository_Create_LunarHoliday(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	def := &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     "EID_AL_FITR",
		NameEN:   "Eid Al Fitr",
		NameAR:   "عيد الفطر",
		Date:     time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC),
		IsManual: false,
	}

	err := repo.Create(ctx, tdb.SQLiteDB, def)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify date fields persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, def.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.Date.IsZero() {
		t.Error("Date should be persisted, got zero value")
	}
}

func TestHolidayDefinitionRepository_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	month := 7
	day := 23
	def := tdb.SeedHolidayDefinition("REVOLUTION_JUL_23", "July 23 Revolution", "ثورة 23 يوليو", &month, &day)

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, def.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.Code != "REVOLUTION_JUL_23" {
		t.Errorf("GetByID() Code = %v, want REVOLUTION_JUL_23", found.Code)
	}
	if found.NameAR != "ثورة 23 يوليو" {
		t.Errorf("GetByID() NameAR = %v, want 'ثورة 23 يوليو'", found.NameAR)
	}
}

func TestHolidayDefinitionRepository_GetByID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, 99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByID() expected nil for non-existent ID, got %+v", found)
	}
}

func TestHolidayDefinitionRepository_GetByCode(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	month := 1
	day := 7
	def := tdb.SeedHolidayDefinition("CHRISTMAS_COPTIC", "Coptic Christmas", "عيد الميلاد المجيد", &month, &day)

	found, err := repo.GetByCode(ctx, tdb.SQLiteDB, "CHRISTMAS_COPTIC")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByCode() returned nil")
	}
	if found.ID != def.ID {
		t.Errorf("GetByCode() ID = %v, want %v", found.ID, def.ID)
	}
}

func TestHolidayDefinitionRepository_GetByCode_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	found, err := repo.GetByCode(ctx, tdb.SQLiteDB, "NONEXISTENT")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByCode() expected nil, got %+v", found)
	}
}

func TestHolidayDefinitionRepository_List(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	// Seed multiple holiday definitions
	month1, day1 := 1, 25
	month2, day2 := 7, 23
	tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month1, &day1)
	tdb.SeedHolidayDefinition("REVOLUTION_JUL_23", "July 23 Revolution", "ثورة 23 يوليو", &month2, &day2)
	tdb.SeedHolidayDefinition("EID_AL_FITR", "Eid Al Fitr", "عيد الفطر", nil, nil)

	defs, err := repo.List(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(defs) != 3 {
		t.Errorf("List() returned %d items, want 3", len(defs))
	}
}

func TestHolidayDefinitionRepository_List_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	defs, err := repo.List(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(defs) != 0 {
		t.Errorf("List() returned %d items, want 0", len(defs))
	}
}

func TestHolidayDefinitionRepository_Create_WithDepartments(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	deptUID := domain.GenerateUID("dep")
	if _, err := tdb.db.ExecContext(ctx, `INSERT INTO departments (uid, code, name_en, is_active) VALUES (?, ?, ?, 1)`, deptUID, "FINANCE", "Finance"); err != nil {
		t.Fatalf("failed to seed department: %v", err)
	}

	def := &domain.HolidayDefinition{
		UID:            domain.GenerateUID("hdef"),
		Code:           "FINANCE_DAY",
		NameEN:         "Finance Day",
		NameAR:         "يوم المالية",
		Date:           time.Date(2026, time.March, 11, 0, 0, 0, 0, time.UTC),
		IsManual:       true,
		DepartmentUIDs: []string{deptUID},
	}

	if err := repo.Create(ctx, tdb.SQLiteDB, def); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, def.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if len(found.DepartmentUIDs) != 1 || found.DepartmentUIDs[0] != deptUID {
		t.Fatalf("unexpected departments: %+v", found.DepartmentUIDs)
	}

	defs, err := repo.ListByDateRangeForDepartment(ctx, tdb.SQLiteDB, def.Date.AddDate(0, 0, -1), def.Date.AddDate(0, 0, 1), deptUID)
	if err != nil {
		t.Fatalf("ListByDateRangeForDepartment() error = %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 scoped holiday, got %d", len(defs))
	}
}
