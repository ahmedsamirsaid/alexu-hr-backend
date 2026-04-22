package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/usecases"
)

// LeaveRequestHandler handles leave request HTTP requests.
type LeaveRequestHandler struct {
	submitUC               *usecases.SubmitLeaveRequestUseCase
	cancelUC               *usecases.CancelLeaveRequestUseCase
	listUC                 *usecases.ListLeaveRequestsUseCase
	getUC                  *usecases.GetLeaveRequestUseCase
	listPendingApprovalsUC *usecases.ListPendingApprovalsUseCase
	approveUC              *usecases.ApproveRequestUseCase
	rejectUC               *usecases.RejectRequestUseCase
	historyUC              *usecases.GetApprovalHistoryUseCase
	getCurrentUserUC       *usecases.GetCurrentUserUseCase
}

func NewLeaveRequestHandler(
	submitUC *usecases.SubmitLeaveRequestUseCase,
	cancelUC *usecases.CancelLeaveRequestUseCase,
	listUC *usecases.ListLeaveRequestsUseCase,
	getUC *usecases.GetLeaveRequestUseCase,
	listPendingApprovalsUC *usecases.ListPendingApprovalsUseCase,
	approveUC *usecases.ApproveRequestUseCase,
	rejectUC *usecases.RejectRequestUseCase,
	historyUC *usecases.GetApprovalHistoryUseCase,
	getCurrentUserUC *usecases.GetCurrentUserUseCase,
) *LeaveRequestHandler {
	return &LeaveRequestHandler{
		submitUC:               submitUC,
		cancelUC:               cancelUC,
		listUC:                 listUC,
		getUC:                  getUC,
		listPendingApprovalsUC: listPendingApprovalsUC,
		approveUC:              approveUC,
		rejectUC:               rejectUC,
		historyUC:              historyUC,
		getCurrentUserUC:       getCurrentUserUC,
	}
}

// Request types

type SubmitLeaveRequestRequest struct {
	LeaveTypeUID string  `json:"leaveTypeUid"`
	StartDate    string  `json:"startDate"`
	EndDate      string  `json:"endDate"`
	Notes        *string `json:"notes,omitempty"`
}

type ApprovalActionRequest struct {
	Comments *string `json:"comments,omitempty"`
}

// Response types

type LeaveRequestResponse struct {
	UID             string  `json:"uid"`
	EmployeeUID     string  `json:"employeeUid"`
	EmployeeName    string  `json:"employeeName,omitempty"`
	LeaveTypeUID    string  `json:"leaveTypeUid"`
	LeaveTypeNameEN string  `json:"leaveTypeNameEn,omitempty"`
	LeaveTypeNameAR string  `json:"leaveTypeNameAr,omitempty"`
	StartDate       string  `json:"startDate"`
	EndDate         string  `json:"endDate"`
	Days            int     `json:"days"`
	Notes           *string `json:"notes,omitempty"`
	Status          string  `json:"status"`
	CurrentStep     int     `json:"currentStep"`
	MaxStep         int     `json:"maxStep"`
	SubmittedAt     string  `json:"submittedAt"`
	DecidedAt       *string `json:"decidedAt,omitempty"`
}

type ListLeaveRequestsResponse struct {
	Requests []LeaveRequestResponse `json:"requests"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

type PendingApprovalResponse struct {
	UID             string  `json:"uid"`
	RequesterUID    string  `json:"requesterUid"`
	RequesterName   string  `json:"requesterName"`
	LeaveTypeUID    string  `json:"leaveTypeUid"`
	LeaveTypeNameEN string  `json:"leaveTypeNameEn"`
	LeaveTypeNameAR string  `json:"leaveTypeNameAr,omitempty"`
	StartDate       string  `json:"startDate"`
	EndDate         string  `json:"endDate"`
	Days            int     `json:"days"`
	Notes           *string `json:"notes,omitempty"`
	CurrentStep     int     `json:"currentStep"`
	MaxStep         int     `json:"maxStep"`
	SubmittedAt     string  `json:"submittedAt"`
}

type ListPendingApprovalsResponse struct {
	Approvals []PendingApprovalResponse `json:"approvals"`
}

type ApprovalHistoryItemResponse struct {
	UID       string  `json:"uid"`
	Action    string  `json:"action"`
	StepOrder *int    `json:"stepOrder,omitempty"`
	ActorUID  string  `json:"actorUid"`
	ActorName string  `json:"actorName"`
	Comments  *string `json:"comments,omitempty"`
	ActedAt   string  `json:"actedAt"`
}

type GetApprovalHistoryResponse struct {
	History []ApprovalHistoryItemResponse `json:"history"`
}

// SubmitLeaveRequest handles POST /api/v1/leave-requests
func (h *LeaveRequestHandler) SubmitLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	// Get current user to find employee UID
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		return
	}

	var req SubmitLeaveRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("leave_request_handler.SubmitLeaveRequest.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.LeaveTypeUID == "" {
		writeError(w, http.StatusBadRequest, "leaveTypeUid is required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startDate format, expected YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid endDate format, expected YYYY-MM-DD")
		return
	}

	input := usecases.SubmitLeaveRequestInput{
		EmployeeUID:  *currentUser.EmployeeUID,
		LeaveTypeUID: req.LeaveTypeUID,
		StartDate:    startDate,
		EndDate:      endDate,
		Notes:        req.Notes,
	}

	output, err := h.submitUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrLeaveTypeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrInsufficientBalance):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidDateRange):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrExceedsConsecutiveDays):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNoDepartmentAssigned):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrOverlappingRequest):
			statusCode = http.StatusConflict
		default:
			slog.Error("leave_request_handler.SubmitLeaveRequest.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	// Handle auto-approved leaves (no approval flow)
	if output.LeaveRecord != nil {
		decidedAt := time.Now().Format("2006-01-02T15:04:05Z")
		writeJSON(w, http.StatusCreated, LeaveRequestResponse{
			UID:          output.LeaveRecord.UID,
			EmployeeUID:  *currentUser.EmployeeUID,
			LeaveTypeUID: req.LeaveTypeUID,
			StartDate:    output.LeaveRecord.StartDate.Format("2006-01-02"),
			EndDate:      output.LeaveRecord.EndDate.Format("2006-01-02"),
			Days:         output.LeaveRecord.Days,
			Notes:        output.LeaveRecord.Notes,
			Status:       "approved",
			CurrentStep:  1,
			MaxStep:      1,
			SubmittedAt:  time.Now().Format("2006-01-02T15:04:05Z"),
			DecidedAt:    &decidedAt,
		})
		return
	}

	// Normal approval flow
	var decidedAt *string
	if output.LeaveRequest.DecidedAt != nil {
		t := output.LeaveRequest.DecidedAt.Format("2006-01-02T15:04:05Z")
		decidedAt = &t
	}

	writeJSON(w, http.StatusCreated, LeaveRequestResponse{
		UID:          output.LeaveRequest.UID,
		EmployeeUID:  output.LeaveRequest.EmployeeUID,
		LeaveTypeUID: output.LeaveRequest.LeaveTypeUID,
		StartDate:    output.LeaveRequest.StartDate.Format("2006-01-02"),
		EndDate:      output.LeaveRequest.EndDate.Format("2006-01-02"),
		Days:         output.LeaveRequest.Days,
		Notes:        output.LeaveRequest.Notes,
		Status:       string(output.ApprovalRequest.Status),
		CurrentStep:  output.ApprovalRequest.CurrentStep,
		MaxStep:      output.ApprovalRequest.MaxStep,
		SubmittedAt:  output.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		DecidedAt:    decidedAt,
	})
}

// CancelLeaveRequest handles POST /api/v1/leave-requests/{uid}/cancel
func (h *LeaveRequestHandler) CancelLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	// Get current user to find employee UID
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	input := usecases.CancelLeaveRequestInput{
		LeaveRequestUID:  uid,
		ActorEmployeeUID: *currentUser.EmployeeUID,
	}

	err = h.cancelUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrLeaveRequestNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrRequestNotPending):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNotRequestOwner):
			statusCode = http.StatusForbidden
		case errors.Is(err, usecases.ErrCannotCancelApproved):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("leave_request_handler.CancelLeaveRequest.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Leave request cancelled successfully",
	})
}

// ListLeaveRequests handles GET /api/v1/leave-requests
func (h *LeaveRequestHandler) ListLeaveRequests(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	var err error

	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	input := usecases.ListLeaveRequestsInput{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	if !claims.HasPermission("*") {
		if claims.IsDepartmentScope() {
			input.ManagedDepartmentUIDs = append([]string(nil), claims.ManagedDepartmentUIDs...)
		} else if claims.IsSelfScope() {
			currentUser, currentUserErr := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
			if currentUserErr != nil || currentUser.EmployeeUID == nil {
				writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
				return
			}

			employeeUID := *currentUser.EmployeeUID
			input.EmployeeUID = &employeeUID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ApprovalRequestStatus(status)
		input.Status = &s
	}

	output, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("leave_request_handler.ListLeaveRequests.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	requests := make([]LeaveRequestResponse, 0, len(output.Requests))
	for _, req := range output.Requests {
		var decidedAt *string
		if req.LeaveRequest.DecidedAt != nil {
			t := req.LeaveRequest.DecidedAt.Format("2006-01-02T15:04:05Z")
			decidedAt = &t
		}

		resp := LeaveRequestResponse{
			UID:          req.LeaveRequest.UID,
			EmployeeUID:  req.LeaveRequest.EmployeeUID,
			LeaveTypeUID: req.LeaveRequest.LeaveTypeUID,
			StartDate:    req.LeaveRequest.StartDate.Format("2006-01-02"),
			EndDate:      req.LeaveRequest.EndDate.Format("2006-01-02"),
			Days:         req.LeaveRequest.Days,
			Notes:        req.LeaveRequest.Notes,
			Status:       string(req.ApprovalRequest.Status),
			CurrentStep:  req.ApprovalRequest.CurrentStep,
			MaxStep:      req.ApprovalRequest.MaxStep,
			SubmittedAt:  req.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
			DecidedAt:    decidedAt,
		}

		requests = append(requests, resp)
	}

	writeJSON(w, http.StatusOK, ListLeaveRequestsResponse{
		Requests: requests,
		Total:    output.Total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetLeaveRequest handles GET /api/v1/leave-requests/{uid}
func (h *LeaveRequestHandler) GetLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.getUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrLeaveRequestNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("leave_request_handler.GetLeaveRequest.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	var decidedAt *string
	if output.LeaveRequest.DecidedAt != nil {
		t := output.LeaveRequest.DecidedAt.Format("2006-01-02T15:04:05Z")
		decidedAt = &t
	}

	resp := LeaveRequestResponse{
		UID:          output.LeaveRequest.UID,
		EmployeeUID:  output.LeaveRequest.EmployeeUID,
		LeaveTypeUID: output.LeaveRequest.LeaveTypeUID,
		StartDate:    output.LeaveRequest.StartDate.Format("2006-01-02"),
		EndDate:      output.LeaveRequest.EndDate.Format("2006-01-02"),
		Days:         output.LeaveRequest.Days,
		Notes:        output.LeaveRequest.Notes,
		Status:       string(output.ApprovalRequest.Status),
		CurrentStep:  output.ApprovalRequest.CurrentStep,
		MaxStep:      output.ApprovalRequest.MaxStep,
		SubmittedAt:  output.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		DecidedAt:    decidedAt,
	}

	if output.Employee != nil {
		resp.EmployeeName = output.Employee.Name
	}
	if output.LeaveType != nil {
		resp.LeaveTypeNameEN = output.LeaveType.NameEN
		resp.LeaveTypeNameAR = output.LeaveType.NameAR
	}

	if !claims.HasPermission("*") {
		if claims.IsDepartmentScope() {
			if output.Employee == nil || output.Employee.DepartmentUID == nil || !claims.HasDepartmentAccess(*output.Employee.DepartmentUID) {
				writeJSONError(w, http.StatusForbidden, "permission_denied", "You can only view leave requests within your scope")
				return
			}
		} else if claims.IsSelfScope() {
			currentUser, currentUserErr := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
			if currentUserErr != nil || currentUser.EmployeeUID == nil || *currentUser.EmployeeUID != output.LeaveRequest.EmployeeUID {
				writeJSONError(w, http.StatusForbidden, "permission_denied", "You can only view leave requests within your scope")
				return
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListPendingApprovals handles GET /api/v1/approvals/pending
func (h *LeaveRequestHandler) ListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	output, err := h.listPendingApprovalsUC.Execute(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("leave_request_handler.ListPendingApprovals.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	approvals := make([]PendingApprovalResponse, 0, len(output.Items))
	for _, pa := range output.Items {
		resp := PendingApprovalResponse{
			UID:          pa.ApprovalRequest.UID,
			RequesterUID: pa.ApprovalRequest.RequesterUID,
			CurrentStep:  pa.ApprovalRequest.CurrentStep,
			MaxStep:      pa.ApprovalRequest.MaxStep,
			SubmittedAt:  pa.ApprovalRequest.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if pa.LeaveRequest != nil {
			resp.LeaveTypeUID = pa.LeaveRequest.LeaveTypeUID
			resp.StartDate = pa.LeaveRequest.StartDate.Format("2006-01-02")
			resp.EndDate = pa.LeaveRequest.EndDate.Format("2006-01-02")
			resp.Days = pa.LeaveRequest.Days
			resp.Notes = pa.LeaveRequest.Notes
		}

		if pa.Employee != nil {
			resp.RequesterName = pa.Employee.Name
		}

		if pa.LeaveType != nil {
			resp.LeaveTypeNameEN = pa.LeaveType.NameEN
			resp.LeaveTypeNameAR = pa.LeaveType.NameAR
		}

		approvals = append(approvals, resp)
	}

	writeJSON(w, http.StatusOK, ListPendingApprovalsResponse{Approvals: approvals})
}

// ApproveRequest handles POST /api/v1/approvals/{uid}/approve
func (h *LeaveRequestHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	// Get current user to find employee UID
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req ApprovalActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		slog.Error("leave_request_handler.ApproveRequest.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.ApproveRequestInput{
		ApprovalRequestUID: uid,
		ActorUserID:        claims.UserID,
		ActorEmployeeUID:   *currentUser.EmployeeUID,
		Comments:           req.Comments,
	}

	output, err := h.approveUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrApprovalRequestNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrRequestNotPending):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNotAuthorizedApprover):
			statusCode = http.StatusForbidden
		case errors.Is(err, usecases.ErrInsufficientBalance):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("leave_request_handler.ApproveRequest.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"uid":                output.ApprovalRequest.UID,
		"status":             string(output.ApprovalRequest.Status),
		"currentStep":        output.ApprovalRequest.CurrentStep,
		"leaveRecordCreated": output.LeaveRecord != nil,
	})
}

// RejectRequest handles POST /api/v1/approvals/{uid}/reject
func (h *LeaveRequestHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	// Get current user to find employee UID
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		return
	}

	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req ApprovalActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		slog.Error("leave_request_handler.RejectRequest.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.RejectRequestInput{
		ApprovalRequestUID: uid,
		ActorUserID:        claims.UserID,
		ActorEmployeeUID:   *currentUser.EmployeeUID,
		Comments:           req.Comments,
	}

	output, err := h.rejectUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrApprovalRequestNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrRequestNotPending):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNotAuthorizedApprover):
			statusCode = http.StatusForbidden
		default:
			slog.Error("leave_request_handler.RejectRequest.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"uid":    output.ApprovalRequest.UID,
		"status": string(output.ApprovalRequest.Status),
	})
}

// GetApprovalHistory handles GET /api/v1/approvals/{uid}/history
func (h *LeaveRequestHandler) GetApprovalHistory(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.historyUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrApprovalRequestNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("leave_request_handler.GetApprovalHistory.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	history := make([]ApprovalHistoryItemResponse, 0, len(output.History))
	for _, item := range output.History {
		history = append(history, ApprovalHistoryItemResponse{
			UID:       item.Action.UID,
			Action:    string(item.Action.Action),
			StepOrder: item.Action.StepOrder,
			ActorUID:  item.ActorUID,
			ActorName: item.ActorName,
			Comments:  item.Action.Comments,
			ActedAt:   item.Action.ActedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, GetApprovalHistoryResponse{
		History: history,
	})
}
