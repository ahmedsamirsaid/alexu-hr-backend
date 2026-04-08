package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestRefreshTokenRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewRefreshTokenRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	// Create user first
	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tokenWithRaw := domain.NewRefreshToken(user.ID, 90)

	err := repo.Create(ctx, tdb.SQLiteDB, &tokenWithRaw.RefreshToken)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if tokenWithRaw.ID == 0 {
		t.Error("expected token ID to be set")
	}
}

func TestRefreshTokenRepository_GetByHash(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewRefreshTokenRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tokenWithRaw := domain.NewRefreshToken(user.ID, 90)
	if err := repo.Create(ctx, tdb.SQLiteDB, &tokenWithRaw.RefreshToken); err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Get by hash
	found, err := repo.GetByHash(ctx, tdb.SQLiteDB, tokenWithRaw.TokenHash)
	if err != nil {
		t.Fatalf("GetByHash() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected to find token")
	}
	if found.UserID != user.ID {
		t.Errorf("UserID = %v, want %v", found.UserID, user.ID)
	}

	// Non-existent hash
	notFound, err := repo.GetByHash(ctx, tdb.SQLiteDB, "nonexistent")
	if err != nil {
		t.Fatalf("GetByHash() error = %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent hash")
	}
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewRefreshTokenRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tokenWithRaw := domain.NewRefreshToken(user.ID, 90)
	if err := repo.Create(ctx, tdb.SQLiteDB, &tokenWithRaw.RefreshToken); err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	if err := repo.Revoke(ctx, tdb.SQLiteDB, tokenWithRaw.ID); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	// Verify
	found, _ := repo.GetByHash(ctx, tdb.SQLiteDB, tokenWithRaw.TokenHash)
	if !found.Revoked {
		t.Error("expected token to be revoked")
	}
}

func TestRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewRefreshTokenRepository()
	userRepo := NewUserRepository()
	ctx := context.Background()

	user := domain.NewUser("+201234567890")
	if err := userRepo.Create(ctx, tdb.SQLiteDB, user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create multiple tokens
	var hashes []string
	for i := 0; i < 3; i++ {
		tokenWithRaw := domain.NewRefreshToken(user.ID, 90)
		if err := repo.Create(ctx, tdb.SQLiteDB, &tokenWithRaw.RefreshToken); err != nil {
			t.Fatalf("failed to create token: %v", err)
		}
		hashes = append(hashes, tokenWithRaw.TokenHash)
	}

	if err := repo.RevokeAllForUser(ctx, tdb.SQLiteDB, user.ID); err != nil {
		t.Fatalf("RevokeAllForUser() error = %v", err)
	}

	// Verify all revoked
	for _, hash := range hashes {
		found, _ := repo.GetByHash(ctx, tdb.SQLiteDB, hash)
		if !found.Revoked {
			t.Errorf("expected token %s to be revoked", hash)
		}
	}
}
