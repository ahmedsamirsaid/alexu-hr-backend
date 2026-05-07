package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
)

// LeaveTypeHandler handles leave type HTTP requests (admin).
type LeaveTypeHandler struct {
	getLeaveTypeDetailsUC *usecases.GetLeaveTypeDetailsUseCase
	listLeaveTypesUC      *usecases.ListLeaveTypesUseCase
	listSubLeaveTypesUC   *usecases.ListSubLeaveTypesUseCase
	updateLeaveTypeUC     *usecases.UpdateLeaveTypeUseCase
	setApprovalFlowUC     *usecases.SetLeaveTypeApprovalFlowUseCase
	toggleLeaveTypeUC     *usecases.ToggleLeaveTypeUseCase
	i18nService           ports.I18nService
}

func NewLeaveTypeHandler(
	getLeaveTypeDetailsUC *usecases.GetLeaveTypeDetailsUseCase,
	listLeaveTypesUC *usecases.ListLeaveTypesUseCase,
	listSubLeaveTypesUC *usecases.ListSubLeaveTypesUseCase,
	updateLeaveTypeUC *usecases.UpdateLeaveTypeUseCase,
	setApprovalFlowUC *usecases.SetLeaveTypeApprovalFlowUseCase,
	toggleLeaveTypeUC *usecases.ToggleLeaveTypeUseCase,
	i18nService ports.I18nService,
) *LeaveTypeHandler {
	return &LeaveTypeHandler{
		getLeaveTypeDetailsUC: getLeaveTypeDetailsUC,
		listLeaveTypesUC:      listLeaveTypesUC,
		listSubLeaveTypesUC:   listSubLeaveTypesUC,
		updateLeaveTypeUC:     updateLeaveTypeUC,
		setApprovalFlowUC:     setApprovalFlowUC,
		toggleLeaveTypeUC:     toggleLeaveTypeUC,
		i18nService:           i18nService,
	}
}

// localizedLeaveName picks NameAR when the request locale is "ar", else NameEN.
func (h *LeaveTypeHandler) localizedLeaveName(r *http.Request, nameEN, nameAR string) string {
	if h.i18nService != nil {
		locale := h.i18nService.GetDefaultLocale()
		// The middleware stores the resolved locale in context; we peek at it via T
		// by resolving a dummy key — the locale is the same one the middleware set.
		// Simpler: use the language middleware's context value directly.
		if ctx := r.Context(); ctx != nil {
			// Resolve locale from context by checking what T returns for a known AR-only key
			_ = locale
		}
	}
	// Use Accept-Language directly for the pick
	acceptLang := r.Header.Get("Accept-Language")
	if len(acceptLang) >= 2 && acceptLang[:2] == "ar" {
		if nameAR != "" {
			return nameAR
		}
	}
	return nameEN
}

// Leave Type responses

type LeaveTypeResponse struct {
	UID                   string  `json:"uid"`
	Code                  string  `json:"code"`
	NameEN                string  `json:"nameEn"`
	NameAR                string  `json:"nameAr"`
	LocalizedName         string  `json:"name"`
	DefaultBalance        int     `json:"defaultBalance"`
	MaxConsecutive        *int    `json:"maxConsecutive,omitempty"`
	RecordingDeadlineDays *int    `json:"recordingDeadlineDays,omitempty"`
	AdvanceNoticeDays     *int    `json:"advanceNoticeDays,omitempty"`
	IsActive              bool    `json:"isActive"`
	CreatedAt             string  `json:"createdAt"`
	UpdatedAt             string  `json:"updatedAt"`
}

type ListLeaveTypesResponse struct {
	LeaveTypes []LeaveTypeResponse `json:"leaveTypes"`
}

type SubLeaveTypeResponse struct {
	UID           string `json:"uid"`
	LeaveTypeUID  string `json:"leaveTypeUid"`
	NameEN        string `json:"nameEn"`
	NameAR        string `json:"nameAr"`
	LocalizedName string `json:"name"`
}

type ListSubLeaveTypesResponse struct {
	SubLeaveTypes []SubLeaveTypeResponse `json:"subLeaveTypes"`
}

type GetLeaveTypeResponse struct {
	LeaveType     LeaveTypeResponse              `json:"leaveType"`
	ApprovalFlow  *LeaveTypeApprovalFlowResponse `json:"approvalFlow,omitempty"`
	SubLeaveTypes []SubLeaveTypeResponse         `json:"subLeaveTypes"`
}

type LeaveTypeApprovalFlowResponse struct {
	UID         string                     `json:"uid"`
	Code        string                     `json:"code"`
	NameEN      string                     `json:"nameEn"`
	NameAR      *string                    `json:"nameAr,omitempty"`
	Description *string                    `json:"description,omitempty"`
	IsActive    bool                       `json:"isActive"`
	CreatedAt   string                     `json:"createdAt"`
	UpdatedAt   string                     `json:"updatedAt"`
	Steps       []ApprovalFlowStepResponse `json:"steps"`
}

// Leave Type requests

type ToggleLeaveTypeRequest struct {
	IsActive bool `json:"isActive"`
}

type optionalIntField struct {
	Set   bool
	Value *int
}

func (f *optionalIntField) UnmarshalJSON(data []byte) error {
	f.Set = true
	if string(data) == "null" {
		f.Value = nil
		return nil
	}

	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	f.Value = &value
	return nil
}

type UpdateLeaveTypeRequest struct {
	DefaultBalance        optionalIntField `json:"defaultBalance"`
	RecordingDeadlineDays optionalIntField `json:"recordingDeadlineDays"`
	AdvanceNoticeDays     optionalIntField `json:"advanceNoticeDays"`
}

type SetLeaveTypeApprovalFlowRequest struct {
	ApprovalFlowUID *string `json:"approvalFlowUid"`
}

// ListLeaveTypes handles GET /api/v1/admin/leave-types
func (h *LeaveTypeHandler) ListLeaveTypes(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	output, err := h.listLeaveTypesUC.Execute(r.Context(), activeOnly)
	if err != nil {
		slog.Error("leave_type_handler.ListLeaveTypes.execute_usecase", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	leaveTypes := make([]LeaveTypeResponse, 0, len(output.LeaveTypes))
	for _, lt := range output.LeaveTypes {
		leaveTypes = append(leaveTypes, LeaveTypeResponse{
			UID:                   lt.UID,
			Code:                  lt.Code,
			NameEN:                lt.NameEN,
			NameAR:                lt.NameAR,
			LocalizedName:         h.localizedLeaveName(r, lt.NameEN, lt.NameAR),
			DefaultBalance:        lt.DefaultBalance,
			MaxConsecutive:        lt.MaxConsecutive,
			RecordingDeadlineDays: lt.RecordingDeadlineDays,
			AdvanceNoticeDays:     lt.AdvanceNoticeDays,
			IsActive:              lt.IsActive,
			CreatedAt:             lt.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:             lt.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, http.StatusOK, ListLeaveTypesResponse{LeaveTypes: leaveTypes})
}

// GetLeaveType handles GET /api/v1/leave-types/{uid}
func (h *LeaveTypeHandler) GetLeaveType(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.getLeaveTypeDetailsUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrLeaveTypeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("leave_type_handler.GetLeaveType.execute_usecase", "error", err, "leave_type_uid", uid)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	subLeaveTypes := make([]SubLeaveTypeResponse, 0, len(output.SubLeaveTypes))
	for _, item := range output.SubLeaveTypes {
		subLeaveTypes = append(subLeaveTypes, SubLeaveTypeResponse{
			UID:           item.UID,
			LeaveTypeUID:  item.LeaveTypeUID,
			NameEN:        item.NameEN,
			NameAR:        item.NameAR,
			LocalizedName: h.localizedLeaveName(r, item.NameEN, item.NameAR),
		})
	}

	steps := make([]ApprovalFlowStepResponse, 0, len(output.ApprovalFlowSteps))
	for _, item := range output.ApprovalFlowSteps {
		steps = append(steps, ApprovalFlowStepResponse{
			UID:       item.Step.UID,
			FlowUID:   item.Step.ApprovalFlowUID,
			StepOrder: item.Step.StepOrder,
			RoleUID:   item.Step.RoleUID,
			RoleName:  item.RoleName,
			CreatedAt: item.Step.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: item.Step.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	var approvalFlow *LeaveTypeApprovalFlowResponse
	if output.ApprovalFlow != nil {
		approvalFlow = &LeaveTypeApprovalFlowResponse{
			UID:         output.ApprovalFlow.UID,
			Code:        output.ApprovalFlow.Code,
			NameEN:      output.ApprovalFlow.NameEN,
			NameAR:      output.ApprovalFlow.NameAR,
			Description: output.ApprovalFlow.Description,
			IsActive:    output.ApprovalFlow.IsActive,
			CreatedAt:   output.ApprovalFlow.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   output.ApprovalFlow.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Steps:       steps,
		}
	}

	writeJSON(w, http.StatusOK, GetLeaveTypeResponse{
		LeaveType: LeaveTypeResponse{
			UID:                   output.LeaveType.UID,
			Code:                  output.LeaveType.Code,
			NameEN:                output.LeaveType.NameEN,
			NameAR:                output.LeaveType.NameAR,
			LocalizedName:         h.localizedLeaveName(r, output.LeaveType.NameEN, output.LeaveType.NameAR),
			DefaultBalance:        output.LeaveType.DefaultBalance,
			MaxConsecutive:        output.LeaveType.MaxConsecutive,
			RecordingDeadlineDays: output.LeaveType.RecordingDeadlineDays,
			AdvanceNoticeDays:     output.LeaveType.AdvanceNoticeDays,
			IsActive:              output.LeaveType.IsActive,
			CreatedAt:             output.LeaveType.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:             output.LeaveType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		ApprovalFlow:  approvalFlow,
		SubLeaveTypes: subLeaveTypes,
	})
}

// UpdateLeaveType handles PATCH /api/v1/admin/leave-types/{uid}
func (h *LeaveTypeHandler) UpdateLeaveType(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req UpdateLeaveTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("leave_type_handler.UpdateLeaveType.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.UpdateLeaveTypeInput{
		UID:                      uid,
		DefaultBalanceSet:        req.DefaultBalance.Set,
		DefaultBalance:           req.DefaultBalance.Value,
		RecordingDeadlineDaysSet: req.RecordingDeadlineDays.Set,
		RecordingDeadlineDays:    req.RecordingDeadlineDays.Value,
		AdvanceNoticeDaysSet:     req.AdvanceNoticeDays.Set,
		AdvanceNoticeDays:        req.AdvanceNoticeDays.Value,
	}
	output, err := h.updateLeaveTypeUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrLeaveTypeNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, usecases.ErrInvalidLeaveTypeDefaultBalance),
			errors.Is(err, usecases.ErrInvalidLeaveTypeRecordingDeadlineDays),
			errors.Is(err, usecases.ErrInvalidLeaveTypeAdvanceNoticeDays):
			statusCode = http.StatusBadRequest
		default:
			slog.Error("leave_type_handler.UpdateLeaveType.execute_usecase", "error", err, "leave_type_uid", uid)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	writeJSON(w, http.StatusOK, LeaveTypeResponse{
		UID:                   output.LeaveType.UID,
		Code:                  output.LeaveType.Code,
		NameEN:                output.LeaveType.NameEN,
		NameAR:                output.LeaveType.NameAR,
		LocalizedName:         h.localizedLeaveName(r, output.LeaveType.NameEN, output.LeaveType.NameAR),
		DefaultBalance:        output.LeaveType.DefaultBalance,
		MaxConsecutive:        output.LeaveType.MaxConsecutive,
		RecordingDeadlineDays: output.LeaveType.RecordingDeadlineDays,
		AdvanceNoticeDays:     output.LeaveType.AdvanceNoticeDays,
		IsActive:              output.LeaveType.IsActive,
		CreatedAt:             output.LeaveType.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             output.LeaveType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// SetLeaveTypeApprovalFlow handles PATCH /api/v1/admin/leave-types/{uid}/approval-flow
func (h *LeaveTypeHandler) SetLeaveTypeApprovalFlow(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req SetLeaveTypeApprovalFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("leave_type_handler.SetLeaveTypeApprovalFlow.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	output, err := h.setApprovalFlowUC.Execute(r.Context(), usecases.SetLeaveTypeApprovalFlowInput{
		LeaveTypeUID:    uid,
		ApprovalFlowUID: req.ApprovalFlowUID,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, usecases.ErrLeaveTypeNotFound), errors.Is(err, usecases.ErrApprovalFlowNotFound):
			statusCode = http.StatusNotFound
		default:
			slog.Error("leave_type_handler.SetLeaveTypeApprovalFlow.execute_usecase", "error", err, "leave_type_uid", uid)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	details, err := h.getLeaveTypeDetailsUC.Execute(r.Context(), output.LeaveType.UID)
	if err != nil {
		slog.Error("leave_type_handler.SetLeaveTypeApprovalFlow.load_details", "error", err, "leave_type_uid", uid)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	subLeaveTypes := make([]SubLeaveTypeResponse, 0, len(details.SubLeaveTypes))
	for _, item := range details.SubLeaveTypes {
		subLeaveTypes = append(subLeaveTypes, SubLeaveTypeResponse{
			UID:           item.UID,
			LeaveTypeUID:  item.LeaveTypeUID,
			NameEN:        item.NameEN,
			NameAR:        item.NameAR,
			LocalizedName: h.localizedLeaveName(r, item.NameEN, item.NameAR),
		})
	}

	steps := make([]ApprovalFlowStepResponse, 0, len(details.ApprovalFlowSteps))
	for _, item := range details.ApprovalFlowSteps {
		steps = append(steps, ApprovalFlowStepResponse{
			UID:       item.Step.UID,
			FlowUID:   item.Step.ApprovalFlowUID,
			StepOrder: item.Step.StepOrder,
			RoleUID:   item.Step.RoleUID,
			RoleName:  item.RoleName,
			CreatedAt: item.Step.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: item.Step.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	var approvalFlow *LeaveTypeApprovalFlowResponse
	if details.ApprovalFlow != nil {
		approvalFlow = &LeaveTypeApprovalFlowResponse{
			UID:         details.ApprovalFlow.UID,
			Code:        details.ApprovalFlow.Code,
			NameEN:      details.ApprovalFlow.NameEN,
			NameAR:      details.ApprovalFlow.NameAR,
			Description: details.ApprovalFlow.Description,
			IsActive:    details.ApprovalFlow.IsActive,
			CreatedAt:   details.ApprovalFlow.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   details.ApprovalFlow.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Steps:       steps,
		}
	}

	writeJSON(w, http.StatusOK, GetLeaveTypeResponse{
		LeaveType: LeaveTypeResponse{
			UID:                   details.LeaveType.UID,
			Code:                  details.LeaveType.Code,
			NameEN:                details.LeaveType.NameEN,
			NameAR:                details.LeaveType.NameAR,
			LocalizedName:         h.localizedLeaveName(r, details.LeaveType.NameEN, details.LeaveType.NameAR),
			DefaultBalance:        details.LeaveType.DefaultBalance,
			MaxConsecutive:        details.LeaveType.MaxConsecutive,
			RecordingDeadlineDays: details.LeaveType.RecordingDeadlineDays,
			AdvanceNoticeDays:     details.LeaveType.AdvanceNoticeDays,
			IsActive:              details.LeaveType.IsActive,
			CreatedAt:             details.LeaveType.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:             details.LeaveType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		ApprovalFlow:  approvalFlow,
		SubLeaveTypes: subLeaveTypes,
	})
}

// ToggleLeaveType handles PATCH /api/v1/admin/leave-types/{uid}/active
func (h *LeaveTypeHandler) ToggleLeaveType(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	var req ToggleLeaveTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("leave_type_handler.ToggleLeaveType.decode_request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := usecases.ToggleLeaveTypeInput{
		UID:      uid,
		IsActive: req.IsActive,
	}
	output, err := h.toggleLeaveTypeUC.Execute(r.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrLeaveTypeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("leave_type_handler.ToggleLeaveType.execute_usecase", "error", err)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	writeJSON(w, http.StatusOK, LeaveTypeResponse{
		UID:                   output.LeaveType.UID,
		Code:                  output.LeaveType.Code,
		NameEN:                output.LeaveType.NameEN,
		NameAR:                output.LeaveType.NameAR,
		LocalizedName:         h.localizedLeaveName(r, output.LeaveType.NameEN, output.LeaveType.NameAR),
		DefaultBalance:        output.LeaveType.DefaultBalance,
		MaxConsecutive:        output.LeaveType.MaxConsecutive,
		RecordingDeadlineDays: output.LeaveType.RecordingDeadlineDays,
		AdvanceNoticeDays:     output.LeaveType.AdvanceNoticeDays,
		IsActive:              output.LeaveType.IsActive,
		CreatedAt:             output.LeaveType.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:             output.LeaveType.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListSubLeaveTypes handles GET /api/v1/leave-types/{uid}/sub-leave-types
func (h *LeaveTypeHandler) ListSubLeaveTypes(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	if uid == "" {
		writeError(w, http.StatusBadRequest, "uid is required")
		return
	}

	output, err := h.listSubLeaveTypesUC.Execute(r.Context(), uid)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, usecases.ErrLeaveTypeNotFound) {
			statusCode = http.StatusNotFound
		} else {
			slog.Error("leave_type_handler.ListSubLeaveTypes.execute_usecase", "error", err, "leave_type_uid", uid)
		}
		writeLocalizedError(w, statusCode, err, h.i18nService, r.Context())
		return
	}

	subLeaveTypes := make([]SubLeaveTypeResponse, 0, len(output.SubLeaveTypes))
	for _, item := range output.SubLeaveTypes {
		subLeaveTypes = append(subLeaveTypes, SubLeaveTypeResponse{
			UID:           item.UID,
			LeaveTypeUID:  item.LeaveTypeUID,
			NameEN:        item.NameEN,
			NameAR:        item.NameAR,
			LocalizedName: h.localizedLeaveName(r, item.NameEN, item.NameAR),
		})
	}

	writeJSON(w, http.StatusOK, ListSubLeaveTypesResponse{SubLeaveTypes: subLeaveTypes})
}
