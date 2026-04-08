package db

import (
	"context"
	"testing"
)

func TestLeaveTypeRepository_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed leave type
	lt := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	// Get by ID
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, lt.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.Code != lt.Code {
		t.Errorf("GetByID() Code = %v, want %v", found.Code, lt.Code)
	}
	if found.DefaultBalance != 7 {
		t.Errorf("GetByID() DefaultBalance = %v, want 7", found.DefaultBalance)
	}
}

func TestLeaveTypeRepository_GetByID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, 99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByID() expected nil for non-existent ID, got %+v", found)
	}
}

func TestLeaveTypeRepository_GetByUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed leave type
	lt := tdb.SeedLeaveType("ANNUAL", "Annual Leave", "السنوية", 21)

	// Get by UID
	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, lt.UID)
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByUID() returned nil")
	}
	if found.ID != lt.ID {
		t.Errorf("GetByUID() ID = %v, want %v", found.ID, lt.ID)
	}
	if found.NameEN != "Annual Leave" {
		t.Errorf("GetByUID() NameEN = %v, want 'Annual Leave'", found.NameEN)
	}
}

func TestLeaveTypeRepository_GetByUID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, "ltype_nonexistent")
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByUID() expected nil for non-existent UID, got %+v", found)
	}
}

func TestLeaveTypeRepository_GetByCode(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed leave type
	lt := tdb.SeedLeaveType("SICK", "Sick Leave", "المرضية", 15)

	// Get by code
	found, err := repo.GetByCode(ctx, tdb.SQLiteDB, "SICK")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByCode() returned nil")
	}
	if found.ID != lt.ID {
		t.Errorf("GetByCode() ID = %v, want %v", found.ID, lt.ID)
	}
	if found.NameAR != "المرضية" {
		t.Errorf("GetByCode() NameAR = %v, want 'المرضية'", found.NameAR)
	}
}

func TestLeaveTypeRepository_GetByCode_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	found, err := repo.GetByCode(ctx, tdb.SQLiteDB, "NONEXISTENT")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByCode() expected nil for non-existent code, got %+v", found)
	}
}

func TestLeaveTypeRepository_List(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed multiple leave types
	tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	tdb.SeedLeaveType("ANNUAL", "Annual Leave", "السنوية", 21)
	tdb.SeedLeaveType("SICK", "Sick Leave", "المرضية", 15)

	// List all (activeOnly=false)
	types, err := repo.List(ctx, tdb.SQLiteDB, false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(types) != 3 {
		t.Errorf("List() returned %d items, want 3", len(types))
	}

	// Verify order (by ID)
	if types[0].Code != "CASUAL" {
		t.Errorf("List()[0].Code = %v, want CASUAL", types[0].Code)
	}
	if types[1].Code != "ANNUAL" {
		t.Errorf("List()[1].Code = %v, want ANNUAL", types[1].Code)
	}
	if types[2].Code != "SICK" {
		t.Errorf("List()[2].Code = %v, want SICK", types[2].Code)
	}
}

func TestLeaveTypeRepository_List_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	types, err := repo.List(ctx, tdb.SQLiteDB, false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(types) != 0 {
		t.Errorf("List() returned %d items on empty database, want 0", len(types))
	}
}

func TestLeaveTypeRepository_List_ActiveOnly(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed leave types
	casual := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	tdb.SeedLeaveType("ANNUAL", "Annual Leave", "السنوية", 21)

	// Deactivate CASUAL
	err := repo.SetActive(ctx, tdb.SQLiteDB, casual.UID, false)
	if err != nil {
		t.Fatalf("SetActive() error = %v", err)
	}

	// List active only
	activeTypes, err := repo.List(ctx, tdb.SQLiteDB, true)
	if err != nil {
		t.Fatalf("List(activeOnly=true) error = %v", err)
	}
	if len(activeTypes) != 1 {
		t.Errorf("List(activeOnly=true) returned %d items, want 1", len(activeTypes))
	}
	if activeTypes[0].Code != "ANNUAL" {
		t.Errorf("List(activeOnly=true)[0].Code = %v, want ANNUAL", activeTypes[0].Code)
	}

	// List all
	allTypes, err := repo.List(ctx, tdb.SQLiteDB, false)
	if err != nil {
		t.Fatalf("List(activeOnly=false) error = %v", err)
	}
	if len(allTypes) != 2 {
		t.Errorf("List(activeOnly=false) returned %d items, want 2", len(allTypes))
	}
}

func TestLeaveTypeRepository_SetActive(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed leave type
	lt := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	// Verify initially active
	found, _ := repo.GetByUID(ctx, tdb.SQLiteDB, lt.UID)
	if !found.IsActive {
		t.Error("leave type should be active initially")
	}

	// Deactivate
	err := repo.SetActive(ctx, tdb.SQLiteDB, lt.UID, false)
	if err != nil {
		t.Fatalf("SetActive(false) error = %v", err)
	}

	// Verify deactivated
	found, _ = repo.GetByUID(ctx, tdb.SQLiteDB, lt.UID)
	if found.IsActive {
		t.Error("leave type should be inactive after SetActive(false)")
	}

	// Reactivate
	err = repo.SetActive(ctx, tdb.SQLiteDB, lt.UID, true)
	if err != nil {
		t.Fatalf("SetActive(true) error = %v", err)
	}

	// Verify reactivated
	found, _ = repo.GetByUID(ctx, tdb.SQLiteDB, lt.UID)
	if !found.IsActive {
		t.Error("leave type should be active after SetActive(true)")
	}
}

func TestLeaveTypeRepository_SetActive_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	err := repo.SetActive(ctx, tdb.SQLiteDB, "ltype_nonexistent", false)
	if err != ErrLeaveTypeNotFound {
		t.Errorf("SetActive() error = %v, want ErrLeaveTypeNotFound", err)
	}
}

func TestLeaveTypeRepository_NullableFields(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed leave type with all nullable fields
	lt := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, lt.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	// max_consecutive should be set (2)
	if found.MaxConsecutive == nil || *found.MaxConsecutive != 2 {
		t.Errorf("GetByID() MaxConsecutive = %v, want 2", found.MaxConsecutive)
	}

	// recording_deadline_days should be set (2)
	if found.RecordingDeadlineDays == nil || *found.RecordingDeadlineDays != 2 {
		t.Errorf("GetByID() RecordingDeadlineDays = %v, want 2", found.RecordingDeadlineDays)
	}

	// advance_notice_days should be nil
	if found.AdvanceNoticeDays != nil {
		t.Errorf("GetByID() AdvanceNoticeDays = %v, want nil", found.AdvanceNoticeDays)
	}
}

func TestLeaveTypeRepository_IsActive(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	// Seed active leave type
	lt := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, lt.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !found.IsActive {
		t.Error("GetByID() IsActive should be true")
	}
}
