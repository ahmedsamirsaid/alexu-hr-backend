package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type AppConfig struct {
	Port              string
	DBPath            string
	DBMigrationsPath  string
	FontPath          string
	LegacySyncEnabled bool
	LegacyDBHost      string
	LegacyDBPort      string
	LegacyDBName      string
	LegacyDBUser      string
	LegacyDBPassword  string
	// Auth config
	AuthEnabled        bool
	DevOTPBypass       bool
	DevBypassOTP       string
	JWTSecret          string
	AccessTokenMinutes int
	RefreshTokenDays   int
	// FCM config
	FCMEnabled            bool
	FCMServiceAccountPath string
	// Scheduler config
	SchedulerEnabled         bool
	ExpiredLeaveGraceDays    int
	SchedulerIntervalHours   int
	HolidaySyncEndpoint      string
	HolidaySyncAPIKey        string
	HolidaySyncCountry       string
	HolidaySyncTimezone      string
	SchedulerIntervalMinutes int
	// Storage config
	MinIOEndpoint              string
	MinIOAccessKey             string
	MinIOSecretKey             string
	MinIOUseSSL                bool
	MinIODocumentsBucket       string
	MinIOAutoCreateBucket      bool
	MinIOUploadExpiryMinutes   int
	MinIODownloadExpiryMinutes int
	I18nLocalesPath            string
	I18nDefaultLocale          string
	I18nSupportedLocales       []string
	I18nHotReload              bool
	GeneralRateLimitPerMinute  int
	RateLimitLockdownDuration  time.Duration
	OTPRequestLimit            int
	OTPRequestWindow           time.Duration
	OTPVerifyLimit             int
	OTPVerifyWindow            time.Duration
	OTPVerifyLockDuration      time.Duration
	LoginRateLimit             int
	LoginRateLimitWindow       time.Duration
}

type jsonConfig struct {
	RateLimit *jsonRateLimitConfig `json:"rateLimit"`
}

type jsonRateLimitConfig struct {
	LockdownMinutes int `json:"lockdownMinutes"`
}

func LoadConfig() *AppConfig {
	jsonCfg := loadJSONConfig()
	lockdownMinutes := 15
	if jsonCfg.RateLimit != nil && jsonCfg.RateLimit.LockdownMinutes > 0 {
		lockdownMinutes = jsonCfg.RateLimit.LockdownMinutes
	}

	return &AppConfig{
		Port:              getEnv("BANU_MUSA_PORT", "8080"),
		DBPath:            getEnv("BANU_MUSA_DB_PATH", "./data/banumusa.db"),
		DBMigrationsPath:  getEnv("BANU_MUSA_DB_MIGRATIONS_PATH", "./adapters/db/migrations"),
		FontPath:          getEnv("BANU_MUSA_FONT_PATH", "./assets/fonts/Noto_Sans_Arabic/static/NotoSansArabic-Regular.ttf"),
		LegacySyncEnabled: getEnvBool("BANU_MUSA_LEGACY_SYNC_ENABLED", false),
		LegacyDBHost:      getEnv("BANU_MUSA_LEGACY_DB_HOST", ""),
		LegacyDBPort:      getEnv("BANU_MUSA_LEGACY_DB_PORT", "1433"),
		LegacyDBName:      getEnv("BANU_MUSA_LEGACY_DB_NAME", ""),
		LegacyDBUser:      getEnv("BANU_MUSA_LEGACY_DB_USER", ""),
		LegacyDBPassword:  getEnv("BANU_MUSA_LEGACY_DB_PASSWORD", ""),
		// Auth config
		AuthEnabled:        getEnvBool("BANU_MUSA_AUTH_ENABLED", true),
		DevOTPBypass:       getEnvBool("BANU_MUSA_DEV_OTP_BYPASS", true),
		DevBypassOTP:       getEnv("BANU_MUSA_DEV_BYPASS_OTP", "112233"),
		JWTSecret:          getEnv("BANU_MUSA_JWT_SECRET", "dev-secret-change-in-production"),
		AccessTokenMinutes: getEnvInt("BANU_MUSA_ACCESS_TOKEN_MINUTES", 10080), // 7 days
		RefreshTokenDays:   getEnvInt("BANU_MUSA_REFRESH_TOKEN_DAYS", 90),
		// FCM config
		FCMEnabled:            getEnvBool("BANU_MUSA_FCM_ENABLED", false),
		FCMServiceAccountPath: getEnv("BANU_MUSA_FCM_SERVICE_ACCOUNT_PATH", "./firebase/staging-banumusa-firebase-adminsdk-fbsvc-9f7f698d1d.json"),
		// Scheduler config
		SchedulerEnabled:         getEnvBool("BANU_MUSA_SCHEDULER_ENABLED", true),
		ExpiredLeaveGraceDays:    getEnvInt("BANU_MUSA_EXPIRED_LEAVE_GRACE_DAYS", 1),
		SchedulerIntervalHours:   getEnvInt("BANU_MUSA_SCHEDULER_INTERVAL_HOURS", 1),
		SchedulerIntervalMinutes: getEnvInt("BANU_MUSA_SCHEDULER_INTERVAL_MINUTES", 1),
		HolidaySyncEndpoint:      getEnv("BANU_MUSA_HOLIDAY_SYNC_ENDPOINT", "https://calendarific.com/api/v2/holidays"),
		HolidaySyncAPIKey:        getEnv("BANU_MUSA_HOLIDAY_SYNC_API_KEY", "ysvLh7UNdrI0mbH9afetHcsaYcDGBiag"),
		HolidaySyncCountry:       getEnv("BANU_MUSA_HOLIDAY_SYNC_COUNTRY", "EG"),
		HolidaySyncTimezone:      getEnv("BANU_MUSA_HOLIDAY_SYNC_TIMEZONE", "Africa/Cairo"),

		// Storage config
		MinIOEndpoint:              getEnv("BANU_MUSA_MINIO_ENDPOINT", "bucket-production-a04a.up.railway.app"),
		MinIOAccessKey:             getEnv("BANU_MUSA_MINIO_ACCESS_KEY", "mU8a1ES7JKx1Q00kox7P2QCqAjiizE4o"),
		MinIOSecretKey:             getEnv("BANU_MUSA_MINIO_SECRET_KEY", "ko5FnnLKOPq9QX51iPeARDEbnCeMbPYl5RC6EGz2LWuuddCa"),
		MinIOUseSSL:                getEnvBool("BANU_MUSA_MINIO_USE_SSL", true),
		MinIODocumentsBucket:       getEnv("BANU_MUSA_MINIO_DOCUMENTS_BUCKET", "documents"),
		MinIOAutoCreateBucket:      getEnvBool("BANU_MUSA_MINIO_AUTO_CREATE_BUCKET", true),
		MinIOUploadExpiryMinutes:   getEnvInt("BANU_MUSA_MINIO_UPLOAD_EXPIRY_MINUTES", 15),
		MinIODownloadExpiryMinutes: getEnvInt("BANU_MUSA_MINIO_DOWNLOAD_EXPIRY_MINUTES", 15),

		// I18n config
		I18nLocalesPath:           getEnv("BANU_MUSA_I18N_LOCALES_PATH", "./locales"),
		I18nDefaultLocale:         getEnv("BANU_MUSA_I18N_DEFAULT_LOCALE", "en"),
		I18nSupportedLocales:      getEnvStringSlice("BANU_MUSA_I18N_SUPPORTED_LOCALES", []string{"en", "ar"}),
		I18nHotReload:             getEnvBool("BANU_MUSA_I18N_HOT_RELOAD", false),
		GeneralRateLimitPerMinute: getEnvInt("BANU_MUSA_RATE_LIMIT_GENERAL_PER_MINUTE", 100),
		RateLimitLockdownDuration: getEnvDurationMinutes("BANU_MUSA_RATE_LIMIT_LOCKDOWN_MINUTES", lockdownMinutes),
		OTPRequestLimit:           getEnvInt("BANU_MUSA_RATE_LIMIT_OTP_REQUEST_LIMIT", 3),
		OTPRequestWindow:          getEnvDurationMinutes("BANU_MUSA_RATE_LIMIT_OTP_REQUEST_WINDOW_MINUTES", 15),
		OTPVerifyLimit:            getEnvInt("BANU_MUSA_RATE_LIMIT_OTP_VERIFY_LIMIT", 5),
		OTPVerifyWindow:           getEnvDurationMinutes("BANU_MUSA_RATE_LIMIT_OTP_VERIFY_WINDOW_MINUTES", 10),
		OTPVerifyLockDuration:     getEnvDurationMinutes("BANU_MUSA_RATE_LIMIT_OTP_VERIFY_LOCK_MINUTES", lockdownMinutes),
		LoginRateLimit:            getEnvInt("BANU_MUSA_RATE_LIMIT_LOGIN_LIMIT", 10),
		LoginRateLimitWindow:      getEnvDurationMinutes("BANU_MUSA_RATE_LIMIT_LOGIN_WINDOW_MINUTES", 15),
	}
}

func loadJSONConfig() jsonConfig {
	var cfg jsonConfig
	paths := []string{
		"config.json",
		filepath.Join(".", "config.json"),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return jsonConfig{}
		}
		return cfg
	}
	return jsonConfig{}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		if value == "" {
			return defaultValue
		}
		// Split by comma and trim whitespace
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) == 0 {
			return defaultValue
		}
		return result
	}
	return defaultValue
}

func getEnvDurationMinutes(key string, defaultMinutes int) time.Duration {
	return time.Duration(getEnvInt(key, defaultMinutes)) * time.Minute
}
