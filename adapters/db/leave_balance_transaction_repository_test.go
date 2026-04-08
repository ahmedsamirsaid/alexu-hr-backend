package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestLeaveBalanceTransactionRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	// Create initial transaction
	notes := "Initial balance allocation"
	tx := domain.NewLeaveBalanceTransaction(balance.ID, domain.TransactionTypeInitial, 7, nil, &employee.ID, &notes)

	err := repo.Create(ctx, tdb.SQLiteDB, tx)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if tx.ID == 0 {
		t.Error("Create() did not set ID")
	}
	if tx.CreatedAt.IsZero() {
		t.Error("Create() did not set CreatedAt")
	}
}

func TestLeaveBalanceTransactionRepository_Create_Deduction(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	// Create a leave record to link to
	leaveRecordRepo := NewLeaveRecordRepository()
	leaveRecord := &domain.LeaveRecord{
		UID:         domain.GenerateUID("leave"),
		EmployeeID:  employee.ID,
		LeaveTypeID: leaveType.ID,
		StartDate:   ParseDate(t, "2026-01-20"),
		EndDate:     ParseDate(t, "2026-01-21"),
		Days:        2,
		RecordedAt:  time.Now(),
		RecordedBy:  &employee.ID,
	}
	if err := leaveRecordRepo.Create(ctx, tdb.SQLiteDB, leaveRecord); err != nil {
		t.Fatalf("Create leave record error = %v", err)
	}

	// Create deduction transaction
	tx := domain.NewLeaveBalanceTransaction(balance.ID, domain.TransactionTypeDeduct, -2, &leaveRecord.ID, &employee.ID, nil)

	err := repo.Create(ctx, tdb.SQLiteDB, tx)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if tx.Days != -2 {
		t.Errorf("Create() Days = %v, want -2", tx.Days)
	}
}

func TestLeaveBalanceTransactionRepository_ListByBalance(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 2)

	// Create multiple transactions
	initialNotes := "Initial balance"
	tx1 := domain.NewLeaveBalanceTransaction(balance.ID, domain.TransactionTypeInitial, 7, nil, &employee.ID, &initialNotes)
	if err := repo.Create(ctx, tdb.SQLiteDB, tx1); err != nil {
		t.Fatalf("Create() tx1 error = %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	tx2 := domain.NewLeaveBalanceTransaction(balance.ID, domain.TransactionTypeDeduct, -2, nil, &employee.ID, nil)
	if err := repo.Create(ctx, tdb.SQLiteDB, tx2); err != nil {
		t.Fatalf("Create() tx2 error = %v", err)
	}

	// List transactions
	txs, err := repo.ListByBalance(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("ListByBalance() error = %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("ListByBalance() returned %d items, want 2", len(txs))
	}

	// Should be ordered by created_at
	if txs[0].TransactionType != domain.TransactionTypeInitial {
		t.Errorf("ListByBalance()[0].TransactionType = %v, want INITIAL", txs[0].TransactionType)
	}
	if txs[1].TransactionType != domain.TransactionTypeDeduct {
		t.Errorf("ListByBalance()[1].TransactionType = %v, want DEDUCT", txs[1].TransactionType)
	}
}

func TestLeaveBalanceTransactionRepository_ListByBalance_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies but no transactions
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	txs, err := repo.ListByBalance(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("ListByBalance() error = %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("ListByBalance() returned %d items, want 0", len(txs))
	}
}

func TestLeaveBalanceTransactionRepository_AllTransactionTypes(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	// Test all transaction types
	types := []domain.TransactionType{
		domain.TransactionTypeInitial,
		domain.TransactionTypeDeduct,
		domain.TransactionTypeRefund,
		domain.TransactionTypeAdjustment,
		domain.TransactionTypeCarryOver,
	}

	for _, txType := range types {
		tx := domain.NewLeaveBalanceTransaction(balance.ID, txType, 1, nil, &employee.ID, nil)
		if err := repo.Create(ctx, tdb.SQLiteDB, tx); err != nil {
			t.Errorf("Create() with type %v error = %v", txType, err)
		}
	}

	// Verify all were created
	txs, err := repo.ListByBalance(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("ListByBalance() error = %v", err)
	}
	if len(txs) != len(types) {
		t.Errorf("ListByBalance() returned %d items, want %d", len(txs), len(types))
	}
}

func TestLeaveBalanceTransactionRepository_NullableFields(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	// Create transaction with all nullable fields as nil
	tx := domain.NewLeaveBalanceTransaction(balance.ID, domain.TransactionTypeAdjustment, 1, nil, nil, nil)
	if err := repo.Create(ctx, tdb.SQLiteDB, tx); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify
	txs, err := repo.ListByBalance(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("ListByBalance() error = %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("ListByBalance() returned %d items, want 1", len(txs))
	}

	if txs[0].LeaveRecordID != nil {
		t.Errorf("LeaveRecordID should be nil, got %v", txs[0].LeaveRecordID)
	}
	if txs[0].CreatedBy != nil {
		t.Errorf("CreatedBy should be nil, got %v", txs[0].CreatedBy)
	}
	if txs[0].Notes != nil {
		t.Errorf("Notes should be nil, got %v", txs[0].Notes)
	}
}

func TestLeaveBalanceTransactionRepository_AuditTrail(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveBalanceTransactionRepository()
	ctx := context.Background()

	// Seed dependencies
	employee := tdb.SeedEmployee("Ahmed Hassan")
	leaveType := tdb.SeedLeaveType("CASUAL", "Casual Leave", "العارضة", 7)
	balance := tdb.SeedLeaveBalance(employee.ID, leaveType.ID, 2026, 7, 0)

	// Create a sequence of transactions simulating balance lifecycle
	transactions := []struct {
		txType domain.TransactionType
		days   int
		notes  string
	}{
		{domain.TransactionTypeInitial, 7, "Initial allocation"},
		{domain.TransactionTypeDeduct, -2, "Leave taken Jan 20-21"},
		{domain.TransactionTypeDeduct, -1, "Leave taken Jan 25"},
		{domain.TransactionTypeRefund, 1, "Cancelled leave Jan 25"},
		{domain.TransactionTypeAdjustment, 2, "Manager adjustment"},
	}

	for _, txData := range transactions {
		notes := txData.notes
		tx := domain.NewLeaveBalanceTransaction(balance.ID, txData.txType, txData.days, nil, &employee.ID, &notes)
		if err := repo.Create(ctx, tdb.SQLiteDB, tx); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		time.Sleep(5 * time.Millisecond) // Ensure ordering
	}

	// Verify audit trail
	txs, err := repo.ListByBalance(ctx, tdb.SQLiteDB, balance.ID)
	if err != nil {
		t.Fatalf("ListByBalance() error = %v", err)
	}

	// Calculate expected final balance
	expectedBalance := 0
	for _, tx := range txs {
		expectedBalance += tx.Days
	}

	// 7 - 2 - 1 + 1 + 2 = 7
	if expectedBalance != 7 {
		t.Errorf("Final calculated balance = %d, want 7", expectedBalance)
	}

	// Verify order maintained
	if len(txs) != 5 {
		t.Errorf("Should have 5 transactions, got %d", len(txs))
	}
}
