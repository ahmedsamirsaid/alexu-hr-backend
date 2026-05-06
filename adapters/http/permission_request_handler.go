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

// PermissionRequestHandler handles excuse/permission request HTTP requests.
type PermissionRequestHandler struct {
	submitUC         *usecases.SubmitPermissionRequestUseCase
	updateUC         *usecases.UpdatePermissionRequestUseCase
	cancelUC         *usecases.CancelPermissionRequestUseCase
	listUC           *usecases.ListPermissionRequestsUseCase
	getUC            *usecases.GetPermissionRequestUseCase
	listPendingUC    *usecases.ListPendingPermissionApprovalsUseCase
	approveUC        *usecases.ApprovePermissionRequestUseCase
	rejectUC         *usecases.RejectPermissionRequestUseCase
	historyUC        *usecases.GetApprovalHistoryUseCase
	eligibilityUC    *usecases.GetPermissionEligibilityUseCase
	getCurrentUserUC *usecases.GetCurrentUserUseCase
}

func NewPermissionRequestHandler(
	submitUC *usecases.SubmitPermissionRequestUseCase,
	updateUC *usecases.UpdatePermissionRequestUseCase,
	cancelUC *usecases.CancelPermissionRequestUseCase,
	listUC *usecases.ListPermissionRequestsUseCase,
	getUC *usecases.GetPermissionRequestUseCase,
	listPendingUC *usecases.ListPendingPermissionApprovalsUseCase,
	approveUC *usecases.ApprovePermissionRequestUseCase,
	rejectUC *usecases.RejectPermissionRequestUseCase,
	historyUC *usecases.GetApprovalHistoryUseCase,
	eligibilityUC *usecases.GetPermissionEligibilityUseCase,
	getCurrentUserUC *usecases.GetCurrentUserUseCase,
) *PermissionRequestHandler {
	return &PermissionRequestHandler{
		submitUC:         submitUC,
		updateUC:         updateUC,
		cancelUC:         cancelUC,
		listUC:           listUC,
		getUC:            getUC,
		listPendingUC:    listPendingUC,
		approveUC:        approveUC,
		rejectUC:         rejectUC,
		historyUC:        historyUC,
		eligibilityUC:    eligibilityUC,
		getCurrentUserUC: getCurrentUserUC,
	}
}

// ---------- DTOs ----------

type SubmitPermissionRequestRequest struct {
	Type           string  `json:"type"`
	PermissionDate string  `json:"permissionDate"`
	StartTime      *string `json:"startTime,omitempty"`
	EndTime        *string `json:"endTime,omitempty"`
	Reason         *string `json:"reason,omitempty"`
}

type UpdatePermissionRequestRequest = SubmitPermissionRequestRequest

type PermissionRequestResponse struct {
	UID                string  `json:"uid"`
	ApprovalRequestUID string  `json:"approvalRequestUid"`
	EmployeeUID        string  `json:"employeeUid"`
	EmployeeName       string  `json:"employeeName,omitempty"`
	Type               string  `json:"type"`
	TypeNameAr         string  `json:"typeNameAr"`
	TypeNameEn         string  `json:"typeNameEn"`
	PermissionDate     string  `json:"permissionDate"`
	StartTime          string  `json:"startTime"`
	EndTime            string  `json:"endTime"`
	Reason             *string `json:"reason,omitempty"`
	Status             string  `json:"status"`
	CurrentStep        int     `json:"currentStep"`
	MaxStep            int     `json:"maxStep"`
	SubmittedAt        string  `json:"submittedAt"`
	DecidedAt          *string `json:"decidedAt,omitempty"`
	// ActorCanApprove is true when the *calling* user is the request's assigned approver
	// (single-step model). Used by the UI to decide whether to render an Approve button.
	ActorCanApprove bool `json:"actorCanApprove"`
}

type ListPermissionRequestsResponse struct {
	Requests []PermissionRequestResponse `json:"requests"`
	Total    int                         `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"pageSize"`
}

type PendingPermissionApprovalResponse struct {
	UID                  string  `json:"uid"`                  // approvalRequest UID
	PermissionRequestUID string  `json:"permissionRequestUid"` // permissionRequest UID
	RequesterUID         string  `json:"requesterUid"`
	RequesterName        string  `json:"requesterName"`
	Type                 string  `json:"type"`
	TypeNameAr           string  `json:"typeNameAr"`
	TypeNameEn           string  `json:"typeNameEn"`
	PermissionDate       string  `json:"permissionDate"`
	StartTime            string  `json:"startTime"`
	EndTime              string  `json:"endTime"`
	Reason               *string `json:"reason,omitempty"`
	CurrentStep          int     `json:"currentStep"`
	MaxStep              int     `json:"maxStep"`
	SubmittedAt          string  `json:"submittedAt"`
}

type ListPendingPermissionApprovalsResponse struct {
	Approvals []PendingPermissionApprovalResponse `json:"approvals"`
}

// ---------- Helpers ----------

func toPermissionRequestResponse(req *domain.PermissionRequest, approval *domain.ApprovalRequest, employee *domain.Employee) PermissionRequestResponse {
	resp := PermissionRequestResponse{
		UID:                req.UID,
		ApprovalRequestUID: req.ApprovalRequestUID,
		EmployeeUID:        req.EmployeeUID,
		Type:               string(req.Type),
		TypeNameAr:         req.Type.NameAR(),
		TypeNameEn:         req.Type.NameEN(),
		PermissionDate:     req.PermissionDate.Format("2006-01-02"),
		StartTime:          req.StartTime,
		EndTime:            req.EndTime,
		Reason:             req.Reason,
		SubmittedAt:        req.SubmittedAt.Format(time.RFC3339),
	}
	if employee != nil {
		resp.EmployeeName = employee.Name
	}
	if req.DecidedAt != nil {
		s := req.DecidedAt.Format(time.RFC3339)
		resp.DecidedAt = &s
	}
	if approval != nil {
		resp.Status = string(approval.Status)
		resp.CurrentStep = approval.CurrentStep
		resp.MaxStep = approval.MaxStep
	}
	return resp
}

// ---------- Handlers ----------

// Submit handles POST /api/v1/permission-requests
func (h *PermissionRequestHandler) Submit(w http.ResponseWriter, r *http.Request) {
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

	var req SubmitPermissionRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	permissionDate, err := time.Parse("2006-01-02", req.PermissionDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid permissionDate format, expected YYYY-MM-DD")
		return
	}

	out, err := h.submitUC.Execute(r.Context(), usecases.SubmitPermissionRequestInput{
		UserUID:        claims.UserUID,
		EmployeeUID:    *currentUser.EmployeeUID,
		Type:           domain.PermissionType(req.Type),
		PermissionDate: permissionDate,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Reason:         req.Reason,
	})
	if err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Submit")
		return
	}

	writeJSON(w, http.StatusCreated, toPermissionRequestResponse(out.PermissionRequest, out.ApprovalRequest, nil))
}

// Update handles PATCH /api/v1/permission-requests/{uid}
func (h *PermissionRequestHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req SubmitPermissionRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	permissionDate, err := time.Parse("2006-01-02", req.PermissionDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid permissionDate format, expected YYYY-MM-DD")
		return
	}

	updated, err := h.updateUC.Execute(r.Context(), usecases.UpdatePermissionRequestInput{
		PermissionRequestUID: uid,
		ActorEmployeeUID:     *currentUser.EmployeeUID,
		Type:                 domain.PermissionType(req.Type),
		PermissionDate:       permissionDate,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Reason:               req.Reason,
	})
	if err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Update")
		return
	}
	writeJSON(w, http.StatusOK, toPermissionRequestResponse(updated, nil, nil))
}

// Cancel handles POST /api/v1/permission-requests/{uid}/cancel
func (h *PermissionRequestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
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
	if err := h.cancelUC.Execute(r.Context(), usecases.CancelPermissionRequestInput{
		PermissionRequestUID: uid,
		ActorEmployeeUID:     *currentUser.EmployeeUID,
	}); err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Cancel")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "permission cancelled"})
}

// Get handles GET /api/v1/permission-requests/{uid}
func (h *PermissionRequestHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.getUC.Execute(r.Context(), usecases.GetPermissionRequestInput{
		UID:         uid,
		ActorUserID: claims.UserID,
	})
	if err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Get")
		return
	}
	resp := toPermissionRequestResponse(out.PermissionRequest, out.ApprovalRequest, out.Employee)
	resp.ActorCanApprove = out.ActorCanApprove
	writeJSON(w, http.StatusOK, resp)
}

// List handles GET /api/v1/permission-requests
func (h *PermissionRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load current user")
		return
	}

	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))

	input := usecases.ListPermissionRequestsInput{Page: page, PageSize: pageSize}
	if v := q.Get("status"); v != "" {
		s := domain.ApprovalRequestStatus(v)
		input.Status = &s
	}
	if v := q.Get("type"); v != "" {
		t := domain.PermissionType(v)
		input.Type = &t
	}

	scope := q.Get("scope") // "self" | "all"
	switch scope {
	case "all":
		// HR / admin — no additional filter; permission middleware already gated this.
	default:
		// Default: requester sees their own.
		if currentUser.EmployeeUID == nil {
			writeJSONError(w, http.StatusBadRequest, "no_employee_linked", "No employee profile linked to user")
			return
		}
		input.EmployeeUID = currentUser.EmployeeUID
	}

	out, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := ListPermissionRequestsResponse{
		Requests: make([]PermissionRequestResponse, 0, len(out.Items)),
		Total:    out.Total,
		Page:     out.Page,
		PageSize: out.PageSize,
	}
	for _, item := range out.Items {
		resp.Requests = append(resp.Requests, toPermissionRequestResponse(item.PermissionRequest, item.ApprovalRequest, item.Employee))
	}
	writeJSON(w, http.StatusOK, resp)
}

// ListByEmployee handles GET /api/v1/employees/{uid}/permission-requests
func (h *PermissionRequestHandler) ListByEmployee(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	employeeUID := r.PathValue("uid")
	if employeeUID == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))

	input := usecases.ListPermissionRequestsInput{
		Page:        page,
		PageSize:    pageSize,
		EmployeeUID: &employeeUID,
	}
	if v := q.Get("status"); v != "" {
		s := domain.ApprovalRequestStatus(v)
		input.Status = &s
	}
	if v := q.Get("type"); v != "" {
		t := domain.PermissionType(v)
		input.Type = &t
	}
	out, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := ListPermissionRequestsResponse{
		Requests: make([]PermissionRequestResponse, 0, len(out.Items)),
		Total:    out.Total,
		Page:     out.Page,
		PageSize: out.PageSize,
	}
	for _, item := range out.Items {
		resp.Requests = append(resp.Requests, toPermissionRequestResponse(item.PermissionRequest, item.ApprovalRequest, item.Employee))
	}
	writeJSON(w, http.StatusOK, resp)
}

// ListPending handles GET /api/v1/permission-approvals/pending
func (h *PermissionRequestHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	out, err := h.listPendingUC.Execute(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := ListPendingPermissionApprovalsResponse{Approvals: make([]PendingPermissionApprovalResponse, 0, len(out.Items))}
	for _, item := range out.Items {
		resp.Approvals = append(resp.Approvals, PendingPermissionApprovalResponse{
			UID:                  item.ApprovalRequest.UID,
			PermissionRequestUID: item.PermissionRequest.UID,
			RequesterUID:         item.Employee.UID,
			RequesterName:        item.Employee.Name,
			Type:                 string(item.PermissionRequest.Type),
			TypeNameAr:           item.PermissionRequest.Type.NameAR(),
			TypeNameEn:           item.PermissionRequest.Type.NameEN(),
			PermissionDate:       item.PermissionRequest.PermissionDate.Format("2006-01-02"),
			StartTime:            item.PermissionRequest.StartTime,
			EndTime:              item.PermissionRequest.EndTime,
			Reason:               item.PermissionRequest.Reason,
			CurrentStep:          item.ApprovalRequest.CurrentStep,
			MaxStep:              item.ApprovalRequest.MaxStep,
			SubmittedAt:          item.PermissionRequest.SubmittedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// Approve handles POST /api/v1/permission-approvals/{uid}/approve
func (h *PermissionRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
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
	var req ApprovalActionRequest
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	out, err := h.approveUC.Execute(r.Context(), usecases.ApprovePermissionRequestInput{
		ApprovalRequestUID: uid,
		ActorUserID:        claims.UserID,
		ActorEmployeeUID:   *currentUser.EmployeeUID,
		Comments:           req.Comments,
	})
	if err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Approve")
		return
	}
	writeJSON(w, http.StatusOK, toPermissionRequestResponse(out.PermissionRequest, out.ApprovalRequest, nil))
}

// Reject handles POST /api/v1/permission-approvals/{uid}/reject
func (h *PermissionRequestHandler) Reject(w http.ResponseWriter, r *http.Request) {
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
	var req ApprovalActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Comments == nil || *req.Comments == "" {
		writeError(w, http.StatusBadRequest, "comments are required when rejecting")
		return
	}
	if _, err := h.rejectUC.Execute(r.Context(), usecases.RejectPermissionRequestInput{
		ApprovalRequestUID: uid,
		ActorUserID:        claims.UserID,
		ActorEmployeeUID:   *currentUser.EmployeeUID,
		Comments:           req.Comments,
	}); err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Reject")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "permission rejected"})
}

// Eligibility handles GET /api/v1/permission-requests/eligibility — returns the requester's
// effective shift, allowed dates per type, and weekly-cap status. Drives the form UI.
func (h *PermissionRequestHandler) Eligibility(w http.ResponseWriter, r *http.Request) {
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
	out, err := h.eligibilityUC.Execute(r.Context(), *currentUser.EmployeeUID)
	if err != nil {
		writePermissionUseCaseError(w, err, "permission_request_handler.Eligibility")
		return
	}

	type typeEligibilityResp struct {
		Type           string   `json:"type"`
		AllowedDates   []string `json:"allowedDates"`
		WeeklyCapHit   bool     `json:"weeklyCapHit"`
		BlockReason    string   `json:"blockReason,omitempty"`
		EffectiveStart string   `json:"effectiveStart,omitempty"`
		EffectiveEnd   string   `json:"effectiveEnd,omitempty"`
	}
	type response struct {
		Shift struct {
			Start        string `json:"start"`
			End          string `json:"end"`
			GraceMinutes int    `json:"graceMinutes"`
		} `json:"shift"`
		Eligibility []typeEligibilityResp `json:"eligibility"`
	}

	resp := response{}
	resp.Shift.Start = out.ShiftStart
	resp.Shift.End = out.ShiftEnd
	resp.Shift.GraceMinutes = out.GraceMinutes
	for _, t := range []domain.PermissionType{
		domain.PermissionTypeMorning, domain.PermissionTypePersonal,
		domain.PermissionTypeOfficial, domain.PermissionTypeHealthInsurance,
	} {
		entry := out.Eligibility[t]
		dates := make([]string, 0, len(entry.AllowedDates))
		for _, d := range entry.AllowedDates {
			dates = append(dates, d.Format("2006-01-02"))
		}
		resp.Eligibility = append(resp.Eligibility, typeEligibilityResp{
			Type:           string(t),
			AllowedDates:   dates,
			WeeklyCapHit:   entry.WeeklyCapHit,
			BlockReason:    entry.BlockReason,
			EffectiveStart: entry.EffectiveStart,
			EffectiveEnd:   entry.EffectiveEnd,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// History handles GET /api/v1/permission-approvals/{uid}/history
func (h *PermissionRequestHandler) History(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}
	out, err := h.historyUC.Execute(r.Context(), uid)
	if err != nil {
		if errors.Is(err, usecases.ErrApprovalRequestNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	historyResp := GetApprovalHistoryResponse{History: make([]ApprovalHistoryItemResponse, 0, len(out.History))}
	for _, item := range out.History {
		a := item.Action
		historyResp.History = append(historyResp.History, ApprovalHistoryItemResponse{
			UID:       a.UID,
			Action:    string(a.Action),
			StepOrder: a.StepOrder,
			ActorUID:  item.ActorUID,
			ActorName: item.ActorName,
			Comments:  a.Comments,
			ActedAt:   a.ActedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, historyResp)
}

// ---------- Error mapping ----------

func writePermissionUseCaseError(w http.ResponseWriter, err error, logTag string) {
	statusCode := http.StatusInternalServerError
	switch {
	case errors.Is(err, usecases.ErrPermissionRequestNotFound),
		errors.Is(err, usecases.ErrApprovalRequestNotFound),
		errors.Is(err, usecases.ErrApprovalFlowNotFound),
		errors.Is(err, usecases.ErrApprovalFlowStepNotFound),
		errors.Is(err, usecases.ErrEmployeeNotFound),
		errors.Is(err, usecases.ErrUserNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, usecases.ErrPermissionTypeInvalid),
		errors.Is(err, usecases.ErrPermissionDeadlineMissed),
		errors.Is(err, usecases.ErrPermissionWindowRequired),
		errors.Is(err, usecases.ErrPermissionWindowOutsideShift),
		errors.Is(err, usecases.ErrPermissionWindowInvalid),
		errors.Is(err, usecases.ErrPermissionOverlap),
		errors.Is(err, usecases.ErrPermissionWeeklyCapExceeded),
		errors.Is(err, usecases.ErrPermissionDateInPast),
		errors.Is(err, usecases.ErrPermissionNotEditable),
		errors.Is(err, usecases.ErrPermissionNotCancellable),
		errors.Is(err, usecases.ErrPermissionAlreadyExpired),
		errors.Is(err, usecases.ErrApprovalFlowHasNoSteps),
		errors.Is(err, usecases.ErrRequestNotPending):
		statusCode = http.StatusBadRequest
	case errors.Is(err, usecases.ErrNotRequestOwner),
		errors.Is(err, usecases.ErrNotAuthorizedApprover):
		statusCode = http.StatusForbidden
	default:
		slog.Error(logTag+".execute_usecase", "error", err)
	}
	writeError(w, statusCode, err.Error())
}

