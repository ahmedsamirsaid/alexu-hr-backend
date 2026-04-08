package usecases_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

func TestExportEmployeesPDF_AllEmployees(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	// Empty font path - will use default Arial
	uc := usecases.NewExportEmployeesPDFUseCase(db, repo, "")

	input := usecases.ExportEmployeesInput{}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Data) == 0 {
		t.Error("expected non-empty data")
	}

	if output.ContentType != "application/pdf" {
		t.Errorf("unexpected content type: %s", output.ContentType)
	}

	if output.Filename == "" {
		t.Error("expected non-empty filename")
	}

	// Verify it's a valid PDF (check magic bytes)
	if !bytes.HasPrefix(output.Data, []byte("%PDF")) {
		t.Error("output does not appear to be a valid PDF")
	}
}

func TestExportEmployeesPDF_FilterByStatus(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesPDFUseCase(db, repo, "")

	activeStatus := domain.EmployeeStatusActive
	input := usecases.ExportEmployeesInput{
		Status: &activeStatus,
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Data) == 0 {
		t.Error("expected non-empty data")
	}

	// Verify it's a valid PDF
	if !bytes.HasPrefix(output.Data, []byte("%PDF")) {
		t.Error("output does not appear to be a valid PDF")
	}
}

func TestExportEmployeesPDF_FilterByDateRange(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesPDFUseCase(db, repo, "")

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

	if len(output.Data) == 0 {
		t.Error("expected non-empty data")
	}

	// Verify it's a valid PDF
	if !bytes.HasPrefix(output.Data, []byte("%PDF")) {
		t.Error("output does not appear to be a valid PDF")
	}
}

func TestExportEmployeesPDF_EmptyResult(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: []*domain.Employee{}}

	uc := usecases.NewExportEmployeesPDFUseCase(db, repo, "")

	input := usecases.ExportEmployeesInput{}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still create valid PDF with "no employees" message
	if len(output.Data) == 0 {
		t.Error("expected non-empty data even for empty result")
	}

	if !bytes.HasPrefix(output.Data, []byte("%PDF")) {
		t.Error("output does not appear to be a valid PDF")
	}
}

func TestExportEmployeesPDF_FilenameFormat(t *testing.T) {
	db := &mockDB{tx: &mockTx{}}
	repo := &mockEmployeeRepoForExport{employees: createTestEmployees()}

	uc := usecases.NewExportEmployeesPDFUseCase(db, repo, "")

	tests := []struct {
		name     string
		input    usecases.ExportEmployeesInput
		contains string
	}{
		{
			name:     "no filters",
			input:    usecases.ExportEmployeesInput{},
			contains: "employees_all",
		},
		{
			name: "status filter",
			input: usecases.ExportEmployeesInput{
				Status: func() *domain.EmployeeStatus { s := domain.EmployeeStatusActive; return &s }(),
			},
			contains: "active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Contains([]byte(output.Filename), []byte(tt.contains)) {
				t.Errorf("expected filename to contain '%s', got '%s'", tt.contains, output.Filename)
			}

			if !bytes.HasSuffix([]byte(output.Filename), []byte(".pdf")) {
				t.Errorf("expected filename to end with '.pdf', got '%s'", output.Filename)
			}
		})
	}
}
