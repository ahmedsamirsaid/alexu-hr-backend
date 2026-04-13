package http

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
)

// AuthMiddleware validates JWT tokens and sets claims in context
func AuthMiddleware(jwtService *JWTService, authEnabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authEnabled {
				// When auth is disabled, use a default admin context
				claims := &JWTClaims{
					UserID:      0,
					UserUID:     "dev_admin",
					Roles:       []string{"admin"},
					Permissions: []string{"*"},
				}
				ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Missing Authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Invalid Authorization header format")
				return
			}

			claims, err := jwtService.ValidateAccessToken(parts[1])
			if err != nil {
				if err == ErrTokenExpired {
					writeJSONError(w, http.StatusUnauthorized, "token_expired", "Access token has expired")
					return
				}
				writeJSONError(w, http.StatusUnauthorized, "invalid_token", "Invalid access token")
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission middleware checks if the user has the required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*JWTClaims)
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
				return
			}

			if !claims.HasPermission(permission) {
				writeJSONError(w, http.StatusForbidden, "permission_denied", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims retrieves JWT claims from the request context
func GetClaims(r *http.Request) *JWTClaims {
	claims, _ := r.Context().Value(ClaimsContextKey).(*JWTClaims)
	return claims
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + code + `","message":"` + message + `"}`))
}
