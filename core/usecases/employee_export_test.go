package usecases_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/xuri/excelize/v2"
)

// mockEmployeeRepoForExport implements EmployeeRepository for export tests
type mockEmployeeRepoForExport struct {
	employees []*domain.Employee
}

func (m *mockEmployeeRepoForExport) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForExport) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForExport) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForExport) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeRepoForExport) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	if filter == nil {
		return m.employees, nil
	}

	var result []*domain.Employee
	for _, emp := range m.employees {
		if filter.Status != nil && emp.Status != *filter.Status {
			continue
		}
		if filter.HireDateFrom != nil && emp.HireDate.Before(*filter.HireDateFrom) {
			continue
		}
		if filter.HireDateTo != nil && emp.HireDate.After(*filter.HireDateTo) {
			continue
		}
		result = append(result, emp)
	}
	return result, nil
}

func (m *mockEmployeeRepoForExport) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForExport) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForExport) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}

func (m *mockEmployeeRepoForExport) Count(ctx context.Context, q ports.Querier) (int, error) {
	return len(m.employees), nil
}

func createTestEmployees() []*domain.Employee {
	email1 := "ahmed@test.com"
	return []*domain.Employee{
		{
			ID:           1,
			UID:          "emp_001",
			Name:         "أحمد محمد",
			Mobile:       "01012345678",
			GovernmentID: "28501011234567",
			UniversityID: "EMP001",
			Email:        &email1,
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
		{
			ID:           3,
			UID:          "emp_003",
			Name:         "عمر حسن",
			Mobile:       "01055555555",
			GovernmentID: "30001011234567",
			UniversityID: "EMP003",
			HireDate:     time.Date(2019, 3, 10, 0, 0, 0, 0, time.UTC),
			Status:       domain.EmployeeStatusInactive,
		},
	}
}

func TestExportEmployees_AllEmployees(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesUseCase(db, repo)

	input := usecases.ExportEmployeesInput{}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Data) == 0 {
		t.Error("expected non-empty data")
	}

	if output.ContentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("unexpected content type: %s", output.ContentType)
	}

	if output.Filename == "" {
		t.Error("expected non-empty filename")
	}

	// Verify Excel content
	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open Excel output: %v", err)
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

	// Should have header + 3 data rows
	if len(rows) != 4 {
		t.Errorf("expected 4 rows (1 header + 3 data), got %d", len(rows))
	}
}

func TestExportEmployees_FilterByStatus(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesUseCase(db, repo)

	activeStatus := domain.EmployeeStatusActive
	input := usecases.ExportEmployeesInput{
		Status: &activeStatus,
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Excel content
	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open Excel output: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}

	// Should have header + 2 active employees
	if len(rows) != 3 {
		t.Errorf("expected 3 rows (1 header + 2 active), got %d", len(rows))
	}

	// Filename should include status filter
	if output.Filename == "" {
		t.Error("expected non-empty filename")
	}
}

func TestExportEmployees_FilterByDateRange(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesUseCase(db, repo)

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)

	input := usecases.ExportEmployeesInput{
		HireDateFrom: &from,
		HireDateTo:   &to,
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify Excel content
	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open Excel output: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}

	// Should have header + 1 employee hired in 2020
	if len(rows) != 2 {
		t.Errorf("expected 2 rows (1 header + 1 in date range), got %d", len(rows))
	}
}

func TestExportEmployees_EmptyResult(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: []*domain.Employee{}}

	uc := usecases.NewExportEmployeesUseCase(db, repo)

	input := usecases.ExportEmployeesInput{}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still create valid Excel file with just headers
	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open Excel output: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}

	// Should have just the header row
	if len(rows) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(rows))
	}
}

func TestExportEmployees_ArabicHeaders(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesUseCase(db, repo)

	input := usecases.ExportEmployeesInput{}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open Excel output: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}

	if len(rows) == 0 {
		t.Fatal("expected at least header row")
	}

	// Check first header is Arabic "الاسم" (Name)
	if len(rows[0]) > 0 && rows[0][0] != "الاسم" {
		t.Errorf("expected first header to be Arabic 'الاسم', got '%s'", rows[0][0])
	}
}
