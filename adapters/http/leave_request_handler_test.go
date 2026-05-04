package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

type mockPermissionRepoForLeaveRequestHandler struct{}

func (m *mockPermissionRepoForLeaveRequestHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Permission, error) {
	return nil, nil
}

func (m *mockPermissionRepoForLeaveRequestHandler) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.Permission, error) {
	return nil, nil
}

func (m *mockPermissionRepoForLeaveRequestHandler) List(ctx context.Context, q ports.Querier) ([]*domain.Permission, error) {
	return nil, nil
}

func (m *mockPermissionRepoForLeaveRequestHandler) GetPermissionsForRole(ctx context.Context, q ports.Querier, roleID int64) ([]*domain.Permission, error) {
	return nil, nil
}

func (m *mockPermissionRepoForLeaveRequestHandler) GetPermissionsForRoles(ctx context.Context, q ports.Querier, roleIDs []int64) (map[int64][]*domain.Permission, error) {
	return map[int64][]*domain.Permission{}, nil
}

type mockUserRepoForLeaveRequestHandler struct {
	userByID map[int64]*domain.User
}

func (m *mockUserRepoForLeaveRequestHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	return m.userByID[id], nil
}

func (m *mockUserRepoForLeaveRequestHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepoForLeaveRequestHandler) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepoForLeaveRequestHandler) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	return nil
}

func (m *mockUserRepoForLeaveRequestHandler) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
	return nil
}

func (m *mockUserRepoForLeaveRequestHandler) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepoForLeaveRequestHandler) Count(ctx context.Context, q ports.Querier) (int, error) {
	return 0, nil
}

func (m *mockUserRepoForLeaveRequestHandler) ExistingPhones(ctx context.Context, q ports.Querier, phones []string) ([]string, error) {
	return nil, nil
}

func (m *mockUserRepoForLeaveRequestHandler) GetByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) (*domain.User, error) {
	return nil, nil
}

type mockEmployeeRepoForLeaveRequestHandler struct {
	employeeByUID map[string]*domain.Employee
}

func (m *mockEmployeeRepoForLeaveRequestHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	return m.employeeByUID[uid], nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.Employee, error) {
	items := make([]*domain.Employee, 0, len(uids))
	for _, uid := range uids {
		if employee := m.employeeByUID[uid]; employee != nil {
			items = append(items, employee)
		}
	}
	return items, nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForLeaveRequestHandler) Count(ctx context.Context, q ports.Querier) (int, error) {
	return 0, nil
}

type mockLeaveRequestRepoForLeaveRequestHandler struct {
	gotFilter ports.LeaveRequestListFilter
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) Create(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	return nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) Update(ctx context.Context, q ports.Querier, request *domain.LeaveRequest) error {
	return nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) ListByEmployee(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeUID string, limit, offset int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) CountByEmployee(ctx context.Context, q ports.Querier, employeeUID string) (int, error) {
	return 0, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) HasOverlapping(ctx context.Context, q ports.Querier, employeeUID string, startDate, endDate time.Time, excludeUID *string) (bool, error) {
	return false, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) List(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter, limit, offset int) ([]*domain.LeaveRequest, error) {
	m.gotFilter = filter
	return []*domain.LeaveRequest{}, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) Count(ctx context.Context, q ports.Querier, filter ports.LeaveRequestListFilter) (int, error) {
	return 0, nil
}

func (m *mockLeaveRequestRepoForLeaveRequestHandler) FindExpiredPending(ctx context.Context, q ports.Querier, graceDays int) ([]*domain.LeaveRequest, error) {
	return nil, nil
}

type mockApprovalRequestRepoForLeaveRequestHandler struct{}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) Create(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	return nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) Update(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	return nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) ListByRequester(ctx context.Context, q ports.Querier, requesterUID string) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) ListPending(ctx context.Context, q ports.Querier) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) ListPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}

func (m *mockApprovalRequestRepoForLeaveRequestHandler) CountPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (int, error) {
	return 0, nil
}

type mockLeaveTypeRepoForLeaveRequestHandler struct{}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) GetSubLeaveTypeByUID(ctx context.Context, q ports.Querier, uid string) (*domain.SubLeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) List(ctx context.Context, q ports.Querier, activeOnly bool) ([]*domain.LeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) ListSubLeaveTypesByLeaveTypeUID(ctx context.Context, q ports.Querier, leaveTypeUID string) ([]*domain.SubLeaveType, error) {
	return nil, nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) Update(ctx context.Context, q ports.Querier, leaveType *domain.LeaveType) error {
	return nil
}

func (m *mockLeaveTypeRepoForLeaveRequestHandler) SetActive(ctx context.Context, q ports.Querier, uid string, isActive bool) error {
	return nil
}

func TestListLeaveRequests_GlobalEmployeeFilterWithoutLinkedEmployee(t *testing.T) {
	leaveRequestRepo := &mockLeaveRequestRepoForLeaveRequestHandler{}
	handler := &LeaveRequestHandler{
		listUC: usecases.NewListLeaveRequestsUseCase(
			nil,
			leaveRequestRepo,
			&mockApprovalRequestRepoForLeaveRequestHandler{},
			&mockLeaveTypeRepoForLeaveRequestHandler{},
		),
		getCurrentUserUC: usecases.NewGetCurrentUserUseCase(
			nil,
			&mockUserRepoForLeaveRequestHandler{
				userByID: map[int64]*domain.User{
					99: {ID: 99, UID: "usr_admin"},
				},
			},
			&mockRoleRepo{},
			&mockPermissionRepoForLeaveRequestHandler{},
			&mockEmployeeRepoForLeaveRequestHandler{},
		),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leave-requests?employeeUid=emp56_ACA_01&page=1&pageSize=10", nil)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      99,
		UserUID:     "usr_admin",
		Permissions: []string{"*"},
		AccessScope: string(domain.RoleScopeGlobal),
	}))

	rr := httptest.NewRecorder()
	handler.ListLeaveRequests(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if leaveRequestRepo.gotFilter.EmployeeUID == nil || *leaveRequestRepo.gotFilter.EmployeeUID != "emp56_ACA_01" {
		t.Fatalf("employeeUID filter = %v, want emp56_ACA_01", leaveRequestRepo.gotFilter.EmployeeUID)
	}
}

func TestListLeaveRequests_DepartmentManagerDefaultsToOwnEmployeeRequests(t *testing.T) {
	ownEmployeeUID := "emp_manager_01"
	leaveRequestRepo := &mockLeaveRequestRepoForLeaveRequestHandler{}
	handler := &LeaveRequestHandler{
		listUC: usecases.NewListLeaveRequestsUseCase(
			nil,
			leaveRequestRepo,
			&mockApprovalRequestRepoForLeaveRequestHandler{},
			&mockLeaveTypeRepoForLeaveRequestHandler{},
		),
		getCurrentUserUC: usecases.NewGetCurrentUserUseCase(
			nil,
			&mockUserRepoForLeaveRequestHandler{
				userByID: map[int64]*domain.User{
					55: {ID: 55, UID: "usr_manager", EmployeeUID: &ownEmployeeUID},
				},
			},
			&mockRoleRepo{},
			&mockPermissionRepoForLeaveRequestHandler{},
			&mockEmployeeRepoForLeaveRequestHandler{},
		),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leave-requests?page=1&pageSize=10", nil)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:                55,
		UserUID:               "usr_manager",
		Permissions:           []string{"leave:request"},
		AccessScope:           string(domain.RoleScopeDepartment),
		ManagedDepartmentUIDs: []string{"dept_aca"},
	}))

	rr := httptest.NewRecorder()
	handler.ListLeaveRequests(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if leaveRequestRepo.gotFilter.EmployeeUID == nil || *leaveRequestRepo.gotFilter.EmployeeUID != ownEmployeeUID {
		t.Fatalf("employeeUID filter = %v, want %s", leaveRequestRepo.gotFilter.EmployeeUID, ownEmployeeUID)
	}
	if leaveRequestRepo.gotFilter.DepartmentUIDs != nil {
		t.Fatalf("department filter = %v, want nil", leaveRequestRepo.gotFilter.DepartmentUIDs)
	}
}

func TestListDepartmentLeaveRequests_UsesDepartmentScopeFilter(t *testing.T) {
	leaveRequestRepo := &mockLeaveRequestRepoForLeaveRequestHandler{}
	handler := &LeaveRequestHandler{
		listUC: usecases.NewListLeaveRequestsUseCase(
			nil,
			leaveRequestRepo,
			&mockApprovalRequestRepoForLeaveRequestHandler{},
			&mockLeaveTypeRepoForLeaveRequestHandler{},
		),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/departments/dept_aca/leave-requests?page=1&pageSize=10", nil)
	req.SetPathValue("departmentUid", "dept_aca")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:                77,
		UserUID:               "usr_manager",
		Permissions:           []string{"leave:approve"},
		AccessScope:           string(domain.RoleScopeDepartment),
		ManagedDepartmentUIDs: []string{"dept_aca"},
	}))

	rr := httptest.NewRecorder()
	handler.ListDepartmentLeaveRequests(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if leaveRequestRepo.gotFilter.EmployeeUID != nil {
		t.Fatalf("employeeUID filter = %v, want nil", leaveRequestRepo.gotFilter.EmployeeUID)
	}
	if len(leaveRequestRepo.gotFilter.DepartmentUIDs) != 1 || leaveRequestRepo.gotFilter.DepartmentUIDs[0] != "dept_aca" {
		t.Fatalf("department filter = %v, want [dept_aca]", leaveRequestRepo.gotFilter.DepartmentUIDs)
	}
}

func TestListDepartmentLeaveRequests_RejectsUnauthorizedDepartmentScope(t *testing.T) {
	leaveRequestRepo := &mockLeaveRequestRepoForLeaveRequestHandler{}
	handler := &LeaveRequestHandler{
		listUC: usecases.NewListLeaveRequestsUseCase(
			nil,
			leaveRequestRepo,
			&mockApprovalRequestRepoForLeaveRequestHandler{},
			&mockLeaveTypeRepoForLeaveRequestHandler{},
		),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/departments/dept_hr/leave-requests?page=1&pageSize=10", nil)
	req.SetPathValue("departmentUid", "dept_hr")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:                77,
		UserUID:               "usr_manager",
		Permissions:           []string{"leave:approve"},
		AccessScope:           string(domain.RoleScopeDepartment),
		ManagedDepartmentUIDs: []string{"dept_aca"},
	}))

	rr := httptest.NewRecorder()
	handler.ListDepartmentLeaveRequests(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}
