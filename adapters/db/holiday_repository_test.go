package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestHolidayDefinitionRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayDefinitionRepository()
	ctx := context.Background()

	month := 1
	day := 25
	def := &domain.HolidayDefinition{
		UID:          domain.GenerateUID("hdef"),
		Code:         "REVOLUTION_JAN_25",
		NameEN:       "January 25 Revolution",
		NameAR:       "ثورة 25 يناير",
		DefaultMonth: &month,
		DefaultDay:   &day,
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

	// Lunar holidays don't have a fixed month/day
	def := &domain.HolidayDefinition{
		UID:          domain.GenerateUID("hdef"),
		Code:         "EID_AL_FITR",
		NameEN:       "Eid Al Fitr",
		NameAR:       "عيد الفطر",
		DefaultMonth: nil,
		DefaultDay:   nil,
	}

	err := repo.Create(ctx, tdb.SQLiteDB, def)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify nil fields persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, def.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.DefaultMonth != nil {
		t.Errorf("DefaultMonth should be nil, got %v", found.DefaultMonth)
	}
	if found.DefaultDay != nil {
		t.Errorf("DefaultDay should be nil, got %v", found.DefaultDay)
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

func TestHolidayInstanceRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	// First create a holiday definition
	month, day := 1, 25
	def := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month, &day)

	// Create instance for 2026
	instance := &domain.HolidayInstance{
		UID:          domain.GenerateUID("hinst"),
		DefinitionID: def.ID,
		Year:         2026,
		ActualDate:   ParseDate(t, "2026-01-25"),
		ObservedDate: ParseDate(t, "2026-01-25"),
		IsConfirmed:  true,
	}

	err := repo.Create(ctx, tdb.SQLiteDB, instance)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if instance.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestHolidayInstanceRepository_Create_DuplicateDefinitionYear(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	// Create holiday definition
	month, day := 1, 25
	def := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month, &day)

	// Create first instance
	instance1 := &domain.HolidayInstance{
		UID:          domain.GenerateUID("hinst"),
		DefinitionID: def.ID,
		Year:         2026,
		ActualDate:   ParseDate(t, "2026-01-25"),
		ObservedDate: ParseDate(t, "2026-01-25"),
		IsConfirmed:  true,
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, instance1); err != nil {
		t.Fatalf("Create() first instance error = %v", err)
	}

	// Try to create duplicate
	instance2 := &domain.HolidayInstance{
		UID:          domain.GenerateUID("hinst"),
		DefinitionID: def.ID,
		Year:         2026,
		ActualDate:   ParseDate(t, "2026-01-25"),
		ObservedDate: ParseDate(t, "2026-01-26"),
		IsConfirmed:  false,
	}
	err := repo.Create(ctx, tdb.SQLiteDB, instance2)
	if err == nil {
		t.Error("Create() expected error for duplicate definition/year, got nil")
	}
}

func TestHolidayInstanceRepository_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	month, day := 1, 25
	def := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month, &day)
	instance := tdb.SeedHolidayInstance(def.ID, 2026, ParseDate(t, "2026-01-25"), ParseDate(t, "2026-01-25"), true)

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, instance.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.Year != 2026 {
		t.Errorf("GetByID() Year = %v, want 2026", found.Year)
	}
	if !found.IsConfirmed {
		t.Error("GetByID() IsConfirmed should be true")
	}
}

func TestHolidayInstanceRepository_GetByID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, 99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByID() expected nil, got %+v", found)
	}
}

func TestHolidayInstanceRepository_GetByDefinitionAndYear(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	month, day := 1, 25
	def := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month, &day)
	instance := tdb.SeedHolidayInstance(def.ID, 2026, ParseDate(t, "2026-01-25"), ParseDate(t, "2026-01-25"), true)

	found, err := repo.GetByDefinitionAndYear(ctx, tdb.SQLiteDB, def.ID, 2026)
	if err != nil {
		t.Fatalf("GetByDefinitionAndYear() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByDefinitionAndYear() returned nil")
	}
	if found.ID != instance.ID {
		t.Errorf("GetByDefinitionAndYear() ID = %v, want %v", found.ID, instance.ID)
	}
}

func TestHolidayInstanceRepository_GetByDefinitionAndYear_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	month, day := 1, 25
	def := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month, &day)

	found, err := repo.GetByDefinitionAndYear(ctx, tdb.SQLiteDB, def.ID, 2026)
	if err != nil {
		t.Fatalf("GetByDefinitionAndYear() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByDefinitionAndYear() expected nil, got %+v", found)
	}
}

func TestHolidayInstanceRepository_ListByYear(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	// Create multiple definitions and instances
	m1, d1 := 1, 25
	m2, d2 := 7, 23
	def1 := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &m1, &d1)
	def2 := tdb.SeedHolidayDefinition("REVOLUTION_JUL_23", "July 23 Revolution", "ثورة 23 يوليو", &m2, &d2)
	def3 := tdb.SeedHolidayDefinition("EID_AL_FITR", "Eid Al Fitr", "عيد الفطر", nil, nil)

	// Create instances for 2026
	tdb.SeedHolidayInstance(def1.ID, 2026, ParseDate(t, "2026-01-25"), ParseDate(t, "2026-01-25"), true)
	tdb.SeedHolidayInstance(def2.ID, 2026, ParseDate(t, "2026-07-23"), ParseDate(t, "2026-07-23"), true)
	tdb.SeedHolidayInstance(def3.ID, 2026, ParseDate(t, "2026-03-20"), ParseDate(t, "2026-03-20"), false)

	// Create instance for 2025
	tdb.SeedHolidayInstance(def1.ID, 2025, ParseDate(t, "2025-01-25"), ParseDate(t, "2025-01-26"), true)

	// List 2026 only
	instances, err := repo.ListByYear(ctx, tdb.SQLiteDB, 2026)
	if err != nil {
		t.Fatalf("ListByYear() error = %v", err)
	}
	if len(instances) != 3 {
		t.Errorf("ListByYear() returned %d items, want 3", len(instances))
	}

	// Should be ordered by observed_date
	if instances[0].ObservedDate.Format("2006-01-02") != "2026-01-25" {
		t.Errorf("ListByYear()[0].ObservedDate = %v, want 2026-01-25", instances[0].ObservedDate)
	}
}

func TestHolidayInstanceRepository_ListByDateRange(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	// Create definitions and instances
	m1, d1 := 1, 25
	m2, d2 := 3, 20
	m3, d3 := 7, 23
	def1 := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &m1, &d1)
	def2 := tdb.SeedHolidayDefinition("EID_AL_FITR", "Eid Al Fitr", "عيد الفطر", &m2, &d2)
	def3 := tdb.SeedHolidayDefinition("REVOLUTION_JUL_23", "July 23 Revolution", "ثورة 23 يوليو", &m3, &d3)

	tdb.SeedHolidayInstance(def1.ID, 2026, ParseDate(t, "2026-01-25"), ParseDate(t, "2026-01-25"), true)
	tdb.SeedHolidayInstance(def2.ID, 2026, ParseDate(t, "2026-03-20"), ParseDate(t, "2026-03-20"), false)
	tdb.SeedHolidayInstance(def3.ID, 2026, ParseDate(t, "2026-07-23"), ParseDate(t, "2026-07-23"), true)

	// Query Q1 2026 (Jan-Mar)
	start := ParseDate(t, "2026-01-01")
	end := ParseDate(t, "2026-03-31")
	instances, err := repo.ListByDateRange(ctx, tdb.SQLiteDB, start, end)
	if err != nil {
		t.Fatalf("ListByDateRange() error = %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("ListByDateRange() returned %d items, want 2", len(instances))
	}
}

func TestHolidayInstanceRepository_Update(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	month, day := 3, 20
	def := tdb.SeedHolidayDefinition("EID_AL_FITR", "Eid Al Fitr", "عيد الفطر", &month, &day)
	instance := tdb.SeedHolidayInstance(def.ID, 2026, ParseDate(t, "2026-03-20"), ParseDate(t, "2026-03-20"), false)

	// Update to confirmed with adjusted observed date
	instance.IsConfirmed = true
	instance.ObservedDate = ParseDate(t, "2026-03-21")
	notes := "Moon sighting confirmed, moved to Sunday"
	instance.Notes = &notes

	err := repo.Update(ctx, tdb.SQLiteDB, instance)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, instance.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !found.IsConfirmed {
		t.Error("Update() IsConfirmed should be true")
	}
	if found.ObservedDate.Format("2006-01-02") != "2026-03-21" {
		t.Errorf("Update() ObservedDate = %v, want 2026-03-21", found.ObservedDate)
	}
	if found.Notes == nil || *found.Notes != notes {
		t.Errorf("Update() Notes = %v, want %v", found.Notes, notes)
	}
}

func TestHolidayInstanceRepository_MovedHoliday(t *testing.T) {
	// Test case where actual date differs from observed date (weekend adjustment)
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	month, day := 10, 6
	def := tdb.SeedHolidayDefinition("ARMED_FORCES_DAY", "Armed Forces Day", "يوم القوات المسلحة", &month, &day)

	// October 6, 2028 falls on Friday (weekend), moved to Thursday
	notes := "Moved to create long weekend"
	instance := &domain.HolidayInstance{
		UID:          domain.GenerateUID("hinst"),
		DefinitionID: def.ID,
		Year:         2028,
		ActualDate:   ParseDate(t, "2028-10-06"),
		ObservedDate: ParseDate(t, "2028-10-05"),
		IsConfirmed:  true,
		Notes:        &notes,
	}

	instRepo := NewHolidayInstanceRepository()
	if err := instRepo.Create(ctx, tdb.SQLiteDB, instance); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Query should use observed_date for working days calculation
	instances, err := repo.ListByDateRange(ctx, tdb.SQLiteDB, ParseDate(t, "2028-10-01"), ParseDate(t, "2028-10-10"))
	if err != nil {
		t.Fatalf("ListByDateRange() error = %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("ListByDateRange() returned %d items, want 1", len(instances))
	}

	// The observed date should be Oct 5 (Thursday), not Oct 6 (Friday)
	if instances[0].ObservedDate.Format("2006-01-02") != "2028-10-05" {
		t.Errorf("ObservedDate = %v, want 2028-10-05", instances[0].ObservedDate)
	}
	if instances[0].ActualDate.Format("2006-01-02") != "2028-10-06" {
		t.Errorf("ActualDate = %v, want 2028-10-06", instances[0].ActualDate)
	}
}

func TestHolidayInstanceRepository_ListByYear_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewHolidayInstanceRepository()
	ctx := context.Background()

	instances, err := repo.ListByYear(ctx, tdb.SQLiteDB, 2026)
	if err != nil {
		t.Fatalf("ListByYear() error = %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("ListByYear() returned %d items, want 0", len(instances))
	}
}

func TestHolidayInstanceRepository_DeleteCascade(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	ctx := context.Background()

	// Create holiday definition
	month, day := 1, 25
	def := tdb.SeedHolidayDefinition("REVOLUTION_JAN_25", "January 25 Revolution", "ثورة 25 يناير", &month, &day)

	// Create instances for multiple years
	tdb.SeedHolidayInstance(def.ID, 2025, ParseDate(t, "2025-01-25"), ParseDate(t, "2025-01-26"), true)
	tdb.SeedHolidayInstance(def.ID, 2026, ParseDate(t, "2026-01-25"), ParseDate(t, "2026-01-25"), true)

	// Delete the definition
	_, err := tdb.db.ExecContext(ctx, "DELETE FROM holiday_definitions WHERE id = ?", def.ID)
	if err != nil {
		t.Fatalf("Delete definition error = %v", err)
	}

	// Instances should be deleted via CASCADE
	repo := NewHolidayInstanceRepository()
	instances, err := repo.ListByYear(ctx, tdb.SQLiteDB, 2026)
	if err != nil {
		t.Fatalf("ListByYear() error = %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("Instances should be deleted after CASCADE, got %d", len(instances))
	}
}
