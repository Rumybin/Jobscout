package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobscout/internal/jobsource"
)

// mockSource implements jobsource.JobSource for testing.
type mockSource struct {
	jobs []jobsource.NormalizedJob
	err  error
}

func (m *mockSource) Search(_ context.Context, _ jobsource.SearchFilters) ([]jobsource.NormalizedJob, error) {
	return m.jobs, m.err
}

func TestSearchHandler_Success(t *testing.T) {
	normalizedJobs := []jobsource.NormalizedJob{
		{
			Source:                     "remotive",
			ExternalID:                "123",
			Title:                     "Backend Engineer",
			CompanyName:               "TechCorp",
			Category:                  "Software Development",
			JobType:                   "full_time",
			CandidateRequiredLocation: "Worldwide",
			SalaryText:                "$80k - $100k",
			ExternalURL:               "https://example.com/123",
			PublicationDate:           "2025-01-15T10:00:00Z",
			Description:               "Backend role.",
		},
	}

	svc := NewService(&mockSource{jobs: normalizedJobs})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/search?keyword=backend&limit=5", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 job, got %d", resp.Total)
	}
	if len(resp.Jobs) != 1 {
		t.Fatalf("expected 1 job in list, got %d", len(resp.Jobs))
	}
	if resp.Jobs[0].ExternalID != "123" {
		t.Errorf("expected ExternalID 123, got %q", resp.Jobs[0].ExternalID)
	}
}

func TestSearchHandler_EmptyResult(t *testing.T) {
	svc := NewService(&mockSource{jobs: []jobsource.NormalizedJob{}})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/search", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("expected 0 total, got %d", resp.Total)
	}
	if len(resp.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(resp.Jobs))
	}
}

func TestSearchHandler_InvalidLimit(t *testing.T) {
	svc := NewService(&mockSource{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/search?limit=invalid", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSearchHandler_LimitTooHigh(t *testing.T) {
	svc := NewService(&mockSource{})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/search?limit=200", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSearchHandler_SourceError(t *testing.T) {
	svc := NewService(&mockSource{err: context.DeadlineExceeded})
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/search", nil)
	rec := httptest.NewRecorder()

	handler.Search(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
