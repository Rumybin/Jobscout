package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobscout/internal/middleware"
)

type mockSummaryService struct {
	summaryFn func(ctx context.Context, userID string) (*SummaryResponse, error)
}

func (m *mockSummaryService) Summary(ctx context.Context, userID string) (*SummaryResponse, error) {
	return m.summaryFn(ctx, userID)
}

func TestHandler_Summary_Success(t *testing.T) {
	svc := &mockSummaryService{
		summaryFn: func(_ context.Context, userID string) (*SummaryResponse, error) {
			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return &SummaryResponse{
				TotalSavedJobs: 4,
				ByStatus: map[string]int64{
					"Applied":  2,
					"Wishlist": 2,
				},
				BySource: map[string]int64{
					"manual":   1,
					"remotive": 3,
				},
				ByCategory: map[string]int64{
					"Software Development": 3,
					"Uncategorized":        1,
				},
				SavedPerMonth: []MonthCount{
					{Month: "2026-05", Count: 1},
					{Month: "2026-06", Count: 3},
				},
			}, nil
		},
	}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp SummaryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.TotalSavedJobs != 4 {
		t.Fatalf("expected 4 saved jobs, got %d", resp.TotalSavedJobs)
	}
	if resp.ByStatus["Applied"] != 2 {
		t.Fatalf("expected 2 applied jobs, got %d", resp.ByStatus["Applied"])
	}
	if len(resp.SavedPerMonth) != 2 || resp.SavedPerMonth[1].Month != "2026-06" {
		t.Fatalf("unexpected saved_per_month: %+v", resp.SavedPerMonth)
	}
}

func TestHandler_Summary_Empty(t *testing.T) {
	svc := &mockSummaryService{
		summaryFn: func(_ context.Context, _ string) (*SummaryResponse, error) {
			return &SummaryResponse{
				ByStatus:      map[string]int64{},
				BySource:      map[string]int64{},
				ByCategory:    map[string]int64{},
				SavedPerMonth: []MonthCount{},
			}, nil
		},
	}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp SummaryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.TotalSavedJobs != 0 {
		t.Fatalf("expected 0 saved jobs, got %d", resp.TotalSavedJobs)
	}
	if len(resp.ByStatus) != 0 || len(resp.SavedPerMonth) != 0 {
		t.Fatalf("expected empty analytics, got %+v", resp)
	}
}

func TestHandler_Summary_Unauthorized(t *testing.T) {
	svc := &mockSummaryService{
		summaryFn: func(_ context.Context, _ string) (*SummaryResponse, error) {
			t.Fatal("service should not be called")
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Summary_ServiceError(t *testing.T) {
	svc := &mockSummaryService{
		summaryFn: func(_ context.Context, _ string) (*SummaryResponse, error) {
			return nil, errors.New("database unavailable")
		},
	}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.Summary(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
