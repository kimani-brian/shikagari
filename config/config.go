package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application-wide configuration values
// loaded from environment variables.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN builds the PostgreSQL Data Source Name string
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Africa/Nairobi",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type UploadConfig struct {
	Dir           string
	MaxFileSizeMB int64
}

// Load reads the .env file (if present) and populates a Config struct.
// In production, variables are injected directly by the runtime environment.
func Load() *Config {
	// Load .env file in non-production environments
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("[config] No .env file found — using system environment variables")
		}
	}

	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Name: getEnv("APP_NAME", "ShikaGari"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "shikagari_user"),
			Password: getEnvRequired("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "shikagari_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:      getEnvRequired("JWT_SECRET"),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 72),
		},
		Upload: UploadConfig{
			Dir:           getEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSizeMB: int64(getEnvInt("MAX_FILE_SIZE_MB", 5)),
		},
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// getEnv returns the value of an env variable or a fallback default.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvRequired panics at startup if a required env variable is missing.
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("[config] FATAL: required environment variable %q is not set", key)
	}
	return value
}

// getEnvInt parses an env variable as an integer, returning a fallback on failure.
func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("[config] WARNING: %q is not a valid integer, using default %d", key, fallback)
		return fallback
	}
	return val
}
