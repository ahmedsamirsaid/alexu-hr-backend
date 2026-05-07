package domain

import "time"

type User struct {
	ID                    int64
	UID                   string
	Phone                 string
	PasswordHash          *string
	EmployeeUID           *string
	ManagedDepartmentUIDs []string
	IsActive              bool
	// PreferredLanguage is the user's preferred locale (e.g. "ar", "en").
	// An empty value defaults to "ar" at the notification-service layer.
	PreferredLanguage     string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Roles                 []Role
}

func NewUser(phone string) *User {
	return &User{
		UID:      GenerateUID("usr"),
		Phone:    phone,
		IsActive: true,
	}
}

func (u *User) HasPermission(code string) bool {
	for _, role := range u.Roles {
		if role.HasAllPermissions() {
			return true // Admin has all permissions
		}
		for _, perm := range role.Permissions {
			if perm.Code == code {
				return true
			}
		}
	}
	return false
}

func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

func (u *User) IsAdmin() bool {
	return u.HasRole("admin")
}

// HasWebPortalAccess returns true if user has any role other than "Employee".
// Users with only the Employee role should use the mobile app instead.
func (u *User) HasWebPortalAccess() bool {
	for _, role := range u.Roles {
		if role.Name != "Employee" {
			return true
		}
	}
	return false
}
