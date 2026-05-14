package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/banumusa/backend/core/usecases"
)

type UserHandler struct {
	listUsersUC  *usecases.ListUsersUseCase
	createUserUC *usecases.CreateUserUseCase
	updateUserUC *usecases.UpdateUserUseCase
	assignRoleUC *usecases.AssignRoleUseCase
	removeRoleUC *usecases.RemoveRoleUseCase
}

func NewUserHandler(
	listUsersUC *usecases.ListUsersUseCase,
	createUserUC *usecases.CreateUserUseCase,
	updateUserUC *usecases.UpdateUserUseCase,
	assignRoleUC *usecases.AssignRoleUseCase,
	removeRoleUC *usecases.RemoveRoleUseCase,
) *UserHandler {
	return &UserHandler{
		listUsersUC:  listUsersUC,
		createUserUC: createUserUC,
		updateUserUC: updateUserUC,
		assignRoleUC: assignRoleUC,
		removeRoleUC: removeRoleUC,
	}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	output, err := h.listUsersUC.Execute(r.Context(), usecases.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		slog.Error("user_handler.ListUsers.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to list users")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

type createUserRequest struct {
	Phone       string  `json:"phone"`
	Password    *string `json:"password"`
	EmployeeUID *string `json:"employeeUid"`
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("user_handler.CreateUser.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Phone == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Phone is required")
		return
	}

	output, err := h.createUserUC.Execute(r.Context(), usecases.CreateUserInput{
		Phone:       req.Phone,
		Password:    req.Password,
		EmployeeUID: req.EmployeeUID,
	})
	if err != nil {
		if errors.Is(err, usecases.ErrPhoneAlreadyExists) {
			writeJSONError(w, http.StatusConflict, "phone_already_exists", "Phone number already registered")
			return
		}
		slog.Error("user_handler.CreateUser.execute_usecase", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to create user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

type updateUserRequest struct {
	Phone             *string `json:"phone"`
	Password          *string `json:"password"`
	EmployeeUID       *string `json:"employeeUid"`
	IsActive          *bool   `json:"isActive"`
	PreferredLanguage *string `json:"preferredLanguage"`
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userUID := r.PathValue("uid")
	if userUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "User UID is required")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("user_handler.UpdateUser.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	err := h.updateUserUC.Execute(r.Context(), usecases.UpdateUserInput{
		UserUID:           userUID,
		Phone:             req.Phone,
		Password:          req.Password,
		EmployeeUID:       req.EmployeeUID,
		IsActive:          req.IsActive,
		PreferredLanguage: req.PreferredLanguage,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrUserNotFound):
			writeJSONError(w, http.StatusNotFound, "user_not_found", "User not found")
		case errors.Is(err, usecases.ErrPhoneAlreadyExists):
			writeJSONError(w, http.StatusConflict, "phone_already_exists", "Phone number already registered")
		default:
			slog.Error("user_handler.UpdateUser.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to update user")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated"})
}

type assignRoleRequest struct {
	RoleUID       string  `json:"roleUid"`
	DepartmentUID *string `json:"departmentUid,omitempty"`
}

func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userUID := r.PathValue("uid")
	if userUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "User UID is required")
		return
	}

	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("user_handler.AssignRole.decode_request", "error", err)
		writeJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.RoleUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "Role UID is required")
		return
	}

	err := h.assignRoleUC.Execute(r.Context(), usecases.AssignRoleInput{
		UserUID:       userUID,
		RoleUID:       req.RoleUID,
		DepartmentUID: req.DepartmentUID,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrUserNotFound):
			writeJSONError(w, http.StatusNotFound, "user_not_found", "User not found")
		case errors.Is(err, usecases.ErrRoleNotFound):
			writeJSONError(w, http.StatusNotFound, "role_not_found", "Role not found")
		case errors.Is(err, usecases.ErrRoleScopeRequired):
			writeJSONError(w, http.StatusBadRequest, "role_scope_required", "This role requires department scope")
		case errors.Is(err, usecases.ErrRoleScopeConflict):
			writeJSONError(w, http.StatusBadRequest, "role_scope_conflict", "This role does not allow department scope")
		case errors.Is(err, usecases.ErrInvalidRoleScopeType):
			writeJSONError(w, http.StatusBadRequest, "invalid_role_scope_type", "Role scope type must be one of: global, department, self")
		default:
			slog.Error("user_handler.AssignRole.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to assign role")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Role assigned"})
}

func (h *UserHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userUID := r.PathValue("uid")
	roleUID := r.PathValue("roleUid")

	if userUID == "" || roleUID == "" {
		writeJSONError(w, http.StatusBadRequest, "validation_error", "User UID and Role UID are required")
		return
	}

	err := h.removeRoleUC.Execute(r.Context(), usecases.RemoveRoleInput{
		UserUID: userUID,
		RoleUID: roleUID,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrUserNotFound):
			writeJSONError(w, http.StatusNotFound, "user_not_found", "User not found")
		case errors.Is(err, usecases.ErrRoleNotFound):
			writeJSONError(w, http.StatusNotFound, "role_not_found", "Role not found")
		default:
			slog.Error("user_handler.RemoveRole.execute_usecase", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "internal_error", "Failed to remove role")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Role removed"})
}
