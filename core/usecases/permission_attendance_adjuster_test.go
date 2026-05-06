package usecases

import (
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

// Confirms the morning-permission late-threshold pipeline with the fixed UTC+2 zone.
// Shift 09:00, grace 15 min, punch 08:08 UTC (= 10:08 Cairo UTC+2).
// With morning permission: threshold = 10:00 UTC+2 + 15 grace = 10:15 UTC+2 = 08:15 UTC.
// Punch 08:08 UTC < 08:15 UTC → NOT late.
func TestMorningPermissionAdjustsLateThresholdWithGrace(t *testing.T) {
	// permissionDate as UTC midnight — same as time.Parse("2006-01-02", ...)
	day := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	shift := &domain.Shift{StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15}

	perm := &domain.PermissionRequest{
		UID:            "prq_morning_test",
		EmployeeUID:    "emp_test",
		Type:           domain.PermissionTypeMorning,
		PermissionDate: day,
		StartTime:      "09:00",
		EndTime:        "10:00",
	}
	adjuster := newPermissionAdjuster([]*domain.PermissionRequest{perm}, shift)

	// workStart built by buildTimeOnDate — uses attendanceLoc (UTC+2)
	workStart := time.Date(2026, 5, 4, 9, 0, 0, 0, attendanceLoc)
	lateThreshold := adjuster.LateThreshold(workStart).Add(time.Duration(shift.GraceMinutes) * time.Minute)

	// Expect 10:15 UTC+2 = 08:15 UTC
	want := time.Date(2026, 5, 4, 10, 15, 0, 0, attendanceLoc)
	if !lateThreshold.Equal(want) {
		t.Fatalf("lateThreshold = %v, want %v", lateThreshold, want)
	}

	// Punch at 08:08 UTC (= 10:08 Cairo), stored as Z in DB
	checkIn := time.Date(2026, 5, 4, 8, 8, 0, 0, time.UTC)
	if checkIn.After(lateThreshold) {
		t.Fatalf("expected 10:08 Cairo check-in to be within late threshold %v, but it was after", lateThreshold)
	}
}

// No permission: punch 08:08 UTC with shift 09:00 UTC+2 + grace 15 → threshold 07:15 UTC → 53 min late.
func TestNoPermissionLateMinutes(t *testing.T) {
	shift := &domain.Shift{StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15}
	adjuster := newPermissionAdjuster(nil, shift)

	workStart := time.Date(2026, 5, 4, 9, 0, 0, 0, attendanceLoc)
	lateThreshold := adjuster.LateThreshold(workStart).Add(time.Duration(shift.GraceMinutes) * time.Minute)

	// threshold = 09:15 UTC+2 = 07:15 UTC
	wantUTC := time.Date(2026, 5, 4, 7, 15, 0, 0, time.UTC)
	if !lateThreshold.Equal(wantUTC) {
		t.Fatalf("lateThreshold = %v, want %v", lateThreshold.UTC(), wantUTC)
	}

	// Punch at 08:08 UTC (10:08 Cairo) → 53 min late
	checkIn := time.Date(2026, 5, 4, 8, 8, 0, 0, time.UTC)
	delta := durationMinutesCeil(checkIn.Sub(lateThreshold))
	if delta != 53 {
		t.Fatalf("late minutes = %d, want 53", delta)
	}
}
