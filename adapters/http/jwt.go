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
	EmployeeUID           *string  `json:"employeeUid,omitempty"`
	Roles                 []string `json:"roles"`
	Permissions           []string `json:"permissions"`
	AccessScope           string   `json:"accessScope"`
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
	effectiveScope := domain.RoleScopeSelf
	highestScopePriority := 0
	managedDeptSet := make(map[string]struct{})

	for i, role := range user.Roles {
		roles[i] = role.Name
		if role.HasAllPermissions() {
			permSet["*"] = true
		}

		for _, perm := range role.Permissions {
			permSet[perm.Code] = true
		}

		normalizedScope := domain.NormalizeRoleScopeType(role.ScopeType)
		priority := domain.RoleScopePriority(normalizedScope)
		if priority > highestScopePriority {
			highestScopePriority = priority
			effectiveScope = normalizedScope
		}
	}

	if highestScopePriority == 0 {
		effectiveScope = domain.RoleScopeSelf
	}

	for _, departmentUID := range user.ManagedDepartmentUIDs {
		managedDeptSet[departmentUID] = struct{}{}
	}

	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}
	managedDepts := make([]string, 0, len(managedDeptSet))
	for departmentUID := range managedDeptSet {
		managedDepts = append(managedDepts, departmentUID)
	}

	claims := JWTClaims{
		UserID:                user.ID,
		UserUID:               user.UID,
		EmployeeUID:           user.EmployeeUID,
		Roles:                 roles,
		Permissions:           permissions,
		AccessScope:           effectiveScope,
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

func (c *JWTClaims) IsGlobalScope() bool {
	if c == nil {
		return false
	}
	return domain.NormalizeRoleScopeType(c.AccessScope) == domain.RoleScopeGlobal
}

func (c *JWTClaims) IsDepartmentScope() bool {
	if c == nil {
		return false
	}
	return domain.NormalizeRoleScopeType(c.AccessScope) == domain.RoleScopeDepartment
}

func (c *JWTClaims) IsSelfScope() bool {
	if c == nil {
		return false
	}
	return domain.NormalizeRoleScopeType(c.AccessScope) == domain.RoleScopeSelf
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
