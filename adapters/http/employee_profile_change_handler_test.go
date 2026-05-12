package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

func TestEmployeeProfileChangeHandler_SubmitApproveAndReject(t *testing.T) {
	db := &profileChangeMockDB{}
	employeeRepo := &profileChangeMockEmployeeRepo{
		employeesByUID: map[string]*domain.Employee{},
	}
	userRepo := &profileChangeMockUserRepo{
		usersByID:          map[int64]*domain.User{},
		usersByEmployeeUID: map[string]*domain.User{},
	}
	roleRepo := newProfileChangeMockRoleRepo()
	permissionRepo := &profileChangeMockPermissionRepo{}
	approvalRequestRepo := &profileChangeMockApprovalRequestRepo{requestsByUID: map[string]*domain.ApprovalRequest{}}
	approvalActionRepo := &profileChangeMockApprovalActionRepo{}
	approvalFlowStepRepo := &profileChangeMockApprovalFlowStepRepo{
		steps: map[string]*domain.ApprovalFlowStep{
			"apf_employee_profile_change:1": {
				UID:             "step_1",
				ApprovalFlowUID: "apf_employee_profile_change",
				StepOrder:       1,
				RoleUID:         "role_information_center",
			},
		},
	}
	changeRequestRepo := &profileChangeMockChangeRequestRepo{
		requestsByUID:         map[string]*domain.EmployeeProfileChangeRequest{},
		requestsByApprovalUID: map[string]*domain.EmployeeProfileChangeRequest{},
		requestsByEmployeeUID: map[string][]*domain.EmployeeProfileChangeRequest{},
	}

	target := &domain.Employee{UID: "emp_target", Name: "Target", FinancialGrade: ptrString("FG1"), MaritalStatus: ptrString("Single")}
	hr := &domain.Employee{UID: "emp_hr", Name: "HR User"}
	ic := &domain.Employee{UID: "emp_ic", Name: "IC User"}
	employeeRepo.employeesByUID[target.UID] = cloneEmployeeForProfileTest(target)
	employeeRepo.employeesByUID[hr.UID] = cloneEmployeeForProfileTest(hr)
	employeeRepo.employeesByUID[ic.UID] = cloneEmployeeForProfileTest(ic)

	hrUser := &domain.User{ID: 1, UID: "usr_hr", Phone: "0101", EmployeeUID: &hr.UID, IsActive: true}
	icUser := &domain.User{ID: 2, UID: "usr_ic", Phone: "0102", EmployeeUID: &ic.UID, IsActive: true}
	userRepo.usersByID[1] = hrUser
	userRepo.usersByID[2] = icUser
	userRepo.usersByEmployeeUID[hr.UID] = hrUser
	userRepo.usersByEmployeeUID[ic.UID] = icUser
	roleRepo.assignRole(1, &domain.Role{UID: "role_hr_staff", Name: "HR Staff"})
	roleRepo.assignRole(2, &domain.Role{UID: "role_information_center", Name: "Information Center"})

	getCurrentUserUC := usecases.NewGetCurrentUserUseCase(db, userRepo, roleRepo, permissionRepo, employeeRepo)
	submitUC := usecases.NewSubmitEmployeeProfileChangeRequestUseCase(db, employeeRepo, roleRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, changeRequestRepo, audit.Noop())
	getUC := usecases.NewGetEmployeeProfileChangeRequestUseCase(db, roleRepo, changeRequestRepo, approvalRequestRepo, employeeRepo)
	listByEmployeeUC := usecases.NewListEmployeeProfileChangeRequestsUseCase(db, roleRepo, changeRequestRepo, getUC)
	listPendingUC := usecases.NewListPendingEmployeeProfileChangeRequestsUseCase(db, changeRequestRepo, roleRepo, getUC)
	approveUC := usecases.NewApproveEmployeeProfileChangeRequestUseCase(db, employeeRepo, changeRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, roleRepo, audit.Noop())
	rejectUC := usecases.NewRejectEmployeeProfileChangeRequestUseCase(db, changeRequestRepo, approvalRequestRepo, approvalActionRepo, approvalFlowStepRepo, roleRepo, audit.Noop())
	historyUC := usecases.NewGetApprovalHistoryUseCase(db, approvalRequestRepo, approvalActionRepo, employeeRepo)
	updateUC := usecases.NewUpdateEmployeeProfileUseCase(db, employeeRepo, userRepo, roleRepo, nil, audit.Noop())

	handler := NewEmployeeProfileChangeHandler(
		submitUC,
		listByEmployeeUC,
		getUC,
		listPendingUC,
		approveUC,
		rejectUC,
		historyUC,
		updateUC,
		getCurrentUserUC,
	)

	submitBody := bytes.NewBufferString(`{"financialGrade":"FG2","idCardValidUntil":"2030-01-01","maritalStatus":"Married","comments":"promotion"}`)
	submitReq := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_target/profile-change-requests", submitBody)
	submitReq.SetPathValue("uid", target.UID)
	submitReq = submitReq.WithContext(context.WithValue(submitReq.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      1,
		UserUID:     hrUser.UID,
		EmployeeUID: &hr.UID,
	}))
	submitRR := httptest.NewRecorder()
	handler.Submit(submitRR, submitReq)

	if submitRR.Code != http.StatusCreated {
		t.Fatalf("submit status = %d, body = %s", submitRR.Code, submitRR.Body.String())
	}
	var submitResp EmployeeProfileChangeRequestResponse
	if err := json.NewDecoder(submitRR.Body).Decode(&submitResp); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if submitResp.Status != "pending" {
		t.Fatalf("submit status = %q, want pending", submitResp.Status)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/profile-change-requests/pending", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      2,
		UserUID:     icUser.UID,
		EmployeeUID: &ic.UID,
	}))
	listRR := httptest.NewRecorder()
	handler.ListPending(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list pending status = %d, body = %s", listRR.Code, listRR.Body.String())
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/api/v1/profile-change-requests/"+submitResp.UID+"/approve", bytes.NewBufferString(`{"comments":"approved"}`))
	approveReq.SetPathValue("requestUid", submitResp.UID)
	approveReq = approveReq.WithContext(context.WithValue(approveReq.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      2,
		UserUID:     icUser.UID,
		EmployeeUID: &ic.UID,
	}))
	approveRR := httptest.NewRecorder()
	handler.Approve(approveRR, approveReq)
	if approveRR.Code != http.StatusOK {
		t.Fatalf("approve status = %d, body = %s", approveRR.Code, approveRR.Body.String())
	}

	updatedTarget := employeeRepo.employeesByUID[target.UID]
	if updatedTarget.FinancialGrade == nil || *updatedTarget.FinancialGrade != "FG2" {
		t.Fatalf("financial grade = %v, want FG2", updatedTarget.FinancialGrade)
	}
	if updatedTarget.MaritalStatus == nil || *updatedTarget.MaritalStatus != "Married" {
		t.Fatalf("marital status = %v, want Married", updatedTarget.MaritalStatus)
	}
	if updatedTarget.IDCardValidUntil == nil || updatedTarget.IDCardValidUntil.Format("2006-01-02") != "2030-01-01" {
		t.Fatalf("id card valid until = %v, want 2030-01-01", updatedTarget.IDCardValidUntil)
	}

	rejectTarget := &domain.Employee{UID: "emp_target_2", Name: "Reject Target", FinancialGrade: ptrString("FG1"), MaritalStatus: ptrString("Single")}
	employeeRepo.employeesByUID[rejectTarget.UID] = cloneEmployeeForProfileTest(rejectTarget)
	rejectSubmitReq := httptest.NewRequest(http.MethodPost, "/api/v1/employees/"+rejectTarget.UID+"/profile-change-requests", bytes.NewBufferString(`{"financialGrade":"FG9"}`))
	rejectSubmitReq.SetPathValue("uid", rejectTarget.UID)
	rejectSubmitReq = rejectSubmitReq.WithContext(context.WithValue(rejectSubmitReq.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      1,
		UserUID:     hrUser.UID,
		EmployeeUID: &hr.UID,
	}))
	rejectSubmitRR := httptest.NewRecorder()
	handler.Submit(rejectSubmitRR, rejectSubmitReq)
	if rejectSubmitRR.Code != http.StatusCreated {
		t.Fatalf("reject submit status = %d, body = %s", rejectSubmitRR.Code, rejectSubmitRR.Body.String())
	}
	var rejectSubmitResp EmployeeProfileChangeRequestResponse
	if err := json.NewDecoder(rejectSubmitRR.Body).Decode(&rejectSubmitResp); err != nil {
		t.Fatalf("decode reject submit response: %v", err)
	}

	rejectReq := httptest.NewRequest(http.MethodPost, "/api/v1/profile-change-requests/"+rejectSubmitResp.UID+"/reject", bytes.NewBufferString(`{"comments":"rejected"}`))
	rejectReq.SetPathValue("requestUid", rejectSubmitResp.UID)
	rejectReq = rejectReq.WithContext(context.WithValue(rejectReq.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      2,
		UserUID:     icUser.UID,
		EmployeeUID: &ic.UID,
	}))
	rejectRR := httptest.NewRecorder()
	handler.Reject(rejectRR, rejectReq)
	if rejectRR.Code != http.StatusOK {
		t.Fatalf("reject status = %d, body = %s", rejectRR.Code, rejectRR.Body.String())
	}
	if got := employeeRepo.employeesByUID[rejectTarget.UID].FinancialGrade; got == nil || *got != "FG1" {
		t.Fatalf("rejected request mutated employee financial grade = %v, want FG1", got)
	}
}

func TestEmployeeProfileChangeHandler_DirectUpdate(t *testing.T) {
	db := &profileChangeMockDB{}
	employeeRepo := &profileChangeMockEmployeeRepo{
		employeesByUID: map[string]*domain.Employee{
			"emp_target": {UID: "emp_target", Name: "Target", FinancialGrade: ptrString("FG1"), MaritalStatus: ptrString("Single")},
			"emp_ic":     {UID: "emp_ic", Name: "IC User"},
		},
	}
	userRepo := &profileChangeMockUserRepo{
		usersByID:          map[int64]*domain.User{},
		usersByEmployeeUID: map[string]*domain.User{},
	}
	roleRepo := newProfileChangeMockRoleRepo()
	permissionRepo := &profileChangeMockPermissionRepo{}
	icUser := &domain.User{ID: 2, UID: "usr_ic", Phone: "0102", EmployeeUID: ptrString("emp_ic"), IsActive: true}
	userRepo.usersByID[2] = icUser
	userRepo.usersByEmployeeUID["emp_ic"] = icUser
	roleRepo.assignRole(2, &domain.Role{UID: "role_information_center", Name: "Information Center"})

	handler := NewEmployeeProfileChangeHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewUpdateEmployeeProfileUseCase(db, employeeRepo, userRepo, roleRepo, nil, audit.Noop()),
		usecases.NewGetCurrentUserUseCase(db, userRepo, roleRepo, permissionRepo, employeeRepo),
	)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/employees/emp_target", bytes.NewBufferString(`{"financialGrade":"FG3","maritalStatus":"Divorced"}`))
	req.SetPathValue("uid", "emp_target")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      2,
		UserUID:     icUser.UID,
		EmployeeUID: ptrString("emp_ic"),
	}))
	rr := httptest.NewRecorder()
	handler.UpdateEmployee(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", rr.Code, rr.Body.String())
	}
	target := employeeRepo.employeesByUID["emp_target"]
	if target.FinancialGrade == nil || *target.FinancialGrade != "FG3" {
		t.Fatalf("financial grade = %v, want FG3", target.FinancialGrade)
	}
	if target.MaritalStatus == nil || *target.MaritalStatus != "Divorced" {
		t.Fatalf("marital status = %v, want Divorced", target.MaritalStatus)
	}
}

func TestEmployeeProfileChangeHandler_DirectUpdate_WithDocuments(t *testing.T) {
	db := &profileChangeMockDB{}
	employeeRepo := &profileChangeMockEmployeeRepo{
		employeesByUID: map[string]*domain.Employee{
			"emp_target": {UID: "emp_target", Name: "Target"},
			"emp_ic":     {UID: "emp_ic", Name: "IC User"},
		},
	}
	userRepo := &profileChangeMockUserRepo{
		usersByID:          map[int64]*domain.User{},
		usersByEmployeeUID: map[string]*domain.User{},
	}
	roleRepo := newProfileChangeMockRoleRepo()
	permissionRepo := &profileChangeMockPermissionRepo{}
	icUser := &domain.User{ID: 2, UID: "usr_ic", Phone: "0102", EmployeeUID: ptrString("emp_ic"), IsActive: true}
	userRepo.usersByID[2] = icUser
	userRepo.usersByEmployeeUID["emp_ic"] = icUser
	roleRepo.assignRole(2, &domain.Role{UID: "role_information_center", Name: "Information Center"})

	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)
	handler := NewEmployeeProfileChangeHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewUpdateEmployeeProfileUseCase(db, employeeRepo, userRepo, roleRepo, uploadUC, audit.Noop()),
		usecases.NewGetCurrentUserUseCase(db, userRepo, roleRepo, permissionRepo, employeeRepo),
	)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/employees/emp_target", bytes.NewBufferString(`{
		"documents": [
			{
				"documentType": "personal_photo",
				"fileName": "photo.png",
				"contentType": "image/png"
			},
			{
				"documentType": "decision_file",
				"fileName": "decision.pdf",
				"contentType": "application/pdf"
			}
		]
	}`))
	req.SetPathValue("uid", "emp_target")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      2,
		UserUID:     icUser.UID,
		EmployeeUID: ptrString("emp_ic"),
	}))
	rr := httptest.NewRecorder()
	handler.UpdateEmployee(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", rr.Code, rr.Body.String())
	}

	var response UpdateEmployeeProfileResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Documents) != 2 {
		t.Fatalf("documents len = %d, want 2", len(response.Documents))
	}

	target := employeeRepo.employeesByUID["emp_target"]
	if target.PersonalPhotoURL == nil || *target.PersonalPhotoURL == "" {
		t.Fatalf("personal photo url not updated: %+v", target.PersonalPhotoURL)
	}
	if target.DecisionFileURL == nil || *target.DecisionFileURL == "" {
		t.Fatalf("decision file url not updated: %+v", target.DecisionFileURL)
	}
}

func TestEmployeeProfileChangeHandler_Submit_AllowsHRStaff(t *testing.T) {
	db := &profileChangeMockDB{}
	employeeRepo := &profileChangeMockEmployeeRepo{
		employeesByUID: map[string]*domain.Employee{
			"emp_target": {UID: "emp_target", Name: "Target", FinancialGrade: ptrString("FG1")},
			"emp_hr":     {UID: "emp_hr", Name: "HR Staff"},
		},
	}
	userRepo := &profileChangeMockUserRepo{
		usersByID:          map[int64]*domain.User{},
		usersByEmployeeUID: map[string]*domain.User{},
	}
	roleRepo := newProfileChangeMockRoleRepo()
	roleRepo.assignRole(3, &domain.Role{UID: "role_hr_staff", Name: "HR Staff"})
	permissionRepo := &profileChangeMockPermissionRepo{}
	hrStaffUser := &domain.User{ID: 3, UID: "usr_hr_staff", Phone: "0103", EmployeeUID: ptrString("emp_hr"), IsActive: true}
	userRepo.usersByID[3] = hrStaffUser
	userRepo.usersByEmployeeUID["emp_hr"] = hrStaffUser

	approvalRequestRepo := &profileChangeMockApprovalRequestRepo{requestsByUID: map[string]*domain.ApprovalRequest{}}
	changeRequestRepo := &profileChangeMockChangeRequestRepo{
		requestsByUID:         map[string]*domain.EmployeeProfileChangeRequest{},
		requestsByApprovalUID: map[string]*domain.EmployeeProfileChangeRequest{},
		requestsByEmployeeUID: map[string][]*domain.EmployeeProfileChangeRequest{},
	}
	approvalFlowStepRepo := &profileChangeMockApprovalFlowStepRepo{
		steps: map[string]*domain.ApprovalFlowStep{
			"apf_employee_profile_change:1": {
				UID:             "step_1",
				ApprovalFlowUID: "apf_employee_profile_change",
				StepOrder:       1,
				RoleUID:         "role_information_center",
			},
		},
	}
	submitUC := usecases.NewSubmitEmployeeProfileChangeRequestUseCase(
		db,
		employeeRepo,
		roleRepo,
		approvalRequestRepo,
		&profileChangeMockApprovalActionRepo{},
		approvalFlowStepRepo,
		changeRequestRepo,
		audit.Noop(),
	)
	getUC := usecases.NewGetEmployeeProfileChangeRequestUseCase(db, roleRepo, changeRequestRepo, approvalRequestRepo, employeeRepo)

	handler := NewEmployeeProfileChangeHandler(
		submitUC,
		nil,
		getUC,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewGetCurrentUserUseCase(db, userRepo, roleRepo, permissionRepo, employeeRepo),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_target/profile-change-requests", bytes.NewBufferString(`{"financialGrade":"FG2"}`))
	req.SetPathValue("uid", "emp_target")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      3,
		UserUID:     hrStaffUser.UID,
		EmployeeUID: ptrString("emp_hr"),
	}))
	rr := httptest.NewRecorder()
	handler.Submit(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("submit status = %d, body = %s", rr.Code, rr.Body.String())
	}
}

func TestEmployeeProfileChangeHandler_Submit_RejectsDepartmentManager(t *testing.T) {
	db := &profileChangeMockDB{}
	employeeRepo := &profileChangeMockEmployeeRepo{
		employeesByUID: map[string]*domain.Employee{
			"emp_target": {UID: "emp_target", Name: "Target", FinancialGrade: ptrString("FG1")},
			"emp_mgr":    {UID: "emp_mgr", Name: "HR Manager"},
		},
	}
	userRepo := &profileChangeMockUserRepo{
		usersByID:          map[int64]*domain.User{},
		usersByEmployeeUID: map[string]*domain.User{},
	}
	roleRepo := newProfileChangeMockRoleRepo()
	roleRepo.assignRole(3, &domain.Role{UID: "role_department_manager", Name: "Department Manager"})
	permissionRepo := &profileChangeMockPermissionRepo{}
	managerUser := &domain.User{ID: 3, UID: "usr_mgr", Phone: "0103", EmployeeUID: ptrString("emp_mgr"), IsActive: true}
	userRepo.usersByID[3] = managerUser
	userRepo.usersByEmployeeUID["emp_mgr"] = managerUser

	approvalRequestRepo := &profileChangeMockApprovalRequestRepo{requestsByUID: map[string]*domain.ApprovalRequest{}}
	changeRequestRepo := &profileChangeMockChangeRequestRepo{
		requestsByUID:         map[string]*domain.EmployeeProfileChangeRequest{},
		requestsByApprovalUID: map[string]*domain.EmployeeProfileChangeRequest{},
		requestsByEmployeeUID: map[string][]*domain.EmployeeProfileChangeRequest{},
	}
	approvalFlowStepRepo := &profileChangeMockApprovalFlowStepRepo{
		steps: map[string]*domain.ApprovalFlowStep{
			"apf_employee_profile_change:1": {
				UID:             "step_1",
				ApprovalFlowUID: "apf_employee_profile_change",
				StepOrder:       1,
				RoleUID:         "role_information_center",
			},
		},
	}
	submitUC := usecases.NewSubmitEmployeeProfileChangeRequestUseCase(
		db,
		employeeRepo,
		roleRepo,
		approvalRequestRepo,
		&profileChangeMockApprovalActionRepo{},
		approvalFlowStepRepo,
		changeRequestRepo,
		audit.Noop(),
	)
	getUC := usecases.NewGetEmployeeProfileChangeRequestUseCase(db, roleRepo, changeRequestRepo, approvalRequestRepo, employeeRepo)

	handler := NewEmployeeProfileChangeHandler(
		submitUC,
		nil,
		getUC,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewGetCurrentUserUseCase(db, userRepo, roleRepo, permissionRepo, employeeRepo),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_target/profile-change-requests", bytes.NewBufferString(`{"financialGrade":"FG2"}`))
	req.SetPathValue("uid", "emp_target")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      3,
		UserUID:     managerUser.UID,
		EmployeeUID: ptrString("emp_mgr"),
	}))
	rr := httptest.NewRecorder()
	handler.Submit(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("submit status = %d, body = %s", rr.Code, rr.Body.String())
	}
}

type profileChangeMockDB struct{}

func (db *profileChangeMockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	return db, nil
}
func (db *profileChangeMockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (db *profileChangeMockDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
func (db *profileChangeMockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (db *profileChangeMockDB) Commit() error   { return nil }
func (db *profileChangeMockDB) Rollback() error { return nil }

type profileChangeMockEmployeeRepo struct {
	employeesByUID map[string]*domain.Employee
}

func (m *profileChangeMockEmployeeRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}
func (m *profileChangeMockEmployeeRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	if employee := m.employeesByUID[uid]; employee != nil {
		return cloneEmployeeForProfileTest(employee), nil
	}
	return nil, nil
}
func (m *profileChangeMockEmployeeRepo) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.Employee, error) {
	out := make([]*domain.Employee, 0, len(uids))
	for _, uid := range uids {
		if employee := m.employeesByUID[uid]; employee != nil {
			out = append(out, cloneEmployeeForProfileTest(employee))
		}
	}
	return out, nil
}
func (m *profileChangeMockEmployeeRepo) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	m.employeesByUID[employee.UID] = cloneEmployeeForProfileTest(employee)
	return nil
}
func (m *profileChangeMockEmployeeRepo) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	m.employeesByUID[employee.UID] = cloneEmployeeForProfileTest(employee)
	return nil
}
func (m *profileChangeMockEmployeeRepo) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	return nil, nil
}
func (m *profileChangeMockEmployeeRepo) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}
func (m *profileChangeMockEmployeeRepo) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}
func (m *profileChangeMockEmployeeRepo) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}
func (m *profileChangeMockEmployeeRepo) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.employeesByUID), nil
}

type profileChangeMockUserRepo struct {
	usersByID          map[int64]*domain.User
	usersByEmployeeUID map[string]*domain.User
}

func (m *profileChangeMockUserRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	return m.usersByID[id], nil
}
func (m *profileChangeMockUserRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	return nil, nil
}
func (m *profileChangeMockUserRepo) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	return nil, nil
}
func (m *profileChangeMockUserRepo) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	return nil
}
func (m *profileChangeMockUserRepo) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
	if user.EmployeeUID != nil {
		m.usersByEmployeeUID[*user.EmployeeUID] = user
	}
	m.usersByID[user.ID] = user
	return nil
}
func (m *profileChangeMockUserRepo) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (m *profileChangeMockUserRepo) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.usersByID), nil
}
func (m *profileChangeMockUserRepo) ExistingPhones(ctx context.Context, q ports.Querier, phones []string) ([]string, error) {
	return nil, nil
}
func (m *profileChangeMockUserRepo) GetByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) (*domain.User, error) {
	return m.usersByEmployeeUID[employeeUID], nil
}

type profileChangeMockRoleRepo struct {
	rolesByUser              map[int64][]*domain.Role
	rolesByUID               map[string]*domain.Role
	managedDepartmentsByUser map[int64][]string
}

func newProfileChangeMockRoleRepo() *profileChangeMockRoleRepo {
	return &profileChangeMockRoleRepo{
		rolesByUser:              map[int64][]*domain.Role{},
		rolesByUID:               map[string]*domain.Role{},
		managedDepartmentsByUser: map[int64][]string{},
	}
}

func (m *profileChangeMockRoleRepo) assignRole(userID int64, role *domain.Role) {
	m.rolesByUser[userID] = append(m.rolesByUser[userID], role)
	m.rolesByUID[role.UID] = role
}

func (m *profileChangeMockRoleRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Role, error) {
	return nil, nil
}
func (m *profileChangeMockRoleRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Role, error) {
	return m.rolesByUID[uid], nil
}
func (m *profileChangeMockRoleRepo) GetByName(ctx context.Context, q ports.Querier, name string) (*domain.Role, error) {
	return nil, nil
}
func (m *profileChangeMockRoleRepo) Create(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}
func (m *profileChangeMockRoleRepo) Update(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}
func (m *profileChangeMockRoleRepo) List(ctx context.Context, q ports.Querier) ([]*domain.Role, error) {
	return nil, nil
}
func (m *profileChangeMockRoleRepo) Delete(ctx context.Context, q ports.Querier, id int64) error {
	return nil
}
func (m *profileChangeMockRoleRepo) GetRolesForUser(ctx context.Context, q ports.Querier, userID int64) ([]*domain.Role, error) {
	return m.rolesByUser[userID], nil
}
func (m *profileChangeMockRoleRepo) AssignRoleToUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}
func (m *profileChangeMockRoleRepo) RemoveRoleFromUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}
func (m *profileChangeMockRoleRepo) AssignPermissionToRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}
func (m *profileChangeMockRoleRepo) RemovePermissionFromRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}
func (m *profileChangeMockRoleRepo) SetRolePermissions(ctx context.Context, q ports.Querier, roleID int64, permissionIDs []int64) error {
	return nil
}
func (m *profileChangeMockRoleRepo) AssignRoleToUserWithDepartment(ctx context.Context, q ports.Querier, userID, roleID int64, departmentUID *string) error {
	return nil
}
func (m *profileChangeMockRoleRepo) GetUsersByRoleAndDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID *string) ([]*domain.User, error) {
	return nil, nil
}
func (m *profileChangeMockRoleRepo) IsUserAuthorizedApprover(ctx context.Context, q ports.Querier, userID int64, roleUID string, departmentUID string) (bool, error) {
	for _, role := range m.rolesByUser[userID] {
		if role.UID == roleUID {
			return true, nil
		}
	}
	return false, nil
}
func (m *profileChangeMockRoleRepo) RemoveRoleFromUserForDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID string) error {
	return nil
}
func (m *profileChangeMockRoleRepo) GetDepartmentManager(ctx context.Context, q ports.Querier, departmentUID string) (*domain.User, error) {
	return nil, nil
}
func (m *profileChangeMockRoleRepo) GetManagedDepartmentUIDs(ctx context.Context, q ports.Querier, userID int64) ([]string, error) {
	return m.managedDepartmentsByUser[userID], nil
}

type profileChangeMockPermissionRepo struct{}

func (m *profileChangeMockPermissionRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Permission, error) {
	return nil, nil
}
func (m *profileChangeMockPermissionRepo) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.Permission, error) {
	return nil, nil
}
func (m *profileChangeMockPermissionRepo) List(ctx context.Context, q ports.Querier) ([]*domain.Permission, error) {
	return nil, nil
}
func (m *profileChangeMockPermissionRepo) GetPermissionsForRole(ctx context.Context, q ports.Querier, roleID int64) ([]*domain.Permission, error) {
	return nil, nil
}
func (m *profileChangeMockPermissionRepo) GetPermissionsForRoles(ctx context.Context, q ports.Querier, roleIDs []int64) (map[int64][]*domain.Permission, error) {
	return map[int64][]*domain.Permission{}, nil
}

type profileChangeMockApprovalRequestRepo struct {
	requestsByUID map[string]*domain.ApprovalRequest
}

func (m *profileChangeMockApprovalRequestRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalRequest, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalRequestRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalRequest, error) {
	return m.requestsByUID[uid], nil
}
func (m *profileChangeMockApprovalRequestRepo) Create(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	m.requestsByUID[request.UID] = request
	return nil
}
func (m *profileChangeMockApprovalRequestRepo) Update(ctx context.Context, q ports.Querier, request *domain.ApprovalRequest) error {
	m.requestsByUID[request.UID] = request
	return nil
}
func (m *profileChangeMockApprovalRequestRepo) ListByRequester(ctx context.Context, q ports.Querier, requesterUID string) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalRequestRepo) ListPending(ctx context.Context, q ports.Querier) ([]*domain.ApprovalRequest, error) {
	items := []*domain.ApprovalRequest{}
	for _, request := range m.requestsByUID {
		if request.Status == domain.ApprovalRequestStatusPending {
			items = append(items, request)
		}
	}
	return items, nil
}
func (m *profileChangeMockApprovalRequestRepo) ListPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) ([]*domain.ApprovalRequest, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalRequestRepo) CountPendingByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (int, error) {
	return 0, nil
}

type profileChangeMockApprovalActionRepo struct {
	actionsByRequest map[string][]*domain.ApprovalAction
}

func (m *profileChangeMockApprovalActionRepo) ensure() {
	if m.actionsByRequest == nil {
		m.actionsByRequest = map[string][]*domain.ApprovalAction{}
	}
}
func (m *profileChangeMockApprovalActionRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalAction, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalActionRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalAction, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalActionRepo) Create(ctx context.Context, q ports.Querier, action *domain.ApprovalAction) error {
	m.ensure()
	m.actionsByRequest[action.ApprovalRequestUID] = append(m.actionsByRequest[action.ApprovalRequestUID], action)
	return nil
}
func (m *profileChangeMockApprovalActionRepo) ListByRequest(ctx context.Context, q ports.Querier, approvalRequestUID string) ([]*domain.ApprovalAction, error) {
	m.ensure()
	return m.actionsByRequest[approvalRequestUID], nil
}

type profileChangeMockApprovalFlowStepRepo struct {
	steps map[string]*domain.ApprovalFlowStep
}

func (m *profileChangeMockApprovalFlowStepRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.ApprovalFlowStep, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalFlowStepRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.ApprovalFlowStep, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalFlowStepRepo) Create(ctx context.Context, q ports.Querier, step *domain.ApprovalFlowStep) error {
	return nil
}
func (m *profileChangeMockApprovalFlowStepRepo) Update(ctx context.Context, q ports.Querier, step *domain.ApprovalFlowStep) error {
	return nil
}
func (m *profileChangeMockApprovalFlowStepRepo) Delete(ctx context.Context, q ports.Querier, uid string) error {
	return nil
}
func (m *profileChangeMockApprovalFlowStepRepo) ListByFlow(ctx context.Context, q ports.Querier, approvalFlowUID string) ([]*domain.ApprovalFlowStep, error) {
	return nil, nil
}
func (m *profileChangeMockApprovalFlowStepRepo) CountByFlow(ctx context.Context, q ports.Querier, approvalFlowUID string) (int, error) {
	return 0, nil
}
func (m *profileChangeMockApprovalFlowStepRepo) GetByFlowAndStep(ctx context.Context, q ports.Querier, approvalFlowUID string, stepOrder int) (*domain.ApprovalFlowStep, error) {
	return m.steps[approvalFlowUID+":"+string(rune(stepOrder+'0'))], nil
}
func (m *profileChangeMockApprovalFlowStepRepo) HasPendingRequestsAtStep(ctx context.Context, q ports.Querier, stepUID string) (bool, error) {
	return false, nil
}

type profileChangeMockChangeRequestRepo struct {
	requestsByUID         map[string]*domain.EmployeeProfileChangeRequest
	requestsByApprovalUID map[string]*domain.EmployeeProfileChangeRequest
	requestsByEmployeeUID map[string][]*domain.EmployeeProfileChangeRequest
}

func (m *profileChangeMockChangeRequestRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.EmployeeProfileChangeRequest, error) {
	return m.requestsByUID[uid], nil
}
func (m *profileChangeMockChangeRequestRepo) GetByApprovalRequestUID(ctx context.Context, q ports.Querier, approvalRequestUID string) (*domain.EmployeeProfileChangeRequest, error) {
	return m.requestsByApprovalUID[approvalRequestUID], nil
}
func (m *profileChangeMockChangeRequestRepo) Create(ctx context.Context, q ports.Querier, request *domain.EmployeeProfileChangeRequest) error {
	m.requestsByUID[request.UID] = request
	m.requestsByApprovalUID[request.ApprovalRequestUID] = request
	m.requestsByEmployeeUID[request.EmployeeUID] = append(m.requestsByEmployeeUID[request.EmployeeUID], request)
	return nil
}
func (m *profileChangeMockChangeRequestRepo) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.EmployeeProfileChangeRequest, error) {
	return m.requestsByEmployeeUID[employeeUID], nil
}
func (m *profileChangeMockChangeRequestRepo) ListPending(ctx context.Context, q ports.Querier) ([]*domain.EmployeeProfileChangeRequest, error) {
	items := []*domain.EmployeeProfileChangeRequest{}
	for _, request := range m.requestsByUID {
		items = append(items, request)
	}
	return items, nil
}
func (m *profileChangeMockChangeRequestRepo) HasPendingForEmployee(ctx context.Context, q ports.Querier, employeeUID string) (bool, error) {
	return len(m.requestsByEmployeeUID[employeeUID]) > 0, nil
}

func cloneEmployeeForProfileTest(employee *domain.Employee) *domain.Employee {
	if employee == nil {
		return nil
	}
	cloned := *employee
	if employee.FinancialGrade != nil {
		value := *employee.FinancialGrade
		cloned.FinancialGrade = &value
	}
	if employee.MaritalStatus != nil {
		value := *employee.MaritalStatus
		cloned.MaritalStatus = &value
	}
	if employee.IDCardValidUntil != nil {
		value := *employee.IDCardValidUntil
		cloned.IDCardValidUntil = &value
	}
	return &cloned
}

func ptrString(value string) *string {
	return &value
}
