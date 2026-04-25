package db

import (
	"context"
	"testing"
)

func TestLeaveTypeRepository_SubLeaveTypes(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewLeaveTypeRepository()
	ctx := context.Background()

	leaveType := tdb.SeedLeaveType("SPECIAL", "Special Leave", "إجازة خاصة", 10)
	first := tdb.SeedSubLeaveType(leaveType.UID, "Type A", "النوع أ")
	second := tdb.SeedSubLeaveType(leaveType.UID, "Type B", "النوع ب")

	found, err := repo.GetSubLeaveTypeByUID(ctx, tdb.SQLiteDB, first.UID)
	if err != nil {
		t.Fatalf("GetSubLeaveTypeByUID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetSubLeaveTypeByUID() returned nil")
	}
	if found.LeaveTypeUID != leaveType.UID {
		t.Fatalf("GetSubLeaveTypeByUID() leaveTypeUID = %s, want %s", found.LeaveTypeUID, leaveType.UID)
	}

	items, err := repo.ListSubLeaveTypesByLeaveTypeUID(ctx, tdb.SQLiteDB, leaveType.UID)
	if err != nil {
		t.Fatalf("ListSubLeaveTypesByLeaveTypeUID() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListSubLeaveTypesByLeaveTypeUID() len = %d, want 2", len(items))
	}
	if items[0].UID != first.UID {
		t.Fatalf("ListSubLeaveTypesByLeaveTypeUID()[0] = %s, want %s", items[0].UID, first.UID)
	}
	if items[1].UID != second.UID {
		t.Fatalf("ListSubLeaveTypesByLeaveTypeUID()[1] = %s, want %s", items[1].UID, second.UID)
	}
}
