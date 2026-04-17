package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

func TestMakeInclusiveEndDate_MidnightExpandsToEndOfDay(t *testing.T) {
	input := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	got := makeInclusiveEndDate(input)
	want := time.Date(2026, 4, 30, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

	if !got.Equal(want) {
		t.Fatalf("makeInclusiveEndDate(%s) = %s, want %s", input, got, want)
	}
}

func TestMakeInclusiveEndDate_NonMidnightStaysUnchanged(t *testing.T) {
	input := time.Date(2026, 4, 30, 13, 45, 30, 500, time.UTC)
	got := makeInclusiveEndDate(input)

	if !got.Equal(input) {
		t.Fatalf("makeInclusiveEndDate(%s) = %s, want unchanged", input, got)
	}
}

func TestAttendanceRecordRepository_ListDailyByDepartmentUID_IncludesCheckoutOnlyDay(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	ctx := context.Background()
	repo := NewAttendanceRecordRepository()

	departmentUID := "dep_test"
	_, err := tdb.ExecContext(ctx, `
		INSERT INTO departments (uid, code, name_en, name_ar, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		departmentUID, "D001", "Operations", "Operations", true, time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("seed department: %v", err)
	}

	employee := tdb.SeedEmployee("Alice")
	employee.DepartmentUID = &departmentUID
	if err := NewEmployeeRepository().Update(ctx, tdb.SQLiteDB, employee); err != nil {
		t.Fatalf("assign employee department: %v", err)
	}

	device := domain.NewAttendanceDevice("192.168.1.10", 4370, "Front Gate", "HQ", "SN-1")
	if err := NewAttendanceDeviceRepository().Create(ctx, tdb.SQLiteDB, device); err != nil {
		t.Fatalf("seed attendance device: %v", err)
	}

	record := domain.NewAttendanceRecord(
		employee.UID,
		device.UID,
		"device_user_checkout_only",
		time.Date(2026, 4, 10, 17, 30, 0, 0, time.UTC),
		domain.AttendancePunchTypeCheckOut,
		nil,
	)
	created, err := repo.Create(ctx, tdb.SQLiteDB, record)
	if err != nil {
		t.Fatalf("seed attendance record: %v", err)
	}
	if !created {
		t.Fatal("expected attendance record to be inserted")
	}

	filter := ports.DepartmentAttendanceLogsFilter{
		StartDate: ptrTime(time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)),
		EndDate:   ptrTime(time.Date(2026, 4, 10, 23, 59, 59, 0, time.UTC)),
	}
	params := ports.ListParams{
		Page:      1,
		PageSize:  20,
		SortBy:    "date",
		SortOrder: ports.SortOrderDesc,
	}

	groups, err := repo.ListDailyByDepartmentUID(ctx, tdb.SQLiteDB, departmentUID, filter, params)
	if err != nil {
		t.Fatalf("ListDailyByDepartmentUID() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("len(groups) = %d, want 1", len(groups))
	}
	if groups[0].CheckIn != nil {
		t.Fatalf("CheckIn = %v, want nil", groups[0].CheckIn)
	}
	if groups[0].CheckOut == nil || !groups[0].CheckOut.Equal(record.PunchedAt) {
		t.Fatalf("CheckOut = %v, want %v", groups[0].CheckOut, record.PunchedAt)
	}
	if groups[0].CheckOutDevice == nil || *groups[0].CheckOutDevice != device.Name {
		t.Fatalf("CheckOutDevice = %v, want %q", groups[0].CheckOutDevice, device.Name)
	}
	if groups[0].CheckOutDeviceUID == nil || *groups[0].CheckOutDeviceUID != device.UID {
		t.Fatalf("CheckOutDeviceUID = %v, want %q", groups[0].CheckOutDeviceUID, device.UID)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
