package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestUserRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")

	err := repo.Create(ctx, tdb.SQLiteDB, user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
}

func TestUserRepository_GetByPhone(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewUserRepository()
	ctx := context.Background()

	// Create user
	user := domain.NewUser("+201234567890")
	if err := repo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Get by phone
	found, err := repo.GetByPhone(ctx, tdb.SQLiteDB, "+201234567890")
	if err != nil {
		t.Fatalf("GetByPhone() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected to find user")
	}
	if found.UID != user.UID {
		t.Errorf("UID = %v, want %v", found.UID, user.UID)
	}

	// Non-existent phone
	notFound, err := repo.GetByPhone(ctx, tdb.SQLiteDB, "+201111111111")
	if err != nil {
		t.Fatalf("GetByPhone() error = %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent phone")
	}
}

func TestUserRepository_GetByUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")
	if err := repo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	found, err := repo.GetByUID(ctx, tdb.SQLiteDB, user.UID)
	if err != nil {
		t.Fatalf("GetByUID() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected to find user")
	}
	if found.Phone != user.Phone {
		t.Errorf("Phone = %v, want %v", found.Phone, user.Phone)
	}
}

func TestUserRepository_Update(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")
	if err := repo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Update phone
	user.Phone = "+201111111111"
	user.IsActive = false
	if err := repo.Update(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	found, _ := repo.GetByID(ctx, tdb.SQLiteDB, user.ID)
	if found.Phone != "+201111111111" {
		t.Errorf("Phone = %v, want +201111111111", found.Phone)
	}
	if found.IsActive != false {
		t.Error("expected IsActive = false")
	}
}

func TestUserRepository_List(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewUserRepository()
	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := domain.NewUser("+2012345678" + string(rune('0'+i)))
		if err := repo.Create(ctx, tdb.SQLiteDB, user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	// List with pagination
	users, err := repo.List(ctx, tdb.SQLiteDB, 3, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 3 {
		t.Errorf("got %d users, want 3", len(users))
	}

	// Page 2
	users, err = repo.List(ctx, tdb.SQLiteDB, 3, 3)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 2 {
		t.Errorf("got %d users, want 2", len(users))
	}
}

func TestUserRepository_Count(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewUserRepository()
	ctx := context.Background()

	// Create users
	for i := 0; i < 3; i++ {
		user := domain.NewUser("+2012345678" + string(rune('0'+i)))
		if err := repo.Create(ctx, tdb.SQLiteDB, user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	count, err := repo.Count(ctx, tdb.SQLiteDB)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d, want 3", count)
	}
}
