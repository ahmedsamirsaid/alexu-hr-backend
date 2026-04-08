package db

import (
	"context"
	"testing"
)

func TestWeekendConfigRepository_List(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	// Seed weekend config (Egypt: Friday=5, Saturday=6)
	tdb.SeedWeekendConfig(5, 6)

	configs, err := repo.List(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("List() returned %d items, want 2", len(configs))
	}

	// Should be ordered by day_of_week
	if configs[0].DayOfWeek != 5 {
		t.Errorf("List()[0].DayOfWeek = %v, want 5 (Friday)", configs[0].DayOfWeek)
	}
	if configs[1].DayOfWeek != 6 {
		t.Errorf("List()[1].DayOfWeek = %v, want 6 (Saturday)", configs[1].DayOfWeek)
	}
}

func TestWeekendConfigRepository_List_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	configs, err := repo.List(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("List() returned %d items, want 0", len(configs))
	}
}

func TestWeekendConfigRepository_GetWeekendDays(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	// Seed weekend config (Egypt: Friday=5, Saturday=6)
	tdb.SeedWeekendConfig(5, 6)

	days, err := repo.GetWeekendDays(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("GetWeekendDays() error = %v", err)
	}
	if len(days) != 2 {
		t.Errorf("GetWeekendDays() returned %d items, want 2", len(days))
	}

	if days[0] != 5 {
		t.Errorf("GetWeekendDays()[0] = %v, want 5", days[0])
	}
	if days[1] != 6 {
		t.Errorf("GetWeekendDays()[1] = %v, want 6", days[1])
	}
}

func TestWeekendConfigRepository_GetWeekendDays_Empty(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	days, err := repo.GetWeekendDays(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("GetWeekendDays() error = %v", err)
	}
	if len(days) != 0 {
		t.Errorf("GetWeekendDays() returned %d items, want 0", len(days))
	}
}

func TestWeekendConfigRepository_DifferentWeekends(t *testing.T) {
	// Test with different weekend configurations (e.g., Western: Sat-Sun)
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	// Western weekend (Saturday=6, Sunday=0)
	tdb.SeedWeekendConfig(0, 6)

	days, err := repo.GetWeekendDays(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("GetWeekendDays() error = %v", err)
	}

	// Should be ordered by day_of_week
	if days[0] != 0 {
		t.Errorf("GetWeekendDays()[0] = %v, want 0 (Sunday)", days[0])
	}
	if days[1] != 6 {
		t.Errorf("GetWeekendDays()[1] = %v, want 6 (Saturday)", days[1])
	}
}

func TestWeekendConfigRepository_UIDGeneration(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	// Seed weekend config
	tdb.SeedWeekendConfig(5, 6)

	configs, err := repo.List(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Each config should have a unique UID
	uids := make(map[string]bool)
	for _, cfg := range configs {
		if cfg.UID == "" {
			t.Error("UID should not be empty")
		}
		if uids[cfg.UID] {
			t.Errorf("Duplicate UID found: %s", cfg.UID)
		}
		uids[cfg.UID] = true
	}
}

func TestWeekendConfigRepository_SingleDayWeekend(t *testing.T) {
	// Some regions have only one weekend day
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewWeekendConfigRepository()
	ctx := context.Background()

	// Single day weekend (Friday only)
	tdb.SeedWeekendConfig(5)

	days, err := repo.GetWeekendDays(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("GetWeekendDays() error = %v", err)
	}
	if len(days) != 1 {
		t.Errorf("GetWeekendDays() returned %d items, want 1", len(days))
	}
	if days[0] != 5 {
		t.Errorf("GetWeekendDays()[0] = %v, want 5", days[0])
	}
}
