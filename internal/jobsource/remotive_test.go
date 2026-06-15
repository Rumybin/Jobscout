package jobsource

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemotiveSource_Search_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Jobs []remotiveJob `json:"jobs"`
		}{
			Jobs: []remotiveJob{
				{
					ID:                        "123",
					URL:                       "https://example.com/123",
					Title:                     "Backend Engineer",
					CompanyName:               "TechCorp",
					Category:                  "Software Development",
					JobType:                   "full_time",
					CandidateRequiredLocation: "Worldwide",
					Salary:                    "",
					PublicationDate:           "2025-01-15T10:00:00Z",
					Description:               "Backend role.",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	src := NewRemotive(server.URL)
	jobs, err := src.Search(context.Background(), SearchFilters{Keyword: "backend"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Title != "Backend Engineer" {
		t.Errorf("expected Backend Engineer, got %q", jobs[0].Title)
	}
	if jobs[0].Source != "remotive" {
		t.Errorf("expected remotive source, got %q", jobs[0].Source)
	}
}

func TestRemotiveSource_Search_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Jobs []remotiveJob `json:"jobs"`
		}{Jobs: []remotiveJob{}})
	}))
	defer server.Close()

	src := NewRemotive(server.URL)
	jobs, err := src.Search(context.Background(), SearchFilters{Keyword: "nomatches"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

func TestRemotiveSource_Search_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	src := NewRemotive(server.URL)
	_, err := src.Search(context.Background(), SearchFilters{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
