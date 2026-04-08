package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestOTPRepository_Create(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewOTPRepository()
	ctx := context.Background()

	otp := domain.NewOTPCode("+201234567890")

	err := repo.Create(ctx, tdb.SQLiteDB, otp)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if otp.ID == 0 {
		t.Error("expected OTP ID to be set")
	}
}

func TestOTPRepository_GetLatestByPhone(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewOTPRepository()
	ctx := context.Background()

	phone := "+201234567890"

	// Create multiple OTPs
	otp1 := domain.NewOTPCode(phone)
	otp1.Code = "111111"
	if err := repo.Create(ctx, tdb.SQLiteDB, otp1); err != nil {
		t.Fatalf("failed to create OTP: %v", err)
	}

	otp2 := domain.NewOTPCode(phone)
	otp2.Code = "222222"
	if err := repo.Create(ctx, tdb.SQLiteDB, otp2); err != nil {
		t.Fatalf("failed to create OTP: %v", err)
	}

	// Get latest
	latest, err := repo.GetLatestByPhone(ctx, tdb.SQLiteDB, phone)
	if err != nil {
		t.Fatalf("GetLatestByPhone() error = %v", err)
	}
	if latest == nil {
		t.Fatal("expected to find OTP")
	}
	if latest.Code != "222222" {
		t.Errorf("Code = %v, want 222222", latest.Code)
	}
}

func TestOTPRepository_MarkUsed(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewOTPRepository()
	ctx := context.Background()

	otp := domain.NewOTPCode("+201234567890")
	if err := repo.Create(ctx, tdb.SQLiteDB, otp); err != nil {
		t.Fatalf("failed to create OTP: %v", err)
	}

	if err := repo.MarkUsed(ctx, tdb.SQLiteDB, otp.ID); err != nil {
		t.Fatalf("MarkUsed() error = %v", err)
	}

	// Verify
	latest, _ := repo.GetLatestByPhone(ctx, tdb.SQLiteDB, otp.Phone)
	if !latest.Used {
		t.Error("expected OTP to be marked as used")
	}
}
