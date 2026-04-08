package domain

import "time"

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
