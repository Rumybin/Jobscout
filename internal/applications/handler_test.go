package applications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jobscout/internal/httputil"
	"jobscout/internal/middleware"
)

type mockService struct {
	saveFn func(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error)
}

func (m *mockService) SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
	return m.saveFn(ctx, userID, req)
}

func TestHandler_Save_Success_External(t *testing.T) {
	svc := &mockService{
		saveFn: func(_ context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
			return &ApplicationResponse{
				ID:          "fake-uuid",
				UserID:      userID,
				Source:      req.Source,
				ExternalID:  req.ExternalID,
				Title:       req.Title,
				CompanyName: req.CompanyName,
				Status:      "Wishlist",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			}, nil
		},
	}
	h := NewHandler(svc)

	body := SaveApplicationRequest{
		Source:      "remotive",
		ExternalID:  "ext123",
		Title:       "Backend Engineer",
		CompanyName: "RemoteCo",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewReader(payload))
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp ApplicationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Source != "remotive" || resp.Title != "Backend Engineer" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_Save_Success_Manual_WithoutSource(t *testing.T) {
	svc := &mockService{
		saveFn: func(_ context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
			// Simulate server-side behavior: source defaults to "manual", external_id is generated
			mockExternalID := "00000000-0000-0000-0000-000000000001"
			if req.Source == "" {
				req.Source = "manual"
			}
			if req.ExternalID == "" {
				req.ExternalID = mockExternalID
			}
			return &ApplicationResponse{
				ID:          "fake-uuid-2",
				UserID:      userID,
				Source:      req.Source,
				ExternalID:  req.ExternalID,
				Title:       req.Title,
				CompanyName: req.CompanyName,
				Status:      "Wishlist",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			}, nil
		},
	}
	h := NewHandler(svc)

	body := SaveApplicationRequest{
		Title:       "Manual Entry",
		CompanyName: "Some Company",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewReader(payload))
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp ApplicationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Source != "manual" {
		t.Fatalf("expected source 'manual', got '%s'", resp.Source)
	}
	if resp.Title != "Manual Entry" {
		t.Fatalf("expected title 'Manual Entry', got '%s'", resp.Title)
	}
}

func TestHandler_Save_Success_Manual_WithSource(t *testing.T) {
	svc := &mockService{
		saveFn: func(_ context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
			// Simulate server-side: when source is "manual", external_id is generated
			mockExternalID := "00000000-0000-0000-0000-000000000002"
			if req.Source == "manual" && req.ExternalID == "" {
				req.ExternalID = mockExternalID
			}
			return &ApplicationResponse{
				ID:          "fake-uuid-3",
				UserID:      userID,
				Source:      req.Source,
				ExternalID:  req.ExternalID,
				Title:       req.Title,
				CompanyName: req.CompanyName,
				Status:      "Wishlist",
				CreatedAt:   "2025-01-01T00:00:00Z",
				UpdatedAt:   "2025-01-01T00:00:00Z",
			}, nil
		},
	}
	h := NewHandler(svc)

	body := SaveApplicationRequest{
		Source:      "manual",
		Title:       "Manual With Source",
		CompanyName: "Another Co",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewReader(payload))
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_Save_InvalidJSON(t *testing.T) {
	svc := &mockService{
		saveFn: func(_ context.Context, _ string, _ SaveApplicationRequest) (*ApplicationResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewReader([]byte(`{invalid}`)))
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Save_Unauthorized(t *testing.T) {
	svc := &mockService{
		saveFn: func(_ context.Context, _ string, _ SaveApplicationRequest) (*ApplicationResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	body := SaveApplicationRequest{
		Source:      "remotive",
		ExternalID:  "ext123",
		Title:       "Backend Engineer",
		CompanyName: "RemoteCo",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewReader(payload))
	// No user ID in context
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Save_Manual_MissingTitle(t *testing.T) {
	// This tests that the service validation catches missing title even for manual entries
	svc := &mockService{
		saveFn: func(_ context.Context, _ string, _ SaveApplicationRequest) (*ApplicationResponse, error) {
			return nil, httputil.ValidationError{Message: "title is required"}
		},
	}
	h := NewHandler(svc)

	body := SaveApplicationRequest{
		CompanyName: "Some Company",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewReader(payload))
	req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, "user-1"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "title is required") {
		t.Fatalf("expected validation error about missing title, got: %s", rr.Body.String())
	}
}
