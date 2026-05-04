package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type effectiveApprovalChain struct {
	requester       *domain.Employee
	effectiveSteps  []*domain.ApprovalFlowStep
	matchedFlowStep *domain.ApprovalFlowStep
	allSteps        []*domain.ApprovalFlowStep
}

type approvalChainResolver struct {
	employeeRepo         ports.EmployeeRepository
	userRepo             ports.UserRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	roleRepo             ports.RoleRepository
}

func newApprovalChainResolver(
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
) *approvalChainResolver {
	return &approvalChainResolver{
		employeeRepo:         employeeRepo,
		userRepo:             userRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		roleRepo:             roleRepo,
	}
}

func (r *approvalChainResolver) resolveForSubmit(ctx context.Context, q ports.Querier, approvalFlowUID string, requester *domain.Employee) (*effectiveApprovalChain, error) {
	steps, err := r.approvalFlowStepRepo.ListByFlow(ctx, q, approvalFlowUID)
	if err != nil {
		return nil, err
	}

	var matchedStep *domain.ApprovalFlowStep
	if requester != nil {
		requesterUser, err := r.userRepo.GetByEmployeeUID(ctx, q, requester.UID)
		if err != nil {
			return nil, err
		}
		if requesterUser != nil {
			matchedStep, err = r.findHighestMatchingFlowStep(ctx, q, requesterUser.ID, requester, steps)
			if err != nil {
				return nil, err
			}
		}
	}

	return &effectiveApprovalChain{
		requester:       requester,
		effectiveSteps:  filterEffectiveApprovalSteps(steps, matchedStep),
		matchedFlowStep: matchedStep,
		allSteps:        steps,
	}, nil
}

func (r *approvalChainResolver) resolveForApprovalRequest(ctx context.Context, q ports.Querier, approvalRequest *domain.ApprovalRequest) (*effectiveApprovalChain, error) {
	requester, err := r.employeeRepo.GetByUID(ctx, q, approvalRequest.RequesterUID)
	if err != nil {
		return nil, err
	}

	steps, err := r.approvalFlowStepRepo.ListByFlow(ctx, q, approvalRequest.ApprovalFlowUID)
	if err != nil {
		return nil, err
	}

	if requester == nil || requester.DepartmentUID == nil {
		return &effectiveApprovalChain{requester: requester, effectiveSteps: steps, allSteps: steps}, nil
	}

	requesterUser, err := r.userRepo.GetByEmployeeUID(ctx, q, requester.UID)
	if err != nil {
		return nil, err
	}
	if requesterUser == nil {
		return &effectiveApprovalChain{requester: requester, effectiveSteps: steps, allSteps: steps}, nil
	}

	matchedStep, err := r.findHighestMatchingFlowStep(ctx, q, requesterUser.ID, requester, steps)
	if err != nil {
		return nil, err
	}

	return &effectiveApprovalChain{
		requester:       requester,
		effectiveSteps:  filterEffectiveApprovalSteps(steps, matchedStep),
		matchedFlowStep: matchedStep,
		allSteps:        steps,
	}, nil
}

func (r *approvalChainResolver) findHighestMatchingFlowStep(ctx context.Context, q ports.Querier, userID int64, requester *domain.Employee, steps []*domain.ApprovalFlowStep) (*domain.ApprovalFlowStep, error) {
	if requester == nil || requester.DepartmentUID == nil {
		return nil, nil
	}

	for _, step := range steps {
		ok, err := r.roleRepo.IsUserAuthorizedApprover(ctx, q, userID, step.RoleUID, *requester.DepartmentUID)
		if err != nil {
			return nil, err
		}
		if ok {
			matched := step
			for _, nextStep := range steps {
				if nextStep.StepOrder <= matched.StepOrder {
					continue
				}
				ok, err := r.roleRepo.IsUserAuthorizedApprover(ctx, q, userID, nextStep.RoleUID, *requester.DepartmentUID)
				if err != nil {
					return nil, err
				}
				if ok {
					matched = nextStep
				}
			}
			return matched, nil
		}
	}

	return nil, nil
}

func filterEffectiveApprovalSteps(steps []*domain.ApprovalFlowStep, matchedStep *domain.ApprovalFlowStep) []*domain.ApprovalFlowStep {
	if matchedStep == nil {
		return steps
	}

	effective := make([]*domain.ApprovalFlowStep, 0, len(steps))
	for _, step := range steps {
		if step.StepOrder > matchedStep.StepOrder {
			effective = append(effective, step)
		}
	}
	return effective
}

func effectiveApprovalStepAt(chain *effectiveApprovalChain, logicalStep int) *domain.ApprovalFlowStep {
	if chain == nil || logicalStep <= 0 || logicalStep > len(chain.effectiveSteps) {
		return nil
	}
	return chain.effectiveSteps[logicalStep-1]
}
