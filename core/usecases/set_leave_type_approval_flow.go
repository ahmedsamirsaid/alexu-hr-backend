package usecases

import (
	"context"

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
}

func NewSetLeaveTypeApprovalFlowUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
	approvalFlowRepo ports.ApprovalFlowRepository,
) *SetLeaveTypeApprovalFlowUseCase {
	return &SetLeaveTypeApprovalFlowUseCase{
		db:               db,
		leaveTypeRepo:    leaveTypeRepo,
		approvalFlowRepo: approvalFlowRepo,
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

	return &SetLeaveTypeApprovalFlowOutput{
		LeaveType:    leaveType,
		ApprovalFlow: approvalFlow,
	}, nil
}
