package usecases

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/banumusa/backend/core/ports"
	"github.com/go-pdf/fpdf"
)

// ExportEmployeesPDFUseCase handles exporting employees to PDF format.
type ExportEmployeesPDFUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	fontPath     string
}

// NewExportEmployeesPDFUseCase creates a new PDF export use case.
func NewExportEmployeesPDFUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	fontPath string,
) *ExportEmployeesPDFUseCase {
	return &ExportEmployeesPDFUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		fontPath:     fontPath,
	}
}

// Column configuration for PDF table.
type pdfColumn struct {
	header string
	width  float64
}

// PDF column definitions.
var pdfColumns = []pdfColumn{
	{header: "Name", width: 45},
	{header: "Mobile", width: 30},
	{header: "Gov ID", width: 35},
	{header: "Univ ID", width: 25},
	{header: "Email", width: 50},
	{header: "Hire Date", width: 25},
	{header: "Status", width: 20},
}

// Execute exports employees to a PDF file.
func (uc *ExportEmployeesPDFUseCase) Execute(ctx context.Context, input ExportEmployeesInput) (*ExportEmployeesOutput, error) {
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

	// Create PDF (A4 Landscape)
	pdf := fpdf.New("L", "mm", "A4", "")

	// Try to add Arabic font if available
	hasArabicFont := false
	if uc.fontPath != "" {
		pdf.AddUTF8Font("NotoSansArabic", "", uc.fontPath)
		hasArabicFont = true
	}

	// Set up header/footer functions
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, "Employee List")
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 10, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04")))
		pdf.Ln(15)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("")
	pdf.AddPage()

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(224, 224, 224)
	for _, col := range pdfColumns {
		pdf.CellFormat(col.width, 8, col.header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 8)
	if hasArabicFont {
		pdf.SetFont("NotoSansArabic", "", 8)
	}
	pdf.SetFillColor(255, 255, 255)

	for i, emp := range employees {
		// Alternate row colors
		if i%2 == 1 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		// Check if we need a new page
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Repeat header
			pdf.SetFont("Arial", "B", 9)
			pdf.SetFillColor(224, 224, 224)
			for _, col := range pdfColumns {
				pdf.CellFormat(col.width, 8, col.header, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Arial", "", 8)
			if hasArabicFont {
				pdf.SetFont("NotoSansArabic", "", 8)
			}
		}

		// Row data
		name := truncate(emp.Name, 20)
		email := ""
		if emp.Email != nil {
			email = truncate(*emp.Email, 25)
		}
		status := string(emp.Status)

		pdf.CellFormat(pdfColumns[0].width, 7, name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColumns[1].width, 7, emp.Mobile, "1", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColumns[2].width, 7, emp.GovernmentID, "1", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColumns[3].width, 7, emp.UniversityID, "1", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColumns[4].width, 7, email, "1", 0, "L", true, 0, "")
		pdf.CellFormat(pdfColumns[5].width, 7, emp.HireDate.Format("2006-01-02"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(pdfColumns[6].width, 7, status, "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
	}

	// Handle empty result
	if len(employees) == 0 {
		pdf.SetFont("Arial", "I", 10)
		pdf.Ln(10)
		pdf.Cell(0, 10, "No employees found matching the specified filters.")
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &ExportEmployeesOutput{
		Data:        buf.Bytes(),
		Filename:    generateFilename(input, "pdf"),
		ContentType: "application/pdf",
	}, nil
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
