package applications

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
func (s *Service) SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
	req.Source = strings.TrimSpace(req.Source)
	req.Title = strings.TrimSpace(req.Title)
	req.CompanyName = strings.TrimSpace(req.CompanyName)

	isManual := req.Source == "" || req.Source == "manual"

	if isManual {
		if req.Source == "" {
			req.Source = "manual"
		}
		req.ExternalID = uuid.New().String()
	} else {
		if req.ExternalID == "" {
			return nil, httputil.ValidationError{Message: "external_id is required for non-manual sources"}
		}
	}

	if req.Title == "" {
		return nil, httputil.ValidationError{Message: "title is required"}
	}
	if req.CompanyName == "" {
		return nil, httputil.ValidationError{Message: "company_name is required"}
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "Wishlist"
	}
	if !validStatuses[status] {
		return nil, httputil.ValidationError{Message: fmt.Sprintf("invalid status: %s. Allowed: Wishlist, Applied, Screening, Technical Test, Interview, Offer, Rejected, Withdrawn", status)}
	}

	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	pubDate := pgtype.Timestamptz{}
	if req.PublicationDate != nil && *req.PublicationDate != "" {
		t, err := time.Parse(time.RFC3339, *req.PublicationDate)
		if err != nil {
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.ConflictError{Message: "job already saved"}
		}
		return nil, fmt.Errorf("failed to save application: %w", err)
	}

	return mapApplicationToResponse(app), nil
}

// GetApplication returns a single application by ID, scoped to the user.
func (s *Service) GetApplication(ctx context.Context, userID, appID string) (*ApplicationResponse, error) {
	var userUUID, appUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	if err := appUUID.Scan(appID); err != nil {
		return nil, httputil.ValidationError{Message: "invalid application ID"}
	}

	app, err := s.queries.GetApplicationByID(ctx, sqlc.GetApplicationByIDParams{
		ID:     appUUID,
		UserID: userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return mapApplicationToResponse(app), nil
}

// ListApplications returns all applications for the user.
func (s *Service) ListApplications(ctx context.Context, userID string) ([]*ApplicationResponse, error) {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	apps, err := s.queries.GetApplicationsByUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	responses := make([]*ApplicationResponse, len(apps))
	for i, app := range apps {
		responses[i] = mapApplicationToResponse(app)
	}

	return responses, nil
}

// UpdateApplication updates editable fields of an application, scoped to the user.
func (s *Service) UpdateApplication(ctx context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error) {
	var userUUID, appUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	if err := appUUID.Scan(appID); err != nil {
		return nil, httputil.ValidationError{Message: "invalid application ID"}
	}

	title := ""
	if req.Title != nil {
		title = *req.Title
	}
	companyName := ""
	if req.CompanyName != nil {
		companyName = *req.CompanyName
	}
	category := ""
	if req.Category != nil {
		category = *req.Category
	}
	jobType := ""
	if req.JobType != nil {
		jobType = *req.JobType
	}
	location := ""
	if req.CandidateRequiredLocation != nil {
		location = *req.CandidateRequiredLocation
	}
	salary := ""
	if req.SalaryText != nil {
		salary = *req.SalaryText
	}
	url := ""
	if req.ExternalURL != nil {
		url = *req.ExternalURL
	}
	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	pubDate := pgtype.Timestamptz{}
	if req.PublicationDate != nil && *req.PublicationDate != "" {
		t, err := time.Parse(time.RFC3339, *req.PublicationDate)
		if err != nil {
			t, err = time.Parse("2006-01-02", *req.PublicationDate)
			if err != nil {
				return nil, httputil.ValidationError{Message: "invalid publication_date format, use RFC3339 or YYYY-MM-DD"}
			}
		}
		if err := pubDate.Scan(t); err != nil {
			return nil, httputil.ValidationError{Message: "invalid publication_date"}
		}
	}

	arg := sqlc.UpdateApplicationParams{
		ID:                              appUUID,
		UserID:                          userUUID,
		UpdateTitle:                     req.Title != nil,
		Title:                           title,
		UpdateCompanyName:               req.CompanyName != nil,
		CompanyName:                     companyName,
		UpdateCategory:                  req.Category != nil,
		Category:                        category,
		UpdateJobType:                   req.JobType != nil,
		JobType:                         jobType,
		UpdateCandidateRequiredLocation: req.CandidateRequiredLocation != nil,
		CandidateRequiredLocation:       location,
		UpdateSalaryText:                req.SalaryText != nil,
		SalaryText:                      salary,
		UpdateExternalUrl:               req.ExternalURL != nil,
		ExternalUrl:                     url,
		UpdatePublicationDate:           req.PublicationDate != nil,
		PublicationDate:                 pubDate,
		UpdateDescription:               req.Description != nil,
		Description:                     description,
	}

	app, err := s.queries.UpdateApplication(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		}
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return mapApplicationToResponse(app), nil
}

// DeleteApplication deletes an application by ID, scoped to the user.
func (s *Service) DeleteApplication(ctx context.Context, userID, appID string) error {
	var userUUID, appUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	if err := appUUID.Scan(appID); err != nil {
		return httputil.ValidationError{Message: "invalid application ID"}
	}

	_, err := s.queries.DeleteApplication(ctx, sqlc.DeleteApplicationParams{
		ID:     appUUID,
		UserID: userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httputil.NotFoundError{Message: "application not found"}
		}
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return nil
}

// UpdateApplicationStatus updates only the status field of an application, scoped to the user.
func (s *Service) UpdateApplicationStatus(ctx context.Context, userID, appID string, newStatus string) (*ApplicationResponse, error) {
	var userUUID, appUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	if err := appUUID.Scan(appID); err != nil {
		return nil, httputil.ValidationError{Message: "invalid application ID"}
	}

	newStatus = strings.TrimSpace(newStatus)
	if newStatus == "" {
		return nil, httputil.ValidationError{Message: "status is required"}
	}
	if !validStatuses[newStatus] {
		return nil, httputil.ValidationError{Message: fmt.Sprintf("invalid status: %s. Allowed: Wishlist, Applied, Screening, Technical Test, Interview, Offer, Rejected, Withdrawn", newStatus)}
	}

	arg := sqlc.UpdateApplicationStatusParams{
		ID:     appUUID,
		UserID: userUUID,
		Status: newStatus,
	}

	app, err := s.queries.UpdateApplicationStatus(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		}
		return nil, fmt.Errorf("failed to update application status: %w", err)
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
