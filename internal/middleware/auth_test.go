package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobscout/internal/auth"
	"jobscout/internal/config"
)

func TestAuth_ValidToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	token, err := auth.GenerateToken("user-1", "test@example.com", cfg.JWTSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		userID := UserIDFromContext(r.Context())
		if userID != "user-1" {
			t.Fatalf("expected userID 'user-1', got '%s'", userID)
		}
		email := EmailFromContext(r.Context())
		if email != "test@example.com" {
			t.Fatalf("expected email 'test@example.com', got '%s'", email)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(cfg)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware(next).ServeHTTP(rec, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := Auth(cfg)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()

	middleware(next).ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := Auth(cfg)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	rec := httptest.NewRecorder()

	middleware(next).ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := Auth(cfg)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	middleware(next).ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()
	if id := UserIDFromContext(ctx); id != "" {
		t.Fatalf("expected empty string, got '%s'", id)
	}
}

func TestEmailFromContext_NotFound(t *testing.T) {
	ctx := context.Background()
	if email := EmailFromContext(ctx); email != "" {
		t.Fatalf("expected empty string, got '%s'", email)
	}
}
