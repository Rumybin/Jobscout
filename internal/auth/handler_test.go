package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockAuthService implements AuthService for testing.
type mockAuthService struct {
	registerFn func(ctx context.Context, email, password string) (*AuthResponse, error)
	loginFn    func(ctx context.Context, email, password string) (*AuthResponse, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, password string) (*AuthResponse, error) {
	return m.registerFn(ctx, email, password)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	return m.loginFn(ctx, email, password)
}

func TestRegister_Success(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(ctx context.Context, email, password string) (*AuthResponse, error) {
			return &AuthResponse{Token: "valid-token"}, nil
		},
	}
	h := NewHandler(svc)

	body := `{"email":"test@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token != "valid-token" {
		t.Fatalf("expected token 'valid-token', got '%s'", resp.Token)
	}
}

func TestRegister_Conflict(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(ctx context.Context, email, password string) (*AuthResponse, error) {
			return nil, fmt.Errorf("email already exists")
		},
	}
	h := NewHandler(svc)

	body := `{"email":"existing@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestRegister_BadRequest(t *testing.T) {
	svc := &mockAuthService{}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{invalid}`))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(ctx context.Context, email, password string) (*AuthResponse, error) {
			return &AuthResponse{Token: "login-token"}, nil
		},
	}
	h := NewHandler(svc)

	body := `{"email":"test@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token != "login-token" {
		t.Fatalf("expected token 'login-token', got '%s'", resp.Token)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(ctx context.Context, email, password string) (*AuthResponse, error) {
			return nil, fmt.Errorf("invalid email or password")
		},
	}
	h := NewHandler(svc)

	body := `{"email":"test@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestLogin_BadRequest(t *testing.T) {
	svc := &mockAuthService{}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`invalid json`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
