package usecases

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/xuri/excelize/v2"
)

var (
	ErrAttendanceImportInvalidFile    = errors.New("invalid file type: only xlsx files are supported")
	ErrAttendanceImportFileTooLarge   = errors.New("file exceeds maximum size of 20MB")
	ErrAttendanceImportEmptyFile      = errors.New("file contains no data rows")
	ErrAttendanceImportInvalidHeaders = errors.New("invalid or missing column headers — expected at least: الرقم, الإسم, الميعاد, نوع الحركه")
	ErrAttendanceImportValidation     = errors.New("import validation failed")
)

// AttendanceImportError represents a single validation error in the import file.
type AttendanceImportError struct {
	Row    int    `json:"row"`
	Column string `json:"column"`
	Value  string `json:"value"`
	Error  string `json:"error"`
}

// ImportAttendanceLogsInput is the input for the attendance import use case.
type ImportAttendanceLogsInput struct {
	File     io.Reader
	FileSize int64
}

// ImportedAttendanceLog represents a successfully imported attendance record.
type ImportedAttendanceLog struct {
	UID         string `json:"uid"`
	EmployeeUID string `json:"employeeUid"`
	PunchType   string `json:"punchType"`
	PunchedAt   string `json:"punchedAt"`
	Row         int    `json:"row"`
}

// ImportAttendanceLogsOutput is the result of the attendance import operation.
type ImportAttendanceLogsOutput struct {
	Success     bool                    `json:"success"`
	Imported    int                     `json:"imported,omitempty"`
	Skipped     int                     `json:"skipped,omitempty"`
	Records     []ImportedAttendanceLog `json:"records,omitempty"`
	Errors      []AttendanceImportError `json:"errors,omitempty"`
	TotalRows   int                     `json:"totalRows"`
	ValidRows   int                     `json:"validRows"`
	InvalidRows int                     `json:"invalidRows"`
}

// ImportAttendanceLogsUseCase handles bulk attendance log import from Excel files.
type ImportAttendanceLogsUseCase struct {
	db           ports.DB
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	deviceRepo   ports.AttendanceDeviceRepository
	auditor      audit.Auditor
}

// NewImportAttendanceLogsUseCase creates a new attendance import use case.
func NewImportAttendanceLogsUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	employeeRepo ports.EmployeeRepository,
	deviceRepo ports.AttendanceDeviceRepository,
	auditor audit.Auditor,
) *ImportAttendanceLogsUseCase {
	return &ImportAttendanceLogsUseCase{
		db:           db,
		recordRepo:   recordRepo,
		employeeRepo: employeeRepo,
		deviceRepo:   deviceRepo,
		auditor:      auditor,
	}
}

const attendanceImportMaxFileSize = 20 * 1024 * 1024 // 20MB

// Expected attendance import column headers (Arabic).
var expectedAttendanceHeaders = []string{
	"رقم البصمه", "الرقم", "الإسم", "الميعاد", "نوع الحركه",
}

// parsedAttendanceRow holds the parsed data from a single Excel row.
type parsedAttendanceRow struct {
	deviceNumber   string
	governmentID   string
	employeeName   string
	punchedAt      time.Time
	punchType      domain.AttendancePunchType
	row            int
	employeeUID    string // resolved after DB lookup
	deviceUID      string // resolved or default
	deviceUserID   string // resolved from employee
}

// Execute performs the attendance log import from an Excel file.
func (uc *ImportAttendanceLogsUseCase) Execute(ctx context.Context, input ImportAttendanceLogsInput) (*ImportAttendanceLogsOutput, error) {
	if input.FileSize > attendanceImportMaxFileSize {
		return nil, ErrAttendanceImportFileTooLarge
	}

	f, err := excelize.OpenReader(input.File)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAttendanceImportInvalidFile, err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrAttendanceImportEmptyFile
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(rows) < 2 {
		return nil, ErrAttendanceImportEmptyFile
	}

	// Validate headers
	headerRow := rows[0]
	indices, err := parseAttendanceHeaders(headerRow)
	if err != nil {
		return nil, ErrAttendanceImportInvalidHeaders
	}

	// Parse and validate data rows
	dataRows := rows[1:]
	parsed, validationErrors := uc.parseAttendanceRows(dataRows, indices)

	totalRows := len(dataRows)
	invalidRows := len(validationErrors)
	validRows := totalRows - invalidRows

	// Filter out empty rows from totalRows and validRows
	emptyRows := 0
	for _, row := range dataRows {
		if isRowEmpty(row) {
			emptyRows++
		}
	}
	totalRows -= emptyRows
	validRows -= emptyRows

	if len(validationErrors) > 0 {
		return &ImportAttendanceLogsOutput{
			Success:     false,
			Errors:      validationErrors,
			TotalRows:   totalRows,
			ValidRows:   validRows,
			InvalidRows: invalidRows,
		}, ErrAttendanceImportValidation
	}

	// Resolve employees and devices from DB
	dbErrors, err := uc.resolveEmployeesAndDevices(ctx, parsed)
	if err != nil {
		return nil, fmt.Errorf("database lookup failed: %w", err)
	}

	if len(dbErrors) > 0 {
		return &ImportAttendanceLogsOutput{
			Success:     false,
			Errors:      dbErrors,
			TotalRows:   totalRows,
			ValidRows:   validRows - len(dbErrors),
			InvalidRows: len(dbErrors),
		}, ErrAttendanceImportValidation
	}

	// Begin transaction for bulk insert
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	imported := make([]ImportedAttendanceLog, 0, len(parsed))
	skipped := 0

	for _, p := range parsed {
		record := domain.NewAttendanceRecord(
			p.employeeUID,
			p.deviceUID,
			p.deviceUserID,
			p.punchedAt.UTC(),
			p.punchType,
			nil,
		)

		created, err := uc.recordRepo.Create(ctx, tx, record)
		if err != nil {
			if isAttendanceRecordConflictError(err) {
				skipped++
				continue
			}
			return nil, fmt.Errorf("failed to create attendance record at row %d: %w", p.row, err)
		}
		if !created {
			skipped++
			continue
		}

		imported = append(imported, ImportedAttendanceLog{
			UID:         record.UID,
			EmployeeUID: record.EmployeeUID,
			PunchType:   string(record.PunchType),
			PunchedAt:   record.PunchedAt.Format(time.RFC3339),
			Row:         p.row,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Audit log for the import
	go uc.auditor.From(ctx).
		Did(audit.ActionImport).
		On(EntityAttendanceImport, "bulk_import").
		WithMeta("total_rows", totalRows).
		WithMeta("imported", len(imported)).
		WithMeta("skipped", skipped).
		Save(ctx)

	return &ImportAttendanceLogsOutput{
		Success:     true,
		Imported:    len(imported),
		Skipped:     skipped,
		Records:     imported,
		TotalRows:   totalRows,
		ValidRows:   totalRows,
		InvalidRows: 0,
	}, nil
}

const EntityAttendanceImport = "attendance_import"

func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseAttendanceHeaders(headers []string) (map[string]int, error) {
	required := map[string]bool{
		"الرقم":       true,
		"الإسم":       true,
		"الميعاد":     true,
		"نوع الحركه": true,
	}
	indices := make(map[string]int)

	for i, h := range headers {
		indices[strings.TrimSpace(h)] = i
	}

	for req := range required {
		if _, ok := indices[req]; !ok {
			return nil, ErrAttendanceImportInvalidHeaders
		}
	}
	return indices, nil
}

func (uc *ImportAttendanceLogsUseCase) parseAttendanceRows(rows [][]string, indices map[string]int) ([]*parsedAttendanceRow, []AttendanceImportError) {
	var parsed []*parsedAttendanceRow
	var validationErrors []AttendanceImportError

	for i, row := range rows {
		rowNum := i + 2 // 1-indexed, plus header row

		if isRowEmpty(row) {
			continue // Skip completely empty rows
		}

		getValue := func(colName string) string {
			idx, ok := indices[colName]
			if ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		deviceNumber := getValue("رقم البصمه")
		governmentID := getValue("الرقم")
		employeeName := getValue("الإسم")
		timeStr := getValue("الميعاد")
		punchTypeStr := getValue("نوع الحركه")

		rowHasErrors := false
		addError := func(column, value, message string) {
			validationErrors = append(validationErrors, AttendanceImportError{
				Row:    rowNum,
				Column: column,
				Value:  value,
				Error:  message,
			})
			rowHasErrors = true
		}

		// Validate required fields
		if governmentID == "" {
			addError("الرقم", "", "الرقم (employee number) is required")
		}
		if employeeName == "" {
			addError("الإسم", "", "الإسم (employee name) is required")
		}
		if timeStr == "" {
			addError("الميعاد", "", "الميعاد (time) is required")
		}
		if punchTypeStr == "" {
			addError("نوع الحركه", "", "نوع الحركه (punch type) is required")
		}

		// Parse time: format is "ص 09:54 01/04/2026" or "م 02:31 01/04/2026"
		var punchedAt time.Time
		if timeStr != "" {
			var err error
			punchedAt, err = parseArabicDateTime(timeStr)
			if err != nil {
				addError("الميعاد", timeStr, fmt.Sprintf("Invalid time format: %v. Expected format: ص HH:MM DD/MM/YYYY or م HH:MM DD/MM/YYYY", err))
			}
		}

		// Parse punch type: حضور = check_in, إنصراف = check_out
		var punchType domain.AttendancePunchType
		if punchTypeStr != "" {
			switch punchTypeStr {
			case "حضور":
				punchType = domain.AttendancePunchTypeCheckIn
			case "إنصراف":
				punchType = domain.AttendancePunchTypeCheckOut
			case "check_in":
				punchType = domain.AttendancePunchTypeCheckIn
			case "check_out":
				punchType = domain.AttendancePunchTypeCheckOut
			default:
				addError("نوع الحركه", punchTypeStr, "Invalid punch type. Expected: حضور (check-in) or إنصراف (check-out)")
			}
		}

		if rowHasErrors {
			continue
		}

		parsed = append(parsed, &parsedAttendanceRow{
			deviceNumber: deviceNumber,
			governmentID: governmentID,
			employeeName: employeeName,
			punchedAt:    punchedAt,
			punchType:    punchType,
			row:          rowNum,
		})
	}

	return parsed, validationErrors
}

// parseArabicDateTime parses the Arabic date/time format used by attendance devices.
// Format: "ص HH:MM DD/MM/YYYY" (AM) or "م HH:MM DD/MM/YYYY" (PM)
var arabicDateTimeRegex = regexp.MustCompile(`^(ص|م)\s+(\d{1,2}):(\d{2})\s+(\d{2})/(\d{2})/(\d{4})$`)

func parseArabicDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	matches := arabicDateTimeRegex.FindStringSubmatch(s)
	if matches == nil {
		return time.Time{}, fmt.Errorf("does not match expected pattern")
	}

	ampm := matches[1]
	hourStr := matches[2]
	minuteStr := matches[3]
	dayStr := matches[4]
	monthStr := matches[5]
	yearStr := matches[6]

	// Parse components
	var hour, minute, day, month, year int
	fmt.Sscanf(hourStr, "%d", &hour)
	fmt.Sscanf(minuteStr, "%d", &minute)
	fmt.Sscanf(dayStr, "%d", &day)
	fmt.Sscanf(monthStr, "%d", &month)
	fmt.Sscanf(yearStr, "%d", &year)

	// Convert to 24-hour format
	// ص = AM (صباح), م = PM (مساء)
	if ampm == "م" && hour < 12 {
		hour += 12
	} else if ampm == "ص" && hour == 12 {
		hour = 0
	}

	// Load Cairo timezone
	loc, err := time.LoadLocation("Africa/Cairo")
	if err != nil {
		loc = time.FixedZone("Africa/Cairo", 2*60*60)
	}

	t := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)
	return t, nil
}

// resolveEmployeesAndDevices looks up employees by government_id and sets device info.
func (uc *ImportAttendanceLogsUseCase) resolveEmployeesAndDevices(ctx context.Context, parsed []*parsedAttendanceRow) ([]AttendanceImportError, error) {
	var validationErrors []AttendanceImportError

	// Get all employees for lookup
	employees, err := uc.employeeRepo.List(ctx, uc.db, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	// Build lookup maps
	empByGovernmentID := make(map[string]*domain.Employee)
	for _, emp := range employees {
		empByGovernmentID[emp.GovernmentID] = emp
	}

	// Get first available device as default (for imported records)
	defaultDeviceUID := "adev_main_gate" // fallback default

	for _, p := range parsed {
		emp, ok := empByGovernmentID[p.governmentID]
		if !ok {
			validationErrors = append(validationErrors, AttendanceImportError{
				Row:    p.row,
				Column: "الرقم",
				Value:  p.governmentID,
				Error:  fmt.Sprintf("No employee found with government ID '%s'", p.governmentID),
			})
			continue
		}

		p.employeeUID = emp.UID

		// Use the employee's university ID as the device_user_id (consistent with manual creation)
		p.deviceUserID = emp.UniversityID
		if p.deviceUserID == "" {
			p.deviceUserID = emp.GovernmentID
		}

		// Use the device number from the Excel to try to match a device, or use default
		p.deviceUID = defaultDeviceUID
	}

	return validationErrors, nil
}
