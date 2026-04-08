package usecases

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// GenerateImportTemplateOutput contains the template file data.
type GenerateImportTemplateOutput struct {
	Data        []byte
	Filename    string
	ContentType string
}

// GenerateImportTemplateUseCase creates an Excel template for employee import.
type GenerateImportTemplateUseCase struct{}

// NewGenerateImportTemplateUseCase creates a new template generator use case.
func NewGenerateImportTemplateUseCase() *GenerateImportTemplateUseCase {
	return &GenerateImportTemplateUseCase{}
}

// Template column headers and sample data.
var templateHeaders = []string{
	"name", "mobile", "government_id", "university_id", "email", "hire_date", "status",
}

var sampleRows = [][]string{
	{
		"أحمد محمد علي",           // Ahmed Mohamed Ali
		"01012345678",
		"28501011234567",
		"EMP001",
		"ahmed.ali@example.com",
		"2024-01-15",
		"active",
	},
	{
		"فاطمة إبراهيم حسن",        // Fatima Ibrahim Hassan
		"01098765432",
		"29002021234567",
		"EMP002",
		"fatima.hassan@example.com",
		"2024-03-01",
		"active",
	},
}

// Execute generates the import template Excel file.
func (uc *GenerateImportTemplateUseCase) Execute() (*GenerateImportTemplateOutput, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Employees"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Write headers
	for i, header := range templateHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style headers
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
	})
	f.SetRowStyle(sheetName, 1, 1, headerStyle)

	// Write sample data rows
	for i, row := range sampleRows {
		rowNum := i + 2
		for j, val := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, rowNum)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	// Set column widths
	columnWidths := map[string]float64{
		"A": 25, // name
		"B": 15, // mobile
		"C": 18, // government_id
		"D": 12, // university_id
		"E": 30, // email
		"F": 12, // hire_date
		"G": 10, // status
	}
	for col, width := range columnWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// Add data validation for status column
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "G2:G1000"
	dv.SetDropList([]string{"active", "inactive"})
	f.AddDataValidation(sheetName, dv)

	// Add comment to header row explaining required fields
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "A1",
		Author: "System",
		Text:   "Required: Employee name (Arabic or English)",
	})
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "B1",
		Author: "System",
		Text:   "Required: Egyptian mobile number (e.g., 01012345678)",
	})
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "C1",
		Author: "System",
		Text:   "Required: 14-digit national ID",
	})
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "D1",
		Author: "System",
		Text:   "Required: University employee ID",
	})
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "E1",
		Author: "System",
		Text:   "Optional: Email address",
	})
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "F1",
		Author: "System",
		Text:   "Required: Date format YYYY-MM-DD or DD/MM/YYYY",
	})
	f.AddComment(sheetName, excelize.Comment{
		Cell:   "G1",
		Author: "System",
		Text:   "Optional: 'active' (default) or 'inactive'",
	})

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write template file: %w", err)
	}

	return &GenerateImportTemplateOutput{
		Data:        buf.Bytes(),
		Filename:    "employee_import_template.xlsx",
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}, nil
}
