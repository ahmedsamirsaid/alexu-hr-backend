package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

const maxUploadSize = 20 * 1024 * 1024 // 20MB

// EmployeeHandler handles employee HTTP requests.
type EmployeeHandler struct {
	getUC        *usecases.GetEmployeeUseCase
	listUC       *usecases.ListEmployeesUseCase
	importUC     *usecases.ImportEmployeesUseCase
	exportUC     *usecases.ExportEmployeesUseCase
	exportPDFUC  *usecases.ExportEmployeesPDFUseCase
	templateUC   *usecases.GenerateImportTemplateUseCase
	assignDeptUC *usecases.AssignEmployeeDepartmentUseCase
	removeDeptUC *usecases.RemoveEmployeeDepartmentUseCase
}

// NewEmployeeHandler creates a new employee handler.
func NewEmployeeHandler(
	getUC *usecases.GetEmployeeUseCase,
	listUC *usecases.ListEmployeesUseCase,
	importUC *usecases.ImportEmployeesUseCase,
	exportUC *usecases.ExportEmployeesUseCase,
	exportPDFUC *usecases.ExportEmployeesPDFUseCase,
	templateUC *usecases.GenerateImportTemplateUseCase,
	assignDeptUC *usecases.AssignEmployeeDepartmentUseCase,
	removeDeptUC *usecases.RemoveEmployeeDepartmentUseCase,
) *EmployeeHandler {
	return &EmployeeHandler{
		getUC:        getUC,
		listUC:       listUC,
		importUC:     importUC,
		exportUC:     exportUC,
		exportPDFUC:  exportPDFUC,
		templateUC:   templateUC,
		assignDeptUC: assignDeptUC,
		removeDeptUC: removeDeptUC,
	}
}

// GetEmployeeResponse represents the get employee API response.
type GetEmployeeResponse struct {
	UID           string  `json:"uid"`
	Name          string  `json:"name"`
	Mobile        string  `json:"mobile"`
	GovernmentID  string  `json:"governmentId"`
	UniversityID  string  `json:"universityId"`
	Email         *string `json:"email,omitempty"`
	HireDate      string  `json:"hireDate"`
	Status        string  `json:"status"`
	Type          string  `json:"type"`
	SubType       string  `json:"subType"`
	DepartmentUID *string `json:"departmentUid,omitempty"`
	ShiftUID      *string `json:"shiftUid,omitempty"`
}

// EmployeeListResponse represents the list employees API response.
type EmployeeListResponse struct {
	Employees []EmployeeListItemResponse `json:"employees"`
}

// EmployeeUserInfoResponse represents linked user account info in the response.
type EmployeeUserInfoResponse struct {
	UserUID  string   `json:"userUid"`
	Phone    string   `json:"phone"`
	IsActive bool     `json:"isActive"`
	Roles    []string `json:"roles"`
}

// EmployeeListItemResponse represents an employee in the list response.
type EmployeeListItemResponse struct {
	UID           string                    `json:"uid"`
	Name          string                    `json:"name"`
	Mobile        string                    `json:"mobile"`
	GovernmentID  string                    `json:"governmentId"`
	UniversityID  string                    `json:"universityId"`
	Email         *string                   `json:"email,omitempty"`
	HireDate      string                    `json:"hireDate"`
	Status        string                    `json:"status"`
	Type          string                    `json:"type"`
	SubType       string                    `json:"subType"`
	DepartmentUID *string                   `json:"departmentUid,omitempty"`
	ShiftUID      *string                   `json:"shiftUid,omitempty"`
	User          *EmployeeUserInfoResponse `json:"user,omitempty"`
}

// GetEmployee handles GET /api/v1/employees/{uid}
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	input := usecases.GetEmployeeInput{UID: uid}
	output, err := h.getUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.GetEmployee.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	if isSelfScopedClaims(claims) {
		if claims.EmployeeUID == nil || *claims.EmployeeUID != output.UID {
			writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
			return
		}
	}

	if isScopedDepartmentClaims(claims) {
		if output.DepartmentUID == nil || !claims.HasDepartmentAccess(*output.DepartmentUID) {
			writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this employee is not permitted")
			return
		}
	}

	writeJSON(w, http.StatusOK, GetEmployeeResponse{
		UID:           output.UID,
		Name:          output.Name,
		Mobile:        output.Mobile,
		GovernmentID:  output.GovernmentID,
		UniversityID:  output.UniversityID,
		Email:         output.Email,
		HireDate:      output.HireDate.Format("2006-01-02"),
		Status:        string(output.Status),
		Type:          string(output.Type),
		SubType:       string(output.SubType),
		DepartmentUID: output.DepartmentUID,
		ShiftUID:      output.ShiftUID,
	})
}

// ListEmployees handles GET /api/v1/employees
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	input, err := parseListFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if isScopedDepartmentClaims(claims) {
		input.ManagedDepartmentUIDs = append([]string(nil), claims.ManagedDepartmentUIDs...)
	}

	output, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("employee_handler.ListEmployees.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if isSelfScopedClaims(claims) {
		if claims.EmployeeUID == nil {
			writeJSON(w, http.StatusOK, EmployeeListResponse{Employees: []EmployeeListItemResponse{}})
			return
		}

		filtered := output.Employees[:0]
		for _, emp := range output.Employees {
			if emp.UID == *claims.EmployeeUID {
				filtered = append(filtered, emp)
			}
		}
		output.Employees = filtered
	}

	employees := make([]EmployeeListItemResponse, 0, len(output.Employees))
	for _, emp := range output.Employees {
		item := EmployeeListItemResponse{
			UID:           emp.UID,
			Name:          emp.Name,
			Mobile:        emp.Mobile,
			GovernmentID:  emp.GovernmentID,
			UniversityID:  emp.UniversityID,
			Email:         emp.Email,
			HireDate:      emp.HireDate.Format("2006-01-02"),
			Status:        string(emp.Status),
			Type:          string(emp.Type),
			SubType:       string(emp.SubType),
			DepartmentUID: emp.DepartmentUID,
			ShiftUID:      emp.ShiftUID,
		}
		if emp.User != nil {
			item.User = &EmployeeUserInfoResponse{
				UserUID:  emp.User.UserUID,
				Phone:    emp.User.Phone,
				IsActive: emp.User.IsActive,
				Roles:    emp.User.Roles,
			}
		}
		employees = append(employees, item)
	}

	writeJSON(w, http.StatusOK, EmployeeListResponse{Employees: employees})
}

// ImportEmployees handles POST /api/v1/employees/import
func (h *EmployeeHandler) ImportEmployees(w http.ResponseWriter, r *http.Request) {
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		if err.Error() == "http: request body too large" {
			writeError(w, http.StatusRequestEntityTooLarge, "file exceeds maximum size of 20MB")
			return
		}
		writeError(w, http.StatusBadRequest, "failed to parse form: "+err.Error())
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Validate file extension
	if !isXLSXFile(header.Filename) {
		writeError(w, http.StatusBadRequest, "only .xlsx files are supported")
		return
	}

	input := usecases.ImportEmployeesInput{
		File:     file,
		FileSize: header.Size,
	}

	output, err := h.importUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrInvalidFileType):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrFileTooLarge):
			statusCode = http.StatusRequestEntityTooLarge
		case errors.Is(err, usecases.ErrEmptyFile):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidHeaders):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrImportValidation):
			// Return validation errors with 400 status but include the output
			writeJSON(w, http.StatusBadRequest, output)
			return
		default:
			slog.Error("employee_handler.ImportEmployees.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, output)
}

// ExportEmployees handles GET /api/v1/employees/export
func (h *EmployeeHandler) ExportEmployees(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if isScopedDepartmentClaims(claims) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Export is not permitted for scoped department users")
		return
	}

	input, err := parseExportFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.exportUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("employee_handler.ExportEmployees.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

// ExportEmployeesPDF handles GET /api/v1/employees/export/pdf
func (h *EmployeeHandler) ExportEmployeesPDF(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if isScopedDepartmentClaims(claims) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Export is not permitted for scoped department users")
		return
	}

	input, err := parseExportFilters(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.exportPDFUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("employee_handler.ExportEmployeesPDF.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

// DownloadImportTemplate handles GET /api/v1/employees/import/template
func (h *EmployeeHandler) DownloadImportTemplate(w http.ResponseWriter, r *http.Request) {
	output, err := h.templateUC.Execute()
	if err != nil {
		slog.Error("employee_handler.DownloadImportTemplate.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeFileResponse(w, output.Data, output.Filename, output.ContentType)
}

// AssignDepartmentRequest represents the request body for assigning a department.
type AssignDepartmentRequest struct {
	DepartmentUID string  `json:"departmentUid"`
	ShiftUID      *string `json:"shiftUid,omitempty"`
}

// AssignDepartment handles PUT /api/v1/employees/{uid}/department
func (h *EmployeeHandler) AssignDepartment(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req AssignDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("employee_handler.AssignDepartment.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DepartmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}

	input := usecases.AssignEmployeeDepartmentInput{
		EmployeeUID:   uid,
		DepartmentUID: req.DepartmentUID,
		ShiftUID:      req.ShiftUID,
	}
	err := h.assignDeptUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrDepartmentInactive):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("employee_handler.AssignDepartment.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveDepartment handles DELETE /api/v1/employees/{uid}/department
func (h *EmployeeHandler) RemoveDepartment(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	input := usecases.RemoveEmployeeDepartmentInput{
		EmployeeUID: uid,
	}
	err := h.removeDeptUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("employee_handler.RemoveDepartment.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseListFilters(r *http.Request) (usecases.ListEmployeesInput, error) {
	var input usecases.ListEmployeesInput

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.EmployeeStatus(statusStr)
		switch status {
		case domain.EmployeeStatusActive, domain.EmployeeStatusInactive, domain.EmployeeStatusTerminated:
			input.Status = &status
		default:
			return input, errors.New("invalid status value")
		}
	}

	if dateStr := r.URL.Query().Get("hire_date_from"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_from format, expected YYYY-MM-DD")
		}
		input.HireDateFrom = &date
	}

	if dateStr := r.URL.Query().Get("hire_date_to"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_to format, expected YYYY-MM-DD")
		}
		input.HireDateTo = &date
	}

	if role := r.URL.Query().Get("role"); role != "" {
		input.Role = &role
	}

	return input, nil
}

func parseExportFilters(r *http.Request) (usecases.ExportEmployeesInput, error) {
	var input usecases.ExportEmployeesInput

	// Parse status filter
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := domain.EmployeeStatus(statusStr)
		switch status {
		case domain.EmployeeStatusActive, domain.EmployeeStatusInactive, domain.EmployeeStatusTerminated:
			input.Status = &status
		default:
			return input, errors.New("invalid status value")
		}
	}

	// Parse hire_date_from filter
	if dateStr := r.URL.Query().Get("hire_date_from"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_from format, expected YYYY-MM-DD")
		}
		input.HireDateFrom = &date
	}

	// Parse hire_date_to filter
	if dateStr := r.URL.Query().Get("hire_date_to"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return input, errors.New("invalid hire_date_to format, expected YYYY-MM-DD")
		}
		input.HireDateTo = &date
	}

	return input, nil
}

func isXLSXFile(filename string) bool {
	if len(filename) < 5 {
		return false
	}
	ext := filename[len(filename)-5:]
	return ext == ".xlsx"
}

func isScopedDepartmentClaims(claims *JWTClaims) bool {
	return claims != nil && !claims.HasPermission("*") && claims.IsDepartmentScope()
}

func isSelfScopedClaims(claims *JWTClaims) bool {
	return claims != nil && !claims.HasPermission("*") && claims.IsSelfScope()
}

func writeFileResponse(w http.ResponseWriter, data []byte, filename, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
