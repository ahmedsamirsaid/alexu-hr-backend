package usecases_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type mockAutoRejectDB struct {
	tx *mockAutoRejectTx
}

func (m *mockAutoRejectDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	return m.tx, nil
}

func (m *mockAutoRejectDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockAutoRejectDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockAutoRejectDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

type mockAutoRejectTx struct {
	committed  bool
	rolledBack bool
}

func (m *mockAutoRejectTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockAutoRejectTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockAutoRejectTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *mockAutoRejectTx) Commit() error {
	m.committed = true
	return nil
}

func (m *mockAutoRejectTx) Rollback() error {
	m.rolledBack = true
	return nil
}

type mockLeaveRequestRepoForAutoReject struct {
	expiredRequests []*domain.LeaveRequest
	updatedRequests []*domain.LeaveRequest
}

func (m *mockLeaveRequestRepoForAutoReject) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForAutoReject) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForAutoReject) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForAutoReject) Create(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	return nil
}

func (m *mockLeaveRequestRepoForAutoReject) Update(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	m.updatedRequests = append(m.updatedRequests, request)
	return nil
}

func (m *mockLeaveRequestRepoForAutoReject) ListByEmployee(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForAutoReject) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeUID string, limit, offset int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForAutoReject) CountByEmployee(ctx context.Context, q ports.Querier, employeeUID string) (int, error) {
	return 0, nil
}

func (m *mockLeaveRequestRepoForAutoReject) HasOverlapping(ctx context.Context, q ports.Querier, employeeUID string, startDate, endDate time.Time, excludeUID *string) (bool, error) {
	return false, nil
}

func (m *mockLeaveRequestRepoForAutoReject) List(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter, limit, offset int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForAutoReject) Count(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter) (int, error) {
	return 0, nil
}

func (m *mockLeaveRequestRepoForAutoReject) FindExpiredPending(ctx context.Context, q ports.Querier, graceDays int) ([]*domain.LeaveRequest, error) {
	return m.expiredRequests, nil
}

type mockApprovalRequestRepoForAutoReject struct {
	requests map[string]*domain.ApprovalRequest
	updated  []*domain.ApprovalRequest
}

func (m *mockApprovalRequestRepoForAutoReject) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForAutoReject) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalRequest, error) {
	if r, ok := m.requests[uid]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockApprovalRequestRepoForAutoReject) Create(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	return nil
}

func (m *mockApprovalRequestRepoForAutoReject) Update(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	m.updated = append(m.updated, request)
	return nil
}

func (m *mockApprovalRequestRepoForAutoReject) ListByRequester(ctx context.Context, q ports.Querier, requesterUID string) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForAutoReject) ListPending(ctx context.Context, q ports.Querier) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForAutoReject) ListPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForAutoReject) CountPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (int, error) {
	return 0, nil
}

type mockApprovalActionRepoForAutoReject struct {
	actions []*domain.ApprovalAction
}

func (m *mockApprovalActionRepoForAutoReject) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalAction, error) {
	return nil, nil
}

func (m *mockApprovalActionRepoForAutoReject) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalAction, error) {
	return nil, nil
}

func (m *mockApprovalActionRepoForAutoReject) Create(ctx context.Context, q ports.Querier, action *domain.ApprovalAction) error {
	m.actions = append(m.actions, action)
	return nil
}

func (m *mockApprovalActionRepoForAutoReject) ListByRequest(ctx context.Context, q ports.Querier, approvalRequestUID string) ([]*domain.ApprovalAction, error) {
	return nil, nil
}

func TestAutoRejectExpiredRequestsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name             string
		graceDays        int
		expiredRequests  []*domain.LeaveRequest
		approvalRequests map[string]*domain.ApprovalRequest
		expectedRejected int
		verifyActions    func(t *testing.T, actions []*domain.ApprovalAction)
	}{
		{
			name:      "rejects expired pending requests",
			graceDays: 1,
			expiredRequests: []*domain.LeaveRequest{
				{
					UID:                "lreq_001",
					EmployeeUID:        "emp_001",
					StartDate:          time.Now().AddDate(0, 0, -5),
					ApprovalRequestUID: "areq_001",
				},
			},
			approvalRequests: map[string]*domain.ApprovalRequest{
				"areq_001": {
					UID:    "areq_001",
					Status: domain.ApprovalRequestStatusPending,
				},
			},
			expectedRejected: 1,
			verifyActions: func(t *testing.T, actions []*domain.ApprovalAction) {
				if len(actions) != 1 {
					t.Errorf("expected 1 action, got %d", len(actions))
					return
				}
				if actions[0].ActorUID != usecases.SystemActorUID {
					t.Errorf("expected actor_uid 'system', got %s", actions[0].ActorUID)
				}
				if actions[0].Action != domain.ApprovalActionTypeReject {
					t.Errorf("expected action 'reject', got %s", actions[0].Action)
				}
				if actions[0].Comments == nil || *actions[0].Comments != "Auto-rejected: leave start date has passed" {
					t.Errorf("unexpected comments: %v", actions[0].Comments)
				}
			},
		},
		{
			name:             "does nothing when no expired requests",
			graceDays:        1,
			expiredRequests:  []*domain.LeaveRequest{},
			approvalRequests: map[string]*domain.ApprovalRequest{},
			expectedRejected: 0,
		},
		{
			name:      "rejects multiple expired requests",
			graceDays: 1,
			expiredRequests: []*domain.LeaveRequest{
				{
					UID:                "lreq_001",
					EmployeeUID:        "emp_001",
					StartDate:          time.Now().AddDate(0, 0, -10),
					ApprovalRequestUID: "areq_001",
				},
				{
					UID:                "lreq_002",
					EmployeeUID:        "emp_001",
					StartDate:          time.Now().AddDate(0, 0, -5),
					ApprovalRequestUID: "areq_002",
				},
			},
			approvalRequests: map[string]*domain.ApprovalRequest{
				"areq_001": {
					UID:    "areq_001",
					Status: domain.ApprovalRequestStatusPending,
				},
				"areq_002": {
					UID:    "areq_002",
					Status: domain.ApprovalRequestStatusPending,
				},
			},
			expectedRejected: 2,
		},
		{
			name:      "skips already approved requests (race condition)",
			graceDays: 1,
			expiredRequests: []*domain.LeaveRequest{
				{
					UID:                "lreq_001",
					EmployeeUID:        "emp_001",
					StartDate:          time.Now().AddDate(0, 0, -5),
					ApprovalRequestUID: "areq_001",
				},
			},
			approvalRequests: map[string]*domain.ApprovalRequest{
				"areq_001": {
					UID:    "areq_001",
					Status: domain.ApprovalRequestStatusApproved, // Already approved
				},
			},
			expectedRejected: 0, // Should not count as rejected since it was already approved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &mockAutoRejectTx{}
			db := &mockAutoRejectDB{tx: tx}
			leaveRequestRepo := &mockLeaveRequestRepoForAutoReject{
				expiredRequests: tt.expiredRequests,
			}
			approvalRequestRepo := &mockApprovalRequestRepoForAutoReject{
				requests: tt.approvalRequests,
			}
			approvalActionRepo := &mockApprovalActionRepoForAutoReject{}

			uc := usecases.NewAutoRejectExpiredRequestsUseCase(
				db,
				leaveRequestRepo,
				approvalRequestRepo,
				approvalActionRepo,
				tt.graceDays,
			)

			output, err := uc.Execute(context.Background())
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if output.RejectedCount != tt.expectedRejected {
				t.Errorf("expected %d rejections, got %d", tt.expectedRejected, output.RejectedCount)
			}

			if tt.verifyActions != nil {
				tt.verifyActions(t, approvalActionRepo.actions)
			}
		})
	}
}
