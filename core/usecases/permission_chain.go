package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

func resolvePermissionFirstStep(
	ctx context.Context,
	q ports.Querier,
	roleRepo ports.RoleRepository,
	steps []*domain.ApprovalFlowStep,
	requester *domain.Employee,
	requesterUserUID string,
) (int, error) {
	if len(steps) == 0 {
		return 0, ErrApprovalFlowHasNoSteps
	}

	for _, step := range steps {
		var deptUID *string
		if requester != nil && requester.DepartmentUID != nil {
			deptUID = requester.DepartmentUID
		}
		users, err := roleRepo.GetUsersByRoleAndDepartment(ctx, q, step.RoleUID, deptUID)
		if err != nil {
			return 0, err
		}

		hasOther := false
		for _, u := range users {
			if u.UID != requesterUserUID {
				hasOther = true
				break
			}
		}
		if hasOther {
			return step.StepOrder, nil
		}
	}

	return steps[len(steps)-1].StepOrder, nil
}
