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
}

func (m *mockEmployeeRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	m.createdEmployees = append(m.createdEmployees, employee)
	return nil
}

func (m *mockEmployeeRepo) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
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

// mockUserRepo implements UserRepository for handler tests
type mockUserRepo struct {
	users        []*domain.User
	createdUsers []*domain.User
}

func (m *mockUserRepo) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	user.ID = int64(len(m.createdUsers) + 1)
	m.createdUsers = append(m.createdUsers, user)
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
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
		},
	}
}

func TestImportEmployees_Integration_Success(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepo{}

	listUC := usecases.NewListEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "ahmed@test.com", "2024-01-15", "active"},
		{"فاطمة علي", "01098765432", "29002021234567", "EMP002", "", "2024-03-01", "active"},
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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status"}
	rows := [][]string{
		{"", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active"}, // Missing name
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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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
	importUC := usecases.NewImportEmployeesUseCase(db, repo, &mockUserRepo{}, defaultMockRoleRepo())
	exportUC := usecases.NewExportEmployeesUseCase(db, repo)
	exportPDFUC := usecases.NewExportEmployeesPDFUseCase(db, repo, "")
	templateUC := usecases.NewGenerateImportTemplateUseCase()

	getUC := usecases.NewGetEmployeeUseCase(db, repo)
	handler := NewEmployeeHandler(getUC, listUC, importUC, exportUC, exportPDFUC, templateUC, nil, nil)

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

// Ensure unused import doesn't cause issues
var _ = io.Discard
