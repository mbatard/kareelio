package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort          string
	PostgresHost        string
	PostgresPort        string
	PostgresDB          string
	PostgresUser        string
	PostgresPassword    string
	SessionSecret       string
	SessionCookieSameSite string
	SessionDurationHours int
	DefaultAdminEmail   string
	DefaultAdminPassword string
	CorsOrigin          string
	DBMigrate           bool
	AppPublicURL        string
	RegistrationEnabled bool
	SMTPHost            string
	SMTPPort            string
	SMTPUsername         string
	SMTPPassword        string
	SMTPFrom            string
	VerificationTokenTTLHours int
}

func Load() *Config {
	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		PostgresHost:        getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:        getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:          getEnv("POSTGRES_DB", "kareelio"),
		PostgresUser:        getEnv("POSTGRES_USER", "kareelio"),
		PostgresPassword:    getEnv("POSTGRES_PASSWORD", "changeme"),
		SessionSecret:       getEnv("SESSION_SECRET", "dev-secret-change-in-production"),
		SessionCookieSameSite: getEnv("SESSION_COOKIE_SAMESITE", "lax"),
		SessionDurationHours: getEnvInt("SESSION_DURATION_HOURS", 72),
		DefaultAdminEmail:   getEnv("DEFAULT_ADMIN_EMAIL", "admin@kareelio.local"),
		DefaultAdminPassword: getEnv("DEFAULT_ADMIN_PASSWORD", "admin"),
		CorsOrigin:          getEnv("CORS_ORIGIN", "http://localhost:5173"),
		DBMigrate:           getEnvBool("DB_MIGRATE", true),
		AppPublicURL:        getEnv("APP_PUBLIC_URL", "http://localhost:5173"),
		RegistrationEnabled: getEnvBool("REGISTRATION_ENABLED", true),
		SMTPHost:            getEnv("SMTP_HOST", ""),
		SMTPPort:            getEnv("SMTP_PORT", "587"),
		SMTPUsername:         getEnv("SMTP_USERNAME", ""),
		SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:            getEnv("SMTP_FROM", ""),
		VerificationTokenTTLHours: getEnvInt("VERIFICATION_TOKEN_TTL_HOURS", 24),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.PostgresHost +
		" port=" + c.PostgresPort +
		" dbname=" + c.PostgresDB +
		" user=" + c.PostgresUser +
		" password=" + c.PostgresPassword +
		" sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
