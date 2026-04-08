package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

const DefaultRefreshTokenDays = 90

type RefreshToken struct {
	ID        int64
	TokenHash string
	UserID    int64
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

type RefreshTokenWithRaw struct {
	RefreshToken
	RawToken string
}

func NewRefreshToken(userID int64, expiryDays int) *RefreshTokenWithRaw {
	if expiryDays <= 0 {
		expiryDays = DefaultRefreshTokenDays
	}

	rawToken := generateRandomToken()
	hash := hashToken(rawToken)

	return &RefreshTokenWithRaw{
		RefreshToken: RefreshToken{
			TokenHash: hash,
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour),
			Revoked:   false,
		},
		RawToken: rawToken,
	}
}

func generateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func HashToken(token string) string {
	return hashToken(token)
}

func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *RefreshToken) IsValid() bool {
	return !r.Revoked && !r.IsExpired()
}
