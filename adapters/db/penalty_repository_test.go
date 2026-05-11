package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestPenaltyRepository_CreateAndListByEmployeeUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewPenaltyRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Penalty Employee")

	first := &domain.Penalty{
		EmployeeUID:            employee.UID,
		PenaltyType:            "warning",
		PenaltyReason:          "Late attendance",
		PenaltyDecisionNumber:  "DEC-1",
		PenaltyDecisionDate:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PenaltyDecisionFileURL: "minio://documents/penalties/dec-1.pdf",
	}
	second := &domain.Penalty{
		EmployeeUID:            employee.UID,
		PenaltyType:            "deduction",
		PenaltyReason:          "Repeated absence",
		PenaltyDecisionNumber:  "DEC-2",
		PenaltyDecisionDate:    time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
		PenaltyDecisionFileURL: "minio://documents/penalties/dec-2.pdf",
	}

	if err := repo.Create(ctx, tdb.SQLiteDB, first); err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, second); err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	penalties, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}

	if len(penalties) != 2 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 2", len(penalties))
	}
	if penalties[0].PenaltyDecisionNumber != "DEC-2" {
		t.Fatalf("first listed penalty = %q, want %q", penalties[0].PenaltyDecisionNumber, "DEC-2")
	}
	if penalties[1].PenaltyDecisionNumber != "DEC-1" {
		t.Fatalf("second listed penalty = %q, want %q", penalties[1].PenaltyDecisionNumber, "DEC-1")
	}
}

func TestPenaltyRepository_CreateRemovalFiltersPenaltyFromList(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewPenaltyRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Penalty Employee")

	penalty := &domain.Penalty{
		EmployeeUID:            employee.UID,
		PenaltyType:            "warning",
		PenaltyReason:          "Late attendance",
		PenaltyDecisionNumber:  "DEC-1",
		PenaltyDecisionDate:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PenaltyDecisionFileURL: "minio://documents/penalties/dec-1.pdf",
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, penalty); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.CreateRemoval(ctx, tdb.SQLiteDB, &domain.PenaltyRemoval{
		PenaltyID:                        penalty.ID,
		PenaltyRemovalType:               "withdrawal",
		PenaltyRemovalNumber:             "WD-1",
		PenaltyRemovalDate:               time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC),
		Notes:                            "Penalty withdrawn",
		PenaltyWithdrawalDecisionFileURL: "minio://documents/penalties/withdrawal.pdf",
	}); err != nil {
		t.Fatalf("CreateRemoval() error = %v", err)
	}

	penalties, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}
	if len(penalties) != 0 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 0", len(penalties))
	}
}

func TestPenaltyRepository_DeleteByEmployeeUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewPenaltyRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Penalty Employee")

	if err := repo.Create(ctx, tdb.SQLiteDB, &domain.Penalty{
		EmployeeUID:            employee.UID,
		PenaltyType:            "warning",
		PenaltyReason:          "Issue",
		PenaltyDecisionNumber:  "DEC-1",
		PenaltyDecisionDate:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PenaltyDecisionFileURL: "minio://documents/penalties/dec-1.pdf",
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.DeleteByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID); err != nil {
		t.Fatalf("DeleteByEmployeeUID() error = %v", err)
	}

	penalties, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}
	if len(penalties) != 0 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 0", len(penalties))
	}
}
