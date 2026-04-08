package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestLeaveRecordRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	record := &domain.LeaveRecord{
		UID:         domain.GenerateUID("leave"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		StartDate:   ParseDate(t, "2026-01-20"),
		EndDate:     ParseDate(t, "2026-01-21"),
		Days:        2,
		RecordedAt:  time.Now(),
		RecordedBy:  &employee.ID,
	}

	err := repo.Create(ctx, tdb.SQLiteDB, record)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if record.ID == 0 {
		t.Error("Create() did not set ID")
	}
	if record.CreatedAt.IsZero() {
		t.Error("Create() did not set CreatedAt")
	}
}

func TestLeaveRecordRepository_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies and create record
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	notes := "Family emergency"
	record := &domain.LeaveRecord{
		UID:         domain.GenerateUID("leave"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		StartDate:   ParseDate(t, "2026-01-20"),
		EndDate:     ParseDate(t, "2026-01-21"),
		Days:        2,
		RecordedAt:  time.Now(),
		RecordedBy:  &employee.ID,
		Notes:       &notes,
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, record); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get by ID
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, record.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.Days != 2 {
		t.Errorf("GetByID() Days = %v, want 2", found.Days)
	}
	if found.Notes == nil || *found.Notes != notes {
		t.Errorf("GetByID() Notes = %v, want %v", found.Notes, notes)
	}
}

func TestLeaveRecordRepository_GetByID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, 99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByID() expected nil for non-existent ID, got %+v", found)
	}
}

func TestLeaveRecordRepository_GetByUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies and create record
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	record := &domain.LeaveRecord{
		UID:         domain.GenerateUID("leave"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		StartDate:   ParseDate(t, "2026-01-20"),
		EndDate:     ParseDate(t, "2026-01-20"),
		Days:        1,
		RecordedAt:  time.Now(),
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, record); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get by UID
	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, record.UID)
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByUID() returned nil")
	}
	if found.ID != record.ID {
		t.Errorf("GetByUID() ID = %v, want %v", found.ID, record.ID)
	}
}

func TestLeaveRecordRepository_GetByUID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, "leave_nonexistent")
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByUID() expected nil for non-existent UID, got %+v", found)
	}
}

func TestLeaveRecordRepository_ListByEmployee(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	// Create multiple records
	records := []struct {
		startDate string
		endDate   string
		days      int
	}{
		{"2026-01-10", "2026-01-10", 1},
		{"2026-01-15", "2026-01-16", 2},
		{"2026-01-20", "2026-01-21", 2},
	}

	for _, r := range records {
		record := &domain.LeaveRecord{
			UID:         domain.GenerateUID("leave"),
			EmployeeID:  employee.ID,
			LeaveTypeID: leaveType.ID,
			StartDate:   ParseDate(t, r.startDate),
			EndDate:     ParseDate(t, r.endDate),
			Days:        r.days,
			RecordedAt:  time.Now(),
		}
		if err := repo.Create(ctx, tdb.SQLiteDB, record); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// List by employee
	found, err := repo.ListByEmployee(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("ListByEmployee() error = %v", err)
	}
	if len(found) != 3 {
		t.Errorf("ListByEmployee() returned %d items, want 3", len(found))
	}

	// Should be ordered by start_date DESC
	if found[0].StartDate.Format("2006-01-02") != "2026-01-20" {
		t.Errorf("ListByEmployee()[0].StartDate = %v, want 2026-01-20", found[0].StartDate)
	}
}

func TestLeaveRecordRepository_ListByEmployee_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	employee := tdb.SeedEmployee("Ahmed Hassan")

	found, err := repo.ListByEmployee(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("ListByEmployee() error = %v", err)
	}
	if len(found) != 0 {
		t.Errorf("ListByEmployee() returned %d items, want 0", len(found))
	}
}

func TestLeaveRecordRepository_ListByEmployeeAndDateRange(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	// Create records in different date ranges
	dates := []struct {
		startDate string
		endDate   string
	}{
		{"2026-01-05", "2026-01-05"},  // Outside range (before)
		{"2026-01-10", "2026-01-11"},  // Inside range
		{"2026-01-15", "2026-01-15"},  // Inside range
		{"2026-01-25", "2026-01-26"},  // Outside range (after)
	}

	for _, d := range dates {
		record := &domain.LeaveRecord{
			UID:         domain.GenerateUID("leave"),
			EmployeeID:  employee.ID,
			LeaveTypeID: leaveType.ID,
			StartDate:   ParseDate(t, d.startDate),
			EndDate:     ParseDate(t, d.endDate),
			Days:        1,
			RecordedAt:  time.Now(),
		}
		if err := repo.Create(ctx, tdb.SQLiteDB, record); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// Query date range
	start := ParseDate(t, "2026-01-10")
	end := ParseDate(t, "2026-01-20")
	found, err := repo.ListByEmployeeAndDateRange(ctx, tdb.SQLiteDB, employee.ID, start, end)
	if err != nil {
		t.Fatalf("ListByEmployeeAndDateRange() error = %v", err)
	}
	if len(found) != 2 {
		t.Errorf("ListByEmployeeAndDateRange() returned %d items, want 2", len(found))
	}
}

func TestLeaveRecordRepository_ListByEmployeeAndType(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	casualType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	annualType := tdb.SeedLeaveType("ANNUAL", "Annual Leave", "السنوية", 21)

	// Create records with different leave types
	for i := 0; i < 3; i++ {
		record := &domain.LeaveRecord{
			UID:         domain.GenerateUID("leave"),
			EmployeeID:  employee.ID,
			LeaveTypeID: casualType.ID,
			StartDate:   ParseDate(t, "2026-01-10"),
			EndDate:     ParseDate(t, "2026-01-10"),
			Days:        1,
			RecordedAt:  time.Now(),
		}
		repo.Create(ctx, tdb.SQLiteDB, record)
	}
	for i := 0; i < 2; i++ {
		record := &domain.LeaveRecord{
			UID:         domain.GenerateUID("leave"),
			EmployeeID:  employee.ID,
			LeaveTypeID: annualType.ID,
			StartDate:   ParseDate(t, "2026-02-10"),
			EndDate:     ParseDate(t, "2026-02-10"),
			Days:        1,
			RecordedAt:  time.Now(),
		}
		repo.Create(ctx, tdb.SQLiteDB, record)
	}

	// List casual leave only
	found, err := repo.ListByEmployeeAndType(ctx, tdb.SQLiteDB, employee.ID, casualType.ID)
	if err != nil {
		t.Fatalf("ListByEmployeeAndType() error = %v", err)
	}
	if len(found) != 3 {
		t.Errorf("ListByEmployeeAndType() returned %d items, want 3", len(found))
	}

	// List annual leave only
	found, err = repo.ListByEmployeeAndType(ctx, tdb.SQLiteDB, employee.ID, annualType.ID)
	if err != nil {
		t.Fatalf("ListByEmployeeAndType() error = %v", err)
	}
	if len(found) != 2 {
		t.Errorf("ListByEmployeeAndType() returned %d items, want 2", len(found))
	}
}

func TestLeaveRecordRepository_Create_WithNullRecordedBy(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	record := &domain.LeaveRecord{
		UID:         domain.GenerateUID("leave"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		StartDate:   ParseDate(t, "2026-01-20"),
		EndDate:     ParseDate(t, "2026-01-20"),
		Days:        1,
		RecordedAt:  time.Now(),
		RecordedBy:  nil, // Not recorded by anyone specific
	}

	if err := repo.Create(ctx, tdb.SQLiteDB, record); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, record.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.RecordedBy != nil {
		t.Errorf("GetByID() RecordedBy = %v, want nil", found.RecordedBy)
	}
}

func TestLeaveRecordRepository_Transaction(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveRecordRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	// Start transaction
	tx, err := tdb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}

	// Create record in transaction
	record := &domain.LeaveRecord{
		UID:         domain.GenerateUID("leave"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		StartDate:   ParseDate(t, "2026-01-20"),
		EndDate:     ParseDate(t, "2026-01-20"),
		Days:        1,
		RecordedAt:  time.Now(),
	}
	if err := repo.Create(ctx, tx, record); err != nil {
		tx.Rollback()
		t.Fatalf("Create() in tx error = %v", err)
	}

	// Rollback
	tx.Rollback()

	// Verify not persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, record.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Error("Record should not exist after rollback")
	}
}
