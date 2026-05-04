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
	updateRejectedUC       *usecases.UpdateRejectedLeaveRequestUseCase
	cancelUC               *usecases.CancelLeaveRequestUseCase
	listUC                 *usecases.ListLeaveRequestsUseCase
	getUC                  *usecases.GetLeaveRequestUseCase
	getEmployeeUC          *usecases.GetEmployeeUseCase
	listPendingApprovalsUC *usecases.ListPendingApprovalsUseCase
	approveUC              *usecases.ApproveRequestUseCase
	rejectUC               *usecases.RejectRequestUseCase
	historyUC              *usecases.GetApprovalHistoryUseCase
	getCurrentUserUC       *usecases.GetCurrentUserUseCase
}

func NewLeaveRequestHandler(
	submitUC *usecases.SubmitLeaveRequestUseCase,
	updateRejectedUC *usecases.UpdateRejectedLeaveRequestUseCase,
	cancelUC *usecases.CancelLeaveRequestUseCase,
	listUC *usecases.ListLeaveRequestsUseCase,
	getUC *usecases.GetLeaveRequestUseCase,
	getEmployeeUC *usecases.GetEmployeeUseCase,
	listPendingApprovalsUC *usecases.ListPendingApprovalsUseCase,
	approveUC *usecases.ApproveRequestUseCase,
	rejectUC *usecases.RejectRequestUseCase,
	historyUC *usecases.GetApprovalHistoryUseCase,
	getCurrentUserUC *usecases.GetCurrentUserUseCase,
) *LeaveRequestHandler {
	return &LeaveRequestHandler{
		submitUC:               submitUC,
		updateRejectedUC:       updateRejectedUC,
		cancelUC:               cancelUC,
		listUC:                 listUC,
		getUC:                  getUC,
		getEmployeeUC:          getEmployeeUC,
		listPendingApprovalsUC: listPendingApprovalsUC,
		approveUC:              approveUC,
		rejectUC:               rejectUC,
		historyUC:              historyUC,
		getCurrentUserUC:       getCurrentUserUC,
	}
}

// Request types

type SubmitLeaveRequestRequest struct {
	EmployeeUID         *string                          `json:"employeeUid,omitempty"`
	LeaveTypeUID        string                           `json:"leaveTypeUid"`
	SubLeaveTypeUID     *string                          `json:"subLeaveTypeUid,omitempty"`
	SubLeaveTypeUIDV2   *string                          `json:"sub_leave_type_uid,omitempty"`
	OtherSubLeaveName   *string                          `json:"otherSubLeaveName,omitempty"`
	OtherSubLeaveNameV2 *string                          `json:"other_sub_leave_name,omitempty"`
	StartDate           string                           `json:"startDate"`
	EndDate             string                           `json:"endDate"`
	Notes               *string                          `json:"notes,omitempty"`
	StudyDestination    *string                          `json:"studyDestination,omitempty"`
	StudyDestinationV2  *string                          `json:"study_destination,omitempty"`
	Assignment          *string                          `json:"assignment,omitempty"`
	AssignmentCountry   *string                          `json:"assignmentCountry,omitempty"`
	AssignmentCountryV2 *string                          `json:"assignment_country,omitempty"`
	SpouseWorkCountry   *string                          `json:"spouseWorkCountry,omitempty"`
	SpouseWorkCountryV2 *string                          `json:"spouse_work_country,omitempty"`
	DocumentsAttached   bool                             `json:"documents_attached"`
	Documents           []SubmitLeaveRequestDocumentItem `json:"documents,omitempty"`
}

type SubmitLeaveRequestDocumentItem struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type UpdateLeaveRequestRequest struct {
	LeaveTypeUID        *string                          `json:"leaveTypeUid,omitempty"`
	SubLeaveTypeUID     *string                          `json:"subLeaveTypeUid,omitempty"`
	SubLeaveTypeUIDV2   *string                          `json:"sub_leave_type_uid,omitempty"`
	StartDate           *string                          `json:"startDate,omitempty"`
	EndDate             *string                          `json:"endDate,omitempty"`
	Notes               *string                          `json:"notes,omitempty"`
	StudyDestination    *string                          `json:"studyDestination,omitempty"`
	StudyDestinationV2  *string                          `json:"study_destination,omitempty"`
	Assignment          *string                          `json:"assignment,omitempty"`
	AssignmentCountry   *string                          `json:"assignmentCountry,omitempty"`
	AssignmentCountryV2 *string                          `json:"assignment_country,omitempty"`
	SpouseWorkCountry   *string                          `json:"spouseWorkCountry,omitempty"`
	SpouseWorkCountryV2 *string                          `json:"spouse_work_country,omitempty"`
	Documents           []SubmitLeaveRequestDocumentItem `json:"documents,omitempty"`
}

type ApprovalActionRequest struct {
	Comments *string `json:"comments,omitempty"`
}

// Response types

type LeaveRequestResponse struct {
	UID                string                               `json:"uid"`
	ApprovalRequestUID string                               `json:"approvalRequestUid"`
	EmployeeUID        string                               `json:"employeeUid"`
	EmployeeName       string                               `json:"employeeName,omitempty"`
	LeaveTypeUID       string                               `json:"leaveTypeUid"`
	LeaveTypeNameEN    string                               `json:"leaveTypeNameEn,omitempty"`
	LeaveTypeNameAR    string                               `json:"leaveTypeNameAr,omitempty"`
	SubLeaveTypeUID    *string                              `json:"subLeaveTypeUid,omitempty"`
	OtherSubLeaveName  *string                              `json:"otherSubLeaveName,omitempty"`
	SubLeaveTypeNameEN string                               `json:"subLeaveTypeNameEn,omitempty"`
	SubLeaveTypeNameAR string                               `json:"subLeaveTypeNameAr,omitempty"`
	StartDate          string                               `json:"startDate"`
	EndDate            string                               `json:"endDate"`
	Days               int                                  `json:"days"`
	Notes              *string                              `json:"notes,omitempty"`
	StudyDestination   *string                              `json:"studyDestination,omitempty"`
	Assignment         *string                              `json:"assignment,omitempty"`
	AssignmentCountry  *string                              `json:"assignmentCountry,omitempty"`
	SpouseWorkCountry  *string                              `json:"spouseWorkCountry,omitempty"`
	Status             string                               `json:"status"`
	CurrentStep        int                                  `json:"currentStep"`
	MaxStep            int                                  `json:"maxStep"`
	SubmittedAt        string                               `json:"submittedAt"`
	DecidedAt          *string                              `json:"decidedAt,omitempty"`
	Documents          []LeaveRequestDocumentUploadResponse `json:"documents,omitempty"`
}

type LeaveRequestDocumentUploadResponse struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"objectKey"`
}

type ListLeaveRequestsResponse struct {
	Requests []LeaveRequestResponse `json:"requests"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

type PendingApprovalResponse struct {
	UID             string  `json:"uid"`
	LeaveRequestUID string  `json:"leaveRequestUid"`
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

	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil {
		slog.Error("leave_request_handler.SubmitLeaveRequest.get_current_user", "error", err, "user_id", claims.UserID)
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	var employeeUID string
	switch {
	case req.EmployeeUID != nil && *req.EmployeeUID != "":
		employeeUID = *req.EmployeeUID

		if currentUser.EmployeeUID != nil && *currentUser.EmployeeUID == employeeUID {
			break
		}

		if claims.HasPermission("*") || claims.IsGlobalScope() {
			break
		}

		if claims.IsDepartmentScope() {
			employee, getEmployeeErr := h.getEmployeeUC.Execute(r.Context(), usecases.GetEmployeeInput{UID: employeeUID})
			if getEmployeeErr != nil {
				switch {
				case errors.Is(getEmployeeErr, usecases.ErrEmployeeNotFound):
					writeJSONError(w, http.StatusBadRequest, "employee_not_found", "Employee not found")
				default:
					slog.Error("leave_request_handler.SubmitLeaveRequest.get_target_employee", "error", getEmployeeErr, "employee_uid", employeeUID)
					writeJSONError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
				}
				return
			}
			if employee.DepartmentUID == nil || !claims.HasDepartmentAccess(*employee.DepartmentUID) {
				writeJSONError(w, http.StatusForbidden, "permission_denied", "You can only submit leave requests within your scope")
				return
			}
			break
		}

		writeJSONError(w, http.StatusForbidden, "permission_denied", "You can only submit leave requests within your scope")
		return
	case currentUser.EmployeeUID == nil:
		writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
		return
	default:
		employeeUID = *currentUser.EmployeeUID
	}

	input := usecases.SubmitLeaveRequestInput{
		UserUID:           claims.UserUID,
		EmployeeUID:       employeeUID,
		ActorEmployeeUID:  currentUser.EmployeeUID,
		LeaveTypeUID:      req.LeaveTypeUID,
		SubLeaveTypeUID:   firstNonNilString(req.SubLeaveTypeUID, req.SubLeaveTypeUIDV2),
		OtherSubLeaveName: firstNonNilString(req.OtherSubLeaveName, req.OtherSubLeaveNameV2),
		StartDate:         startDate,
		EndDate:           endDate,
		Notes:             req.Notes,
		StudyDestination:  firstNonNilString(req.StudyDestination, req.StudyDestinationV2),
		Assignment:        req.Assignment,
		AssignmentCountry: firstNonNilString(req.AssignmentCountry, req.AssignmentCountryV2),
		SpouseWorkCountry: firstNonNilString(req.SpouseWorkCountry, req.SpouseWorkCountryV2),
		DocumentsAttached: req.DocumentsAttached,
		Documents:         make([]usecases.SubmitLeaveRequestDocumentInput, 0, len(req.Documents)),
	}
	for _, document := range req.Documents {
		input.Documents = append(input.Documents, usecases.SubmitLeaveRequestDocumentInput{
			FileName:    document.FileName,
			ContentType: document.ContentType,
		})
	}

	output, err := h.submitUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrLeaveTypeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrSubLeaveTypeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrSubLeaveTypeDoesNotBelongToLeaveType):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInsufficientBalance):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidDateRange):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrLeaveRequestOutsideDeadline):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrExceedsConsecutiveDays):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNoDepartmentAssigned):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrOverlappingRequest):
			statusCode = http.StatusConflict
		case errors.Is(err, usecases.ErrLeaveRequestDocumentsRequired):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrLeaveRequestDocumentsUnsupported):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidFilename):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidContentType):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("leave_request_handler.SubmitLeaveRequest.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	// Handle auto-approved leaves (no approval flow)
	if output.LeaveRecord != nil && output.LeaveRequest != nil && output.ApprovalRequest != nil {
		var decidedAt *string
		if output.LeaveRequest.DecidedAt != nil {
			t := output.LeaveRequest.DecidedAt.Format("2006-01-02T15:04:05Z")
			decidedAt = &t
		}
		writeJSON(w, http.StatusCreated, LeaveRequestResponse{
			UID:                output.LeaveRequest.UID,
			ApprovalRequestUID: output.ApprovalRequest.UID,
			EmployeeUID:        output.LeaveRequest.EmployeeUID,
			LeaveTypeUID:       output.LeaveRequest.LeaveTypeUID,
			SubLeaveTypeUID:    output.LeaveRequest.SubLeaveTypeUID,
			OtherSubLeaveName:  output.LeaveRequest.OtherSubLeaveName,
			StartDate:          output.LeaveRequest.StartDate.Format("2006-01-02"),
			EndDate:            output.LeaveRequest.EndDate.Format("2006-01-02"),
			Days:               output.LeaveRequest.Days,
			Notes:              output.LeaveRequest.Notes,
			StudyDestination:   output.LeaveRequest.StudyDestination,
			Assignment:         output.LeaveRequest.Assignment,
			AssignmentCountry:  output.LeaveRequest.AssignmentCountry,
			SpouseWorkCountry:  output.LeaveRequest.SpouseWorkCountry,
			Status:             string(output.ApprovalRequest.Status),
			CurrentStep:        output.ApprovalRequest.CurrentStep,
			MaxStep:            output.ApprovalRequest.MaxStep,
			SubmittedAt:        output.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
			DecidedAt:          decidedAt,
			Documents:          toLeaveRequestDocumentResponses(output.Documents),
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
		UID:               output.LeaveRequest.UID,
		EmployeeUID:       output.LeaveRequest.EmployeeUID,
		LeaveTypeUID:      output.LeaveRequest.LeaveTypeUID,
		SubLeaveTypeUID:   output.LeaveRequest.SubLeaveTypeUID,
		OtherSubLeaveName: output.LeaveRequest.OtherSubLeaveName,
		StartDate:         output.LeaveRequest.StartDate.Format("2006-01-02"),
		EndDate:           output.LeaveRequest.EndDate.Format("2006-01-02"),
		Days:              output.LeaveRequest.Days,
		Notes:             output.LeaveRequest.Notes,
		StudyDestination:  output.LeaveRequest.StudyDestination,
		Assignment:        output.LeaveRequest.Assignment,
		AssignmentCountry: output.LeaveRequest.AssignmentCountry,
		SpouseWorkCountry: output.LeaveRequest.SpouseWorkCountry,
		Status:            string(output.ApprovalRequest.Status),
		CurrentStep:       output.ApprovalRequest.CurrentStep,
		MaxStep:           output.ApprovalRequest.MaxStep,
		SubmittedAt:       output.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		DecidedAt:         decidedAt,
		Documents:         toLeaveRequestDocumentResponses(output.Documents),
	})
}

func toLeaveRequestDocumentResponses(documents []usecases.SubmitLeaveRequestDocumentOutput) []LeaveRequestDocumentUploadResponse {
	if len(documents) == 0 {
		return nil
	}

	response := make([]LeaveRequestDocumentUploadResponse, 0, len(documents))
	for _, document := range documents {
		response = append(response, LeaveRequestDocumentUploadResponse{
			FileName:  document.FileName,
			URL:       document.URL,
			Method:    document.Method,
			Bucket:    document.Bucket,
			ObjectKey: document.ObjectKey,
		})
	}

	return response
}

func firstNonNilString(values ...*string) *string {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
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

// UpdateRejectedLeaveRequest handles PATCH /api/v1/leave-requests/{uid}
func (h *LeaveRequestHandler) UpdateRejectedLeaveRequest(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

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

	var req UpdateLeaveRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("leave_request_handler.UpdateRejectedLeaveRequest.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var startDate *time.Time
	if req.StartDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid startDate format, expected YYYY-MM-DD")
			return
		}
		startDate = &parsed
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid endDate format, expected YYYY-MM-DD")
			return
		}
		endDate = &parsed
	}

	input := usecases.UpdateRejectedLeaveRequestInput{
		LeaveRequestUID:   uid,
		ActorUserUID:      claims.UserUID,
		ActorEmployeeUID:  *currentUser.EmployeeUID,
		LeaveTypeUID:      req.LeaveTypeUID,
		SubLeaveTypeUID:   firstNonNilString(req.SubLeaveTypeUID, req.SubLeaveTypeUIDV2),
		StartDate:         startDate,
		EndDate:           endDate,
		Notes:             req.Notes,
		StudyDestination:  firstNonNilString(req.StudyDestination, req.StudyDestinationV2),
		Assignment:        req.Assignment,
		AssignmentCountry: firstNonNilString(req.AssignmentCountry, req.AssignmentCountryV2),
		SpouseWorkCountry: firstNonNilString(req.SpouseWorkCountry, req.SpouseWorkCountryV2),
		Documents:         make([]usecases.SubmitLeaveRequestDocumentInput, 0, len(req.Documents)),
	}
	for _, document := range req.Documents {
		input.Documents = append(input.Documents, usecases.SubmitLeaveRequestDocumentInput{
			FileName:    document.FileName,
			ContentType: document.ContentType,
		})
	}

	output, err := h.updateRejectedUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrLeaveRequestNotFound), errors.Is(err, usecases.ErrApprovalRequestNotFound), errors.Is(err, usecases.ErrLeaveRequestDocumentNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrNotRequestOwner):
			statusCode = http.StatusForbidden
		case errors.Is(err, usecases.ErrRequestNotRejected), errors.Is(err, usecases.ErrLeaveTypeNoApprovalRequired), errors.Is(err, usecases.ErrApprovalFlowMismatch):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrInvalidDateRange), errors.Is(err, usecases.ErrNoWorkingDays), errors.Is(err, usecases.ErrOverlappingRequest), errors.Is(err, usecases.ErrNoDepartmentAssigned), errors.Is(err, usecases.ErrSubLeaveTypeNotFound), errors.Is(err, usecases.ErrSubLeaveTypeDoesNotBelongToLeaveType), errors.Is(err, usecases.ErrInvalidFilename), errors.Is(err, usecases.ErrInvalidContentType):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("leave_request_handler.UpdateRejectedLeaveRequest.execute_usecase", "error", err)
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
		UID:                output.LeaveRequest.UID,
		ApprovalRequestUID: output.LeaveRequest.ApprovalRequestUID,
		EmployeeUID:        output.LeaveRequest.EmployeeUID,
		LeaveTypeUID:       output.LeaveRequest.LeaveTypeUID,
		SubLeaveTypeUID:    output.LeaveRequest.SubLeaveTypeUID,
		OtherSubLeaveName:  output.LeaveRequest.OtherSubLeaveName,
		StartDate:          output.LeaveRequest.StartDate.Format("2006-01-02"),
		EndDate:            output.LeaveRequest.EndDate.Format("2006-01-02"),
		Days:               output.LeaveRequest.Days,
		Notes:              output.LeaveRequest.Notes,
		StudyDestination:   output.LeaveRequest.StudyDestination,
		Assignment:         output.LeaveRequest.Assignment,
		AssignmentCountry:  output.LeaveRequest.AssignmentCountry,
		SpouseWorkCountry:  output.LeaveRequest.SpouseWorkCountry,
		Status:             string(output.ApprovalRequest.Status),
		CurrentStep:        output.ApprovalRequest.CurrentStep,
		MaxStep:            output.ApprovalRequest.MaxStep,
		SubmittedAt:        output.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		DecidedAt:          decidedAt,
		Documents:          toLeaveRequestDocumentResponses(output.Documents),
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListLeaveRequests handles GET /api/v1/leave-requests
func (h *LeaveRequestHandler) ListLeaveRequests(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

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

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ApprovalRequestStatus(status)
		input.Status = &s
	}

	requestedEmployeeUID := r.URL.Query().Get("employeeUid")
	switch {
	case requestedEmployeeUID != "":
		if claims.HasPermission("*") || claims.IsGlobalScope() {
			input.EmployeeUID = &requestedEmployeeUID
			break
		}

		if claims.IsDepartmentScope() {
			employee, err := h.getEmployeeUC.Execute(r.Context(), usecases.GetEmployeeInput{UID: requestedEmployeeUID})
			if err != nil {
				switch {
				case errors.Is(err, usecases.ErrEmployeeNotFound):
					writeJSONError(w, http.StatusBadRequest, "employee_not_found", "Employee not found")
				default:
					slog.Error("leave_request_handler.ListLeaveRequests.get_target_employee", "error", err, "employee_uid", requestedEmployeeUID)
					writeJSONError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
				}
				return
			}
			if employee.DepartmentUID == nil || !claims.HasDepartmentAccess(*employee.DepartmentUID) {
				writeJSONError(w, http.StatusForbidden, "permission_denied", "You can only view leave requests within your scope")
				return
			}
			input.EmployeeUID = &requestedEmployeeUID
			break
		}

		currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
		if err != nil || currentUser.EmployeeUID == nil {
			writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
			return
		}
		if *currentUser.EmployeeUID != requestedEmployeeUID {
			writeJSONError(w, http.StatusForbidden, "permission_denied", "You can only view leave requests within your scope")
			return
		}
		input.EmployeeUID = &requestedEmployeeUID
	case claims.HasPermission("*") || claims.IsGlobalScope():
		// No additional filter.
	default:
		currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
		if err != nil || currentUser.EmployeeUID == nil {
			writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
			return
		}
		employeeUID := *currentUser.EmployeeUID
		input.EmployeeUID = &employeeUID
	}

	output, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("leave_request_handler.ListLeaveRequests.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, buildListLeaveRequestsResponse(output, page, pageSize))
}

// ListDepartmentLeaveRequests handles GET /api/v1/departments/{departmentUid}/leave-requests
func (h *LeaveRequestHandler) ListDepartmentLeaveRequests(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	departmentUID := r.PathValue("departmentUid")
	if departmentUID == "" {
		writeError(w, http.StatusBadRequest, "departmentUid is required")
		return
	}
	if !canAccessDepartment(claims, departmentUID) {
		writeJSONError(w, http.StatusForbidden, "permission_denied", "Access to this department is not permitted")
		return
	}

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
		ManagedDepartmentUIDs: []string{departmentUID},
		Limit:                 pageSize,
		Offset:                (page - 1) * pageSize,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ApprovalRequestStatus(status)
		input.Status = &s
	}

	output, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		slog.Error("leave_request_handler.ListDepartmentLeaveRequests.execute_usecase", "error", err, "department_uid", departmentUID)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, buildListLeaveRequestsResponse(output, page, pageSize))
}

func buildListLeaveRequestsResponse(output *usecases.ListLeaveRequestsOutput, page, pageSize int) ListLeaveRequestsResponse {
	requests := make([]LeaveRequestResponse, 0, len(output.Requests))
	for _, req := range output.Requests {
		var decidedAt *string
		if req.LeaveRequest.DecidedAt != nil {
			t := req.LeaveRequest.DecidedAt.Format("2006-01-02T15:04:05Z")
			decidedAt = &t
		}

		resp := LeaveRequestResponse{
			UID:                req.LeaveRequest.UID,
			ApprovalRequestUID: req.LeaveRequest.ApprovalRequestUID,
			EmployeeUID:        req.LeaveRequest.EmployeeUID,
			LeaveTypeUID:       req.LeaveRequest.LeaveTypeUID,
			SubLeaveTypeUID:    req.LeaveRequest.SubLeaveTypeUID,
			OtherSubLeaveName:  req.LeaveRequest.OtherSubLeaveName,
			StartDate:          req.LeaveRequest.StartDate.Format("2006-01-02"),
			EndDate:            req.LeaveRequest.EndDate.Format("2006-01-02"),
			Days:               req.LeaveRequest.Days,
			Notes:              req.LeaveRequest.Notes,
			StudyDestination:   req.LeaveRequest.StudyDestination,
			Assignment:         req.LeaveRequest.Assignment,
			AssignmentCountry:  req.LeaveRequest.AssignmentCountry,
			SpouseWorkCountry:  req.LeaveRequest.SpouseWorkCountry,
			Status:             string(req.ApprovalRequest.Status),
			CurrentStep:        req.ApprovalRequest.CurrentStep,
			MaxStep:            req.ApprovalRequest.MaxStep,
			SubmittedAt:        req.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
			DecidedAt:          decidedAt,
		}
		if req.LeaveType != nil {
			resp.LeaveTypeNameEN = req.LeaveType.NameEN
			resp.LeaveTypeNameAR = req.LeaveType.NameAR
		}
		if req.SubLeaveType != nil {
			resp.SubLeaveTypeNameEN = req.SubLeaveType.NameEN
			resp.SubLeaveTypeNameAR = req.SubLeaveType.NameAR
		}

		requests = append(requests, resp)
	}

	return ListLeaveRequestsResponse{
		Requests: requests,
		Total:    output.Total,
		Page:     page,
		PageSize: pageSize,
	}
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
		UID:                output.LeaveRequest.UID,
		ApprovalRequestUID: output.LeaveRequest.ApprovalRequestUID,
		EmployeeUID:        output.LeaveRequest.EmployeeUID,
		LeaveTypeUID:       output.LeaveRequest.LeaveTypeUID,
		SubLeaveTypeUID:    output.LeaveRequest.SubLeaveTypeUID,
		OtherSubLeaveName:  output.LeaveRequest.OtherSubLeaveName,
		StartDate:          output.LeaveRequest.StartDate.Format("2006-01-02"),
		EndDate:            output.LeaveRequest.EndDate.Format("2006-01-02"),
		Days:               output.LeaveRequest.Days,
		Notes:              output.LeaveRequest.Notes,
		StudyDestination:   output.LeaveRequest.StudyDestination,
		Assignment:         output.LeaveRequest.Assignment,
		AssignmentCountry:  output.LeaveRequest.AssignmentCountry,
		SpouseWorkCountry:  output.LeaveRequest.SpouseWorkCountry,
		Status:             string(output.ApprovalRequest.Status),
		CurrentStep:        output.ApprovalRequest.CurrentStep,
		MaxStep:            output.ApprovalRequest.MaxStep,
		SubmittedAt:        output.LeaveRequest.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		DecidedAt:          decidedAt,
		Documents:          toLeaveRequestDocumentResponses(output.Documents),
	}

	if output.Employee != nil {
		resp.EmployeeName = output.Employee.Name
	}
	if output.LeaveType != nil {
		resp.LeaveTypeNameEN = output.LeaveType.NameEN
		resp.LeaveTypeNameAR = output.LeaveType.NameAR
	}
	if output.SubLeaveType != nil {
		resp.SubLeaveTypeNameEN = output.SubLeaveType.NameEN
		resp.SubLeaveTypeNameAR = output.SubLeaveType.NameAR
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
			resp.LeaveRequestUID = pa.LeaveRequest.UID
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
