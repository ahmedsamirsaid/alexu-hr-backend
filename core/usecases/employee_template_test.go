package usecases_test

import (
	"bytes"
	"testing"

	"github.com/banumusa/backend/core/usecases"
	"github.com/xuri/excelize/v2"
)

func TestGenerateImportTemplate(t *testing.T) {
	uc := usecases.NewGenerateImportTemplateUseCase()

	output, err := uc.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Data) == 0 {
		t.Error("expected non-empty data")
	}

	if output.ContentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("unexpected content type: %s", output.ContentType)
	}

	if output.Filename != "employee_import_template.xlsx" {
		t.Errorf("unexpected filename: %s", output.Filename)
	}
}

func TestGenerateImportTemplate_HasHeaders(t *testing.T) {
	uc := usecases.NewGenerateImportTemplateUseCase()

	output, err := uc.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open template: %v", err)
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

	if len(rows) < 1 {
		t.Fatal("expected at least header row")
	}

	expectedHeaders := []string{"name", "mobile", "government_id", "university_id", "email", "hire_date", "status"}
	headerRow := rows[0]

	if len(headerRow) < len(expectedHeaders) {
		t.Fatalf("expected %d headers, got %d", len(expectedHeaders), len(headerRow))
	}

	for i, expected := range expectedHeaders {
		if headerRow[i] != expected {
			t.Errorf("header %d: expected '%s', got '%s'", i, expected, headerRow[i])
		}
	}
}

func TestGenerateImportTemplate_HasSampleData(t *testing.T) {
	uc := usecases.NewGenerateImportTemplateUseCase()

	output, err := uc.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("failed to open template: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}

	// Should have header + 2 sample rows
	if len(rows) < 3 {
		t.Errorf("expected at least 3 rows (header + 2 samples), got %d", len(rows))
	}

	// Check sample data contains Arabic names
	if len(rows) >= 2 && len(rows[1]) > 0 {
		name := rows[1][0]
		// Check if name contains Arabic characters
		hasArabic := false
		for _, r := range name {
			if r >= 0x0600 && r <= 0x06FF {
				hasArabic = true
				break
			}
		}
		if !hasArabic {
			t.Errorf("expected Arabic sample name, got '%s'", name)
		}
	}
}

func TestGenerateImportTemplate_ValidExcelFile(t *testing.T) {
	uc := usecases.NewGenerateImportTemplateUseCase()

	output, err := uc.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the output is a valid Excel file by parsing it
	f, err := excelize.OpenReader(bytes.NewReader(output.Data))
	if err != nil {
		t.Fatalf("output is not a valid Excel file: %v", err)
	}
	defer f.Close()

	// Verify we can read all sheets and rows without errors
	for _, sheet := range f.GetSheetList() {
		_, err := f.GetRows(sheet)
		if err != nil {
			t.Errorf("failed to read sheet %s: %v", sheet, err)
		}
	}
}
