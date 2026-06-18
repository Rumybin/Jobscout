package auth

import (
	"context"
	"strings"
	"testing"

	"jobscout/internal/config"
	"jobscout/internal/testdb"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	pool := testdb.Open(t)
	cfg := &config.Config{JWTSecret: "test-secret"}

	return NewService(pool, cfg)
}

func TestService_RegisterAndLogin(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	registerResp, err := svc.Register(ctx, "user@example.com", "secret123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if registerResp.Token == "" {
		t.Fatal("expected register token")
	}

	registerClaims, err := ValidateToken(registerResp.Token, "test-secret")
	if err != nil {
		t.Fatalf("validate register token: %v", err)
	}
	if registerClaims.Email != "user@example.com" {
		t.Fatalf("expected user@example.com, got %q", registerClaims.Email)
	}

	loginResp, err := svc.Login(ctx, "user@example.com", "secret123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("expected login token")
	}

	loginClaims, err := ValidateToken(loginResp.Token, "test-secret")
	if err != nil {
		t.Fatalf("validate login token: %v", err)
	}
	if loginClaims.Subject != registerClaims.Subject {
		t.Fatalf("expected same user id, got %q and %q", registerClaims.Subject, loginClaims.Subject)
	}
}

func TestService_RegisterValidation(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	cases := []struct {
		name     string
		email    string
		password string
		want     string
	}{
		{name: "missing email", password: "secret123", want: "email is required"},
		{name: "missing password", email: "user@example.com", want: "password is required"},
		{name: "short password", email: "user@example.com", password: "123", want: "password must be at least 6 characters"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(ctx, tc.email, tc.password)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestService_RegisterDuplicateEmail(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Register(ctx, "duplicate@example.com", "secret123"); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err := svc.Register(ctx, "duplicate@example.com", "secret123")
	if err == nil {
		t.Fatal("expected duplicate email error")
	}
	if !strings.Contains(err.Error(), "failed to create user") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_LoginInvalidCredentials(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Register(ctx, "login@example.com", "secret123"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	cases := []struct {
		name     string
		email    string
		password string
		want     string
	}{
		{name: "missing email", password: "secret123", want: "email is required"},
		{name: "missing password", email: "login@example.com", want: "password is required"},
		{name: "wrong email", email: "missing@example.com", password: "secret123", want: "invalid email or password"},
		{name: "wrong password", email: "login@example.com", password: "wrong123", want: "invalid email or password"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Login(ctx, tc.email, tc.password)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestService_GetUser(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	resp, err := svc.Register(ctx, "me@example.com", "secret123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	claims, err := ValidateToken(resp.Token, "test-secret")
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	me, err := svc.GetUser(ctx, claims.Subject)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if me.ID != claims.Subject {
		t.Fatalf("expected id %q, got %q", claims.Subject, me.ID)
	}
	if me.Email != "me@example.com" {
		t.Fatalf("expected me@example.com, got %q", me.Email)
	}
}

func TestService_GetUserErrors(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	cases := []struct {
		name   string
		userID string
		want   string
	}{
		{name: "invalid id", userID: "not-a-uuid", want: "invalid user id"},
		{name: "missing user", userID: "00000000-0000-0000-0000-000000000001", want: "user not found"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.GetUser(ctx, tc.userID)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %q", tc.want, err.Error())
			}
		})
	}
}
