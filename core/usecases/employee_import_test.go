package usecases_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/xuri/excelize/v2"
)

// mockEmployeeRepoForImport implements EmployeeRepository for import tests
type mockEmployeeRepoForImport struct {
	employees        []*domain.Employee
	existingGovIDs   []string
	existingMobiles  []string
	existingUniIDs   []string
	createdEmployees []*domain.Employee
}

// mockUserRepoForImport implements UserRepository for import tests
type mockUserRepoForImport struct {
	users          []*domain.User
	existingPhones []string
	createdUsers   []*domain.User
}

func (m *mockUserRepoForImport) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepoForImport) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepoForImport) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepoForImport) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	user.ID = int64(len(m.createdUsers) + 1)
	m.createdUsers = append(m.createdUsers, user)
	return nil
}

func (m *mockUserRepoForImport) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
	return nil
}

func (m *mockUserRepoForImport) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.User, error) {
	return m.users, nil
}

func (m *mockUserRepoForImport) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.users), nil
}

func (m *mockUserRepoForImport) ExistingPhones(ctx context.Context, q ports.Querier, phones []string) ([]string, error) {
	return m.existingPhones, nil
}

func (m *mockUserRepoForImport) GetByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) (*domain.User, error) {
	for _, u := range m.users {
		if u.EmployeeUID != nil && *u.EmployeeUID == employeeUID {
			return u, nil
		}
	}
	return nil, nil
}

// mockRoleRepoForImport implements RoleRepository for import tests
type mockRoleRepoForImport struct {
	roles         []*domain.Role
	assignedRoles map[int64][]int64 // userID -> roleIDs
}

func (m *mockRoleRepoForImport) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRoleRepoForImport) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.UID == uid {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRoleRepoForImport) GetByName(ctx context.Context, q ports.Querier, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockRoleRepoForImport) Create(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepoForImport) Update(ctx context.Context, q ports.Querier, role *domain.Role) error {
	return nil
}

func (m *mockRoleRepoForImport) List(ctx context.Context, q ports.Querier) ([]*domain.Role, error) {
	return m.roles, nil
}

func (m *mockRoleRepoForImport) Delete(ctx context.Context, q ports.Querier, id int64) error {
	return nil
}

func (m *mockRoleRepoForImport) GetRolesForUser(ctx context.Context, q ports.Querier, userID int64) ([]*domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepoForImport) AssignRoleToUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	if m.assignedRoles == nil {
		m.assignedRoles = make(map[int64][]int64)
	}
	m.assignedRoles[userID] = append(m.assignedRoles[userID], roleID)
	return nil
}

func (m *mockRoleRepoForImport) RemoveRoleFromUser(ctx context.Context, q ports.Querier, userID, roleID int64) error {
	return nil
}

func (m *mockRoleRepoForImport) AssignPermissionToRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}

func (m *mockRoleRepoForImport) RemovePermissionFromRole(ctx context.Context, q ports.Querier, roleID, permissionID int64) error {
	return nil
}

func (m *mockRoleRepoForImport) SetRolePermissions(ctx context.Context, q ports.Querier, roleID int64, permissionIDs []int64) error {
	return nil
}

func (m *mockRoleRepoForImport) AssignRoleToUserWithDepartment(ctx context.Context, q ports.Querier, userID, roleID int64, departmentUID *string) error {
	return nil
}

func (m *mockRoleRepoForImport) GetUsersByRoleAndDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID *string) ([]*domain.User, error) {
	return nil, nil
}

func (m *mockRoleRepoForImport) IsUserAuthorizedApprover(ctx context.Context, q ports.Querier, userID int64, roleUID string, departmentUID string) (bool, error) {
	return false, nil
}

func (m *mockRoleRepoForImport) RemoveRoleFromUserForDepartment(ctx context.Context, q ports.Querier, roleUID string, departmentUID string) error {
	return nil
}

func (m *mockRoleRepoForImport) GetDepartmentManager(ctx context.Context, q ports.Querier, departmentUID string) (*domain.User, error) {
	return nil, nil
}

func (m *mockRoleRepoForImport) GetManagedDepartmentUIDs(ctx context.Context, q ports.Querier, userID int64) ([]string, error) {
	_ = ctx
	_ = q
	_ = userID
	return nil, nil
}

// newTestImportUseCase creates a use case with default mocks for testing
func newTestImportUseCase(db *mockDB, empRepo *mockEmployeeRepoForImport, userRepo *mockUserRepoForImport, roleRepo *mockRoleRepoForImport) *usecases.ImportEmployeesUseCase {
	if userRepo == nil {
		userRepo = &mockUserRepoForImport{}
	}
	if roleRepo == nil {
		roleRepo = &mockRoleRepoForImport{
			roles: []*domain.Role{
				{ID: 1, UID: "role_employee", Name: "Employee", IsSystem: true},
			},
		}
	}
	return usecases.NewImportEmployeesUseCase(db, empRepo, userRepo, roleRepo)
}

func (m *mockEmployeeRepoForImport) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForImport) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForImport) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	m.createdEmployees = append(m.createdEmployees, employee)
	return nil
}

func (m *mockEmployeeRepoForImport) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForImport) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	return m.employees, nil
}

func (m *mockEmployeeRepoForImport) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return m.existingGovIDs, nil
}

func (m *mockEmployeeRepoForImport) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return m.existingMobiles, nil
}

func (m *mockEmployeeRepoForImport) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return m.existingUniIDs, nil
}

func (m *mockEmployeeRepoForImport) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.employees), nil
}

// createTestExcelFile creates a test Excel file with the given data
func createTestExcelFile(t *testing.T, headers []string, rows [][]string) *bytes.Buffer {
	t.Helper()

	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Write headers
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Write data rows
	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("failed to write test Excel file: %v", err)
	}

	return &buf
}

func TestImportEmployees_ValidFile(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "ahmed@test.com", "2024-01-15", "active", "permanent", "normal"},
		{"فاطمة علي", "01098765432", "29002021234567", "EMP002", "", "2024-03-01", "active", "temporary", "contract_employees"},
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}
	userRepo := &mockUserRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, userRepo, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Errorf("expected success=true, got false")
	}

	if output.Imported != 2 {
		t.Errorf("expected 2 imported, got %d", output.Imported)
	}

	if len(empRepo.createdEmployees) != 2 {
		t.Errorf("expected 2 created employees, got %d", len(empRepo.createdEmployees))
	}
	if len(empRepo.createdEmployees) == 2 {
		if empRepo.createdEmployees[0].Type != domain.EmployeeTypePermanent || empRepo.createdEmployees[0].SubType != domain.EmployeeSubTypeNormal {
			t.Errorf("expected first employee classification permanent/normal, got %s/%s", empRepo.createdEmployees[0].Type, empRepo.createdEmployees[0].SubType)
		}
		if empRepo.createdEmployees[1].Type != domain.EmployeeTypeTemporary || empRepo.createdEmployees[1].SubType != domain.EmployeeSubTypeContractEmployees {
			t.Errorf("expected second employee classification temporary/contract_employees, got %s/%s", empRepo.createdEmployees[1].Type, empRepo.createdEmployees[1].SubType)
		}
	}

	// Verify users were created
	if len(userRepo.createdUsers) != 2 {
		t.Errorf("expected 2 created users, got %d", len(userRepo.createdUsers))
	}

	if !mockTx.committed {
		t.Error("expected transaction to be committed")
	}
}

func TestImportEmployees_MissingRequiredFields(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"},    // Missing name
		{"فاطمة علي", "", "29002021234567", "EMP002", "", "2024-03-01", "active", "permanent", "normal"},      // Missing mobile
		{"عمر حسن", "01055555555", "", "EMP003", "", "2024-03-01", "active", "permanent", "normal"},           // Missing gov ID
		{"سارة أحمد", "01066666666", "30001011234567", "", "", "2024-03-01", "active", "permanent", "normal"}, // Missing uni ID
		{"محمد علي", "01077777777", "31001011234567", "EMP005", "", "", "active", "permanent", "normal"},      // Missing hire date
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrImportValidation) {
		t.Fatalf("expected ErrImportValidation, got %v", err)
	}

	if output.Success {
		t.Error("expected success=false")
	}

	if len(output.Errors) != 5 {
		t.Errorf("expected 5 validation errors, got %d", len(output.Errors))
	}
}

func TestImportEmployees_InvalidFormats(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "invalid_mobile", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"},  // Invalid mobile
		{"فاطمة علي", "01098765432", "29002021234567", "EMP002", "", "invalid-date", "active", "permanent", "normal"},   // Invalid date
		{"عمر حسن", "01055555555", "30001011234567", "EMP003", "", "2024-03-01", "unknown", "temporary", "invalid_sub"}, // Invalid status and subtype
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrImportValidation) {
		t.Fatalf("expected ErrImportValidation, got %v", err)
	}

	if output.Success {
		t.Error("expected success=false")
	}

	if len(output.Errors) < 3 {
		t.Errorf("expected at least 3 validation errors, got %d", len(output.Errors))
	}
}

func TestImportEmployees_InFileDuplicates(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"},
		{"فاطمة علي", "01012345678", "29002021234567", "EMP002", "", "2024-03-01", "active", "temporary", "contract_employees"},  // Duplicate mobile
		{"عمر حسن", "01055555555", "28501011234567", "EMP003", "", "2024-03-01", "active", "permanent", "special_needs"},         // Duplicate gov ID
		{"سارة أحمد", "01066666666", "30001011234567", "EMP001", "", "2024-03-01", "active", "temporary", "comprehensive_bonus"}, // Duplicate uni ID
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrImportValidation) {
		t.Fatalf("expected ErrImportValidation, got %v", err)
	}

	if output.Success {
		t.Error("expected success=false")
	}

	// Should have errors for duplicate mobile, gov ID, and uni ID
	if len(output.Errors) < 3 {
		t.Errorf("expected at least 3 duplicate errors, got %d", len(output.Errors))
	}
}

func TestImportEmployees_DatabaseDuplicates(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"},
		{"فاطمة علي", "01098765432", "29002021234567", "EMP002", "", "2024-03-01", "active", "temporary", "contract_employees"},
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{
		existingGovIDs:  []string{"28501011234567"}, // First employee's gov ID already exists
		existingMobiles: []string{"01098765432"},    // Second employee's mobile already exists
	}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrImportValidation) {
		t.Fatalf("expected ErrImportValidation, got %v", err)
	}

	if output.Success {
		t.Error("expected success=false")
	}

	if len(output.Errors) != 2 {
		t.Errorf("expected 2 database duplicate errors, got %d", len(output.Errors))
	}
}

func TestImportEmployees_EmptyFile(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{} // No data rows

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrEmptyFile) {
		t.Fatalf("expected ErrEmptyFile, got %v", err)
	}
}

func TestImportEmployees_InvalidHeaders(t *testing.T) {
	headers := []string{"wrong", "headers", "here"}
	rows := [][]string{
		{"data1", "data2", "data3"},
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrInvalidHeaders) {
		t.Fatalf("expected ErrInvalidHeaders, got %v", err)
	}
}

func TestImportEmployees_FileTooLarge(t *testing.T) {
	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     bytes.NewReader([]byte{}),
		FileSize: 25 * 1024 * 1024, // 25MB - exceeds 20MB limit
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestImportEmployees_CreatesUsersWithEmployeeRole(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "ahmed@test.com", "2024-01-15", "active", "permanent", "normal"},
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}
	userRepo := &mockUserRepoForImport{}
	roleRepo := &mockRoleRepoForImport{
		roles:         []*domain.Role{{ID: 1, UID: "role_employee", Name: "Employee", IsSystem: true}},
		assignedRoles: make(map[int64][]int64),
	}

	uc := usecases.NewImportEmployeesUseCase(db, empRepo, userRepo, roleRepo)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Errorf("expected success=true, got false")
	}

	// Verify user was created
	if len(userRepo.createdUsers) != 1 {
		t.Fatalf("expected 1 created user, got %d", len(userRepo.createdUsers))
	}

	// Verify user has correct phone
	user := userRepo.createdUsers[0]
	if user.Phone != "01012345678" {
		t.Errorf("expected user phone '01012345678', got '%s'", user.Phone)
	}

	// Verify user is linked to employee
	if user.EmployeeUID == nil {
		t.Error("expected user to be linked to employee")
	}

	// Verify user is active
	if !user.IsActive {
		t.Error("expected user to be active")
	}

	// Verify Employee role was assigned
	if len(roleRepo.assignedRoles[user.ID]) != 1 {
		t.Errorf("expected 1 role assigned, got %d", len(roleRepo.assignedRoles[user.ID]))
	}

	if roleRepo.assignedRoles[user.ID][0] != 1 {
		t.Errorf("expected Employee role (ID=1) assigned, got %d", roleRepo.assignedRoles[user.ID][0])
	}
}

func TestImportEmployees_DuplicateUserPhoneError(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"},
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}
	userRepo := &mockUserRepoForImport{
		existingPhones: []string{"01012345678"}, // Phone already exists in users table
		users: []*domain.User{
			{ID: 1, Phone: "01012345678", EmployeeUID: nil}, // User without employee link - conflict!
		},
	}

	uc := newTestImportUseCase(db, empRepo, userRepo, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, usecases.ErrImportValidation) {
		t.Fatalf("expected ErrImportValidation, got %v", err)
	}

	if output.Success {
		t.Error("expected success=false")
	}

	if len(output.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(output.Errors))
	}

	// Verify error message mentions phone number
	if output.Errors[0].Column != "mobile" {
		t.Errorf("expected error on mobile column, got: %s", output.Errors[0].Column)
	}
}

func TestImportEmployees_DateFormats(t *testing.T) {
	headers := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type"}
	rows := [][]string{
		{"أحمد محمد", "01012345678", "28501011234567", "EMP001", "", "2024-01-15", "active", "permanent", "normal"},             // YYYY-MM-DD
		{"فاطمة علي", "01098765432", "29002021234567", "EMP002", "", "15/01/2024", "active", "temporary", "contract_employees"}, // DD/MM/YYYY
	}

	buf := createTestExcelFile(t, headers, rows)

	mockTx := &mockTx{}
	db := &mockDB{tx: mockTx}
	empRepo := &mockEmployeeRepoForImport{}

	uc := newTestImportUseCase(db, empRepo, nil, nil)

	input := usecases.ImportEmployeesInput{
		File:     buf,
		FileSize: int64(buf.Len()),
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Errorf("expected success=true, got false with errors: %v", output.Errors)
	}

	if output.Imported != 2 {
		t.Errorf("expected 2 imported, got %d", output.Imported)
	}

	// Verify dates were parsed correctly
	if len(empRepo.createdEmployees) >= 2 {
		expectedDate1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		if !empRepo.createdEmployees[0].HireDate.Equal(expectedDate1) {
			t.Errorf("expected hire date %v, got %v", expectedDate1, empRepo.createdEmployees[0].HireDate)
		}
		if !empRepo.createdEmployees[1].HireDate.Equal(expectedDate1) {
			t.Errorf("expected hire date %v, got %v", expectedDate1, empRepo.createdEmployees[1].HireDate)
		}
	}
}
