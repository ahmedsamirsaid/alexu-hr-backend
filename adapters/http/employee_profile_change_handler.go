package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/banumusa/backend/core/usecases"
)

type EmployeeProfileChangeHandler struct {
	submitUC         *usecases.SubmitEmployeeProfileChangeRequestUseCase
	listByEmpUC      *usecases.ListEmployeeProfileChangeRequestsUseCase
	getUC            *usecases.GetEmployeeProfileChangeRequestUseCase
	listPendingUC    *usecases.ListPendingEmployeeProfileChangeRequestsUseCase
	approveUC        *usecases.ApproveEmployeeProfileChangeRequestUseCase
	rejectUC         *usecases.RejectEmployeeProfileChangeRequestUseCase
	historyUC        *usecases.GetApprovalHistoryUseCase
	updateEmpUC      *usecases.UpdateEmployeeProfileUseCase
	getCurrentUserUC *usecases.GetCurrentUserUseCase
}

func NewEmployeeProfileChangeHandler(
	submitUC *usecases.SubmitEmployeeProfileChangeRequestUseCase,
	listByEmpUC *usecases.ListEmployeeProfileChangeRequestsUseCase,
	getUC *usecases.GetEmployeeProfileChangeRequestUseCase,
	listPendingUC *usecases.ListPendingEmployeeProfileChangeRequestsUseCase,
	approveUC *usecases.ApproveEmployeeProfileChangeRequestUseCase,
	rejectUC *usecases.RejectEmployeeProfileChangeRequestUseCase,
	historyUC *usecases.GetApprovalHistoryUseCase,
	updateEmpUC *usecases.UpdateEmployeeProfileUseCase,
	getCurrentUserUC *usecases.GetCurrentUserUseCase,
) *EmployeeProfileChangeHandler {
	return &EmployeeProfileChangeHandler{
		submitUC:         submitUC,
		listByEmpUC:      listByEmpUC,
		getUC:            getUC,
		listPendingUC:    listPendingUC,
		approveUC:        approveUC,
		rejectUC:         rejectUC,
		historyUC:        historyUC,
		updateEmpUC:      updateEmpUC,
		getCurrentUserUC: getCurrentUserUC,
	}
}

type EmployeeProfileChangeRequestResponse struct {
	UID                       string  `json:"uid"`
	ApprovalRequestUID        string  `json:"approvalRequestUid"`
	Status                    string  `json:"status"`
	CurrentStep               int     `json:"currentStep"`
	MaxStep                   int     `json:"maxStep"`
	EmployeeUID               string  `json:"employeeUid"`
	EmployeeName              string  `json:"employeeName"`
	SubmittedByUID            string  `json:"submittedByUid"`
	SubmittedByName           string  `json:"submittedByName"`
	CurrentFinancialGrade     *string `json:"currentFinancialGrade,omitempty"`
	CurrentIDCardValidUntil   *string `json:"currentIdCardValidUntil,omitempty"`
	CurrentMaritalStatus      *string `json:"currentMaritalStatus,omitempty"`
	RequestedFinancialGrade   *string `json:"requestedFinancialGrade,omitempty"`
	RequestedIDCardValidUntil *string `json:"requestedIdCardValidUntil,omitempty"`
	RequestedMaritalStatus    *string `json:"requestedMaritalStatus,omitempty"`
	Comments                  *string `json:"comments,omitempty"`
	CreatedAt                 string  `json:"createdAt"`
}

type ListEmployeeProfileChangeRequestsResponse struct {
	Requests []EmployeeProfileChangeRequestResponse `json:"requests"`
}

type EmployeeProfileChangeHistoryResponse struct {
	Request EmployeeProfileChangeRequestResponse `json:"request"`
	History []ApprovalHistoryItemResponse        `json:"history"`
}

func (h *EmployeeProfileChangeHandler) Submit(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}

	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeError(w, http.StatusBadRequest, "No employee profile linked to user")
		return
	}

	employeeUID := r.PathValue("uid")
	if strings.TrimSpace(employeeUID) == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var payload map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input, err := parseSubmitEmployeeProfileChangeRequest(employeeUID, claims.UserID, *currentUser.EmployeeUID, payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.submitUC.Execute(r.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			status = http.StatusNotFound
		case errors.Is(err, usecases.ErrNoProfileChangesRequested), errors.Is(err, usecases.ErrEmployeeProfileChangeAlreadyPending):
			status = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNotAuthorizedApprover):
			status = http.StatusForbidden
		default:
			slog.Error("employee_profile_change_handler.Submit.execute_usecase", "error", err)
		}
		writeError(w, status, err.Error())
		return
	}

	details, err := h.getUC.Execute(r.Context(), claims.UserID, output.Request.UID)
	if err != nil {
		slog.Error("employee_profile_change_handler.Submit.get_request", "error", err, "request_uid", output.Request.UID)
		writeError(w, http.StatusInternalServerError, "failed to load created request")
		return
	}
	writeJSON(w, http.StatusCreated, toEmployeeProfileChangeRequestResponse(details))
}

func (h *EmployeeProfileChangeHandler) ListByEmployee(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	employeeUID := r.PathValue("uid")
	if strings.TrimSpace(employeeUID) == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}
	items, err := h.listByEmpUC.Execute(r.Context(), claims.UserID, employeeUID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrNotAuthorizedApprover) {
			status = http.StatusForbidden
		} else {
			slog.Error("employee_profile_change_handler.ListByEmployee.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, status, err.Error())
		return
	}
	resp := make([]EmployeeProfileChangeRequestResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toEmployeeProfileChangeRequestResponse(item))
	}
	writeJSON(w, http.StatusOK, ListEmployeeProfileChangeRequestsResponse{Requests: resp})
}

func (h *EmployeeProfileChangeHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	requestUID := r.PathValue("requestUid")
	if strings.TrimSpace(requestUID) == "" {
		writeError(w, http.StatusBadRequest, "requestUid is required")
		return
	}
	item, err := h.getUC.Execute(r.Context(), claims.UserID, requestUID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeProfileChangeRequestNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, usecases.ErrNotAuthorizedApprover) {
			status = http.StatusForbidden
		} else {
			slog.Error("employee_profile_change_handler.Get.execute_usecase", "error", err, "request_uid", requestUID)
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toEmployeeProfileChangeRequestResponse(item))
}

func (h *EmployeeProfileChangeHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	items, err := h.listPendingUC.Execute(r.Context(), claims.UserID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrNotAuthorizedApprover) {
			status = http.StatusForbidden
		} else {
			slog.Error("employee_profile_change_handler.ListPending.execute_usecase", "error", err)
		}
		writeError(w, status, err.Error())
		return
	}
	resp := make([]EmployeeProfileChangeRequestResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toEmployeeProfileChangeRequestResponse(item))
	}
	writeJSON(w, http.StatusOK, ListEmployeeProfileChangeRequestsResponse{Requests: resp})
}

func (h *EmployeeProfileChangeHandler) Approve(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeError(w, http.StatusBadRequest, "No employee profile linked to user")
		return
	}
	requestUID := r.PathValue("requestUid")
	var req ApprovalActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	output, err := h.approveUC.Execute(r.Context(), usecases.ApproveEmployeeProfileChangeRequestInput{
		RequestUID:       requestUID,
		ActorUserID:      claims.UserID,
		ActorEmployeeUID: *currentUser.EmployeeUID,
		Comments:         req.Comments,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeProfileChangeRequestNotFound), errors.Is(err, usecases.ErrApprovalRequestNotFound):
			status = http.StatusNotFound
		case errors.Is(err, usecases.ErrRequestNotPending), errors.Is(err, usecases.ErrEmployeeNotFound):
			status = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNotAuthorizedApprover):
			status = http.StatusForbidden
		default:
			slog.Error("employee_profile_change_handler.Approve.execute_usecase", "error", err, "request_uid", requestUID)
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"uid":         output.Request.UID,
		"status":      string(output.ApprovalRequest.Status),
		"currentStep": output.ApprovalRequest.CurrentStep,
	})
}

func (h *EmployeeProfileChangeHandler) Reject(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeError(w, http.StatusBadRequest, "No employee profile linked to user")
		return
	}
	requestUID := r.PathValue("requestUid")
	var req ApprovalActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	approvalRequest, err := h.rejectUC.Execute(r.Context(), usecases.RejectEmployeeProfileChangeRequestInput{
		RequestUID:       requestUID,
		ActorUserID:      claims.UserID,
		ActorEmployeeUID: *currentUser.EmployeeUID,
		Comments:         req.Comments,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeProfileChangeRequestNotFound), errors.Is(err, usecases.ErrApprovalRequestNotFound):
			status = http.StatusNotFound
		case errors.Is(err, usecases.ErrRequestNotPending):
			status = http.StatusBadRequest
		case errors.Is(err, usecases.ErrNotAuthorizedApprover):
			status = http.StatusForbidden
		default:
			slog.Error("employee_profile_change_handler.Reject.execute_usecase", "error", err, "request_uid", requestUID)
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"uid":    requestUID,
		"status": string(approvalRequest.Status),
	})
}

func (h *EmployeeProfileChangeHandler) History(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	requestUID := r.PathValue("requestUid")
	if strings.TrimSpace(requestUID) == "" {
		writeError(w, http.StatusBadRequest, "requestUid is required")
		return
	}
	item, err := h.getUC.Execute(r.Context(), claims.UserID, requestUID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrEmployeeProfileChangeRequestNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, usecases.ErrNotAuthorizedApprover) {
			status = http.StatusForbidden
		}
		writeError(w, status, err.Error())
		return
	}
	history, err := h.historyUC.Execute(r.Context(), item.Request.ApprovalRequestUID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrApprovalRequestNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	respHistory := make([]ApprovalHistoryItemResponse, 0, len(history.History))
	for _, action := range history.History {
		respHistory = append(respHistory, ApprovalHistoryItemResponse{
			Action:    string(action.Action.Action),
			StepOrder: action.Action.StepOrder,
			ActorUID:  action.ActorUID,
			ActorName: action.ActorName,
			Comments:  action.Action.Comments,
			ActedAt:   action.Action.ActedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, EmployeeProfileChangeHistoryResponse{
		Request: toEmployeeProfileChangeRequestResponse(item),
		History: respHistory,
	})
}

func (h *EmployeeProfileChangeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
		return
	}
	currentUser, err := h.getCurrentUserUC.Execute(r.Context(), usecases.GetCurrentUserInput{UserID: claims.UserID})
	if err != nil || currentUser.EmployeeUID == nil {
		writeError(w, http.StatusBadRequest, "No employee profile linked to user")
		return
	}
	employeeUID := r.PathValue("uid")
	if strings.TrimSpace(employeeUID) == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	changes, err := parseEmployeeProfilePatchPayload(payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := h.updateEmpUC.Execute(r.Context(), usecases.UpdateEmployeeProfileInput{
		EmployeeUID:      employeeUID,
		ActorUserID:      claims.UserID,
		ActorEmployeeUID: *currentUser.EmployeeUID,
		Changes:          changes,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrEmployeeNotFound):
			status = http.StatusNotFound
		case errors.Is(err, usecases.ErrNotAuthorizedApprover):
			status = http.StatusForbidden
		case errors.Is(err, usecases.ErrNoProfileChangesRequested),
			errors.Is(err, usecases.ErrEmployeeMobileAlreadyExists),
			errors.Is(err, usecases.ErrGovernmentIDAlreadyExists),
			errors.Is(err, usecases.ErrUniversityIDAlreadyExists),
			errors.Is(err, usecases.ErrPhoneAlreadyExists),
			errors.Is(err, usecases.ErrInvalidEmployeeClassification):
			status = http.StatusBadRequest
		default:
			slog.Error("employee_profile_change_handler.UpdateEmployee.execute_usecase", "error", err, "employee_uid", employeeUID)
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"uid": output.Employee.UID})
}

func toEmployeeProfileChangeRequestResponse(item *usecases.EmployeeProfileChangeRequestDetails) EmployeeProfileChangeRequestResponse {
	resp := EmployeeProfileChangeRequestResponse{
		UID:                       item.Request.UID,
		ApprovalRequestUID:        item.Request.ApprovalRequestUID,
		EmployeeUID:               item.Request.EmployeeUID,
		SubmittedByUID:            item.Request.SubmittedByEmployeeUID,
		CurrentFinancialGrade:     item.Request.CurrentFinancialGrade,
		CurrentIDCardValidUntil:   formatOptionalDatePtr(item.Request.CurrentIDCardValidUntil),
		CurrentMaritalStatus:      item.Request.CurrentMaritalStatus,
		RequestedFinancialGrade:   item.Request.RequestedFinancialGrade,
		RequestedIDCardValidUntil: formatOptionalDatePtr(item.Request.RequestedIDCardValidUntil),
		RequestedMaritalStatus:    item.Request.RequestedMaritalStatus,
		Comments:                  item.Request.Comments,
		CreatedAt:                 item.Request.CreatedAt.Format(time.RFC3339),
	}
	if item.ApprovalRequest != nil {
		resp.Status = string(item.ApprovalRequest.Status)
		resp.CurrentStep = item.ApprovalRequest.CurrentStep
		resp.MaxStep = item.ApprovalRequest.MaxStep
	}
	if item.Employee != nil {
		resp.EmployeeName = item.Employee.Name
	}
	if item.SubmittedBy != nil {
		resp.SubmittedByName = item.SubmittedBy.Name
	}
	return resp
}

func formatOptionalDatePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func parseSubmitEmployeeProfileChangeRequest(employeeUID string, actorUserID int64, actorEmployeeUID string, payload map[string]json.RawMessage) (usecases.SubmitEmployeeProfileChangeRequestInput, error) {
	input := usecases.SubmitEmployeeProfileChangeRequestInput{
		EmployeeUID:      employeeUID,
		ActorUserID:      actorUserID,
		ActorEmployeeUID: actorEmployeeUID,
	}
	for key, raw := range payload {
		switch key {
		case "financialGrade":
			input.SetFinancialGrade = true
			value, err := decodeOptionalString(raw)
			if err != nil {
				return input, err
			}
			input.RequestedFinancialGrade = value
		case "idCardValidUntil":
			input.SetIDCardValidUntil = true
			value, err := decodeOptionalDate(raw)
			if err != nil {
				return input, err
			}
			input.RequestedIDCardValidUntil = value
		case "maritalStatus":
			input.SetMaritalStatus = true
			value, err := decodeOptionalString(raw)
			if err != nil {
				return input, err
			}
			input.RequestedMaritalStatus = value
		case "comments":
			value, err := decodeOptionalString(raw)
			if err != nil {
				return input, err
			}
			input.Comments = value
		default:
			return input, errors.New("unsupported field " + key)
		}
	}
	return input, nil
}

func parseEmployeeProfilePatchPayload(payload map[string]json.RawMessage) (map[string]any, error) {
	changes := make(map[string]any, len(payload))
	for key, raw := range payload {
		var (
			value any
			err   error
		)
		switch key {
		case "name", "mobile", "governmentId", "universityId", "status", "type", "subType":
			value, err = decodeRequiredString(raw)
		case "email", "telephoneNumber", "gender", "religion", "maritalStatus", "address", "placeOfBirth", "placeOfResidence", "policeStation",
			"academicLevel", "educationalQualification", "universityName", "faculty", "specialization", "appointmentDecisionNumber",
			"appointmentType", "departmentName", "grade", "workEntity", "employeeFileNumber", "insuranceNumber", "employmentStatus",
			"militaryStatus", "medicalCadre", "personalPhotoUrl", "nationalCardImageUrl", "qualificationCertificateImageUrl", "cvUrl",
			"memberNumber", "insuranceCode", "jobGroup", "qualitativeGroup", "jobTitleAtLevel", "jobTitleBeforePlacement", "financialGrade",
			"previousFinancialGrade", "jobLevel", "natureOfAppointment", "notes", "decisionFileUrl", "departmentUid", "shiftUid":
			value, err = decodeOptionalStringValue(raw)
		case "hireDate":
			value, err = decodeRequiredDateValue(raw)
		case "dateOfBirth", "idCardValidUntil", "subscriptionDate", "actualAppointmentReappointmentDate", "appointmentDecisionDate", "decisionDate", "gradeGrantDate":
			value, err = decodeOptionalDateValue(raw)
		case "personWithSpecialNeeds", "solidarityFund", "reappointment", "appointmentSeniorityOrGradeWithdrawal":
			value, err = decodeBool(raw)
		case "yearObtained":
			value, err = decodeOptionalIntValue(raw)
		case "decisionNumber":
			value, err = decodeOptionalInt64Value(raw)
		default:
			return nil, errors.New("unsupported field " + key)
		}
		if err != nil {
			return nil, err
		}
		changes[key] = value
	}
	return changes, nil
}

func decodeRequiredString(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", errors.New("invalid string value")
	}
	return value, nil
}

func decodeOptionalString(raw json.RawMessage) (*string, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("invalid string value")
	}
	return &value, nil
}

func decodeOptionalStringValue(raw json.RawMessage) (any, error) {
	return decodeOptionalString(raw)
}

func decodeRequiredDateValue(raw json.RawMessage) (any, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("invalid date value")
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	return parsed, nil
}

func decodeOptionalDate(raw json.RawMessage) (*time.Time, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("invalid date value")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	return &parsed, nil
}

func decodeOptionalDateValue(raw json.RawMessage) (any, error) {
	return decodeOptionalDate(raw)
}

func decodeBool(raw json.RawMessage) (any, error) {
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("invalid boolean value")
	}
	return value, nil
}

func decodeOptionalIntValue(raw json.RawMessage) (any, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("invalid integer value")
	}
	return value, nil
}

func decodeOptionalInt64Value(raw json.RawMessage) (any, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("invalid integer value")
	}
	return value, nil
}
