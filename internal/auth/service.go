package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"jobscout/internal/config"
	"jobscout/internal/database/sqlc"
)

// Service handles authentication business logic.
type Service struct {
	pool   *pgxpool.Pool
	secret string
}

// NewService creates a new auth service.
func NewService(pool *pgxpool.Pool, cfg *config.Config) *Service {
	return &Service{
		pool:   pool,
		secret: cfg.JWTSecret,
	}
}

// Register creates a new user with the given email and plaintext password.
// It returns a JWT on success.
func (s *Service) Register(ctx context.Context, email, password string) (*AuthResponse, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	q := sqlc.New(s.pool)
	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := GenerateToken(idToString(user.ID), email, s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{Token: token}, nil
}

// Login verifies the email and password and returns a JWT on success.
func (s *Service) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	q := sqlc.New(s.pool)
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := GenerateToken(idToString(user.ID), email, s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{Token: token}, nil
}

// GetUser returns the user data for the given user ID.
func (s *Service) GetUser(ctx context.Context, userID string) (*MeResponse, error) {
	uid, err := stringToPgUUID(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	q := sqlc.New(s.pool)
	user, err := q.GetUserByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &MeResponse{
		ID:        idToString(user.ID),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

func idToString(uuid pgtype.UUID) string {
	src := uuid.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", src[0:4], src[4:6], src[6:8], src[8:10], src[10:16])
}

func stringToPgUUID(s string) (pgtype.UUID, error) {
	var uid pgtype.UUID
	if err := uid.Scan(s); err != nil {
		return uid, err
	}
	return uid, nil
}
