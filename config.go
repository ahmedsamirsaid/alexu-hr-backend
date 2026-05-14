package main

import (
	"os"
	"strconv"
	"strings"
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
	// PostgreSQL config
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
	// pgBouncer config
	PgBouncerEnabled bool
	PgBouncerHost    string
	PgBouncerPort    string
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
	// Seed config
	SeedDevData bool
	// Storage config
	MinIOEndpoint                string
	MinIOAccessKey               string
	MinIOSecretKey               string
	MinIOUseSSL                  bool
	MinIODocumentsBucket         string
	MinIOAutoCreateBucket        bool
	MinIOUploadExpiryMinutes     int
	MinIODownloadExpiryMinutes   int
	I18nLocalesPath      string
	I18nDefaultLocale    string
	I18nSupportedLocales []string
	I18nHotReload        bool
}

func LoadConfig() *AppConfig {
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
			// PostgreSQL config
		PostgresHost:     getEnv("BANU_MUSA_POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("BANU_MUSA_POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("BANU_MUSA_POSTGRES_USER", "banumusa"),
		PostgresPassword: getEnv("BANU_MUSA_POSTGRES_PASSWORD", "banumusa-secret"),
		PostgresDB:       getEnv("BANU_MUSA_POSTGRES_DB", "banumusa"),
		PostgresSSLMode:  getEnv("BANU_MUSA_POSTGRES_SSLMODE", "disable"),
		// pgBouncer config
		PgBouncerEnabled: getEnvBool("BANU_MUSA_PGBOUNCER_ENABLED", false),
		PgBouncerHost:    getEnv("BANU_MUSA_PGBOUNCER_HOST", "localhost"),
		PgBouncerPort:    getEnv("BANU_MUSA_PGBOUNCER_PORT", "6432"),
		// Auth config
		AuthEnabled:        getEnvBool("BANU_MUSA_AUTH_ENABLED", true),
		DevOTPBypass:       getEnvBool("BANU_MUSA_DEV_OTP_BYPASS", true),
		DevBypassOTP:       getEnv("BANU_MUSA_DEV_BYPASS_OTP", "112233"),
		JWTSecret:          getEnv("BANU_MUSA_JWT_SECRET", "dev-secret-change-in-production"),
		AccessTokenMinutes: getEnvInt("BANU_MUSA_ACCESS_TOKEN_MINUTES", 10080),
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

		// Seed config
		SeedDevData: getEnvBool("BANU_MUSA_SEED_DEV", true),

		// Storage config
		MinIOEndpoint:              getEnv("BANU_MUSA_MINIO_ENDPOINT", "bucket-production-62aa.up.railway.app"),
		MinIOAccessKey:             getEnv("BANU_MUSA_MINIO_ACCESS_KEY", "hgZN6FTEHjaeg0nJX8H5NmCc1IHOIcqp"),
		MinIOSecretKey:             getEnv("BANU_MUSA_MINIO_SECRET_KEY", "vkiYoESOd5vr0wdo1J34t9JOzphCkoh3mBMqQ7V739UOCr4K"),
		MinIOUseSSL:                getEnvBool("BANU_MUSA_MINIO_USE_SSL", true),
		MinIODocumentsBucket:       getEnv("BANU_MUSA_MINIO_DOCUMENTS_BUCKET", "documents"),
		MinIOAutoCreateBucket:      getEnvBool("BANU_MUSA_MINIO_AUTO_CREATE_BUCKET", true),
		MinIOUploadExpiryMinutes:   getEnvInt("BANU_MUSA_MINIO_UPLOAD_EXPIRY_MINUTES", 15),
		MinIODownloadExpiryMinutes: getEnvInt("BANU_MUSA_MINIO_DOWNLOAD_EXPIRY_MINUTES", 15),

		// I18n config
		I18nLocalesPath:      getEnv("BANU_MUSA_I18N_LOCALES_PATH", "./locales"),
		I18nDefaultLocale:    getEnv("BANU_MUSA_I18N_DEFAULT_LOCALE", "en"),
		I18nSupportedLocales: getEnvStringSlice("BANU_MUSA_I18N_SUPPORTED_LOCALES", []string{"en", "ar"}),
		I18nHotReload:        getEnvBool("BANU_MUSA_I18N_HOT_RELOAD", false),
	}
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
