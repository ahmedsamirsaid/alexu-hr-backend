package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestLeaveBalanceRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	balance := &domain.LeaveBalance{
		UID:         domain.GenerateUID("lbal"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		Year:        2026,
		TotalDays:   7,
		UsedDays:    0,
	}

	err := repo.Create(ctx, tdb.SQLiteDB, balance)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if balance.ID == 0 {
		t.Error("Create() did not set ID")
	}
	if balance.CreatedAt.IsZero() {
		t.Error("Create() did not set CreatedAt")
	}
}

func TestLeaveBalanceRepository_Create_DuplicateEmployeeTypeYear(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	// Create first balance
	balance1 := &domain.LeaveBalance{
		UID:         domain.GenerateUID("lbal"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		Year:        2026,
		TotalDays:   7,
		UsedDays:    0,
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, balance1); err != nil {
		t.Fatalf("Create() first balance error = %v", err)
	}

	// Try to create duplicate
	balance2 := &domain.LeaveBalance{
		UID:         domain.GenerateUID("lbal"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		Year:        2026,
		TotalDays:   7,
		UsedDays:    0,
	}
	err := repo.Create(ctx, tdb.SQLiteDB, balance2)
	if err == nil {
		t.Error("Create() expected error for duplicate employee/type/year, got nil")
	}
}

func TestLeaveBalanceRepository_GetByID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies and balance
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 2)

	// Get by ID
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByID() returned nil")
	}
	if found.TotalDays != 7 {
		t.Errorf("GetByID() TotalDays = %v, want 7", found.TotalDays)
	}
	if found.UsedDays != 2 {
		t.Errorf("GetByID() UsedDays = %v, want 2", found.UsedDays)
	}
}

func TestLeaveBalanceRepository_GetByID_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, 99999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByID() expected nil for non-existent ID, got %+v", found)
	}
}

func TestLeaveBalanceRepository_GetByEmployeeAndTypeAndYear(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies and balance
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 3)

	// Get by employee/type/year
	found, err := repo.GetByEmployeeAndTypeAndYear(ctx, tdb.SQLiteDB, employee.ID, leaveType.ID, 2026)
	if err != nil {
		t.Fatalf("GetByEmployeeAndTypeAndYear() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByEmployeeAndTypeAndYear() returned nil")
	}
	if found.ID != balance.ID {
		t.Errorf("GetByEmployeeAndTypeAndYear() ID = %v, want %v", found.ID, balance.ID)
	}
	if found.UsedDays != 3 {
		t.Errorf("GetByEmployeeAndTypeAndYear() UsedDays = %v, want 3", found.UsedDays)
	}
}

func TestLeaveBalanceRepository_GetByEmployeeAndTypeAndYear_NotFound(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed employee and leave type but no balance
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)

	found, err := repo.GetByEmployeeAndTypeAndYear(ctx, tdb.SQLiteDB, employee.ID, leaveType.ID, 2026)
	if err != nil {
		t.Fatalf("GetByEmployeeAndTypeAndYear() error = %v", err)
	}
	if found != nil {
		t.Errorf("GetByEmployeeAndTypeAndYear() expected nil, got %+v", found)
	}
}

func TestLeaveBalanceRepository_Update(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies and balance
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	// Update used days
	balance.UsedDays = 2
	err := repo.Update(ctx, tdb.SQLiteDB, balance)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update persisted
	found, err := repo.GetByID(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.UsedDays != 2 {
		t.Errorf("Update() UsedDays = %v, want 2", found.UsedDays)
	}
}

func TestLeaveBalanceRepository_ListByEmployee(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	casualType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	annualType := tdb.SeedLeaveType("ANNUAL", "Annual Leave", "السنوية", 21)

	// Create balances for different years and types
	tdb.SeedLeaveBalance(employee.ID, casualType.ID, 2025, 7, 5)
	tdb.SeedLeaveBalance(employee.ID, casualType.ID, 2026, 7, 2)
	tdb.SeedLeaveBalance(employee.ID, annualType.ID, 2026, 21, 10)

	// List by employee
	balances, err := repo.ListByEmployee(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("ListByEmployee() error = %v", err)
	}
	if len(balances) != 3 {
		t.Errorf("ListByEmployee() returned %d items, want 3", len(balances))
	}

	// Should be ordered by year DESC, then leave_type_id
	if balances[0].Year != 2026 {
		t.Errorf("ListByEmployee()[0].Year = %v, want 2026", balances[0].Year)
	}
}

func TestLeaveBalanceRepository_ListByEmployeeAndYear(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	casualType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	annualType := tdb.SeedLeaveType("ANNUAL", "Annual Leave", "السنوية", 21)

	// Create balances
	tdb.SeedLeaveBalance(employee.ID, casualType.ID, 2025, 7, 5)
	tdb.SeedLeaveBalance(employee.ID, casualType.ID, 2026, 7, 2)
	tdb.SeedLeaveBalance(employee.ID, annualType.ID, 2026, 21, 10)

	// List by employee and year
	balances, err := repo.ListByEmployeeAndYear(ctx, tdb.SQLiteDB, employee.ID, 2026)
	if err != nil {
		t.Fatalf("ListByEmployeeAndYear() error = %v", err)
	}
	if len(balances) != 2 {
		t.Errorf("ListByEmployeeAndYear() returned %d items, want 2", len(balances))
	}

	// All should be 2026
	for _, b := range balances {
		if b.Year != 2026 {
			t.Errorf("ListByEmployeeAndYear() contains balance with year %d, want 2026", b.Year)
		}
	}
}

func TestLeaveBalanceRepository_ListByEmployee_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed employee but no balances
	employee := tdb.SeedEmployee("Ahmed Hassan")

	balances, err := repo.ListByEmployee(ctx, tdb.SQLiteDB, employee.ID)
	if err != nil {
		t.Fatalf("ListByEmployee() error = %v", err)
	}
	if len(balances) != 0 {
		t.Errorf("ListByEmployee() returned %d items, want 0", len(balances))
	}
}

func TestLeaveBalanceRepository_RemainingBalance(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceRepository()
	ctx := context.Background()

	// Seed dependencies and balance
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 3)

	found, err := repo.GetByID(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	remaining := found.TotalDays - found.UsedDays
	if remaining != 4 {
		t.Errorf("Remaining balance = %d, want 4", remaining)
	}
}
