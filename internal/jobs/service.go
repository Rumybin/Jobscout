package jobs

import (
	"context"
	"jobscout/internal/jobsource"
)

// Service holds job search business logic.
type Service struct {
	source jobsource.JobSource
}

// NewService creates a new job search service.
func NewService(source jobsource.JobSource) *Service {
	return &Service{source: source}
}

// Search performs a job search using the configured external job source.
func (s *Service) Search(ctx context.Context, filters jobsource.SearchFilters) (*SearchResponse, error) {
	jobs, err := s.source.Search(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &SearchResponse{
		Jobs:  jobs,
		Total: len(jobs),
	}, nil
}
