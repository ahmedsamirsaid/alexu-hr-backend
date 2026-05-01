package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetLeaveTypeDetailsOutput struct {
	LeaveType         *domain.LeaveType
	ApprovalFlow      *domain.ApprovalFlow
	ApprovalFlowSteps []ApprovalFlowStepDetails
	SubLeaveTypes     []*domain.SubLeaveType
}

type GetLeaveTypeDetailsUseCase struct {
	db               ports.DB
	leaveTypeRepo    ports.LeaveTypeRepository
	approvalFlowRepo ports.ApprovalFlowRepository
	stepRepo         ports.ApprovalFlowStepRepository
	roleRepo         ports.RoleRepository
}

func NewGetLeaveTypeDetailsUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
	approvalFlowRepo ports.ApprovalFlowRepository,
	stepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
) *GetLeaveTypeDetailsUseCase {
	return &GetLeaveTypeDetailsUseCase{
		db:               db,
		leaveTypeRepo:    leaveTypeRepo,
		approvalFlowRepo: approvalFlowRepo,
		stepRepo:         stepRepo,
		roleRepo:         roleRepo,
	}
}

func (uc *GetLeaveTypeDetailsUseCase) Execute(ctx context.Context, leaveTypeUID string) (*GetLeaveTypeDetailsOutput, error) {
	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, leaveTypeUID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}

	var approvalFlow *domain.ApprovalFlow
	var approvalFlowSteps []ApprovalFlowStepDetails
	if leaveType.ApprovalFlowUID != nil && *leaveType.ApprovalFlowUID != "" {
		approvalFlow, err = uc.approvalFlowRepo.GetByUID(ctx, uc.db, *leaveType.ApprovalFlowUID)
		if err != nil {
			return nil, err
		}
		if approvalFlow != nil {
			steps, err := uc.stepRepo.ListByFlow(ctx, uc.db, approvalFlow.UID)
			if err != nil {
				return nil, err
			}
			approvalFlowSteps = make([]ApprovalFlowStepDetails, 0, len(steps))
			for _, step := range steps {
				detail := ApprovalFlowStepDetails{Step: step}
				role, err := uc.roleRepo.GetByUID(ctx, uc.db, step.RoleUID)
				if err != nil {
					return nil, err
				}
				if role != nil {
					detail.RoleName = role.Name
				}
				approvalFlowSteps = append(approvalFlowSteps, detail)
			}
		}
	}

	subLeaveTypes, err := uc.leaveTypeRepo.ListSubLeaveTypesByLeaveTypeUID(ctx, uc.db, leaveTypeUID)
	if err != nil {
		return nil, err
	}

	return &GetLeaveTypeDetailsOutput{
		LeaveType:         leaveType,
		ApprovalFlow:      approvalFlow,
		ApprovalFlowSteps: approvalFlowSteps,
		SubLeaveTypes:     subLeaveTypes,
	}, nil
}
