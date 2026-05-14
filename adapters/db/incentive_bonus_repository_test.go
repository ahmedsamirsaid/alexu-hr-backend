package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestIncentiveBonusRepository_CreateAndListByEmployeeUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewIncentiveBonusRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Bonus Employee")

	first := &domain.IncentiveBonus{
		EmployeeUID:      employee.UID,
		BonusDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		DecisionNumber:   "BON-1",
		DecisionDate:     time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
		DecisionImageURL: "minio://documents/bonus/bon-1.pdf",
	}
	second := &domain.IncentiveBonus{
		EmployeeUID:      employee.UID,
		BonusDate:        time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
		DecisionNumber:   "BON-2",
		DecisionDate:     time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
		DecisionImageURL: "minio://documents/bonus/bon-2.pdf",
	}

	if err := repo.Create(ctx, tdb.SQLiteDB, first); err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, second); err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	bonuses, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}

	if len(bonuses) != 2 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 2", len(bonuses))
	}
	if bonuses[0].DecisionNumber != "BON-2" {
		t.Fatalf("first listed bonus = %q, want %q", bonuses[0].DecisionNumber, "BON-2")
	}
	if bonuses[1].DecisionNumber != "BON-1" {
		t.Fatalf("second listed bonus = %q, want %q", bonuses[1].DecisionNumber, "BON-1")
	}
}

func TestIncentiveBonusRepository_DeleteByEmployeeUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewIncentiveBonusRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Bonus Employee")

	if err := repo.Create(ctx, tdb.SQLiteDB, &domain.IncentiveBonus{
		EmployeeUID:      employee.UID,
		BonusDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		DecisionNumber:   "BON-1",
		DecisionDate:     time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
		DecisionImageURL: "minio://documents/bonus/bon-1.pdf",
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.DeleteByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID); err != nil {
		t.Fatalf("DeleteByEmployeeUID() error = %v", err)
	}

	bonuses, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}
	if len(bonuses) != 0 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 0", len(bonuses))
	}
}
