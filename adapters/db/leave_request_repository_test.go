package db

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestLeaveRequestRepository_FindExpiredPending(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	ctx := context.Background()
	repo := NewLeaveRequestRepository()

	// Seed required data
	employee := tdb.SeedEmployee("Test Employee")
	leaveType := tdb.SeedLeaveType("CAS", "Casual Leave", "إجازة عارضة", 21)

	// Seed an approval flow
	approvalFlow := seedApprovalFlow(t, tdb)

	tests := []struct {
		name          string
		graceDays     int
		setupRequests func() // creates leave requests with different statuses/dates
		expectedCount int
		expectedUIDs  []string
	}{
		{
			name:      "finds expired pending requests beyond grace period",
			graceDays: 1,
			setupRequests: func() {
				// Request with start date 5 days ago (should be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now().AddDate(0, 0, -5), time.Now().AddDate(0, 0, -3), "pending", "expired_1")

				// Request with start date today (should NOT be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now(), time.Now().AddDate(0, 0, 2), "pending", "current")
			},
			expectedCount: 1,
		},
		{
			name:      "skips approved requests",
			graceDays: 1,
			setupRequests: func() {
				// Approved request with past start date (should NOT be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now().AddDate(0, 0, -5), time.Now().AddDate(0, 0, -3), "approved", "approved_past")
			},
			expectedCount: 0,
		},
		{
			name:      "skips rejected requests",
			graceDays: 1,
			setupRequests: func() {
				// Rejected request with past start date (should NOT be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now().AddDate(0, 0, -5), time.Now().AddDate(0, 0, -3), "rejected", "rejected_past")
			},
			expectedCount: 0,
		},
		{
			name:      "respects grace period",
			graceDays: 3,
			setupRequests: func() {
				// Request 2 days old (within 3-day grace, should NOT be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, 0), "pending", "within_grace")

				// Request 5 days old (beyond 3-day grace, should be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now().AddDate(0, 0, -5), time.Now().AddDate(0, 0, -3), "pending", "beyond_grace")
			},
			expectedCount: 1,
		},
		{
			name:      "returns empty for no expired requests",
			graceDays: 1,
			setupRequests: func() {
				// Future request (should NOT be found)
				createLeaveRequest(t, tdb, employee.UID, leaveType.UID, approvalFlow.UID,
					time.Now().AddDate(0, 0, 5), time.Now().AddDate(0, 0, 7), "pending", "future")
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous test data
			clearLeaveRequests(t, tdb)
			clearApprovalRequests(t, tdb)

			tt.setupRequests()

			results, err := repo.FindExpiredPending(ctx, tdb.SQLiteDB, tt.graceDays)
			if err != nil {
				t.Fatalf("FindExpiredPending failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
				for _, r := range results {
					t.Logf("  found: %s (start: %s)", r.UID, r.StartDate.Format("2006-01-02"))
				}
			}
		})
	}
}

func TestLeaveRequestRepository_CreateAndGetByUID_WithSubLeaveType(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	ctx := context.Background()
	repo := NewLeaveRequestRepository()

	employee := tdb.SeedEmployee("Test Employee")
	leaveType := tdb.SeedLeaveType("SPECIAL", "Special Leave", "إجازة خاصة", 10)
	subLeaveType := tdb.SeedSubLeaveType(leaveType.UID, "Type A", "النوع أ")
	approvalFlow := seedApprovalFlow(t, tdb)

	approvalRequestUID := domain.GenerateUID("apr")
	now := time.Now()
	_, err := tdb.db.ExecContext(ctx, `
		INSERT INTO approval_requests (uid, requester_uid, approval_flow_uid, status, current_step, max_step, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		approvalRequestUID, employee.UID, approvalFlow.UID, "pending", 1, 1, now, now)
	if err != nil {
		t.Fatalf("failed to seed approval request: %v", err)
	}

	request := domain.NewLeaveRequest(
		employee.UID,
		leaveType.UID,
		&subLeaveType.UID,
		time.Now(),
		time.Now().AddDate(0, 0, 1),
		2,
		nil,
		nil,
		nil,
		nil,
		nil,
		approvalRequestUID,
	)

	if err := repo.Create(ctx, tdb.SQLiteDB, request); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, request.UID)
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetByUID() returned nil")
	}
	if found.SubLeaveTypeUID == nil {
		t.Fatal("GetByUID() subLeaveTypeUID is nil")
	}
	if *found.SubLeaveTypeUID != subLeaveType.UID {
		t.Fatalf("GetByUID() subLeaveTypeUID = %s, want %s", *found.SubLeaveTypeUID, subLeaveType.UID)
	}
}

func seedApprovalFlow(t *testing.T, tdb *TestDB) *domain.ApprovalFlow {
	t.Helper()

	flow := &domain.ApprovalFlow{
		UID:       domain.GenerateUID("flow"),
		Code:      "TEST",
		NameEN:    "Test Flow",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()
	_, err := tdb.db.ExecContext(ctx, `
		INSERT INTO approval_flows (uid, code, name_en, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		flow.UID, flow.Code, flow.NameEN, flow.IsActive, flow.CreatedAt, flow.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to seed approval flow: %v", err)
	}

	return flow
}

func createLeaveRequest(t *testing.T, tdb *TestDB, employeeUID, leaveTypeUID, approvalFlowUID string, startDate, endDate time.Time, status, notes string) *domain.LeaveRequest {
	t.Helper()

	ctx := context.Background()
	now := time.Now()

	// Create approval request first
	approvalRequestUID := domain.GenerateUID("areq")
	_, err := tdb.db.ExecContext(ctx, `
		INSERT INTO approval_requests (uid, requester_uid, approval_flow_uid, status, current_step, max_step, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		approvalRequestUID, employeeUID, approvalFlowUID, status, 1, 1, now, now)
	if err != nil {
		t.Fatalf("failed to create approval request: %v", err)
	}

	// Create leave request
	leaveRequestUID := domain.GenerateUID("lreq")
	var decidedAt *time.Time
	if status != "pending" {
		d := now
		decidedAt = &d
	}

	_, err = tdb.db.ExecContext(ctx, `
		INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		leaveRequestUID, employeeUID, leaveTypeUID, startDate, endDate, 3, notes, now, decidedAt, approvalRequestUID, now, now)
	if err != nil {
		t.Fatalf("failed to create leave request: %v", err)
	}

	return &domain.LeaveRequest{
		UID:                leaveRequestUID,
		EmployeeUID:        employeeUID,
		LeaveTypeUID:       leaveTypeUID,
		StartDate:          startDate,
		EndDate:            endDate,
		Days:               3,
		Notes:              &notes,
		SubmittedAt:        now,
		DecidedAt:          decidedAt,
		ApprovalRequestUID: approvalRequestUID,
	}
}

func clearLeaveRequests(t *testing.T, tdb *TestDB) {
	t.Helper()
	_, err := tdb.db.Exec("DELETE FROM leave_requests")
	if err != nil {
		t.Fatalf("failed to clear leave requests: %v", err)
	}
}

func clearApprovalRequests(t *testing.T, tdb *TestDB) {
	t.Helper()
	_, err := tdb.db.Exec("DELETE FROM approval_requests")
	if err != nil {
		t.Fatalf("failed to clear approval requests: %v", err)
	}
}
