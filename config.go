package main

import (
	"os"
	"strconv"
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
	SchedulerEnabled       bool
	ExpiredLeaveGraceDays  int
	SchedulerIntervalHours int
	HolidaySyncEndpoint    string
	HolidaySyncAPIKey      string
	HolidaySyncCountry     string
	HolidaySyncTimezone    string
	SchedulerIntervalMinutes int
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
		// Auth config
		AuthEnabled:        getEnvBool("BANU_MUSA_AUTH_ENABLED", true),
		DevOTPBypass:       getEnvBool("BANU_MUSA_DEV_OTP_BYPASS", true),
		DevBypassOTP:       getEnv("BANU_MUSA_DEV_BYPASS_OTP", "112233"),
		JWTSecret:          getEnv("BANU_MUSA_JWT_SECRET", "dev-secret-change-in-production"),
		AccessTokenMinutes: getEnvInt("BANU_MUSA_ACCESS_TOKEN_MINUTES", 60),
		RefreshTokenDays:   getEnvInt("BANU_MUSA_REFRESH_TOKEN_DAYS", 90),
		// FCM config
		FCMEnabled:            getEnvBool("BANU_MUSA_FCM_ENABLED", true),
		FCMServiceAccountPath: getEnv("BANU_MUSA_FCM_SERVICE_ACCOUNT_PATH", "./firebase/staging-banumusa-firebase-adminsdk-fbsvc-9f7f698d1d.json"),
		// Scheduler config
		SchedulerEnabled:       getEnvBool("BANU_MUSA_SCHEDULER_ENABLED", true),
		ExpiredLeaveGraceDays:  getEnvInt("BANU_MUSA_EXPIRED_LEAVE_GRACE_DAYS", 1),
		SchedulerIntervalHours: getEnvInt("BANU_MUSA_SCHEDULER_INTERVAL_HOURS", 1),
		SchedulerIntervalMinutes: getEnvInt("BANU_MUSA_SCHEDULER_INTERVAL_MINUTES", 1),
		HolidaySyncEndpoint:    getEnv("BANU_MUSA_HOLIDAY_SYNC_ENDPOINT", "https://calendarific.com/api/v2/holidays"),
		HolidaySyncAPIKey:      getEnv("BANU_MUSA_HOLIDAY_SYNC_API_KEY", "ysvLh7UNdrI0mbH9afetHcsaYcDGBiag"),
		HolidaySyncCountry:     getEnv("BANU_MUSA_HOLIDAY_SYNC_COUNTRY", "EG"),
		HolidaySyncTimezone:    getEnv("BANU_MUSA_HOLIDAY_SYNC_TIMEZONE", "Africa/Cairo"),
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
