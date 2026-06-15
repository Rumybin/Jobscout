package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_Valid(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	email := "test@example.com"
	secret := "test-secret"

	tokenStr, err := GenerateToken(userID, email, secret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	email := "test@example.com"
	secret := "test-secret"

	tokenStr, _ := GenerateToken(userID, email, secret)

	claims, err := ValidateToken(tokenStr, secret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.Subject != userID {
		t.Fatalf("expected subject %s, got %s", userID, claims.Subject)
	}
	if claims.Email != email {
		t.Fatalf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Issuer != "jobscout" {
		t.Fatalf("expected issuer 'jobscout', got %s", claims.Issuer)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret"

	claims := TokenClaims{
		Email: "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "jobscout",
			Subject:   "550e8400-e29b-41d4-a716-446655440000",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-80 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = ValidateToken(tokenStr, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	email := "test@example.com"
	secret := "test-secret"

	tokenStr, _ := GenerateToken(userID, email, secret)

	_, err := ValidateToken(tokenStr, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	_, err := ValidateToken("not-a-valid-jwt", "secret")
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := ValidateToken("", "secret")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}
