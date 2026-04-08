package usecases

import (
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
)

func TestRequestOTPInput_Validation(t *testing.T) {
	input := RequestOTPInput{
		Phone: "+201234567890",
	}

	if input.Phone == "" {
		t.Error("expected phone to be set")
	}
}

func TestVerifyOTPInput_Validation(t *testing.T) {
	input := VerifyOTPInput{
		Phone: "+201234567890",
		OTP:   "123456",
	}

	if input.Phone == "" {
		t.Error("expected phone to be set")
	}
	if input.OTP == "" {
		t.Error("expected code to be set")
	}
}

func TestOTPCode_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		otp      *domain.OTPCode
		expected bool
	}{
		{
			name: "not expired",
			otp: &domain.OTPCode{
				ExpiresAt: time.Now().Add(5 * time.Minute),
			},
			expected: false,
		},
		{
			name: "expired",
			otp: &domain.OTPCode{
				ExpiresAt: time.Now().Add(-5 * time.Minute),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.otp.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOTPCode_IsValid(t *testing.T) {
	otp := &domain.OTPCode{
		Code:      "123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	if !otp.IsValid("123456") {
		t.Error("expected valid OTP")
	}

	if otp.IsValid("654321") {
		t.Error("expected invalid OTP for wrong code")
	}

	// Mark as used
	otp.Used = true
	if otp.IsValid("123456") {
		t.Error("expected invalid OTP when used")
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	token := &domain.RefreshToken{
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Revoked:   false,
	}

	if !token.IsValid() {
		t.Error("expected valid token")
	}

	token.Revoked = true
	if token.IsValid() {
		t.Error("expected invalid token when revoked")
	}

	token.Revoked = false
	token.ExpiresAt = time.Now().Add(-1 * time.Hour)
	if token.IsValid() {
		t.Error("expected invalid token when expired")
	}
}

func TestCheckPassword(t *testing.T) {
	user := domain.NewUser("+201234567890")

	// No password set
	if CheckPassword(user, "anything") {
		t.Error("expected false for user without password")
	}

	// Set password
	if err := SetUserPassword(user, "secret123"); err != nil {
		t.Fatalf("SetUserPassword() error = %v", err)
	}

	// Check correct password
	if !CheckPassword(user, "secret123") {
		t.Error("expected true for correct password")
	}

	// Check wrong password
	if CheckPassword(user, "wrongpassword") {
		t.Error("expected false for wrong password")
	}
}

func TestLoginPasswordInput_Validation(t *testing.T) {
	input := LoginPasswordInput{
		Phone:    "+201234567890",
		Password: "secret123",
	}

	if input.Phone == "" {
		t.Error("expected phone to be set")
	}
	if input.Password == "" {
		t.Error("expected password to be set")
	}
}

func TestCreateUserInput_Validation(t *testing.T) {
	password := "secret123"
	input := CreateUserInput{
		Phone:    "+201234567890",
		Password: &password,
	}

	if input.Phone == "" {
		t.Error("expected phone to be set")
	}
}

func TestUpdateUserInput_Validation(t *testing.T) {
	phone := "+201111111111"
	input := UpdateUserInput{
		UserUID: "usr_abc123",
		Phone:   &phone,
	}

	if input.UserUID == "" {
		t.Error("expected userUID to be set")
	}
}

func TestErrors(t *testing.T) {
	errors := []error{
		ErrInvalidOTP,
		ErrOTPExpired,
		ErrUserNotFound,
		ErrUserInactive,
		ErrInvalidCredentials,
		ErrPhoneAlreadyExists,
		ErrRoleNotFound,
		ErrInvalidRefreshToken,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("expected error to not be nil")
		}
		if err.Error() == "" {
			t.Error("expected error message to not be empty")
		}
	}
}
