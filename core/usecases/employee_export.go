package usecases

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/xuri/excelize/v2"
)

// ExportEmployeesInput defines the filters for exporting employees.
type ExportEmployeesInput struct {
	Status       *domain.EmployeeStatus
	HireDateFrom *time.Time
	HireDateTo   *time.Time
}

// ExportEmployeesOutput contains the exported file data.
type ExportEmployeesOutput struct {
	Data        []byte
	Filename    string
	ContentType string
}

// ExportEmployeesUseCase handles exporting employees to Excel format.
type ExportEmployeesUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
}

// NewExportEmployeesUseCase creates a new export use case.
func NewExportEmployeesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
) *ExportEmployeesUseCase {
	return &ExportEmployeesUseCase{
		db:           db,
		employeeRepo: employeeRepo,
	}
}

// Arabic column headers for export.
var arabicHeaders = []string{
	"الاسم",          // Name
	"رقم الهاتف",     // Mobile
	"الرقم القومي",   // Government ID
	"الرقم الجامعي",  // University ID
	"البريد الإلكتروني", // Email
	"تاريخ التعيين",   // Hire Date
	"الحالة",         // Status
}

// Execute exports employees to an Excel file.
func (uc *ExportEmployeesUseCase) Execute(ctx context.Context, input ExportEmployeesInput) (*ExportEmployeesOutput, error) {
	// Build filter
	filter := &ports.EmployeeListFilter{
		Status:       input.Status,
		HireDateFrom: input.HireDateFrom,
		HireDateTo:   input.HireDateTo,
	}

	// Query employees
	employees, err := uc.employeeRepo.List(ctx, uc.db, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "الموظفين" // "Employees" in Arabic
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Remove default sheet

	// Set RTL direction for the sheet
	rightToLeft := true
	if err := f.SetSheetView(sheetName, 0, &excelize.ViewOptions{RightToLeft: &rightToLeft}); err != nil {
		return nil, fmt.Errorf("failed to set RTL: %w", err)
	}

	// Write headers
	for i, header := range arabicHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style headers (bold)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetRowStyle(sheetName, 1, 1, headerStyle)

	// Write data rows
	for i, emp := range employees {
		rowNum := i + 2 // Start from row 2 (after headers)

		f.SetCellValue(sheetName, cellName(1, rowNum), emp.Name)
		f.SetCellValue(sheetName, cellName(2, rowNum), emp.Mobile)
		f.SetCellValue(sheetName, cellName(3, rowNum), emp.GovernmentID)
		f.SetCellValue(sheetName, cellName(4, rowNum), emp.UniversityID)
		if emp.Email != nil {
			f.SetCellValue(sheetName, cellName(5, rowNum), *emp.Email)
		}
		f.SetCellValue(sheetName, cellName(6, rowNum), emp.HireDate.Format("2006-01-02"))
		f.SetCellValue(sheetName, cellName(7, rowNum), translateStatus(emp.Status))
	}

	// Auto-fit column widths (approximate)
	columnWidths := []float64{25, 15, 18, 15, 30, 15, 12}
	for i, width := range columnWidths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, width)
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return &ExportEmployeesOutput{
		Data:        buf.Bytes(),
		Filename:    generateFilename(input, "xlsx"),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}, nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func translateStatus(status domain.EmployeeStatus) string {
	switch status {
	case domain.EmployeeStatusActive:
		return "نشط" // Active
	case domain.EmployeeStatusInactive:
		return "غير نشط" // Inactive
	case domain.EmployeeStatusTerminated:
		return "منتهي" // Terminated
	default:
		return string(status)
	}
}

func generateFilename(input ExportEmployeesInput, ext string) string {
	today := time.Now().Format("2006-01-02")
	parts := []string{"employees"}

	if input.Status != nil {
		parts = append(parts, string(*input.Status))
	} else if input.HireDateFrom != nil || input.HireDateTo != nil {
		if input.HireDateFrom != nil && input.HireDateTo != nil {
			parts = append(parts, fmt.Sprintf("%s_to_%s",
				input.HireDateFrom.Format("2006-01-02"),
				input.HireDateTo.Format("2006-01-02")))
		} else if input.HireDateFrom != nil {
			parts = append(parts, fmt.Sprintf("from_%s", input.HireDateFrom.Format("2006-01-02")))
		} else {
			parts = append(parts, fmt.Sprintf("to_%s", input.HireDateTo.Format("2006-01-02")))
		}
	} else {
		parts = append(parts, "all")
	}

	parts = append(parts, today)

	filename := ""
	for i, part := range parts {
		if i > 0 {
			filename += "_"
		}
		filename += part
	}

	return filename + "." + ext
}
