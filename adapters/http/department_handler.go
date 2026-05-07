package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

// DepartmentHandler handles department HTTP requests (admin).
type DepartmentHandler struct {
	listDeptUC   *usecases.ListDepartmentsUseCase
	getDeptUC    *usecases.GetDepartmentUseCase
	createDeptUC *usecases.CreateDepartmentUseCase
	updateDeptUC *usecases.UpdateDepartmentUseCase
	assignMgrUC  *usecases.AssignDepartmentManagerUseCase
	removeMgrUC  *usecases.RemoveDepartmentManagerUseCase
	i18nService  ports.I18nService
}

func NewDepartmentHandler(
	listDeptUC *usecases.ListDepartmentsUseCase,
	getDeptUC *usecases.GetDepartmentUseCase,
	createDeptUC *usecases.CreateDepartmentUseCase,
	updateDeptUC *usecases.UpdateDepartmentUseCase,
	assignMgrUC *usecases.AssignDepartmentManagerUseCase,
	removeMgrUC *usecases.RemoveDepartmentManagerUseCase,
	i18nService ports.I18nService,
) *DepartmentHandler {
	return &DepartmentHandler{
		listDeptUC:   listDeptUC,
		getDeptUC:    getDeptUC,
		createDeptUC: createDeptUC,
		updateDeptUC: updateDeptUC,
		assignMgrUC:  assignMgrUC,
		removeMgrUC:  removeMgrUC,
		i18nService:  i18nService,
	}
}

// localizedDeptName picks NameAR when the request locale is "ar", otherwise NameEN.
// NameAR is a pointer since it's optional; falls back to NameEN if nil.
func (h *DepartmentHandler) localizedDeptName(r *http.Request, nameEN string, nameAR *string) string {
	acceptLang := r.Header.Get("Accept-Language")
	if len(acceptLang) >= 2 && acceptLang[:2] == "ar" && nameAR != nil && *nameAR != "" {
		return *nameAR
	}
	return nameEN
}

// Response types

type DepartmentResponse struct {
	UID             string  `json:"uid"`
	Code            string  `json:"code"`
	NameEN          string  `json:"nameEn"`
	NameAR          *string `json:"nameAr,omitempty"`
	LocalizedName   string  `json:"name"`
	IsActive        bool    `json:"isActive"`
	DefaultShiftUID *string `json:"defaultShiftUid,omitempty"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type DepartmentManagerResponse struct {
	UID   string  `json:"uid"`
	Phone string  `json:"phone"`
	Name  *string `json:"name,omitempty"`
}

type DepartmentDetailResponse struct {
	UID             string                     `json:"uid"`
	Code            string                     `json:"code"`
	NameEN          string                     `json:"nameEn"`
	NameAR          *string                    `json:"nameAr,omitempty"`
	LocalizedName   string                     `json:"name"`
	IsActive        bool                       `json:"isActive"`
	DefaultShiftUID *string                    `json:"defaultShiftUid,omitempty"`
	Manager         *DepartmentManagerResponse `json:"manager"`
	CreatedAt       string                     `json:"createdAt"`
	UpdatedAt       string                     `json:"updatedAt"`
}

type ListDepartmentsResponse struct {
	Departments []DepartmentResponse `json:"departments"`
}

// Request types

type CreateDepartmentRequest struct {
	Code            string  `json:"code"`
	NameEN          string  `json:"nameEn"`
	NameAR          *string `json:"nameAr,omitempty"`
	DefaultShiftUID *string `json:"defaultShiftUid,omitempty"`
}

type UpdateDepartmentRequest struct {
	Code            *string `json:"code,omitempty"`
	NameEN          *string `json:"nameEn,omitempty"`
	NameAR          *string `json:"nameAr,omitempty"`
	IsActive        *bool   `json:"isActive,omitempty"`
	DefaultShiftUID *string `json:"defaultShiftUid,omitempty"`
}

type AssignManagerRequest struct {
	UserUID string `json:"userUid"`
}

// ListDepartments handles GET /api/v1/admin/departments
func (h *DepartmentHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	activeOnly := r.URL.Query().Get("active") == "true"

	output, err := h.listDeptUC.Execute(r.Context(), activeOnly)
	if err != nil {
		slog.Error("department_handler.ListDepartments.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	departments := make([]DepartmentResponse, 0, len(output.Departments))
	for _, d := range output.Departments {
		if !canAccessDepartment(claims, d.UID) {
			continue
		}

		departments = append(departments, DepartmentResponse{
			UID:             d.UID,
			Code:            d.Code,
			NameEN:          d.NameEN,
			NameAR:          d.NameAR,
			LocalizedName:   h.localizedDeptName(r, d.NameEN, d.NameAR),
			IsActive:        d.IsActive,
			DefaultShiftUID: d.DefaultShiftUID,
			CreatedAt:       d.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, ListDepartmentsResponse{Departments: departments})
}

// GetDepartment handles GET /api/v1/admin/departments/{uid}
func (h *DepartmentHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	if !canAccessDepartment(claims, uid) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

	output, err := h.getDeptUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrDepartmentNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("department_handler.GetDepartment.execute_usecase", "error", err)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	var manager *DepartmentManagerResponse
	if output.Manager != nil && output.Manager.User != nil {
		manager = &DepartmentManagerResponse{
			UID:   output.Manager.User.UID,
			Phone: output.Manager.User.Phone,
		}
		if output.Manager.Employee != nil {
			manager.Name = &output.Manager.Employee.Name
		}
	}

	writeJSON(w, http.StatusOK, DepartmentDetailResponse{
		UID:             output.Department.UID,
		Code:            output.Department.Code,
		NameEN:          output.Department.NameEN,
		NameAR:          output.Department.NameAR,
		LocalizedName:   h.localizedDeptName(r, output.Department.NameEN, output.Department.NameAR),
		IsActive:        output.Department.IsActive,
		DefaultShiftUID: output.Department.DefaultShiftUID,
		Manager:         manager,
		CreatedAt:       output.Department.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       output.Department.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CreateDepartment handles POST /api/v1/admin/departments
func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("department_handler.CreateDepartment.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}
	if req.NameEN == "" {
		writeError(w, http.StatusBadRequest, "nameEn is required")
		return
	}

	input := usecases.CreateDepartmentInput{
		Code:            req.Code,
		NameEN:          req.NameEN,
		NameAR:          req.NameAR,
		DefaultShiftUID: req.DefaultShiftUID,
	}
	output, err := h.createDeptUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrDepartmentCodeExists) {
			statusCode = http.StatusConflict
		} else {
			slog.Error("department_handler.CreateDepartment.execute_usecase", "error", err)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	writeJSON(w, http.StatusCreated, DepartmentResponse{
		UID:             output.Department.UID,
		Code:            output.Department.Code,
		NameEN:          output.Department.NameEN,
		NameAR:          output.Department.NameAR,
		LocalizedName:   h.localizedDeptName(r, output.Department.NameEN, output.Department.NameAR),
		IsActive:        output.Department.IsActive,
		DefaultShiftUID: output.Department.DefaultShiftUID,
		CreatedAt:       output.Department.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       output.Department.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateDepartment handles PATCH /api/v1/admin/departments/{uid}
func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	if !canAccessDepartment(claims, uid) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("department_handler.UpdateDepartment.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.UpdateDepartmentInput{
		UID:             uid,
		Code:            req.Code,
		NameEN:          req.NameEN,
		NameAR:          req.NameAR,
		IsActive:        req.IsActive,
		DefaultShiftUID: req.DefaultShiftUID,
	}
	output, err := h.updateDeptUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrDepartmentCodeExists):
			statusCode = http.StatusConflict
		default:
			slog.Error("department_handler.UpdateDepartment.execute_usecase", "error", err)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	writeJSON(w, http.StatusOK, DepartmentResponse{
		UID:             output.Department.UID,
		Code:            output.Department.Code,
		NameEN:          output.Department.NameEN,
		NameAR:          output.Department.NameAR,
		LocalizedName:   h.localizedDeptName(r, output.Department.NameEN, output.Department.NameAR),
		IsActive:        output.Department.IsActive,
		DefaultShiftUID: output.Department.DefaultShiftUID,
		CreatedAt:       output.Department.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       output.Department.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// AssignManager handles POST /api/v1/admin/departments/{uid}/manager
func (h *DepartmentHandler) AssignManager(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	if !canAccessDepartment(claims, uid) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

	var req AssignManagerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("department_handler.AssignManager.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserUID == "" {
		writeError(w, http.StatusBadRequest, "userUid is required")
		return
	}

	input := usecases.AssignDepartmentManagerInput{
		DepartmentUID: uid,
		UserUID:       req.UserUID,
	}
	err := h.assignMgrUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrDepartmentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrUserNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrDepartmentInactive):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("department_handler.AssignManager.execute_usecase", "error", err)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveManager handles DELETE /api/v1/admin/departments/{uid}/manager
func (h *DepartmentHandler) RemoveManager(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	if !canAccessDepartment(claims, uid) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

	input := usecases.RemoveDepartmentManagerInput{
		DepartmentUID: uid,
	}
	err := h.removeMgrUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrDepartmentNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("department_handler.RemoveManager.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func canAccessDepartment(claims *JWTClaims, departmentUID string) bool {
	if claims == nil {
		return false
	}
	if claims.HasPermission("*") || claims.IsGlobalScope() {
		return true
	}
	if claims.IsDepartmentScope() {
		return claims.HasDepartmentAccess(departmentUID)
	}
	return false
}
