package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

type RoleHandler struct {
	listRolesUC       *usecases.ListRolesUseCase
	createRoleUC      *usecases.CreateRoleUseCase
	setPermissionsUC  *usecases.SetRolePermissionsUseCase
	setRoleScopeUC    *usecases.SetRoleScopeUseCase
	listPermissionsUC *usecases.ListPermissionsUseCase
}

func NewRoleHandler(
	listRolesUC *usecases.ListRolesUseCase,
	createRoleUC *usecases.CreateRoleUseCase,
	setPermissionsUC *usecases.SetRolePermissionsUseCase,
	setRoleScopeUC *usecases.SetRoleScopeUseCase,
	listPermissionsUC *usecases.ListPermissionsUseCase,
) *RoleHandler {
	return &RoleHandler{
		listRolesUC:       listRolesUC,
		createRoleUC:      createRoleUC,
		setPermissionsUC:  setPermissionsUC,
		setRoleScopeUC:    setRoleScopeUC,
		listPermissionsUC: listPermissionsUC,
	}
}

func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	output, err := h.listRolesUC.Execute(r.Context())
	if err != nil {
		slog.Error("role_handler.ListRoles.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list roles")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ScopeType   string `json:"scopeType"`
}

func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("role_handler.CreateRole.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}

	output, err := h.createRoleUC.Execute(r.Context(), usecases.CreateRoleInput{
		Name:        req.Name,
		Description: req.Description,
		ScopeType:   req.ScopeType,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrRoleNameExists) {
			writeJSONError(w, http.StatusConflict, "role_name_exists", "Role with this name already exists")
			return
		}
		if errors.Is(err, usecases.ErrInvalidRoleScopeType) {
			writeJSONError(w, http.StatusBadRequest, "invalid_role_scope_type", "Role scope type must be one of: global, department, self")
			return
		}
		slog.Error("role_handler.CreateRole.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to create role")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

type setPermissionsRequest struct {
	PermissionUIDs []string `json:"permissionUids"`
}

func (h *RoleHandler) SetPermissions(w http.ResponseWriter, r *http.Request) {
	roleUID := r.PathValue("uid")
	if roleUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Role UID is required")
		return
	}

	var req setPermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("role_handler.SetPermissions.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	err := h.setPermissionsUC.Execute(r.Context(), usecases.SetRolePermissionsInput{
		RoleUID:        roleUID,
		PermissionUIDs: req.PermissionUIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrRoleNotFound):
			writeJSONError(w, http.StatusNotFound, "role_not_found", "Role not found")
		case errors.Is(err, usecases.ErrCannotModifySystemRole):
			writeJSONError(w, http.StatusForbidden, "cannot_modify_system_role", "Cannot modify system role permissions")
		default:
			slog.Error("role_handler.SetPermissions.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to set permissions")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Permissions updated"})
}

type setRoleScopeRequest struct {
	ScopeType string `json:"scopeType"`
}

func (h *RoleHandler) SetScope(w http.ResponseWriter, r *http.Request) {
	roleUID := r.PathValue("uid")
	if roleUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Role UID is required")
		return
	}

	var req setRoleScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("role_handler.SetScope.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	err := h.setRoleScopeUC.Execute(r.Context(), usecases.SetRoleScopeInput{
		RoleUID:   roleUID,
		ScopeType: req.ScopeType,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrRoleNotFound):
			writeJSONError(w, http.StatusNotFound, "role_not_found", "Role not found")
		case errors.Is(err, usecases.ErrCannotModifySystemRole):
			writeJSONError(w, http.StatusForbidden, "cannot_modify_system_role", "Cannot modify system role scope")
		case errors.Is(err, usecases.ErrInvalidRoleScopeType):
			writeJSONError(w, http.StatusBadRequest, "invalid_role_scope_type", "Role scope type must be one of: global, department, self")
		default:
			slog.Error("role_handler.SetScope.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to set role scope")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Role scope updated"})
}

func (h *RoleHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	output, err := h.listPermissionsUC.Execute(r.Context())
	if err != nil {
		slog.Error("role_handler.ListPermissions.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list permissions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
