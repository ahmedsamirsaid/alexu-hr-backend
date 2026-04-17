package http

import (
	"errors"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type JWTClaims struct {
	UserID                int64    `json:"userId"`
	UserUID               string   `json:"userUid"`
	Roles                 []string `json:"roles"`
	Permissions           []string `json:"permissions"`
	ManagedDepartmentUIDs []string `json:"managedDepartmentUids"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey      []byte
	accessTokenTTL time.Duration
}

func NewJWTService(secretKey string, accessTokenMinutes int) *JWTService {
	if accessTokenMinutes <= 0 {
		accessTokenMinutes = 60
	}

	return &JWTService{
		secretKey:      []byte(secretKey),
		accessTokenTTL: time.Duration(accessTokenMinutes) * time.Minute,
	}
}

func (s *JWTService) GenerateAccessToken(user *domain.User) (string, error) {
	roles := make([]string, len(user.Roles))
	permSet := make(map[string]bool)

	for i, role := range user.Roles {
		roles[i] = role.Name
		if role.HasAllPermissions() {
			permSet["*"] = true
		} else {
			for _, perm := range role.Permissions {
				permSet[perm.Code] = true
			}
		}
	}

	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}
	managedDepts := append([]string(nil), user.ManagedDepartmentUIDs...)

	claims := JWTClaims{
		UserID:                user.ID,
		UserUID:               user.UID,
		Roles:                 roles,
		Permissions:           permissions,
		ManagedDepartmentUIDs: managedDepts,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (c *JWTClaims) HasPermission(code string) bool {
	for _, p := range c.Permissions {
		if p == "*" || p == code {
			return true
		}
	}
	return false
}

func (c *JWTClaims) HasDepartmentAccess(departmentUID string) bool {
	for _, managedUID := range c.ManagedDepartmentUIDs {
		if managedUID == departmentUID {
			return true
		}
	}
	return false
}

// HasWebPortalAccess returns true if user has any role other than "Employee".
func (c *JWTClaims) HasWebPortalAccess() bool {
	for _, role := range c.Roles {
		if role != "Employee" {
			return true
		}
	}
	return false
}
