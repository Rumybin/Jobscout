package applications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobscout/internal/middleware"
)

type mockService struct {
	saveFn func(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error)
}

func (m *mockService) SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
	return m.saveFn(ctx, userID, req)
}

func TestHandler_Save_Success(t *testing.T) {
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
