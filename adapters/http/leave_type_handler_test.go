package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type mockLeaveTypeRepoForHandler struct {
	leaveTypeByUID map[string]*domain.LeaveType
	subLeaveTypes  map[string][]*domain.SubLeaveType
}

type mockApprovalFlowRepoForHandler struct {
	flowByUID map[string]*domain.ApprovalFlow
}

type mockApprovalFlowStepRepoForHandler struct {
	stepsByFlowUID map[string][]*domain.ApprovalFlowStep
}

type mockRoleRepoForLeaveTypeHandler struct {
	roles []*domain.Role
}

func (m *mockLeaveTypeRepoForHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveType, error) {
	if m.leaveTypeByUID == nil {
		return nil, nil
	}
	return m.leaveTypeByUID[uid], nil
}

func (m *mockLeaveTypeRepoForHandler) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForHandler) GetSubLeaveTypeByUID(ctx context.Context, q ports.Querier, uid string) (*domain.SubLeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForHandler) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForHandler) ListSubLeaveTypesByLeaveTypeUID(ctx context.Context, q ports.Querier, leaveTypeUID string) ([]*domain.SubLeaveType, error) {
	if m.subLeaveTypes == nil {
		return nil, nil
	}
	return m.subLeaveTypes[leaveTypeUID], nil
}

func (m *mockLeaveTypeRepoForHandler) SetActive(ctx context.Context, q ports.Querier, uid string, isActive bool) error {
	return nil
}

func (m *mockLeaveTypeRepoForHandler) Update(ctx context.Context, q ports.Querier, leaveType *domain.LeaveType) error {
	if m.leaveTypeByUID == nil {
		m.leaveTypeByUID = map[string]*domain.LeaveType{}
	}
	cloned := *leaveType
	m.leaveTypeByUID[leaveType.UID] = &cloned
	return nil
}

func (m *mockApprovalFlowRepoForHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalFlow, error) {
	return nil, nil
}

func (m *mockApprovalFlowRepoForHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalFlow, error) {
	if m.flowByUID == nil {
		return nil, nil
	}
	return m.flowByUID[uid], nil
}

func (m *mockApprovalFlowRepoForHandler) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.ApprovalFlow, error) {
	return nil, nil
}

func (m *mockApprovalFlowRepoForHandler) Create(ctx context.Context, q ports.Querier, flow *domain.ApprovalFlow) error {
	return nil
}

func (m *mockApprovalFlowRepoForHandler) Update(ctx context.Context, q ports.Querier, flow *domain.ApprovalFlow) error {
	return nil
}

func (m *mockApprovalFlowRepoForHandler) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.ApprovalFlow, error) {
	return nil, nil
}

func (m *mockApprovalFlowStepRepoForHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalFlowStep, error) {
	return nil, nil
}

func (m *mockApprovalFlowStepRepoForHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalFlowStep, error) {
	return nil, nil
}

func (m *mockApprovalFlowStepRepoForHandler) Create(ctx context.Context, q ports.Querier, step *domain.ApprovalFlowStep) error {
	return nil
}

func (m *mockApprovalFlowStepRepoForHandler) Update(ctx context.Context, q ports.Querier, step *domain.ApprovalFlowStep) error {
	return nil
}

func (m *mockApprovalFlowStepRepoForHandler) Delete(ctx context.Context, q ports.Querier, uid string) error {
	return nil
}

func (m *mockApprovalFlowStepRepoForHandler) ListByFlow(ctx context.Context, q ports.Querier, approvalFlowUID string) ([]*domain.ApprovalFlowStep, error) {
	if m.stepsByFlowUID == nil {
		return nil, nil
	}
	return m.stepsByFlowUID[approvalFlowUID], nil
}

func (m *mockApprovalFlowStepRepoForHandler) CountByFlow(ctx context.Context, q ports.Querier, approvalFlowUID string) (int, error) {
	return 0, nil
}

func (m *mockApprovalFlowStepRepoForHandler) GetByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (*domain.ApprovalFlowStep, error) {
	return nil, nil
}

func (m *mockApprovalFlowStepRepoForHandler) HasPendingRequestsAtStep(ctx context.Context, q ports.Querier, stepUID string) (bool, error) {
	return false, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Role, error) {
	for _, role := range m.roles {
		if role.UID == uid {
			return role, nil
		}
	}
	return nil, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetByName(ctx context.Context, q ports.Querier, name string) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) Create(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) Update(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) List(ctx context.Context, q ports.Querier) ([]*domain.Role, error) {
	return m.roles, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) Delete(ctx context.Context, q ports.Querier, id int64) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetRolesForUser(ctx context.Context, q ports.Querier, userID int64) ([]*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) AssignRoleToUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) RemoveRoleFromUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) AssignPermissionToRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) RemovePermissionFromRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) SetRolePermissions(ctx context.Context, q ports.Querier, roleID int64, permissionIDs []int64) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) AssignRoleToUserWithDepartment(ctx context.Context, q ports.Querier, userID, roleID int64, departmentUID *string) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetUsersByRoleAndDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID *string) ([]*domain.User, error) {
	return nil, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) IsUserAuthorizedApprover(ctx context.Context, q ports.Querier, userID int64, roleUID string, departmentUID string) (bool, error) {
	return false, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) RemoveRoleFromUserForDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID string) error {
	return nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetDepartmentManager(ctx context.Context, q ports.Querier, departmentUID string) (*domain.User, error) {
	return nil, nil
}

func (m *mockRoleRepoForLeaveTypeHandler) GetManagedDepartmentUIDs(ctx context.Context, q ports.Querier, userID int64) ([]string, error) {
	return nil, nil
}

func TestLeaveTypeHandler_GetLeaveType_Success(t *testing.T) {
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	maxConsecutive := 365
	recordingDeadlineDays := 30
	advanceNoticeDays := 7
	approvalFlowUID := "apf_special"
	nameAR := "موافقة خاصة"
	description := "Requires special approval chain"

	repo := &mockLeaveTypeRepoForHandler{
		leaveTypeByUID: map[string]*domain.LeaveType{
			"lt_special": {
				UID:                   "lt_special",
				Code:                  "SPECIAL",
				NameEN:                "Special Leave",
				NameAR:                "إجازة خاصة",
				DefaultBalance:        365,
				MaxConsecutive:        &maxConsecutive,
				RecordingDeadlineDays: &recordingDeadlineDays,
				AdvanceNoticeDays:     &advanceNoticeDays,
				IsActive:              true,
				ApprovalFlowUID:       &approvalFlowUID,
				CreatedAt:             now,
				UpdatedAt:             now,
			},
		},
		subLeaveTypes: map[string][]*domain.SubLeaveType{
			"lt_special": {
				{
					UID:          "slt_1",
					LeaveTypeUID: "lt_special",
					NameEN:       "Type A",
					NameAR:       "النوع أ",
				},
			},
		},
	}
	approvalFlowRepo := &mockApprovalFlowRepoForHandler{
		flowByUID: map[string]*domain.ApprovalFlow{
			approvalFlowUID: {
				UID:         approvalFlowUID,
				Code:        "SPECIAL_CHAIN",
				NameEN:      "Special Chain",
				NameAR:      &nameAR,
				Description: &description,
				IsActive:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}
	stepRepo := &mockApprovalFlowStepRepoForHandler{
		stepsByFlowUID: map[string][]*domain.ApprovalFlowStep{
			approvalFlowUID: {
				{
					UID:             "afs_1",
					ApprovalFlowUID: approvalFlowUID,
					StepOrder:       1,
					RoleUID:         "role_1",
					CreatedAt:       now,
					UpdatedAt:       now,
				},
			},
		},
	}
	roleRepo := &mockRoleRepoForLeaveTypeHandler{
		roles: []*domain.Role{{UID: "role_1", Name: "Manager"}},
	}

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, stepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	setApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, audit.Noop())
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	handler := NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setApprovalFlowUC, toggleLeaveTypeUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leave-types/lt_special", nil)
	req.SetPathValue("uid", "lt_special")
	rec := httptest.NewRecorder()

	handler.GetLeaveType(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp GetLeaveTypeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.LeaveType.UID != "lt_special" {
		t.Fatalf("leaveType.uid = %q, want %q", resp.LeaveType.UID, "lt_special")
	}
	if len(resp.SubLeaveTypes) != 1 {
		t.Fatalf("len(subLeaveTypes) = %d, want 1", len(resp.SubLeaveTypes))
	}
	if resp.SubLeaveTypes[0].UID != "slt_1" {
		t.Fatalf("subLeaveTypes[0].uid = %q, want %q", resp.SubLeaveTypes[0].UID, "slt_1")
	}
	if resp.ApprovalFlow == nil {
		t.Fatal("approvalFlow is nil, want value")
	}
	if resp.ApprovalFlow.UID != approvalFlowUID {
		t.Fatalf("approvalFlow.uid = %q, want %q", resp.ApprovalFlow.UID, approvalFlowUID)
	}
	if resp.ApprovalFlow.Code != "SPECIAL_CHAIN" {
		t.Fatalf("approvalFlow.code = %q, want %q", resp.ApprovalFlow.Code, "SPECIAL_CHAIN")
	}
	if len(resp.ApprovalFlow.Steps) != 1 {
		t.Fatalf("len(approvalFlow.steps) = %d, want 1", len(resp.ApprovalFlow.Steps))
	}
	if resp.ApprovalFlow.Steps[0].RoleName != "Manager" {
		t.Fatalf("approvalFlow.steps[0].roleName = %q, want %q", resp.ApprovalFlow.Steps[0].RoleName, "Manager")
	}
}

func TestLeaveTypeHandler_GetLeaveType_NotFound(t *testing.T) {
	repo := &mockLeaveTypeRepoForHandler{}
	approvalFlowRepo := &mockApprovalFlowRepoForHandler{}
	stepRepo := &mockApprovalFlowStepRepoForHandler{}
	roleRepo := &mockRoleRepoForLeaveTypeHandler{}

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, stepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	setApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, audit.Noop())
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	handler := NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setApprovalFlowUC, toggleLeaveTypeUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leave-types/missing", nil)
	req.SetPathValue("uid", "missing")
	rec := httptest.NewRecorder()

	handler.GetLeaveType(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestLeaveTypeHandler_UpdateLeaveType_Success(t *testing.T) {
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	recordingDeadlineDays := 2
	advanceNoticeDays := 5
	repo := &mockLeaveTypeRepoForHandler{
		leaveTypeByUID: map[string]*domain.LeaveType{
			"lt_casual": {
				UID:                   "lt_casual",
				Code:                  "CASUAL",
				NameEN:                "Casual Leave",
				NameAR:                "عارضة",
				DefaultBalance:        7,
				RecordingDeadlineDays: &recordingDeadlineDays,
				AdvanceNoticeDays:     &advanceNoticeDays,
				IsActive:              true,
				CreatedAt:             now,
				UpdatedAt:             now,
			},
		},
	}
	approvalFlowRepo := &mockApprovalFlowRepoForHandler{}
	stepRepo := &mockApprovalFlowStepRepoForHandler{}
	roleRepo := &mockRoleRepoForLeaveTypeHandler{}

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, stepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	setApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, audit.Noop())
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	handler := NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setApprovalFlowUC, toggleLeaveTypeUC)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/leave-types/lt_casual", strings.NewReader(`{"defaultBalance":10,"recordingDeadlineDays":4,"advanceNoticeDays":null}`))
	req.SetPathValue("uid", "lt_casual")
	rec := httptest.NewRecorder()

	handler.UpdateLeaveType(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp LeaveTypeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.DefaultBalance != 10 {
		t.Fatalf("defaultBalance = %d, want 10", resp.DefaultBalance)
	}
	if resp.RecordingDeadlineDays == nil || *resp.RecordingDeadlineDays != 4 {
		t.Fatalf("recordingDeadlineDays = %v, want 4", resp.RecordingDeadlineDays)
	}
	if resp.AdvanceNoticeDays != nil {
		t.Fatalf("advanceNoticeDays = %v, want nil", resp.AdvanceNoticeDays)
	}
}

func TestLeaveTypeHandler_UpdateLeaveType_InvalidDefaultBalance(t *testing.T) {
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	repo := &mockLeaveTypeRepoForHandler{
		leaveTypeByUID: map[string]*domain.LeaveType{
			"lt_casual": {
				UID:            "lt_casual",
				Code:           "CASUAL",
				NameEN:         "Casual Leave",
				NameAR:         "عارضة",
				DefaultBalance: 7,
				IsActive:       true,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
	}
	approvalFlowRepo := &mockApprovalFlowRepoForHandler{}
	stepRepo := &mockApprovalFlowStepRepoForHandler{}
	roleRepo := &mockRoleRepoForLeaveTypeHandler{}

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, stepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	setApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, audit.Noop())
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	handler := NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setApprovalFlowUC, toggleLeaveTypeUC)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/leave-types/lt_casual", strings.NewReader(`{"defaultBalance":-1}`))
	req.SetPathValue("uid", "lt_casual")
	rec := httptest.NewRecorder()

	handler.UpdateLeaveType(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLeaveTypeHandler_SetLeaveTypeApprovalFlow_Success(t *testing.T) {
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	repo := &mockLeaveTypeRepoForHandler{
		leaveTypeByUID: map[string]*domain.LeaveType{
			"lt_casual": {
				UID:            "lt_casual",
				Code:           "CASUAL",
				NameEN:         "Casual Leave",
				NameAR:         "عارضة",
				DefaultBalance: 7,
				IsActive:       true,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
	}
	nameAR := "موافقة خاصة"
	description := "Requires special approval chain"
	approvalFlowRepo := &mockApprovalFlowRepoForHandler{
		flowByUID: map[string]*domain.ApprovalFlow{
			"apf_special": {
				UID:         "apf_special",
				Code:        "SPECIAL_CHAIN",
				NameEN:      "Special Chain",
				NameAR:      &nameAR,
				Description: &description,
				IsActive:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}
	stepRepo := &mockApprovalFlowStepRepoForHandler{
		stepsByFlowUID: map[string][]*domain.ApprovalFlowStep{
			"apf_special": {
				{
					UID:             "afs_1",
					ApprovalFlowUID: "apf_special",
					StepOrder:       1,
					RoleUID:         "role_1",
					CreatedAt:       now,
					UpdatedAt:       now,
				},
			},
		},
	}
	roleRepo := &mockRoleRepoForLeaveTypeHandler{
		roles: []*domain.Role{{UID: "role_1", Name: "Manager"}},
	}

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, stepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	setApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, audit.Noop())
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	handler := NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setApprovalFlowUC, toggleLeaveTypeUC)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/leave-types/lt_casual/approval-flow", strings.NewReader(`{"approvalFlowUid":"apf_special"}`))
	req.SetPathValue("uid", "lt_casual")
	rec := httptest.NewRecorder()

	handler.SetLeaveTypeApprovalFlow(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp GetLeaveTypeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ApprovalFlow == nil || resp.ApprovalFlow.UID != "apf_special" {
		t.Fatalf("approvalFlow = %+v, want apf_special", resp.ApprovalFlow)
	}
	if len(resp.ApprovalFlow.Steps) != 1 {
		t.Fatalf("len(approvalFlow.steps) = %d, want 1", len(resp.ApprovalFlow.Steps))
	}
	if repo.leaveTypeByUID["lt_casual"].ApprovalFlowUID == nil || *repo.leaveTypeByUID["lt_casual"].ApprovalFlowUID != "apf_special" {
		t.Fatalf("leaveType approvalFlowUid not updated")
	}
}

func TestLeaveTypeHandler_SetLeaveTypeApprovalFlow_Clear(t *testing.T) {
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	approvalFlowUID := "apf_special"
	repo := &mockLeaveTypeRepoForHandler{
		leaveTypeByUID: map[string]*domain.LeaveType{
			"lt_casual": {
				UID:             "lt_casual",
				Code:            "CASUAL",
				NameEN:          "Casual Leave",
				NameAR:          "عارضة",
				DefaultBalance:  7,
				ApprovalFlowUID: &approvalFlowUID,
				IsActive:        true,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
	}
	approvalFlowRepo := &mockApprovalFlowRepoForHandler{}
	stepRepo := &mockApprovalFlowStepRepoForHandler{}
	roleRepo := &mockRoleRepoForLeaveTypeHandler{}

	getLeaveTypeDetailsUC := usecases.NewGetLeaveTypeDetailsUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, stepRepo, roleRepo)
	listLeaveTypesUC := usecases.NewListLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	listSubLeaveTypesUC := usecases.NewListSubLeaveTypesUseCase(&mockAttendanceDB{}, repo)
	updateLeaveTypeUC := usecases.NewUpdateLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	setApprovalFlowUC := usecases.NewSetLeaveTypeApprovalFlowUseCase(&mockAttendanceDB{}, repo, approvalFlowRepo, audit.Noop())
	toggleLeaveTypeUC := usecases.NewToggleLeaveTypeUseCase(&mockAttendanceDB{}, repo, audit.Noop())
	handler := NewLeaveTypeHandler(getLeaveTypeDetailsUC, listLeaveTypesUC, listSubLeaveTypesUC, updateLeaveTypeUC, setApprovalFlowUC, toggleLeaveTypeUC)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/leave-types/lt_casual/approval-flow", strings.NewReader(`{"approvalFlowUid":null}`))
	req.SetPathValue("uid", "lt_casual")
	rec := httptest.NewRecorder()

	handler.SetLeaveTypeApprovalFlow(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp GetLeaveTypeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ApprovalFlow != nil {
		t.Fatalf("approvalFlow = %+v, want nil", resp.ApprovalFlow)
	}
	if repo.leaveTypeByUID["lt_casual"].ApprovalFlowUID != nil {
		t.Fatalf("leaveType approvalFlowUid = %v, want nil", repo.leaveTypeByUID["lt_casual"].ApprovalFlowUID)
	}
}
