package jobsource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// NewRemotive creates a new RemotiveSource.
func NewRemotive(baseURL string) *RemotiveSource {
	if baseURL == "" {
		baseURL = "https://remotive.com/api/remote-jobs"
	}
	return &RemotiveSource{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// RemotiveSource implements JobSource for the Remotive Public API.
type RemotiveSource struct {
	baseURL string
	client  *http.Client
}

// remotiveJob represents a single job from the Remotive API response.
type remotiveJob struct {
	ID                        string `json:"id"`
	URL                       string `json:"url"`
	Title                     string `json:"title"`
	CompanyName               string `json:"company_name"`
	Category                  string `json:"category"`
	JobType                   string `json:"job_type"`
	CandidateRequiredLocation string `json:"candidate_required_location"`
	Salary                    string `json:"salary"`
	PublicationDate           string `json:"publication_date"`
	Description               string `json:"description"`
}

// Search performs a job search via the Remotive API and returns normalized results.
func (r *RemotiveSource) Search(ctx context.Context, filters SearchFilters) ([]NormalizedJob, error) {
	u, err := url.Parse(r.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid remotive base URL: %w", err)
	}

	q := u.Query()
	if filters.Keyword != "" {
		q.Set("search", filters.Keyword)
	}
	if filters.Category != "" {
		q.Set("category", filters.Category)
	}
	if filters.CompanyName != "" {
		q.Set("company_name", filters.CompanyName)
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create remotive request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("remotive request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remotive returned status %d: %s", resp.StatusCode, string(body))
	}

	var remotiveResp struct {
		Jobs []remotiveJob `json:"jobs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&remotiveResp); err != nil {
		return nil, fmt.Errorf("decode remotive response: %w", err)
	}

	normalized := make([]NormalizedJob, 0, len(remotiveResp.Jobs))
	for _, j := range remotiveResp.Jobs {
		normalized = append(normalized, NormalizedJob{
			Source:                     "remotive",
			ExternalID:                j.ID,
			Title:                     j.Title,
			CompanyName:               j.CompanyName,
			Category:                  j.Category,
			JobType:                   j.JobType,
			CandidateRequiredLocation: j.CandidateRequiredLocation,
			SalaryText:                j.Salary,
			ExternalURL:               j.URL,
			PublicationDate:           j.PublicationDate,
			Description:               j.Description,
		})
	}

	return applyClientSideFilters(normalized, filters), nil
}

// applyClientSideFilters applies additional filtering that the Remotive API does not support.
func applyClientSideFilters(jobs []NormalizedJob, filters SearchFilters) []NormalizedJob {
	if filters.JobType == "" && filters.CandidateRequiredLocation == "" && filters.SalaryText == "" && filters.PostedDate == "" {
		return jobs
	}

	filtered := make([]NormalizedJob, 0, len(jobs))
	for _, j := range jobs {
		if filters.JobType != "" && !strings.EqualFold(j.JobType, filters.JobType) {
			continue
		}
		if filters.CandidateRequiredLocation != "" {
			loc := strings.ToLower(j.CandidateRequiredLocation)
			needle := strings.ToLower(filters.CandidateRequiredLocation)
			if !strings.Contains(loc, needle) {
				continue
			}
		}
		if filters.SalaryText != "" {
			salary := strings.ToLower(j.SalaryText)
			needle := strings.ToLower(filters.SalaryText)
			if !strings.Contains(salary, needle) {
				continue
			}
		}
		if filters.PostedDate != "" && j.PublicationDate < filters.PostedDate {
			continue
		}
		filtered = append(filtered, j)
	}
	return filtered
}