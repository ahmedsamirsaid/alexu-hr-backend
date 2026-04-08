package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/usecases"
)

// ApprovalFlowHandler handles approval flow HTTP requests (admin).
type ApprovalFlowHandler struct {
	listFlowsUC      *usecases.ListApprovalFlowsUseCase
	createFlowUC     *usecases.CreateApprovalFlowUseCase
	updateFlowUC     *usecases.UpdateApprovalFlowUseCase
	listStepsUC      *usecases.ListApprovalFlowStepsUseCase
	createStepUC     *usecases.CreateApprovalFlowStepUseCase
	updateStepUC     *usecases.UpdateApprovalFlowStepUseCase
	deleteStepUC     *usecases.DeleteApprovalFlowStepUseCase
}

func NewApprovalFlowHandler(
	listFlowsUC *usecases.ListApprovalFlowsUseCase,
	createFlowUC *usecases.CreateApprovalFlowUseCase,
	updateFlowUC *usecases.UpdateApprovalFlowUseCase,
	listStepsUC *usecases.ListApprovalFlowStepsUseCase,
	createStepUC *usecases.CreateApprovalFlowStepUseCase,
	updateStepUC *usecases.UpdateApprovalFlowStepUseCase,
	deleteStepUC *usecases.DeleteApprovalFlowStepUseCase,
) *ApprovalFlowHandler {
	return &ApprovalFlowHandler{
		listFlowsUC:  listFlowsUC,
		createFlowUC: createFlowUC,
		updateFlowUC: updateFlowUC,
		listStepsUC:  listStepsUC,
		createStepUC: createStepUC,
		updateStepUC: updateStepUC,
		deleteStepUC: deleteStepUC,
	}
}

// Approval Flow responses

type ApprovalFlowResponse struct {
	UID         string  `json:"uid"`
	Code        string  `json:"code"`
	NameEN      string  `json:"nameEn"`
	NameAR      *string `json:"nameAr,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"isActive"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type ListApprovalFlowsResponse struct {
	Flows []ApprovalFlowResponse `json:"flows"`
}

type ApprovalFlowStepResponse struct {
	UID       string `json:"uid"`
	FlowUID   string `json:"flowUid"`
	StepOrder int    `json:"stepOrder"`
	RoleUID   string `json:"roleUid"`
	RoleName  string `json:"roleName,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ListApprovalFlowStepsResponse struct {
	Steps []ApprovalFlowStepResponse `json:"steps"`
}

// Approval Flow requests

type CreateApprovalFlowRequest struct {
	Code        string  `json:"code"`
	NameEN      string  `json:"nameEn"`
	NameAR      *string `json:"nameAr,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UpdateApprovalFlowRequest struct {
	NameEN      *string `json:"nameEn,omitempty"`
	NameAR      *string `json:"nameAr,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

type CreateApprovalFlowStepRequest struct {
	StepOrder int    `json:"stepOrder"`
	RoleUID   string `json:"roleUid"`
}

type UpdateApprovalFlowStepRequest struct {
	StepOrder *int    `json:"stepOrder,omitempty"`
	RoleUID   *string `json:"roleUid,omitempty"`
}

// ListApprovalFlows handles GET /api/v1/admin/approval-flows
func (h *ApprovalFlowHandler) ListApprovalFlows(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	output, err := h.listFlowsUC.Execute(r.Context(), activeOnly)
	if err != nil {
		slog.Error("approval_flow_handler.ListApprovalFlows.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	flows := make([]ApprovalFlowResponse, 0, len(output.Flows))
	for _, f := range output.Flows {
		flows = append(flows, ApprovalFlowResponse{
			UID:         f.UID,
			Code:        f.Code,
			NameEN:      f.NameEN,
			NameAR:      f.NameAR,
			Description: f.Description,
			IsActive:    f.IsActive,
			CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, ListApprovalFlowsResponse{Flows: flows})
}

// CreateApprovalFlow handles POST /api/v1/admin/approval-flows
func (h *ApprovalFlowHandler) CreateApprovalFlow(w http.ResponseWriter, r *http.Request) {
	var req CreateApprovalFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("approval_flow_handler.CreateApprovalFlow.decode_request", "error", err)
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

	input := usecases.CreateApprovalFlowInput{
		Code:        req.Code,
		NameEN:      req.NameEN,
		NameAR:      req.NameAR,
		Description: req.Description,
	}
	output, err := h.createFlowUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrApprovalFlowCodeExists) {
			statusCode = http.StatusConflict
		} else {
			slog.Error("approval_flow_handler.CreateApprovalFlow.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ApprovalFlowResponse{
		UID:         output.Flow.UID,
		Code:        output.Flow.Code,
		NameEN:      output.Flow.NameEN,
		NameAR:      output.Flow.NameAR,
		Description: output.Flow.Description,
		IsActive:    output.Flow.IsActive,
		CreatedAt:   output.Flow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   output.Flow.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateApprovalFlow handles PATCH /api/v1/admin/approval-flows/{uid}
func (h *ApprovalFlowHandler) UpdateApprovalFlow(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req UpdateApprovalFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("approval_flow_handler.UpdateApprovalFlow.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.UpdateApprovalFlowInput{
		UID:         uid,
		NameEN:      req.NameEN,
		NameAR:      req.NameAR,
		Description: req.Description,
		IsActive:    req.IsActive,
	}
	output, err := h.updateFlowUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrApprovalFlowNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("approval_flow_handler.UpdateApprovalFlow.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApprovalFlowResponse{
		UID:         output.Flow.UID,
		Code:        output.Flow.Code,
		NameEN:      output.Flow.NameEN,
		NameAR:      output.Flow.NameAR,
		Description: output.Flow.Description,
		IsActive:    output.Flow.IsActive,
		CreatedAt:   output.Flow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   output.Flow.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListApprovalFlowSteps handles GET /api/v1/admin/approval-flows/{uid}/steps
func (h *ApprovalFlowHandler) ListApprovalFlowSteps(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.listStepsUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrApprovalFlowNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("approval_flow_handler.ListApprovalFlowSteps.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	steps := make([]ApprovalFlowStepResponse, 0, len(output.Steps))
	for _, s := range output.Steps {
		steps = append(steps, ApprovalFlowStepResponse{
			UID:       s.UID,
			FlowUID:   s.ApprovalFlowUID,
			StepOrder: s.StepOrder,
			RoleUID:   s.RoleUID,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, ListApprovalFlowStepsResponse{Steps: steps})
}

// CreateApprovalFlowStep handles POST /api/v1/admin/approval-flows/{uid}/steps
func (h *ApprovalFlowHandler) CreateApprovalFlowStep(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req CreateApprovalFlowStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("approval_flow_handler.CreateApprovalFlowStep.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RoleUID == "" {
		writeError(w, http.StatusBadRequest, "roleUid is required")
		return
	}

	input := usecases.CreateApprovalFlowStepInput{
		ApprovalFlowUID: uid,
		StepOrder:       req.StepOrder,
		RoleUID:         req.RoleUID,
	}
	output, err := h.createStepUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrApprovalFlowNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrRoleNotFound):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrDuplicateStepOrder):
			statusCode = http.StatusConflict
		default:
			slog.Error("approval_flow_handler.CreateApprovalFlowStep.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ApprovalFlowStepResponse{
		UID:       output.Step.UID,
		FlowUID:   output.Step.ApprovalFlowUID,
		StepOrder: output.Step.StepOrder,
		RoleUID:   output.Step.RoleUID,
		CreatedAt: output.Step.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: output.Step.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateApprovalFlowStep handles PATCH /api/v1/admin/approval-flows/{uid}/steps/{stepUid}
func (h *ApprovalFlowHandler) UpdateApprovalFlowStep(w http.ResponseWriter, r *http.Request) {
	flowUID := r.PathValue("uid")
	stepUID := r.PathValue("stepUid")
	if flowUID == "" || stepUID == "" {
		writeError(w, http.StatusBadRequest, "uid and stepUid are required")
		return
	}

	var req UpdateApprovalFlowStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("approval_flow_handler.UpdateApprovalFlowStep.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.UpdateApprovalFlowStepInput{
		UID:       stepUID,
		StepOrder: req.StepOrder,
		RoleUID:   req.RoleUID,
	}
	output, err := h.updateStepUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrApprovalFlowStepNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrRoleNotFound):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecases.ErrDuplicateStepOrder):
			statusCode = http.StatusConflict
		default:
			slog.Error("approval_flow_handler.UpdateApprovalFlowStep.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApprovalFlowStepResponse{
		UID:       output.Step.UID,
		FlowUID:   output.Step.ApprovalFlowUID,
		StepOrder: output.Step.StepOrder,
		RoleUID:   output.Step.RoleUID,
		CreatedAt: output.Step.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: output.Step.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// DeleteApprovalFlowStep handles DELETE /api/v1/admin/approval-flows/{uid}/steps/{stepUid}
func (h *ApprovalFlowHandler) DeleteApprovalFlowStep(w http.ResponseWriter, r *http.Request) {
	flowUID := r.PathValue("uid")
	stepUID := r.PathValue("stepUid")
	if flowUID == "" || stepUID == "" {
		writeError(w, http.StatusBadRequest, "uid and stepUid are required")
		return
	}

	err := h.deleteStepUC.Execute(r.Context(), stepUID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrApprovalFlowStepNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("approval_flow_handler.DeleteApprovalFlowStep.execute_usecase", "error", err)
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
