package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

func canSubmitOrViewEmployeeProfileChanges(ctx context.Context, q ports.Querier, roleRepo ports.RoleRepository, userID int64) (bool, error) {
	roles, err := roleRepo.GetRolesForUser(ctx, q, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		if role != nil && (role.UID == hrStaffRoleUID || role.UID == informationCenterRoleUID) {
			return true, nil
		}
	}

	return false, nil
}
