package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

// Reproduces the user-reported scenario end-to-end against the real SQLite schema:
// a morning permission for May 4 is created, its approval_request is set to "approved",
// and ListApprovedForEmployeeOnDate must return it. If this passes, the data path
// (employee_uid match, permission_date string match, status filter) is sound.
func TestListApprovedForEmployeeOnDate_ReturnsApprovedMorning(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	ctx := context.Background()
	emp := tdb.SeedEmployee("Test Morning Permission Employee")

	// Seed an approval flow + step so the FK on approval_requests is satisfied.
	tdb.MustExec(
		`INSERT INTO approval_flows (uid, code, name_en, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"apf_test", "test_flow", "test", time.Now(), time.Now(),
	)

	permissionDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.Local)

	approvalUID := domain.GenerateUID("apr")
	tdb.MustExec(
		`INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		approvalUID, "apf_test", emp.UID, 1, 1, string(domain.ApprovalRequestStatusApproved), time.Now(), time.Now(),
	)

	repo := NewPermissionRequestRepository()
	perm := &domain.PermissionRequest{
		UID:                domain.GenerateUID("prq"),
		EmployeeUID:        emp.UID,
		Type:               domain.PermissionTypeMorning,
		PermissionDate:     permissionDate,
		StartTime:          "09:00",
		EndTime:            "10:00",
		ApprovalRequestUID: approvalUID,
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, perm); err != nil {
		t.Fatalf("failed to create permission: %v", err)
	}

	// Query with the same Local-zone date the runtime uses.
	got, err := repo.ListApprovedForEmployeeOnDate(ctx, tdb.SQLiteDB, emp.UID, permissionDate)
	if err != nil {
		t.Fatalf("ListApprovedForEmployeeOnDate: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 approved permission, got %d", len(got))
	}

	// Also query with a UTC-midnight date (what the attendance group repo passes — it parses
	// "2026-05-04" via time.Parse, which yields UTC). This must also match.
	utcDate, _ := time.Parse("2006-01-02", "2026-05-04")
	got, err = repo.ListApprovedForEmployeeOnDate(ctx, tdb.SQLiteDB, emp.UID, utcDate)
	if err != nil {
		t.Fatalf("ListApprovedForEmployeeOnDate (UTC date): %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 approved permission with UTC date, got %d", len(got))
	}
	if got[0].Type != domain.PermissionTypeMorning {
		t.Fatalf("expected morning permission, got %s", got[0].Type)
	}
}

// Confirms that a pending permission is NOT returned (per the user's requirement that
// only approved permissions count).
func TestListApprovedForEmployeeOnDate_ExcludesPending(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	ctx := context.Background()
	emp := tdb.SeedEmployee("Test Pending Employee")

	tdb.MustExec(
		`INSERT INTO approval_flows (uid, code, name_en, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"apf_test", "test_flow", "test", time.Now(), time.Now(),
	)

	permissionDate := time.Date(2026, 5, 4, 0, 0, 0, 0, time.Local)

	approvalUID := domain.GenerateUID("apr")
	tdb.MustExec(
		`INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		approvalUID, "apf_test", emp.UID, 1, 1, string(domain.ApprovalRequestStatusPending), time.Now(), time.Now(),
	)

	repo := NewPermissionRequestRepository()
	perm := &domain.PermissionRequest{
		UID:                domain.GenerateUID("prq"),
		EmployeeUID:        emp.UID,
		Type:               domain.PermissionTypeMorning,
		PermissionDate:     permissionDate,
		StartTime:          "09:00",
		EndTime:            "10:00",
		ApprovalRequestUID: approvalUID,
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, perm); err != nil {
		t.Fatalf("failed to create permission: %v", err)
	}

	got, err := repo.ListApprovedForEmployeeOnDate(ctx, tdb.SQLiteDB, emp.UID, permissionDate)
	if err != nil {
		t.Fatalf("ListApprovedForEmployeeOnDate: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 approved permissions for pending row, got %d", len(got))
	}
}
