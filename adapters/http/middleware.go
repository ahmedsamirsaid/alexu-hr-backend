package http

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/i18n"
	"github.com/banumusa/backend/core/ports"
	"golang.org/x/time/rate"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
)

// LanguageMiddleware extracts locale from Accept-Language header.
func LanguageMiddleware(i18nService ports.I18nService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptLang := r.Header.Get("Accept-Language")
			locale := i18n.ParseAcceptLanguage(acceptLang, i18nService.GetSupportedLocales())

			// Add debug logging for detected locale
			slog.Debug("language.middleware.locale_detected",
				"accept_language", acceptLang,
				"detected_locale", locale,
				"path", r.URL.Path,
			)

			ctx := i18n.WithLocale(r.Context(), locale)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
			actorUID := claims.UserUID
			if claims.EmployeeUID != nil && *claims.EmployeeUID != "" {
				actorUID = *claims.EmployeeUID
			}
			ctx = audit.WithActor(ctx, actorUID)
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

// RequireAnyPermission middleware checks if the user has at least one of the required permissions.
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*JWTClaims)
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "authentication_required", "Not authenticated")
				return
			}

			for _, permission := range permissions {
				if claims.HasPermission(permission) {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeJSONError(w, http.StatusForbidden, "permission_denied", "Insufficient permissions")
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

type ipVisitor struct {
	limiter      *rate.Limiter
	lastSeen     time.Time
	blockedUntil time.Time
}

type IPRateLimiter struct {
	mu               sync.Mutex
	visitors         map[string]*ipVisitor
	limit            rate.Limit
	burst            int
	lockdownDuration time.Duration
	now              func() time.Time
}

func NewIPRateLimiter(requestsPerMinute int, lockdownDuration time.Duration) *IPRateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 100
	}
	if lockdownDuration <= 0 {
		lockdownDuration = 15 * time.Minute
	}
	return &IPRateLimiter{
		visitors:         make(map[string]*ipVisitor),
		limit:            rate.Every(time.Minute / time.Duration(requestsPerMinute)),
		burst:            requestsPerMinute,
		lockdownDuration: lockdownDuration,
		now:              time.Now,
	}
}

func (l *IPRateLimiter) allow(ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	visitor, ok := l.visitors[ip]
	if !ok {
		visitor = &ipVisitor{
			limiter:  rate.NewLimiter(l.limit, l.burst),
			lastSeen: now,
		}
		l.visitors[ip] = visitor
	}
	visitor.lastSeen = now

	if visitor.blockedUntil.After(now) {
		return false, visitor.blockedUntil.Sub(now)
	}

	if !visitor.limiter.Allow() {
		visitor.blockedUntil = now.Add(l.lockdownDuration)
		return false, l.lockdownDuration
	}

	return true, 0
}

func RateLimitMiddleware(limiter *IPRateLimiter, i18nService ports.I18nService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			allowed, retryAfter := limiter.allow(ip)
			if !allowed {
				slog.Warn("http.rate_limit.ip_exceeded", "ip", ip, "path", r.URL.Path)
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
				writeLocalizedError(w, http.StatusTooManyRequests, domain.ErrRateLimitExceeded, i18nService, r.Context())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + code + `","message":"` + message + `"}`))
}
