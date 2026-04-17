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
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Permissions []Permission
}

func NewRole(name, description string) *Role {
	return &Role{
		UID:         GenerateUID("role"),
		Name:        name,
		Description: description,
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
