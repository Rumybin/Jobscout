package analytics

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"jobscout/internal/database/sqlc"
)

// Service handles analytics business logic.
type Service struct {
	queries *sqlc.Queries
}

// NewService creates a new Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		queries: sqlc.New(pool),
	}
}

// Summary returns saved-job analytics for one user.
func (s *Service) Summary(ctx context.Context, userID string) (*SummaryResponse, error) {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	total, err := s.queries.CountApplicationsByUser(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %w", err)
	}

	byStatus, err := s.countByStatus(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	bySource, err := s.countBySource(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	byCategory, err := s.countByCategory(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	savedPerMonth, err := s.countBySavedMonth(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return &SummaryResponse{
		TotalSavedJobs: total,
		ByStatus:       byStatus,
		BySource:       bySource,
		ByCategory:     byCategory,
		SavedPerMonth:  savedPerMonth,
	}, nil
}

func (s *Service) countByStatus(ctx context.Context, userID pgtype.UUID) (map[string]int64, error) {
	rows, err := s.queries.CountApplicationsByStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications by status: %w", err)
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.Total
	}

	return counts, nil
}

func (s *Service) countBySource(ctx context.Context, userID pgtype.UUID) (map[string]int64, error) {
	rows, err := s.queries.CountApplicationsBySource(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications by source: %w", err)
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Source] = row.Total
	}

	return counts, nil
}

func (s *Service) countByCategory(ctx context.Context, userID pgtype.UUID) (map[string]int64, error) {
	rows, err := s.queries.CountApplicationsByCategory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications by category: %w", err)
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Category] = row.Total
	}

	return counts, nil
}

func (s *Service) countBySavedMonth(ctx context.Context, userID pgtype.UUID) ([]MonthCount, error) {
	rows, err := s.queries.CountApplicationsBySavedMonth(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications by saved month: %w", err)
	}

	counts := make([]MonthCount, len(rows))
	for i, row := range rows {
		counts[i] = MonthCount{
			Month: row.Month,
			Count: row.Total,
		}
	}

	return counts, nil
}
