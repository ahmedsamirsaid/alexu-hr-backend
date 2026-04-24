package domain

import (
	"strings"
	"time"
)

type Role struct {
	ID          int64
	UID         string
	Name        string
	Description string
	ScopeType   string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Permissions []Permission
}

const (
	RoleScopeGlobal     = "global"
	RoleScopeDepartment = "department"
	RoleScopeSelf       = "self"
)

func IsValidRoleScopeType(scopeType string) bool {
	switch strings.ToLower(strings.TrimSpace(scopeType)) {
	case RoleScopeGlobal, RoleScopeDepartment, RoleScopeSelf:
		return true
	default:
		return false
	}
}

func NormalizeRoleScopeType(scopeType string) string {
	normalized := strings.ToLower(strings.TrimSpace(scopeType))
	if normalized == "" {
		return RoleScopeGlobal
	}
	if !IsValidRoleScopeType(normalized) {
		return ""
	}
	return normalized
}

func RoleScopePriority(scopeType string) int {
	switch NormalizeRoleScopeType(scopeType) {
	case RoleScopeSelf:
		return 1
	case RoleScopeDepartment:
		return 2
	case RoleScopeGlobal:
		return 3
	default:
		return 0
	}
}

func NewRole(name, description string, scopeType ...string) *Role {
	normalizedScopeType := RoleScopeGlobal
	if len(scopeType) > 0 {
		if normalized := NormalizeRoleScopeType(scopeType[0]); normalized != "" {
			normalizedScopeType = normalized
		}
	}

	return &Role{
		UID:         GenerateUID("role"),
		Name:        name,
		Description: description,
		ScopeType:   normalizedScopeType,
		IsSystem:    false,
	}
}

// HasAllPermissions returns true only for privileged system roles.
// Not every system role should bypass RBAC checks.
func (r Role) HasAllPermissions() bool {
	if !r.IsSystem {
		return false
	}
	return r.UID == "role_admin" || strings.EqualFold(r.Name, "Admin")
}
