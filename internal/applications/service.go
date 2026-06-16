package applications

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jobscout/internal/database/sqlc"
	"jobscout/internal/httputil"
)

var validStatuses = map[string]bool{
	"Wishlist":       true,
	"Applied":        true,
	"Screening":      true,
	"Technical Test": true,
	"Interview":      true,
	"Offer":          true,
	"Rejected":       true,
	"Withdrawn":      true,
}

// Service handles application business logic.
type Service struct {
	queries *sqlc.Queries
}

// NewService creates a new Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		queries: sqlc.New(pool),
	}
}

// SaveApplication saves a new application for the user.
// If the application already exists (unique constraint violation), it returns a ConflictError.
func (s *Service) SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
	// Validate required fields
	req.Source = strings.TrimSpace(req.Source)
	req.Title = strings.TrimSpace(req.Title)
	req.CompanyName = strings.TrimSpace(req.CompanyName)

	if req.Source == "" {
		return nil, httputil.ValidationError{Message: "source is required"}
	}
	if req.Title == "" {
		return nil, httputil.ValidationError{Message: "title is required"}
	}
	if req.CompanyName == "" {
		return nil, httputil.ValidationError{Message: "company_name is required"}
	}

	// Set default status if not provided
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "Wishlist"
	}
	if !validStatuses[status] {
		return nil, httputil.ValidationError{Message: fmt.Sprintf("invalid status: %s. Allowed: Wishlist, Applied, Screening, Technical Test, Interview, Offer, Rejected, Withdrawn", status)}
	}

	// Parse user ID string to pgtype.UUID
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse optional publication date to pgtype.Timestamptz
	var pubDate pgtype.Timestamptz
	if req.PublicationDate != nil && *req.PublicationDate != "" {
		t, err := time.Parse(time.RFC3339, *req.PublicationDate)
		if err != nil {
			// Try ISO date without time
			t, err = time.Parse("2006-01-02", *req.PublicationDate)
			if err != nil {
				return nil, httputil.ValidationError{Message: "invalid publication_date format, use RFC3339 or YYYY-MM-DD"}
			}
		}
		if err := pubDate.Scan(t); err != nil {
			return nil, httputil.ValidationError{Message: "invalid publication_date"}
		}
	}

	arg := sqlc.CreateApplicationParams{
		UserID:                    userUUID,
		Source:                    req.Source,
		ExternalID:                req.ExternalID,
		Title:                     req.Title,
		CompanyName:               req.CompanyName,
		Category:                  req.Category,
		JobType:                   req.JobType,
		CandidateRequiredLocation: req.CandidateRequiredLocation,
		SalaryText:                req.SalaryText,
		ExternalUrl:               req.ExternalURL,
		PublicationDate:           pubDate,
		Description:               req.Description,
		Status:                    status,
	}

	app, err := s.queries.CreateApplication(ctx, arg)
	if err != nil {
		// Check for unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.ConflictError{Message: "job already saved"}
		}
		return nil, fmt.Errorf("failed to save application: %w", err)
	}

	return mapApplicationToResponse(app), nil
}

func mapApplicationToResponse(app sqlc.Application) *ApplicationResponse {
	resp := &ApplicationResponse{
		ID:                        app.ID.String(),
		UserID:                    app.UserID.String(),
		Source:                    app.Source,
		ExternalID:                app.ExternalID,
		Title:                     app.Title,
		CompanyName:               app.CompanyName,
		Category:                  app.Category,
		JobType:                   app.JobType,
		CandidateRequiredLocation: app.CandidateRequiredLocation,
		SalaryText:                app.SalaryText,
		ExternalURL:               app.ExternalUrl,
		Description:               app.Description,
		Status:                    app.Status,
		CreatedAt:                 app.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:                 app.UpdatedAt.Time.Format(time.RFC3339),
	}

	if app.PublicationDate.Valid {
		resp.PublicationDate = strPtr(app.PublicationDate.Time.Format(time.RFC3339))
	}

	return resp
}

func strPtr(s string) *string {
	return &s
}
