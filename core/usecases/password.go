package usecases

import (
	"github.com/banumusa/backend/core/domain"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword verifies the password against the user's stored hash
func CheckPassword(user *domain.User, password string) bool {
	if user.PasswordHash == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	return err == nil
}

// SetUserPassword hashes and sets the password on the user
func SetUserPassword(user *domain.User, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	user.PasswordHash = &hash
	return nil
}
