package jobs

import "jobscout/internal/jobsource"

// SearchResponse wraps the list of normalized jobs with metadata.
type SearchResponse struct {
	Jobs  []jobsource.NormalizedJob `json:"jobs"`
	Total int                       `json:"total"`
}
