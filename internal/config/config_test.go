package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	restoreEnv(t)

	cfg := Load()

	if cfg.AppEnv != "development" {
		t.Fatalf("expected development, got %q", cfg.AppEnv)
	}
	if cfg.AppPort != "8080" {
		t.Fatalf("expected 8080, got %q", cfg.AppPort)
	}
	if cfg.DatabaseURL != "postgres://jobscout:jobscout@localhost:5432/jobscout?sslmode=disable" {
		t.Fatalf("unexpected database URL: %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "change-me-in-production" {
		t.Fatalf("unexpected JWT secret: %q", cfg.JWTSecret)
	}
	if cfg.RemotiveBaseURL != "https://remotive.com/api/remote-jobs" {
		t.Fatalf("unexpected Remotive base URL: %q", cfg.RemotiveBaseURL)
	}
}

func TestLoadOverrides(t *testing.T) {
	restoreEnv(t)
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("REMOTIVE_BASE_URL", "https://example.com/jobs")

	cfg := Load()

	if cfg.AppEnv != "test" {
		t.Fatalf("expected test, got %q", cfg.AppEnv)
	}
	if cfg.AppPort != "9090" {
		t.Fatalf("expected 9090, got %q", cfg.AppPort)
	}
	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test?sslmode=disable" {
		t.Fatalf("unexpected database URL: %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Fatalf("unexpected JWT secret: %q", cfg.JWTSecret)
	}
	if cfg.RemotiveBaseURL != "https://example.com/jobs" {
		t.Fatalf("unexpected Remotive base URL: %q", cfg.RemotiveBaseURL)
	}
}

func restoreEnv(t *testing.T) {
	t.Helper()

	keys := []string{"APP_ENV", "APP_PORT", "DATABASE_URL", "JWT_SECRET", "REMOTIVE_BASE_URL"}
	original := make(map[string]string, len(keys))
	existed := make(map[string]bool, len(keys))

	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		original[key] = value
		existed[key] = ok
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if existed[key] {
				if err := os.Setenv(key, original[key]); err != nil {
					t.Errorf("restore %s: %v", key, err)
				}
				continue
			}
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("unset %s: %v", key, err)
			}
		}
	})
}
