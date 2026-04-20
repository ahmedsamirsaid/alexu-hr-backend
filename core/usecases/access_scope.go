package usecases

import "github.com/banumusa/backend/core/domain"

func determineAccessScope(roles []domain.Role) string {
	hasDepartmentScope := false

	for _, role := range roles {
		switch domain.NormalizeRoleScopeType(role.ScopeType) {
		case domain.RoleScopeGlobal:
			return domain.RoleScopeGlobal
		case domain.RoleScopeDepartment:
			hasDepartmentScope = true
		}
	}

	if hasDepartmentScope {
		return domain.RoleScopeDepartment
	}

	return domain.RoleScopeSelf
}
