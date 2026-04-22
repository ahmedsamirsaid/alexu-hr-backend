package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ApprovalFlowStepDetails struct {
	Step     *domain.ApprovalFlowStep
	RoleName string
}

type GetApprovalFlowOutput struct {
	Flow  *domain.ApprovalFlow
	Steps []ApprovalFlowStepDetails
}

type GetApprovalFlowUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
	stepRepo ports.ApprovalFlowStepRepository
	roleRepo ports.RoleRepository
}

func NewGetApprovalFlowUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
	stepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
) *GetApprovalFlowUseCase {
	return &GetApprovalFlowUseCase{
		db:       db,
		flowRepo: flowRepo,
		stepRepo: stepRepo,
		roleRepo: roleRepo,
	}
}

func (uc *GetApprovalFlowUseCase) Execute(ctx context.Context, flowUID string) (*GetApprovalFlowOutput, error) {
	flow, err := uc.flowRepo.GetByUID(ctx, uc.db, flowUID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrApprovalFlowNotFound
	}

	steps, err := uc.stepRepo.ListByFlow(ctx, uc.db, flowUID)
	if err != nil {
		return nil, err
	}

	stepDetails := make([]ApprovalFlowStepDetails, 0, len(steps))
	for _, step := range steps {
		detail := ApprovalFlowStepDetails{Step: step}
		role, err := uc.roleRepo.GetByUID(ctx, uc.db, step.RoleUID)
		if err != nil {
			return nil, err
		}
		if role != nil {
			detail.RoleName = role.Name
		}
		stepDetails = append(stepDetails, detail)
	}

	return &GetApprovalFlowOutput{
		Flow:  flow,
		Steps: stepDetails,
	}, nil
}
