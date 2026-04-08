package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestRoleRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewRoleRepository()
	ctx := context.Background()

	role := domain.NewRole("hr_manager", "HR Department Manager")

	err := repo.Create(ctx, tdb.SQLiteDB, role)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if role.ID == 0 {
		t.Error("expected role ID to be set")
	}
}

func TestRoleRepository_GetByName(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewRoleRepository()
	ctx := context.Background()

	role := domain.NewRole("hr_manager", "HR Manager")
	if err := repo.Create(ctx, tdb.SQLiteDB, role); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	found, err := repo.GetByName(ctx, tdb.SQLiteDB, "hr_manager")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected to find role")
	}
	if found.UID != role.UID {
		t.Errorf("UID = %v, want %v", found.UID, role.UID)
	}
}

func TestRoleRepository_AssignRoleToUser(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	roleRepo := NewRoleRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	// Create user and role
	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	role := domain.NewRole("test_role", "Test Role")
	if err := roleRepo.Create(ctx, tdb.SQLiteDB, role); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Assign role
	if err := roleRepo.AssignRoleToUser(ctx, tdb.SQLiteDB, user.ID, role.ID); err != nil {
		t.Fatalf("AssignRoleToUser() error = %v", err)
	}

	// Verify by checking user_roles table directly
	var count int
	row := tdb.SQLiteDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_roles WHERE user_id = ? AND role_id = ?", user.ID, role.ID)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to count user_roles: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user_role entry, got %d", count)
	}
}

func TestRoleRepository_RemoveRoleFromUser(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	roleRepo := NewRoleRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	role := domain.NewRole("test_role", "Test Role")
	if err := roleRepo.Create(ctx, tdb.SQLiteDB, role); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Assign and then remove
	if err := roleRepo.AssignRoleToUser(ctx, tdb.SQLiteDB, user.ID, role.ID); err != nil {
		t.Fatalf("AssignRoleToUser() error = %v", err)
	}
	if err := roleRepo.RemoveRoleFromUser(ctx, tdb.SQLiteDB, user.ID, role.ID); err != nil {
		t.Fatalf("RemoveRoleFromUser() error = %v", err)
	}

	roles, _ := roleRepo.GetRolesForUser(ctx, tdb.SQLiteDB, user.ID)
	if len(roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(roles))
	}
}

func TestRoleRepository_GetRolesForUser(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	roleRepo := NewRoleRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	// Create user and role
	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	role := domain.NewRole("test_role", "Test Role")
	if err := roleRepo.Create(ctx, tdb.SQLiteDB, role); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Assign role
	if err := roleRepo.AssignRoleToUser(ctx, tdb.SQLiteDB, user.ID, role.ID); err != nil {
		t.Fatalf("AssignRoleToUser() error = %v", err)
	}

	// Verify via GetRolesForUser
	roles, err := roleRepo.GetRolesForUser(ctx, tdb.SQLiteDB, user.ID)
	if err != nil {
		t.Fatalf("GetRolesForUser() error = %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(roles))
	}
	if roles[0].Name != "test_role" {
		t.Errorf("role name = %v, want test_role", roles[0].Name)
	}
}

func TestRoleRepository_SetRolePermissions(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	roleRepo := NewRoleRepository()
	ctx := context.Background()

	// Create role
	role := domain.NewRole("test_role", "Test Role")
	if err := roleRepo.Create(ctx, tdb.SQLiteDB, role); err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create permissions
	perm1 := tdb.SeedPermission("test:read", "Read test")
	perm2 := tdb.SeedPermission("test:write", "Write test")

	// Set permissions
	if err := roleRepo.SetRolePermissions(ctx, tdb.SQLiteDB, role.ID, []int64{perm1.ID, perm2.ID}); err != nil {
		t.Fatalf("SetRolePermissions() error = %v", err)
	}

	// Verify
	permRepo := NewPermissionRepository()
	perms, err := permRepo.GetPermissionsForRole(ctx, tdb.SQLiteDB, role.ID)
	if err != nil {
		t.Fatalf("GetPermissionsForRole() error = %v", err)
	}
	if len(perms) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(perms))
	}

	// Update permissions (remove one)
	if err := roleRepo.SetRolePermissions(ctx, tdb.SQLiteDB, role.ID, []int64{perm1.ID}); err != nil {
		t.Fatalf("SetRolePermissions() error = %v", err)
	}

	perms, _ = permRepo.GetPermissionsForRole(ctx, tdb.SQLiteDB, role.ID)
	if len(perms) != 1 {
		t.Errorf("expected 1 permission, got %d", len(perms))
	}
}
