package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/xuri/excelize/v2"
)

// Mock implementations for testing

type mockDB struct {
	tx *mockTx
}

func (m *mockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	return m.tx, nil
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

type mockTx struct {
	committed  bool
	rolledBack bool
}

func (m *mockTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (m *mockTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *mockTx) Commit() error {
	m.committed = true
	return nil
}

func (m *mockTx) Rollback() error {
	m.rolledBack = true
	return nil
}

type mockEmployeeRepo struct {
	employees        []*domain.Employee
	createdEmployees []*domain.Employee
	updatedEmployees []*domain.Employee
}

func (m *mockEmployeeRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	for _, employee := range m.employees {
		if employee.UID == uid {
			return employee, nil
		}
	}
	return nil, nil
}

func (m *mockEmployeeRepo) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) GetByUIDs(ctx context.Context, q ports.Querier, uids []string) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	m.createdEmployees = append(m.createdEmployees, employee)
	return nil
}

func (m *mockEmployeeRepo) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	m.updatedEmployees = append(m.updatedEmployees, employee)
	for i, existing := range m.employees {
		if existing.UID == employee.UID {
			m.employees[i] = employee
			break
		}
	}
	return nil
}

func (m *mockEmployeeRepo) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	return m.employees, nil
}

func (m *mockEmployeeRepo) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.employees), nil
}

type mockPenaltyRepo struct {
	penalties map[string][]*domain.Penalty
	removals  map[int64]*domain.PenaltyRemoval
}

func (m *mockPenaltyRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Penalty, error) {
	for _, items := range m.penalties {
		for _, item := range items {
			if item.ID == id {
				copied := *item
				return &copied, nil
			}
		}
	}
	return nil, nil
}

func (m *mockPenaltyRepo) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.Penalty, error) {
	if m.penalties == nil {
		return nil, nil
	}
	items := m.penalties[employeeUID]
	result := make([]*domain.Penalty, 0, len(items))
	for _, item := range items {
		if m.removals != nil {
			if _, removed := m.removals[item.ID]; removed {
				continue
			}
		}
		copied := *item
		result = append(result, &copied)
	}
	return result, nil
}

func (m *mockPenaltyRepo) Create(ctx context.Context, q ports.Querier, penalty *domain.Penalty) error {
	if m.penalties == nil {
		m.penalties = make(map[string][]*domain.Penalty)
	}
	if penalty.ID == 0 {
		var nextID int64 = 1
		for _, items := range m.penalties {
			for _, existing := range items {
				if existing.ID >= nextID {
					nextID = existing.ID + 1
				}
			}
		}
		penalty.ID = nextID
	}
	copied := *penalty
	m.penalties[penalty.EmployeeUID] = append(m.penalties[penalty.EmployeeUID], &copied)
	return nil
}

func (m *mockPenaltyRepo) IsRemoved(ctx context.Context, q ports.Querier, penaltyID int64) (bool, error) {
	if m.removals == nil {
		return false, nil
	}
	_, ok := m.removals[penaltyID]
	return ok, nil
}

func (m *mockPenaltyRepo) CreateRemoval(ctx context.Context, q ports.Querier, removal *domain.PenaltyRemoval) error {
	if m.removals == nil {
		m.removals = make(map[int64]*domain.PenaltyRemoval)
	}
	if removal.ID == 0 {
		removal.ID = int64(len(m.removals) + 1)
	}
	copied := *removal
	m.removals[removal.PenaltyID] = &copied
	return nil
}

func (m *mockPenaltyRepo) DeleteByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) error {
	if m.penalties != nil {
		delete(m.penalties, employeeUID)
	}
	return nil
}

type mockAnnualReportRepo struct {
	reports map[string][]*domain.AnnualReport
}

func (m *mockAnnualReportRepo) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.AnnualReport, error) {
	if m.reports == nil {
		return nil, nil
	}
	items := m.reports[employeeUID]
	result := make([]*domain.AnnualReport, 0, len(items))
	for _, item := range items {
		copied := *item
		result = append(result, &copied)
	}
	return result, nil
}

func (m *mockAnnualReportRepo) Create(ctx context.Context, q ports.Querier, report *domain.AnnualReport) error {
	if m.reports == nil {
		m.reports = make(map[string][]*domain.AnnualReport)
	}
	copied := *report
	m.reports[report.EmployeeUID] = append(m.reports[report.EmployeeUID], &copied)
	return nil
}

func (m *mockAnnualReportRepo) DeleteByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) error {
	if m.reports != nil {
		delete(m.reports, employeeUID)
	}
	return nil
}

type mockIncentiveBonusRepo struct {
	bonuses map[string][]*domain.IncentiveBonus
}

func (m *mockIncentiveBonusRepo) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.IncentiveBonus, error) {
	if m.bonuses == nil {
		return nil, nil
	}
	items := m.bonuses[employeeUID]
	result := make([]*domain.IncentiveBonus, 0, len(items))
	for _, item := range items {
		copied := *item
		result = append(result, &copied)
	}
	return result, nil
}

func (m *mockIncentiveBonusRepo) Create(ctx context.Context, q ports.Querier, bonus *domain.IncentiveBonus) error {
	if m.bonuses == nil {
		m.bonuses = make(map[string][]*domain.IncentiveBonus)
	}
	if bonus.ID == 0 {
		var nextID int64 = 1
		for _, items := range m.bonuses {
			for _, existing := range items {
				if existing.ID >= nextID {
					nextID = existing.ID + 1
				}
			}
		}
		bonus.ID = nextID
	}
	copied := *bonus
	m.bonuses[bonus.EmployeeUID] = append(m.bonuses[bonus.EmployeeUID], &copied)
	return nil
}

func (m *mockIncentiveBonusRepo) DeleteByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) error {
	if m.bonuses != nil {
		delete(m.bonuses, employeeUID)
	}
	return nil
}

type mockObjectStorageService struct{}

func (m *mockObjectStorageService) EnsureBucket(ctx context.Context, bucket string) error {
	return nil
}

func (m *mockObjectStorageService) PresignUpload(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (*ports.PresignedUpload, error) {
	return &ports.PresignedUpload{
		URL:       "https://example.com/upload",
		Method:    http.MethodPut,
		Bucket:    bucket,
		ObjectKey: objectKey,
		ExpiresAt: time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC),
	}, nil
}

func (m *mockObjectStorageService) PresignDownload(ctx context.Context, bucket string, objectKey string, expiry time.Duration, downloadFilename string) (*ports.PresignedDownload, error) {
	return &ports.PresignedDownload{
		URL:       "https://example.com/download",
		Method:    http.MethodGet,
		Bucket:    bucket,
		ObjectKey: objectKey,
		ExpiresAt: time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC),
	}, nil
}

func (m *mockObjectStorageService) StatObject(ctx context.Context, bucket string, objectKey string) (*ports.ObjectInfo, error) {
	return &ports.ObjectInfo{
		Bucket:    bucket,
		ObjectKey: objectKey,
	}, nil
}

// mockUserRepo implements UserRepository for handler tests
type mockUserRepo struct {
	users        []*domain.User
	createdUsers []*domain.User
	updatedUsers []*domain.User
}

func (m *mockUserRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	for _, user := range m.users {
		if user.UID == uid {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Phone == phone {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	user.ID = int64(len(m.createdUsers) + 1)
	m.createdUsers = append(m.createdUsers, user)
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
	m.updatedUsers = append(m.updatedUsers, user)
	for i, existing := range m.users {
		if existing.ID == user.ID {
			m.users[i] = user
			break
		}
	}
	return nil
}

func (m *mockUserRepo) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.User, error) {
	return m.users, nil
}

func (m *mockUserRepo) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.users), nil
}

func (m *mockUserRepo) ExistingPhones(ctx context.Context, q ports.Querier, phones []string) ([]string, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) (*domain.User, error) {
	for _, user := range m.users {
		if user.EmployeeUID != nil && *user.EmployeeUID == employeeUID {
			return user, nil
		}
	}
	return nil, nil
}

// mockRoleRepo implements RoleRepository for handler tests
type mockRoleRepo struct {
	roles []*domain.Role
}

func (m *mockRoleRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) GetByName(ctx context.Context, q ports.Querier, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRoleRepo) Create(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepo) Update(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepo) List(ctx context.Context, q ports.Querier) ([]*domain.Role, error) {
	return m.roles, nil
}

func (m *mockRoleRepo) Delete(ctx context.Context, q ports.Querier, id int64) error {
	return nil
}

func (m *mockRoleRepo) GetRolesForUser(ctx context.Context, q ports.Querier, userID int64) ([]*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) AssignRoleToUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}

func (m *mockRoleRepo) RemoveRoleFromUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}

func (m *mockRoleRepo) AssignPermissionToRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}

func (m *mockRoleRepo) RemovePermissionFromRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}

func (m *mockRoleRepo) SetRolePermissions(ctx context.Context, q ports.Querier, roleID int64, permissionIDs []int64) error {
	return nil
}

func (m *mockRoleRepo) AssignRoleToUserWithDepartment(ctx context.Context, q ports.Querier, userID, roleID int64, departmentUID *string) error {
	return nil
}

func (m *mockRoleRepo) GetUsersByRoleAndDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID *string) ([]*domain.User, error) {
	return nil, nil
}

func (m *mockRoleRepo) IsUserAuthorizedApprover(ctx context.Context, q ports.Querier, userID int64, roleUID string, departmentUID string) (bool, error) {
	return false, nil
}

func (m *mockRoleRepo) RemoveRoleFromUserForDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID string) error {
	return nil
}

func (m *mockRoleRepo) GetDepartmentManager(ctx context.Context, q ports.Querier, departmentUID string) (*domain.User, error) {
	return nil, nil
}

func (m *mockRoleRepo) GetManagedDepartmentUIDs(ctx context.Context, q ports.Querier, userID int64) ([]string, error) {
	_ = ctx
	_ = q
	_ = userID
	return nil, nil
}

// defaultMockRoleRepo returns a role repo with the Employee role
func defaultMockRoleRepo() *mockRoleRepo {
	return &mockRoleRepo{
		roles: []*domain.Role{
			{ID: 1, UID: "role_employee", Name: "Employee", IsSystem: true},
		},
	}
}

// Helper to create multipart form with Excel file
func createMultipartRequest(t *testing.T, fileData []byte) (*http.Request, error) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "employees.xlsx")
	if err != nil {
		return nil, err
	}
	part.Write(fileData)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// Helper to create a test Excel file
func createTestExcel(t *testing.T, headers []string, rows [][]string) []byte {
	t.Helper()

	f := excelize.NewFile()
	sheetName := "Sheet1"

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("failed to write Excel file: %v", err)
	}

	return buf.Bytes()
}

func createTestEmployees() []*domain.Employee {
	email := "test@example.com"
	return []*domain.Employee{
		{
			ID:           1,
			UID:          "emp_001",
			Name:         "أحمد محمد",
			Mobile:       "01012345678",
			GovernmentID: "28501011234567",
			UniversityID: "EMP001",
			Email:        &email,
			HireDate:     time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
			Status:       domain.EmployeeStatusActive,
			Type:         domain.EmployeeTypePermanent,
			SubType:      domain.EmployeeSubTypeNormal,
		},
		{
			ID:           2,
			UID:          "emp_002",
			Name:         "فاطمة علي",
			Mobile:       "01098765432",
			GovernmentID: "29002021234567",
			UniversityID: "EMP002",
			HireDate:     time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			Status:       domain.EmployeeStatusActive,
			Type:         domain.EmployeeTypeTemporary,
			SubType:      domain.EmployeeSubTypeContractEmployees,
		},
	}
}

func TestListEmployees_HRStaffGetsFullListEvenWithSelfScope(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{employees: createTestEmployees()}
	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      10,
		UserUID:     "usr_hr_staff",
		EmployeeUID: ptrString("emp_001"),
		Roles:       []string{"HR Staff"},
		Permissions: []string{"employees:read"},
		AccessScope: domain.RoleScopeSelf,
	}))
	rr := httptest.NewRecorder()

	handler.ListEmployees(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp EmployeeListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Employees) != 2 {
		t.Fatalf("expected 2 employees, got %d", len(resp.Employees))
	}
}

func TestListEmployees_SelfScopedNonHRStaffOnlyGetsSelf(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{employees: createTestEmployees()}
	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      11,
		UserUID:     "usr_employee",
		EmployeeUID: ptrString("emp_001"),
		Roles:       []string{"Employee"},
		Permissions: []string{"employees:read"},
		AccessScope: domain.RoleScopeSelf,
	}))
	rr := httptest.NewRecorder()

	handler.ListEmployees(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp EmployeeListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Employees) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(resp.Employees))
	}
	if resp.Employees[0].UID != "emp_001" {
		t.Fatalf("expected emp_001, got %s", resp.Employees[0].UID)
	}
}

func TestImportEmployees_Integration_Success(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "ahmed@test.com", "2024-01-15", "active", "permanent", "normal"},
		{"فاطمة علي", "01098765432", "29002021234567", "EMP002", "", "2024-03-01", "active", "temporary", "contract_employees"},
	}
	excelData := createTestExcel(t, headers, rows)

	req, err := createMultipartRequest(t, excelData)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ImportEmployees(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var result usecases.ImportEmployeesOutput
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success=true, got false")
	}

	if result.Imported != 2 {
		t.Errorf("expected 2 imported, got %d", result.Imported)
	}
}

func TestImportEmployees_Integration_ValidationErrors(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"}, // Missing name
	}
	excelData := createTestExcel(t, headers, rows)

	req, err := createMultipartRequest(t, excelData)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.ImportEmployees(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var result usecases.ImportEmployeesOutput
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Success {
		t.Error("expected success=false")
	}

	if len(result.Errors) == 0 {
		t.Error("expected validation errors")
	}
}

func TestExportEmployees_Integration_Excel(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{employees: createTestEmployees()}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/export", nil)
	rr := httptest.NewRecorder()

	handler.ExportEmployees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("unexpected content type: %s", contentType)
	}

	if rr.Header().Get("Content-Disposition") == "" {
		t.Error("expected Content-Disposition header")
	}

	// Verify it's a valid Excel file
	_, err := excelize.OpenReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Errorf("response is not valid Excel: %v", err)
	}
}

func TestExportEmployees_Integration_ExcelWithFilters(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{employees: createTestEmployees()}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/export?status=active", nil)
	rr := httptest.NewRecorder()

	handler.ExportEmployees(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestExportEmployees_Integration_PDF(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{employees: createTestEmployees()}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/export/pdf", nil)
	rr := httptest.NewRecorder()

	handler.ExportEmployeesPDF(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("unexpected content type: %s", contentType)
	}

	// Verify it's a valid PDF
	data := rr.Body.Bytes()
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Error("response does not appear to be a valid PDF")
	}
}

func TestExportEmployees_Integration_PDFWithFilters(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{employees: createTestEmployees()}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/export/pdf?status=active&hire_date_from=2020-01-01", nil)
	rr := httptest.NewRecorder()

	handler.ExportEmployeesPDF(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestDownloadTemplate_Integration(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/import/template", nil)
	rr := httptest.NewRecorder()

	handler.DownloadImportTemplate(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("unexpected content type: %s", contentType)
	}

	// Verify it's a valid Excel file with correct structure
	f, err := excelize.OpenReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not valid Excel: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		t.Fatal("expected at least one sheet")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}

	if len(rows) < 3 {
		t.Error("expected header row plus sample data rows")
	}
}

func TestImportEmployees_Integration_InvalidFileType(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	// Create request with non-xlsx file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "employees.csv")
	part.Write([]byte("name,mobile\ntest,123"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler.ImportEmployees(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestImportEmployees_Integration_MissingFile(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo(), audit.Noop())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo, nil)
	handler := NewEmployeeHandler(nil, getUC, listUC, nil, importUC, exportUC, exportPDFUC, templateUC, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	// Create request without file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler.ImportEmployees(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEmployeeHandler_ListPenalties(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{
		employees: []*domain.Employee{
			{UID: "emp_1", Name: "Ahmed"},
		},
	}
	penaltyRepo := &mockPenaltyRepo{
		penalties: map[string][]*domain.Penalty{
			"emp_1": {
				{
					ID:                     1,
					EmployeeUID:            "emp_1",
					PenaltyType:            "warning",
					PenaltyReason:          "Late attendance",
					PenaltyDecisionNumber:  "DEC-1",
					PenaltyDecisionDate:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					PenaltyDecisionFileURL: "minio://documents/penalties/dec-1.pdf",
				},
			},
		},
	}
	downloadUC := usecases.NewGenerateDocumentDownloadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewListEmployeePenaltiesUseCase(db, employeeRepo, penaltyRepo, downloadUC),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/emp_1/penalties", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.ListPenalties(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response EmployeePenaltiesResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Penalties) != 1 {
		t.Fatalf("expected 1 penalty, got %d", len(response.Penalties))
	}
	if response.Penalties[0].ID != 1 {
		t.Fatalf("id = %d, want 1", response.Penalties[0].ID)
	}
	if response.Penalties[0].PenaltyDecisionNumber != "DEC-1" {
		t.Fatalf("penaltyDecisionNumber = %q, want %q", response.Penalties[0].PenaltyDecisionNumber, "DEC-1")
	}
	if response.Penalties[0].PenaltyDecisionFile == nil {
		t.Fatal("expected penaltyDecisionFile preview instructions")
	}
	if response.Penalties[0].PenaltyDecisionFile.Method != http.MethodGet {
		t.Fatalf("penaltyDecisionFile.method = %q, want %q", response.Penalties[0].PenaltyDecisionFile.Method, http.MethodGet)
	}
	if response.Penalties[0].PenaltyDecisionFile.URL == "" {
		t.Fatal("expected penaltyDecisionFile.url")
	}
}

func TestEmployeeHandler_CreatePenalty(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{
		employees: []*domain.Employee{
			{UID: "emp_1", Name: "Ahmed"},
		},
	}
	penaltyRepo := &mockPenaltyRepo{}
	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewCreateEmployeePenaltyUseCase(db, employeeRepo, penaltyRepo, uploadUC),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	body := bytes.NewBufferString(`{"penaltyType":"warning","penaltyReason":"Late attendance","penaltyDecisionNumber":"DEC-1","penaltyDecisionDate":"2026-05-01","penaltyDecisionFile":{"fileName":"decision.pdf","contentType":"application/pdf"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_1/penalties", body)
	req.SetPathValue("employeeUid", "emp_1")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{UserUID: "user_1"}))
	rr := httptest.NewRecorder()

	handler.CreatePenalty(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response CreateEmployeePenaltyResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.PenaltyDecisionFile == nil {
		t.Fatal("expected penaltyDecisionFile upload instructions in response")
	}
	if response.ID == 0 {
		t.Fatal("expected penalty id in response")
	}
	if response.PenaltyDecisionFile.StoredURL == "" {
		t.Fatal("expected storedUrl in upload instructions")
	}

	items := penaltyRepo.penalties["emp_1"]
	if len(items) != 1 {
		t.Fatalf("expected 1 stored penalty, got %d", len(items))
	}
	if items[0].PenaltyDecisionFileURL == "" {
		t.Fatal("expected stored penaltyDecisionFileUrl")
	}
	if items[0].PenaltyDecisionFileURL != response.PenaltyDecisionFile.StoredURL {
		t.Fatalf("stored penaltyDecisionFileUrl = %q", items[0].PenaltyDecisionFileURL)
	}
}

func TestEmployeeHandler_CreatePenaltyRemoval(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	penaltyRepo := &mockPenaltyRepo{
		penalties: map[string][]*domain.Penalty{
			"emp_1": {{
				ID:                     3,
				EmployeeUID:            "emp_1",
				PenaltyType:            "warning",
				PenaltyReason:          "Late attendance",
				PenaltyDecisionNumber:  "DEC-1",
				PenaltyDecisionDate:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				PenaltyDecisionFileURL: "minio://documents/penalties/dec-1.pdf",
			}},
		},
	}
	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewCreateEmployeePenaltyRemovalUseCase(db, penaltyRepo, uploadUC),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	body := bytes.NewBufferString(`{"penaltyRemovalType":"withdrawal","penaltyRemovalNumber":"WD-1","penaltyRemovalDate":"2026-05-08","notes":"Penalty withdrawn","penaltyWithdrawalDecisionFile":{"fileName":"withdrawal.pdf","contentType":"application/pdf"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/penalties/3/removals", body)
	req.SetPathValue("penaltyId", "3")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{UserUID: "user_1"}))
	rr := httptest.NewRecorder()

	handler.CreatePenaltyRemoval(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response EmployeePenaltyRemovalResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.PenaltyID != 3 {
		t.Fatalf("penaltyId = %d, want 3", response.PenaltyID)
	}
	if response.PenaltyWithdrawalDecisionFile == nil {
		t.Fatal("expected upload instructions")
	}
	if response.PenaltyWithdrawalDecisionFile.StoredURL == "" {
		t.Fatal("expected storedUrl in upload instructions")
	}
	if penaltyRepo.removals[3] == nil {
		t.Fatal("expected stored penalty removal")
	}
}

func TestEmployeeHandler_UpdatePenalties(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{
		employees: []*domain.Employee{
			{UID: "emp_1", Name: "Ahmed"},
		},
	}
	penaltyRepo := &mockPenaltyRepo{
		penalties: map[string][]*domain.Penalty{
			"emp_1": {
				{
					ID:                     1,
					EmployeeUID:            "emp_1",
					PenaltyType:            "old",
					PenaltyReason:          "Old",
					PenaltyDecisionNumber:  "OLD-1",
					PenaltyDecisionDate:    time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
					PenaltyDecisionFileURL: "minio://documents/penalties/old.pdf",
				},
			},
		},
	}

	handler := NewEmployeeHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		usecases.NewUpdateEmployeePenaltiesUseCase(db, employeeRepo, penaltyRepo),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	body := bytes.NewBufferString(`{"penalties":[{"penaltyType":"warning","penaltyReason":"Late attendance","penaltyDecisionNumber":"DEC-1","penaltyDecisionDate":"2026-05-01","penaltyDecisionFileUrl":"minio://documents/penalties/dec-1.pdf"},{"penaltyType":"deduction","penaltyReason":"Absence","penaltyDecisionNumber":"DEC-2","penaltyDecisionDate":"2026-05-02","penaltyDecisionFileUrl":"minio://documents/penalties/dec-2.pdf"}]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/employees/emp_1/penalties", body)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.UpdatePenalties(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	items := penaltyRepo.penalties["emp_1"]
	if len(items) != 2 {
		t.Fatalf("expected 2 stored penalties, got %d", len(items))
	}
	if items[0].PenaltyDecisionNumber != "DEC-1" || items[1].PenaltyDecisionNumber != "DEC-2" {
		t.Fatalf("unexpected stored penalties: %+v", items)
	}
}

func TestEmployeeHandler_ListIncentiveBonuses(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	bonusRepo := &mockIncentiveBonusRepo{
		bonuses: map[string][]*domain.IncentiveBonus{
			"emp_1": {{
				ID:               1,
				EmployeeUID:      "emp_1",
				BonusDate:        time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				DecisionNumber:   "BON-1",
				DecisionDate:     time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
				DecisionImageURL: "minio://documents/bonus/bon-1.pdf",
			}},
		},
	}
	downloadUC := usecases.NewGenerateDocumentDownloadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecases.NewListEmployeeIncentiveBonusesUseCase(db, employeeRepo, bonusRepo, downloadUC), nil, nil,
		nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/emp_1/incentive-bonuses", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.ListIncentiveBonuses(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response EmployeeIncentiveBonusesResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Bonuses) != 1 || response.Bonuses[0].ID != 1 {
		t.Fatalf("unexpected bonuses response: %+v", response)
	}
	if response.Bonuses[0].DecisionImage == nil || response.Bonuses[0].DecisionImage.Method != http.MethodGet {
		t.Fatalf("expected decisionImage preview instructions, got %+v", response.Bonuses[0].DecisionImage)
	}
}

func TestEmployeeHandler_CreateIncentiveBonus(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	bonusRepo := &mockIncentiveBonusRepo{}
	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, usecases.NewCreateEmployeeIncentiveBonusUseCase(db, employeeRepo, bonusRepo, uploadUC), nil,
		nil, nil, nil)

	body := bytes.NewBufferString(`{"bonusDate":"2026-05-01","decisionNumber":"BON-1","decisionDate":"2026-05-02","decisionImage":{"fileName":"bonus.pdf","contentType":"application/pdf"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_1/incentive-bonuses", body)
	req.SetPathValue("employeeUid", "emp_1")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{UserUID: "user_1"}))
	rr := httptest.NewRecorder()

	handler.CreateIncentiveBonus(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response CreateEmployeeIncentiveBonusResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.DecisionImage == nil || response.DecisionImage.StoredURL == "" {
		t.Fatalf("expected decisionImage upload instructions, got %+v", response.DecisionImage)
	}
	items := bonusRepo.bonuses["emp_1"]
	if len(items) != 1 || items[0].DecisionImageURL != response.DecisionImage.StoredURL {
		t.Fatalf("unexpected stored bonuses: %+v", items)
	}
}

func TestEmployeeHandler_UpdateIncentiveBonuses(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	bonusRepo := &mockIncentiveBonusRepo{
		bonuses: map[string][]*domain.IncentiveBonus{
			"emp_1": {{
				EmployeeUID:      "emp_1",
				BonusDate:        time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				DecisionNumber:   "OLD-1",
				DecisionDate:     time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
				DecisionImageURL: "minio://documents/bonus/old.pdf",
			}},
		},
	}

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, usecases.NewUpdateEmployeeIncentiveBonusesUseCase(db, employeeRepo, bonusRepo),
		nil, nil, nil)

	body := bytes.NewBufferString(`{"bonuses":[{"bonusDate":"2026-05-01","decisionNumber":"BON-1","decisionDate":"2026-05-02","decisionImageUrl":"minio://documents/bonus/bon-1.pdf"},{"bonusDate":"2026-05-03","decisionNumber":"BON-2","decisionDate":"2026-05-04","decisionImageUrl":"minio://documents/bonus/bon-2.pdf"}]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/employees/emp_1/incentive-bonuses", body)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.UpdateIncentiveBonuses(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	items := bonusRepo.bonuses["emp_1"]
	if len(items) != 2 {
		t.Fatalf("expected 2 stored bonuses, got %d", len(items))
	}
	if items[0].DecisionNumber != "BON-1" || items[1].DecisionNumber != "BON-2" {
		t.Fatalf("unexpected stored bonuses: %+v", items)
	}
}

func TestEmployeeHandler_ListAnnualReports(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	reportRepo := &mockAnnualReportRepo{
		reports: map[string][]*domain.AnnualReport{
			"emp_1": {{
				EmployeeUID:    "emp_1",
				ReportYear:     "2025",
				ReportGrade:    "A",
				Notes:          "Strong performance",
				ReportImageURL: "minio://documents/reports/r1.pdf",
			}},
		},
	}
	downloadUC := usecases.NewGenerateDocumentDownloadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		usecases.NewListEmployeeAnnualReportsUseCase(db, employeeRepo, reportRepo, downloadUC), nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/emp_1/annual-reports", nil)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.ListAnnualReports(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response EmployeeAnnualReportsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Reports) != 1 || response.Reports[0].ReportYear != "2025" {
		t.Fatalf("unexpected reports response: %+v", response)
	}
	if response.Reports[0].ReportImage == nil || response.Reports[0].ReportImage.Method != http.MethodGet {
		t.Fatalf("expected reportImage preview instructions, got %+v", response.Reports[0].ReportImage)
	}
}

func TestEmployeeHandler_CreateAnnualReport(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	reportRepo := &mockAnnualReportRepo{}
	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, usecases.NewCreateEmployeeAnnualReportUseCase(db, employeeRepo, reportRepo, uploadUC), nil)

	body := bytes.NewBufferString(`{"reportYear":"2025","reportGrade":"A","notes":"Strong performance","reportImage":{"fileName":"report.pdf","contentType":"application/pdf"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_1/annual-reports", body)
	req.SetPathValue("employeeUid", "emp_1")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{UserUID: "user_1"}))
	rr := httptest.NewRecorder()

	handler.CreateAnnualReport(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response CreateEmployeeAnnualReportResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ReportImage == nil || response.ReportImage.StoredURL == "" {
		t.Fatalf("expected reportImage upload instructions, got %+v", response.ReportImage)
	}
	items := reportRepo.reports["emp_1"]
	if len(items) != 1 || items[0].ReportImageURL != response.ReportImage.StoredURL {
		t.Fatalf("unexpected stored reports: %+v", items)
	}
}

func TestEmployeeHandler_CreateAnnualReport_NotesOptional(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	reportRepo := &mockAnnualReportRepo{}
	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, usecases.NewCreateEmployeeAnnualReportUseCase(db, employeeRepo, reportRepo, uploadUC), nil)

	body := bytes.NewBufferString(`{"reportYear":"2025","reportGrade":"A","reportImage":{"fileName":"report.pdf","contentType":"application/pdf"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/emp_1/annual-reports", body)
	req.SetPathValue("employeeUid", "emp_1")
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{UserUID: "user_1"}))
	rr := httptest.NewRecorder()

	handler.CreateAnnualReport(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	items := reportRepo.reports["emp_1"]
	if len(items) != 1 {
		t.Fatalf("expected 1 stored report, got %d", len(items))
	}
	if items[0].Notes != "" {
		t.Fatalf("notes = %q, want empty string", items[0].Notes)
	}
}

func TestEmployeeHandler_UpdateAnnualReports(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	employeeRepo := &mockEmployeeRepo{employees: []*domain.Employee{{UID: "emp_1", Name: "Ahmed"}}}
	reportRepo := &mockAnnualReportRepo{
		reports: map[string][]*domain.AnnualReport{
			"emp_1": {{EmployeeUID: "emp_1", ReportYear: "2024", ReportGrade: "B", Notes: "Old", ReportImageURL: "minio://documents/reports/old.pdf"}},
		},
	}

	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, usecases.NewUpdateEmployeeAnnualReportsUseCase(db, employeeRepo, reportRepo))

	body := bytes.NewBufferString(`{"reports":[{"reportYear":"2025","reportGrade":"A","notes":"Strong performance","reportImageUrl":"minio://documents/reports/r1.pdf"},{"reportYear":"2024","reportGrade":"B+","notes":"Good performance","reportImageUrl":"minio://documents/reports/r2.pdf"}]}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/employees/emp_1/annual-reports", body)
	req.SetPathValue("employeeUid", "emp_1")
	rr := httptest.NewRecorder()

	handler.UpdateAnnualReports(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	items := reportRepo.reports["emp_1"]
	if len(items) != 2 || items[0].ReportYear != "2025" || items[1].ReportYear != "2024" {
		t.Fatalf("unexpected stored reports: %+v", items)
	}
}

func TestEmployeeHandler_UpdateOwnProfile(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	uploadUC := usecases.NewGenerateDocumentUploadURLUseCase(&mockObjectStorageService{}, "documents", 15)
	employeeUID := "emp_1"
	oldEmail := "old@example.com"
	oldTelephone := "12345"
	employeeRepo := &mockEmployeeRepo{
		employees: []*domain.Employee{{
			ID:              1,
			UID:             employeeUID,
			Name:            "Old Name",
			Mobile:          "01000000000",
			TelephoneNumber: &oldTelephone,
			Email:           &oldEmail,
		}},
	}
	userRepo := &mockUserRepo{
		users: []*domain.User{{
			ID:          7,
			UID:         "usr_1",
			Phone:       "01000000000",
			EmployeeUID: &employeeUID,
			IsActive:    true,
		}},
	}

	handler := NewEmployeeHandler(
		nil,
		nil,
		nil,
		usecases.NewUpdateOwnEmployeeProfileUseCase(db, employeeRepo, userRepo, uploadUC, audit.Noop()),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	body := bytes.NewBufferString(`{"name":"New Name","mobile":"01099999999","telephoneNumber":"67890","email":"new@example.com","personalPhoto":{"fileName":"profile.jpg","contentType":"image/jpeg"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/employees/me/profile", body)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:      7,
		UserUID:     "usr_1",
		EmployeeUID: &employeeUID,
	}))
	rr := httptest.NewRecorder()

	handler.UpdateOwnProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response UpdateOwnEmployeeProfileResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Name != "New Name" {
		t.Fatalf("name = %q, want %q", response.Name, "New Name")
	}
	if response.Mobile != "01099999999" {
		t.Fatalf("mobile = %q, want %q", response.Mobile, "01099999999")
	}
	if response.TelephoneNumber == nil || *response.TelephoneNumber != "67890" {
		t.Fatalf("telephoneNumber = %v, want 67890", response.TelephoneNumber)
	}
	if response.Email == nil || *response.Email != "new@example.com" {
		t.Fatalf("email = %v, want new@example.com", response.Email)
	}
	if response.PersonalPhoto == nil || response.PersonalPhoto.DocumentType != "personal_photo" || response.PersonalPhoto.StoredURL == "" {
		t.Fatalf("personalPhoto = %+v, want upload instructions", response.PersonalPhoto)
	}
	if response.PersonalPhotoURL == nil || *response.PersonalPhotoURL != response.PersonalPhoto.StoredURL {
		t.Fatalf("personalPhotoUrl = %v, want %q", response.PersonalPhotoURL, response.PersonalPhoto.StoredURL)
	}
	if len(employeeRepo.updatedEmployees) != 1 || employeeRepo.updatedEmployees[0].Mobile != "01099999999" {
		t.Fatalf("employee update not recorded correctly: %+v", employeeRepo.updatedEmployees)
	}
	if employeeRepo.updatedEmployees[0].PersonalPhotoURL == nil || *employeeRepo.updatedEmployees[0].PersonalPhotoURL != response.PersonalPhoto.StoredURL {
		t.Fatalf("employee personal photo not recorded correctly: %+v", employeeRepo.updatedEmployees[0].PersonalPhotoURL)
	}
	if len(userRepo.updatedUsers) != 1 || userRepo.updatedUsers[0].Phone != "01099999999" {
		t.Fatalf("user update not recorded correctly: %+v", userRepo.updatedUsers)
	}
}

func TestEmployeeHandler_UpdateOwnProfile_NoEmployeeLink(t *testing.T) {
	handler := NewEmployeeHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	body := bytes.NewBufferString(`{"name":"New Name","mobile":"01099999999"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/employees/me/profile", body)
	req = req.WithContext(context.WithValue(req.Context(), ClaimsContextKey, &JWTClaims{
		UserID:  7,
		UserUID: "usr_1",
	}))
	rr := httptest.NewRecorder()

	handler.UpdateOwnProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

// Ensure unused import doesn't cause issues
var _ = io.Discard
