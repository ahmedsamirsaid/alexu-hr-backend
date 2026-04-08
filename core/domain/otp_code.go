package domain

import (
	"crypto/rand"
	"fmt"
	"time"
)

const OTPExpiryMinutes = 5

type OTPCode struct {
	ID        int64
	Phone     string
	Code      string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

func NewOTPCode(phone string) *OTPCode {
	return &OTPCode{
		Phone:     phone,
		Code:      generateOTPCode(),
		ExpiresAt: time.Now().Add(OTPExpiryMinutes * time.Minute),
		Used:      false,
	}
}

func generateOTPCode() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	// Convert to 6 digits
	num := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", num%1000000)
}

func (o *OTPCode) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

func (o *OTPCode) IsValid(otp string) bool {
	return !o.Used && !o.IsExpired() && o.Code == otp
}
