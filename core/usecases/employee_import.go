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
	ErrInvalidFileType    = errors.New("invalid file type: only xlsx files are supported")
	ErrFileTooLarge       = errors.New("file exceeds maximum size of 20MB")
	ErrEmptyFile          = errors.New("file contains no data rows")
	ErrInvalidHeaders     = errors.New("invalid or missing column headers")
	ErrImportValidation   = errors.New("import validation failed")
	ErrUserCreationFailed = errors.New("user creation failed")
)

// ImportError represents a single validation error in the import file.
type ImportError struct {
	Row    int    `json:"row"`
	Column string `json:"column"`
	Value  string `json:"value"`
	Error  string `json:"error"`
}

// ImportEmployeesInput is the input for the import use case.
type ImportEmployeesInput struct {
	File     io.Reader
	FileSize int64
}

// ImportedEmployee represents a successfully imported employee.
type ImportedEmployee struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	Row  int    `json:"row"`
}

// ImportEmployeesOutput is the result of the import operation.
type ImportEmployeesOutput struct {
	Success     bool               `json:"success"`
	Imported    int                `json:"imported,omitempty"`
	Employees   []ImportedEmployee `json:"employees,omitempty"`
	Errors      []ImportError      `json:"errors,omitempty"`
	TotalRows   int                `json:"totalRows"`
	ValidRows   int                `json:"validRows"`
	InvalidRows int                `json:"invalidRows"`
}

// ImportEmployeesUseCase handles bulk employee import from Excel files.
type ImportEmployeesUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	userRepo     ports.UserRepository
	roleRepo     ports.RoleRepository
	auditor      audit.Auditor
}

// NewImportEmployeesUseCase creates a new import use case.
func NewImportEmployeesUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *ImportEmployeesUseCase {
	return &ImportEmployeesUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditor:      auditor,
	}
}

// Expected column headers (order matters).
var expectedHeaders = []string{
	"name", "mobile", "government_id", "university_id", "email", "hire_date", "status", "type", "sub_type",
}

const maxFileSize = 20 * 1024 * 1024 // 20MB

// Execute performs the employee import from an Excel file.
func (uc *ImportEmployeesUseCase) Execute(ctx context.Context, input ImportEmployeesInput) (*ImportEmployeesOutput, error) {
	// Validate file size
	if input.FileSize > maxFileSize {
		return nil, ErrFileTooLarge
	}

	// Parse Excel file
	f, err := excelize.OpenReader(input.File)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFileType, err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrEmptyFile
	}
	sheetName := sheets[0]

	// Read all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(rows) < 2 {
		return nil, ErrEmptyFile
	}

	// Validate headers
	headerRow := rows[0]
	if !validateHeaders(headerRow) {
		return nil, ErrInvalidHeaders
	}

	// Parse and validate data rows
	dataRows := rows[1:]
	employees, validationErrors := uc.parseAndValidateRows(dataRows)

	totalRows := len(dataRows)
	invalidRows := len(validationErrors)
	validRows := totalRows - invalidRows

	// If there are validation errors, return them without importing
	if len(validationErrors) > 0 {
		return &ImportEmployeesOutput{
			Success:     false,
			Errors:      validationErrors,
			TotalRows:   totalRows,
			ValidRows:   validRows,
			InvalidRows: invalidRows,
		}, ErrImportValidation
	}

	// Check database uniqueness for employees
	dbErrors, err := uc.checkDatabaseUniqueness(ctx, employees)
	if err != nil {
		return nil, fmt.Errorf("database uniqueness check failed: %w", err)
	}

	if len(dbErrors) > 0 {
		return &ImportEmployeesOutput{
			Success:     false,
			Errors:      dbErrors,
			TotalRows:   totalRows,
			ValidRows:   validRows - len(dbErrors),
			InvalidRows: len(dbErrors),
		}, ErrImportValidation
	}

	// Check for duplicate phones in users table (for auto user creation)
	userPhoneErrors, err := uc.checkUserPhoneUniqueness(ctx, employees)
	if err != nil {
		return nil, fmt.Errorf("user phone uniqueness check failed: %w", err)
	}

	if len(userPhoneErrors) > 0 {
		return &ImportEmployeesOutput{
			Success:     false,
			Errors:      userPhoneErrors,
			TotalRows:   totalRows,
			ValidRows:   validRows - len(userPhoneErrors),
			InvalidRows: len(userPhoneErrors),
		}, ErrImportValidation
	}

	// Begin transaction for bulk insert
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the Employee role for auto-assignment
	employeeRole, err := uc.roleRepo.GetByName(ctx, tx, "Employee")
	if err != nil {
		return nil, fmt.Errorf("failed to get Employee role: %w", err)
	}
	if employeeRole == nil {
		return nil, fmt.Errorf("Employee role not found - run migrations to create it")
	}

	// Insert all employees and create users
	importedEmployees := make([]ImportedEmployee, 0, len(employees))
	for _, emp := range employees {
		// Check if employee already exists (by government ID for re-import scenarios)
		existingEmp, err := uc.findExistingEmployee(ctx, tx, emp.employee)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing employee at row %d: %w", emp.row, err)
		}

		var employeeUID string
		if existingEmp != nil {
			// Update existing employee
			oldName := existingEmp.Name
			oldMobile := existingEmp.Mobile
			oldEmail := ""
			if existingEmp.Email != nil {
				oldEmail = *existingEmp.Email
			}
			oldStatus := existingEmp.Status

			existingEmp.Name = emp.employee.Name
			existingEmp.Mobile = emp.employee.Mobile
			existingEmp.Email = emp.employee.Email
			existingEmp.Status = emp.employee.Status
			existingEmp.Type = emp.employee.Type
			existingEmp.SubType = emp.employee.SubType
			if err := uc.employeeRepo.Update(ctx, tx, existingEmp); err != nil {
				return nil, fmt.Errorf("failed to update employee at row %d: %w", emp.row, err)
			}
			employeeUID = existingEmp.UID

			// Audit log for employee update with field changes
			newEmail := ""
			if emp.employee.Email != nil {
				newEmail = *emp.employee.Email
			}
			go uc.auditor.Actor("system_import").
				Did(audit.ActionUpdate).
				On(audit.EntityEmployee, employeeUID).
				WithMeta("old_name", oldName).
				WithMeta("new_name", emp.employee.Name).
				WithMeta("old_mobile", oldMobile).
				WithMeta("new_mobile", emp.employee.Mobile).
				WithMeta("old_email", oldEmail).
				WithMeta("new_email", newEmail).
				WithMeta("old_status", oldStatus).
				WithMeta("new_status", emp.employee.Status).
				WithMeta("import_row", emp.row).
				Save(ctx)
		} else {
			// Create new employee
			if err := uc.employeeRepo.Create(ctx, tx, emp.employee); err != nil {
				return nil, fmt.Errorf("failed to create employee at row %d: %w", emp.row, err)
			}
			employeeUID = emp.employee.UID

			// Audit log for employee creation
			go uc.auditor.Actor("system_import").
				Did(audit.ActionCreate).
				On(audit.EntityEmployee, employeeUID).
				WithMeta("name", emp.employee.Name).
				WithMeta("mobile", emp.employee.Mobile).
				WithMeta("government_id", emp.employee.GovernmentID).
				WithMeta("university_id", emp.employee.UniversityID).
				WithMeta("import_row", emp.row).
				Save(ctx)
		}

		// Check if user already exists for this employee
		existingUser, err := uc.userRepo.GetByEmployeeUID(ctx, tx, employeeUID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing user at row %d: %w", emp.row, err)
		}

		if existingUser == nil {
			// Create user with employee's mobile as phone
			user := domain.NewUser(emp.employee.Mobile)
			user.EmployeeUID = &employeeUID
			user.IsActive = true

			if err := uc.userRepo.Create(ctx, tx, user); err != nil {
				return nil, fmt.Errorf("failed to create user for employee '%s' at row %d: %w", emp.employee.Name, emp.row, err)
			}

			// Assign Employee role to user
			if err := uc.roleRepo.AssignRoleToUser(ctx, tx, user.ID, employeeRole.ID); err != nil {
				return nil, fmt.Errorf("failed to assign Employee role at row %d: %w", emp.row, err)
			}
		}

		importedEmployees = append(importedEmployees, ImportedEmployee{
			UID:  employeeUID,
			Name: emp.employee.Name,
			Row:  emp.row,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &ImportEmployeesOutput{
		Success:     true,
		Imported:    len(importedEmployees),
		Employees:   importedEmployees,
		TotalRows:   totalRows,
		ValidRows:   totalRows,
		InvalidRows: 0,
	}, nil
}

type parsedEmployee struct {
	employee *domain.Employee
	row      int
}

func validateHeaders(headers []string) bool {
	if len(headers) < len(expectedHeaders) {
		return false
	}
	for i, expected := range expectedHeaders {
		if strings.ToLower(strings.TrimSpace(headers[i])) != expected {
			return false
		}
	}
	return true
}

func (uc *ImportEmployeesUseCase) parseAndValidateRows(rows [][]string) ([]parsedEmployee, []ImportError) {
	var employees []parsedEmployee
	var validationErrors []ImportError

	// Track in-file duplicates
	seenGovernmentIDs := make(map[string]int)
	seenMobiles := make(map[string]int)
	seenUniversityIDs := make(map[string]int)

	for i, row := range rows {
		rowNum := i + 2 // 1-indexed, plus header row

		// Get cell values with defaults for missing columns
		getValue := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		name := getValue(0)
		mobile := getValue(1)
		governmentID := getValue(2)
		universityID := getValue(3)
		email := getValue(4)
		hireDateStr := getValue(5)
		statusStr := getValue(6)
		typeStr := getValue(7)
		subTypeStr := getValue(8)

		rowHasErrors := false
		addError := func(column, value, message string) {
			validationErrors = append(validationErrors, ImportError{
				Row:    rowNum,
				Column: column,
				Value:  value,
				Error:  message,
			})
			rowHasErrors = true
		}

		// Required field validation
		if name == "" {
			addError("name", "", "Required field missing")
		}
		if mobile == "" {
			addError("mobile", "", "Required field missing")
		} else if !isValidMobile(mobile) {
			addError("mobile", mobile, "Invalid mobile number format")
		}
		if governmentID == "" {
			addError("government_id", "", "Required field missing")
		}
		if universityID == "" {
			addError("university_id", "", "Required field missing")
		}

		// Parse hire date
		var hireDate time.Time
		if hireDateStr == "" {
			addError("hire_date", "", "Required field missing")
		} else {
			var err error
			hireDate, err = parseDate(hireDateStr)
			if err != nil {
				addError("hire_date", hireDateStr, "Invalid date format. Use YYYY-MM-DD or DD/MM/YYYY")
			}
		}

		// Parse status (default to active if empty)
		status := domain.EmployeeStatusActive
		if statusStr != "" {
			switch strings.ToLower(statusStr) {
			case "active":
				status = domain.EmployeeStatusActive
			case "inactive":
				status = domain.EmployeeStatusInactive
			default:
				addError("status", statusStr, "Invalid status. Use 'active' or 'inactive'")
			}
		}

		employeeType := domain.EmployeeTypePermanent
		if typeStr != "" {
			employeeType = domain.NormalizeEmployeeType(typeStr)
			if employeeType == "" {
				addError("type", typeStr, "Invalid type. Use 'permanent' or 'temporary'")
			}
		}

		var subType domain.EmployeeSubType
		if subTypeStr == "" {
			addError("sub_type", "", "Required field missing")
		} else {
			subType = domain.NormalizeEmployeeSubType(subTypeStr)
			if subType == "" {
				addError("sub_type", subTypeStr, "Invalid sub type")
			}
		}

		if employeeType != "" && subType != "" && !domain.IsValidEmployeeSubTypeForType(employeeType, subType) {
			addError("sub_type", subTypeStr, "Sub type does not belong to the selected type")
		}

		// Check in-file duplicates
		if mobile != "" {
			if prevRow, exists := seenMobiles[mobile]; exists {
				addError("mobile", mobile, fmt.Sprintf("Duplicate mobile number (also in row %d)", prevRow))
			} else {
				seenMobiles[mobile] = rowNum
			}
		}

		if governmentID != "" {
			if prevRow, exists := seenGovernmentIDs[governmentID]; exists {
				addError("government_id", governmentID, fmt.Sprintf("Duplicate government ID (also in row %d)", prevRow))
			} else {
				seenGovernmentIDs[governmentID] = rowNum
			}
		}

		if universityID != "" {
			if prevRow, exists := seenUniversityIDs[universityID]; exists {
				addError("university_id", universityID, fmt.Sprintf("Duplicate university ID (also in row %d)", prevRow))
			} else {
				seenUniversityIDs[universityID] = rowNum
			}
		}

		if rowHasErrors {
			continue
		}

		// Create employee
		emp := domain.NewEmployee(name, mobile, governmentID, universityID, hireDate)
		emp.Status = status
		emp.Type = employeeType
		emp.SubType = subType
		if email != "" {
			emp.Email = &email
		}

		employees = append(employees, parsedEmployee{
			employee: emp,
			row:      rowNum,
		})
	}

	return employees, validationErrors
}

func (uc *ImportEmployeesUseCase) checkDatabaseUniqueness(ctx context.Context, employees []parsedEmployee) ([]ImportError, error) {
	var validationErrors []ImportError

	// Collect all values to check
	governmentIDs := make([]string, len(employees))
	mobiles := make([]string, len(employees))
	universityIDs := make([]string, len(employees))
	rowByGovernmentID := make(map[string]int)
	rowByMobile := make(map[string]int)
	rowByUniversityID := make(map[string]int)

	for i, emp := range employees {
		governmentIDs[i] = emp.employee.GovernmentID
		mobiles[i] = emp.employee.Mobile
		universityIDs[i] = emp.employee.UniversityID
		rowByGovernmentID[emp.employee.GovernmentID] = emp.row
		rowByMobile[emp.employee.Mobile] = emp.row
		rowByUniversityID[emp.employee.UniversityID] = emp.row
	}

	// Check existing government IDs
	existingGovIDs, err := uc.employeeRepo.ExistingGovernmentIDs(ctx, uc.db, governmentIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range existingGovIDs {
		validationErrors = append(validationErrors, ImportError{
			Row:    rowByGovernmentID[id],
			Column: "government_id",
			Value:  id,
			Error:  "Employee with this government ID already exists",
		})
	}

	// Check existing mobiles
	existingMobiles, err := uc.employeeRepo.ExistingMobiles(ctx, uc.db, mobiles)
	if err != nil {
		return nil, err
	}
	for _, m := range existingMobiles {
		validationErrors = append(validationErrors, ImportError{
			Row:    rowByMobile[m],
			Column: "mobile",
			Value:  m,
			Error:  "Employee with this mobile number already exists",
		})
	}

	// Check existing university IDs
	existingUniIDs, err := uc.employeeRepo.ExistingUniversityIDs(ctx, uc.db, universityIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range existingUniIDs {
		validationErrors = append(validationErrors, ImportError{
			Row:    rowByUniversityID[id],
			Column: "university_id",
			Value:  id,
			Error:  "Employee with this university ID already exists",
		})
	}

	return validationErrors, nil
}

// isValidMobile validates mobile number format.
// Accepts formats like: 01012345678, +201012345678, 201012345678
var mobileRegex = regexp.MustCompile(`^(\+$12$20)$31[0125]\d{8}$`)

func isValidMobile(mobile string) bool {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(mobile, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	return mobileRegex.MatchString(cleaned)
}

// parseDate parses date in YYYY-MM-DD or DD/MM/YYYY format.
func parseDate(s string) (time.Time, error) {
	// Try YYYY-MM-DD first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Try DD/MM/YYYY
	if t, err := time.Parse("02/01/2006", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

// checkUserPhoneUniqueness checks if any employee mobiles already exist as user phones.
func (uc *ImportEmployeesUseCase) checkUserPhoneUniqueness(ctx context.Context, employees []parsedEmployee) ([]ImportError, error) {
	var validationErrors []ImportError

	// Collect all mobiles
	mobiles := make([]string, len(employees))
	rowByMobile := make(map[string]int)
	nameByMobile := make(map[string]string)

	for i, emp := range employees {
		mobiles[i] = emp.employee.Mobile
		rowByMobile[emp.employee.Mobile] = emp.row
		nameByMobile[emp.employee.Mobile] = emp.employee.Name
	}

	// Check existing phones in users table
	existingPhones, err := uc.userRepo.ExistingPhones(ctx, uc.db, mobiles)
	if err != nil {
		return nil, err
	}

	for _, phone := range existingPhones {
		// Check if the existing user is already linked to an employee we're updating
		existingUser, err := uc.userRepo.GetByPhone(ctx, uc.db, phone)
		if err != nil {
			return nil, err
		}

		// If user exists but is not linked to any employee, it's a conflict
		// If user is linked to an employee, we'll handle it in the update flow
		if existingUser != nil && existingUser.EmployeeUID == nil {
			validationErrors = append(validationErrors, ImportError{
				Row:    rowByMobile[phone],
				Column: "mobile",
				Value:  phone,
				Error:  fmt.Sprintf("A user account with phone number '%s' already exists. Employee '%s' cannot be created with a duplicate phone number.", phone, nameByMobile[phone]),
			})
		}
	}

	return validationErrors, nil
}

// findExistingEmployee checks if an employee already exists by government ID.
func (uc *ImportEmployeesUseCase) findExistingEmployee(ctx context.Context, q ports.Querier, emp *domain.Employee) (*domain.Employee, error) {
	// Check by government ID (primary identifier)
	existing, err := uc.employeeRepo.ExistingGovernmentIDs(ctx, q, []string{emp.GovernmentID})
	if err != nil {
		return nil, err
	}

	if len(existing) == 0 {
		return nil, nil
	}

	// Find the employee by iterating (we need the full employee object)
	employees, err := uc.employeeRepo.List(ctx, q, nil)
	if err != nil {
		return nil, err
	}

	for _, e := range employees {
		if e.GovernmentID == emp.GovernmentID {
			return e, nil
		}
	}

	return nil, nil
}
