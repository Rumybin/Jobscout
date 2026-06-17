package jobsource

import "context"

// NormalizedJob represents a job listing normalized from any external source.
type NormalizedJob struct {
	Source                    string `json:"source"`
	ExternalID                string `json:"external_id"`
	Title                     string `json:"title"`
	CompanyName               string `json:"company_name"`
	Category                  string `json:"category"`
	JobType                   string `json:"job_type"`
	CandidateRequiredLocation string `json:"candidate_required_location"`
	SalaryText                string `json:"salary_text"`
	ExternalURL               string `json:"external_url"`
	PublicationDate           string `json:"publication_date"`
	Description               string `json:"description"`
}

// SearchFilters holds all job search parameters.
type SearchFilters struct {
	Keyword                   string `json:"keyword,omitempty"`
	Category                  string `json:"category,omitempty"`
	CompanyName               string `json:"company_name,omitempty"`
	JobType                   string `json:"job_type,omitempty"`
	CandidateRequiredLocation string `json:"candidate_required_location,omitempty"`
	SalaryText                string `json:"salary_text,omitempty"`
	PostedDate                string `json:"posted_date,omitempty"`
	Limit                     int    `json:"limit,omitempty"`
}

// JobSource is the interface that all external job sources implement.
type JobSource interface {
	Search(ctx context.Context, filters SearchFilters) ([]NormalizedJob, error)
}
