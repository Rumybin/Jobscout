package config

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	AppEnv          string
	AppPort         string
	DatabaseURL     string
	JWTSecret       string
	RemotiveBaseURL string
}

// Load reads configuration from environment variables, applying sensible defaults
// where a variable is not set.
func Load() *Config {
	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppPort:         getEnv("APP_PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://jobscout:jobscout@localhost:5432/jobscout?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		RemotiveBaseURL: getEnv("REMOTIVE_BASE_URL", "https://remotive.com/api/remote-jobs"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
