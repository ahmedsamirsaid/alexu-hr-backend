package usecases

import (
	"context"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type SetLeaveTypeApprovalFlowInput struct {
	LeaveTypeUID    string
	ApprovalFlowUID *string
}

type SetLeaveTypeApprovalFlowOutput struct {
	LeaveType    *domain.LeaveType
	ApprovalFlow *domain.ApprovalFlow
}

type SetLeaveTypeApprovalFlowUseCase struct {
	db               ports.DB
	leaveTypeRepo    ports.LeaveTypeRepository
	approvalFlowRepo ports.ApprovalFlowRepository
	auditor          audit.Auditor
}

func NewSetLeaveTypeApprovalFlowUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
	approvalFlowRepo ports.ApprovalFlowRepository,
	auditor audit.Auditor,
) *SetLeaveTypeApprovalFlowUseCase {
	return &SetLeaveTypeApprovalFlowUseCase{
		db:               db,
		leaveTypeRepo:    leaveTypeRepo,
		approvalFlowRepo: approvalFlowRepo,
		auditor:          auditor,
	}
}

func (uc *SetLeaveTypeApprovalFlowUseCase) Execute(ctx context.Context, input SetLeaveTypeApprovalFlowInput) (*SetLeaveTypeApprovalFlowOutput, error) {
	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, input.LeaveTypeUID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}

	// Store old approval flow UID for audit log
	oldApprovalFlowUID := leaveType.ApprovalFlowUID

	var approvalFlow *domain.ApprovalFlow
	if input.ApprovalFlowUID != nil && *input.ApprovalFlowUID != "" {
		approvalFlow, err = uc.approvalFlowRepo.GetByUID(ctx, uc.db, *input.ApprovalFlowUID)
		if err != nil {
			return nil, err
		}
		if approvalFlow == nil {
			return nil, ErrApprovalFlowNotFound
		}
		leaveType.ApprovalFlowUID = &approvalFlow.UID
	} else {
		leaveType.ApprovalFlowUID = nil
	}

	if err := uc.leaveTypeRepo.Update(ctx, uc.db, leaveType); err != nil {
		return nil, err
	}

	// Audit log after successful update
	actorName := audit.ActorFromContext(ctx)
	var actionKey string
	if leaveType.ApprovalFlowUID != nil {
		if oldApprovalFlowUID != nil {
			actionKey = "audit.sentence.change_leave_type_flow"
		} else {
			actionKey = "audit.sentence.assign_leave_type_flow"
		}
	} else {
		actionKey = "audit.sentence.remove_leave_type_flow"
	}

	actionParams := map[string]interface{}{
		"Actor":     actorName,
		"LeaveType": leaveType.NameEN,
	}

	auditBuilder := uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityLeaveType, input.LeaveTypeUID).
		WithMeta("action_key", actionKey).
		WithMeta("action_params", actionParams).
		WithMeta("leave_type_name", leaveType.NameEN)

	if approvalFlow != nil {
		auditBuilder = auditBuilder.WithMeta("approval_flow_name", approvalFlow.NameEN)
	}

	defer auditBuilder.Save(ctx)

	return &SetLeaveTypeApprovalFlowOutput{
		LeaveType:    leaveType,
		ApprovalFlow: approvalFlow,
	}, nil
}
