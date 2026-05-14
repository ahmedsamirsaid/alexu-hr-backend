package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/banumusa/backend/adapters/db"
	httpadapter "github.com/banumusa/backend/adapters/http"
	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/banumusa/backend/core/usecases"
	"github.com/golang-migrate/migrate/v4"
	sqlitemigrate "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type stubI18nService struct{}

func (s stubI18nService) T(ctx context.Context, key string) string { return key }
func (s stubI18nService) TWithParams(ctx context.Context, key string, params map[string]interface{}) string {
	return key
}
func (s stubI18nService) TLocale(locale, key string) string { return key }
func (s stubI18nService) TLocaleWithParams(locale, key string, params map[string]interface{}) string {
	return key
}
func (s stubI18nService) GetSupportedLocales() []string { return []string{"en"} }
func (s stubI18nService) GetDefaultLocale() string      { return "en" }

func TestRateLimitService_RequestWindowAndVerifyLock(t *testing.T) {
	sqliteDB := newIntegrationDB(t)
	repo := db.NewRateLimitRepository()
	service := usecases.NewRateLimitService(sqliteDB, repo, usecases.RateLimitConfig{
		OTPRequestLimit:  3,
		OTPRequestWindow: 80 * time.Millisecond,
		OTPRequestLock:   120 * time.Millisecond,
		OTPVerifyLimit:   5,
		OTPVerifyWindow:  80 * time.Millisecond,
		OTPVerifyLock:    120 * time.Millisecond,
		LoginLock:        120 * time.Millisecond,
		LoginLimit:       10,
		LoginWindow:      time.Minute,
	})

	subject := "10.0.0.1|+201000000000"
	for i := 0; i < 3; i++ {
		result, err := service.AllowOTPRequest(context.Background(), subject)
		if err != nil {
			t.Fatalf("AllowOTPRequest() error = %v", err)
		}
		if !result.Allowed {
			t.Fatalf("AllowOTPRequest() denied request %d unexpectedly", i+1)
		}
	}

	blocked, err := service.AllowOTPRequest(context.Background(), subject)
	if err != nil {
		t.Fatalf("AllowOTPRequest() blocked error = %v", err)
	}
	if blocked.Allowed || blocked.RetryAfter <= 0 {
		t.Fatalf("AllowOTPRequest() = %+v, want denied with retry", blocked)
	}

	time.Sleep(130 * time.Millisecond)
	afterWindow, err := service.AllowOTPRequest(context.Background(), subject)
	if err != nil {
		t.Fatalf("AllowOTPRequest() after window error = %v", err)
	}
	if !afterWindow.Allowed {
		t.Fatalf("AllowOTPRequest() after window = %+v, want allowed", afterWindow)
	}

	for i := 0; i < 5; i++ {
		result, err := service.RegisterOTPVerifyFailure(context.Background(), subject)
		if err != nil {
			t.Fatalf("RegisterOTPVerifyFailure() error = %v", err)
		}
		if !result.Allowed {
			t.Fatalf("RegisterOTPVerifyFailure() denied attempt %d unexpectedly", i+1)
		}
	}

	locked, err := service.RegisterOTPVerifyFailure(context.Background(), subject)
	if err != nil {
		t.Fatalf("RegisterOTPVerifyFailure() lock error = %v", err)
	}
	if locked.Allowed || !locked.Locked || locked.RetryAfter <= 0 {
		t.Fatalf("RegisterOTPVerifyFailure() = %+v, want locked", locked)
	}

	checkLocked, err := service.CheckOTPVerify(context.Background(), subject)
	if err != nil {
		t.Fatalf("CheckOTPVerify() error = %v", err)
	}
	if checkLocked.Allowed || !checkLocked.Locked {
		t.Fatalf("CheckOTPVerify() = %+v, want locked", checkLocked)
	}

	if err := service.ResetOTPVerify(context.Background(), subject); err != nil {
		t.Fatalf("ResetOTPVerify() error = %v", err)
	}
	resetCheck, err := service.CheckOTPVerify(context.Background(), subject)
	if err != nil {
		t.Fatalf("CheckOTPVerify() after reset error = %v", err)
	}
	if !resetCheck.Allowed {
		t.Fatalf("CheckOTPVerify() after reset = %+v, want allowed", resetCheck)
	}
}

func TestRateLimitMiddleware_PerIP(t *testing.T) {
	limiter := httpadapter.NewIPRateLimiter(2, 120*time.Millisecond)
	handler := httpadapter.RateLimitMiddleware(limiter, stubI18nService{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("blocked status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("missing Retry-After header")
	}
}

func TestAuthHandler_LoginRateLimitAndReset(t *testing.T) {
	sqliteDB := newIntegrationDB(t)
	userRepo := db.NewUserRepository()
	roleRepo := db.NewRoleRepository()
	permissionRepo := db.NewPermissionRepository()
	employeeRepo := db.NewEmployeeRepository()
	refreshTokenRepo := db.NewRefreshTokenRepository()
	rateLimitRepo := db.NewRateLimitRepository()
	rateLimitService := usecases.NewRateLimitService(sqliteDB, rateLimitRepo, usecases.RateLimitConfig{
		OTPRequestLimit:  3,
		OTPRequestWindow: time.Minute,
		OTPRequestLock:   time.Minute,
		OTPVerifyLimit:   5,
		OTPVerifyWindow:  time.Minute,
		OTPVerifyLock:    time.Minute,
		LoginLock:        time.Minute,
		LoginLimit:       2,
		LoginWindow:      time.Minute,
	})

	user := domain.NewUser("+201000000111")
	if err := usecases.SetUserPassword(user, "correct-password"); err != nil {
		t.Fatalf("SetUserPassword() error = %v", err)
	}
	if err := userRepo.Create(context.Background(), sqliteDB, user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}
	seedRoleAssignment(t, sqliteDB, user.ID, "Manager")

	jwtService := httpadapter.NewJWTService("test-secret", 60)
	auditor := audit.NewAuditor(sqliteDB, db.NewAuditLogRepository())
	loginUC := usecases.NewLoginPasswordUseCase(sqliteDB, userRepo, roleRepo, permissionRepo, employeeRepo, refreshTokenRepo, jwtService, false, "", 7, auditor)
	handler := httpadapter.NewAuthHandler(nil, nil, loginUC, nil, nil, nil, rateLimitService, stubI18nService{})

	assertLoginStatus := func(password string, want int) {
		t.Helper()
		body := map[string]string{"phone": user.Phone, "password": password}
		payload, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(payload))
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.LoginPassword(rec, req)
		if rec.Code != want {
			t.Fatalf("login(%q) status = %d, want %d, body=%s", password, rec.Code, want, rec.Body.String())
		}
	}

	assertLoginStatus("wrong", http.StatusUnauthorized)
	assertLoginStatus("correct-password", http.StatusOK)
	assertLoginStatus("wrong", http.StatusUnauthorized)
	assertLoginStatus("wrong", http.StatusUnauthorized)
	assertLoginStatus("wrong", http.StatusTooManyRequests)
}

func TestAuthHandler_OTPRequestRateLimit_IsScopedToIPAndPhone(t *testing.T) {
	sqliteDB := newIntegrationDB(t)
	userRepo := db.NewUserRepository()
	otpRepo := db.NewOTPRepository()
	rateLimitRepo := db.NewRateLimitRepository()
	rateLimitService := usecases.NewRateLimitService(sqliteDB, rateLimitRepo, usecases.RateLimitConfig{
		OTPRequestLimit:  3,
		OTPRequestWindow: time.Minute,
		OTPRequestLock:   time.Minute,
		OTPVerifyLimit:   5,
		OTPVerifyWindow:  time.Minute,
		OTPVerifyLock:    time.Minute,
		LoginLock:        time.Minute,
		LoginLimit:       2,
		LoginWindow:      time.Minute,
	})

	user := domain.NewUser("+201000000222")
	if err := userRepo.Create(context.Background(), sqliteDB, user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	requestUC := usecases.NewRequestOTPUseCase(sqliteDB, otpRepo, userRepo)
	handler := httpadapter.NewAuthHandler(requestUC, nil, nil, nil, nil, nil, rateLimitService, stubI18nService{})

	assertRequestStatus := func(ip string, want int) {
		t.Helper()
		body := map[string]string{"phone": user.Phone}
		payload, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/request", bytes.NewReader(payload))
		req.RemoteAddr = ip + ":1234"
		rec := httptest.NewRecorder()
		handler.RequestOTP(rec, req)
		if rec.Code != want {
			t.Fatalf("requestOTP(%s) status = %d, want %d, body=%s", ip, rec.Code, want, rec.Body.String())
		}
	}

	assertRequestStatus("10.0.0.1", http.StatusOK)
	assertRequestStatus("10.0.0.1", http.StatusOK)
	assertRequestStatus("10.0.0.1", http.StatusOK)
	assertRequestStatus("10.0.0.1", http.StatusTooManyRequests)
	assertRequestStatus("10.0.0.2", http.StatusOK)
}

func newIntegrationDB(t *testing.T) *db.SQLiteDB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "integration.db")
	sqliteDB, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteDB() error = %v", err)
	}

	driver, err := sqlitemigrate.WithInstance(sqliteDB.DB(), &sqlitemigrate.Config{})
	if err != nil {
		t.Fatalf("sqlite.WithInstance() error = %v", err)
	}

	migrationsPath := findMigrationsDir(t)
	migrator, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsPath), "sqlite", driver)
	if err != nil {
		t.Fatalf("migrate.NewWithDatabaseInstance() error = %v", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrator.Up() error = %v", err)
	}
	t.Cleanup(func() {
		sqliteDB.Close()
	})

	clearTables(t, sqliteDB.DB())
	return sqliteDB
}

func clearTables(t *testing.T, sqlDB *sql.DB) {
	t.Helper()
	ctx := context.Background()
	if _, err := sqlDB.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	defer func() {
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
			t.Fatalf("enable foreign keys: %v", err)
		}
	}()
	tables := []string{
		"user_roles",
		"role_permissions",
		"refresh_tokens",
		"otp_codes",
		"rate_limit_records",
		"users",
		"roles",
		"permissions",
		"audit_logs",
	}
	for _, table := range tables {
		if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			t.Fatalf("clear table %s: %v", table, err)
		}
	}
}

func seedRoleAssignment(t *testing.T, sqliteDB *db.SQLiteDB, userID int64, roleName string) {
	t.Helper()
	ctx := context.Background()
	now := time.Now()
	roleUID := domain.GenerateUID("role")
	result, err := sqliteDB.ExecContext(ctx, `
		INSERT INTO roles (uid, name, description, is_system, created_at, updated_at)
		VALUES (?, ?, ?, 0, ?, ?)`,
		roleUID, roleName, roleName, now, now,
	)
	if err != nil {
		t.Fatalf("insert role: %v", err)
	}
	roleID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("role LastInsertId: %v", err)
	}
	if _, err := sqliteDB.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id, created_at) VALUES (?, ?, ?)`, userID, roleID, now); err != nil {
		t.Fatalf("insert user role: %v", err)
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	candidates := []string{
		filepath.Join(cwd, "adapters", "db", "migrations"),
		filepath.Join(cwd, "..", "adapters", "db", "migrations"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	t.Fatal("migrations directory not found")
	return ""
}

var _ ports.I18nService = stubI18nService{}
