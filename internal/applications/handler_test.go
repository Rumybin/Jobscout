package applications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"jobscout/internal/httputil"
	"jobscout/internal/middleware"
)

type mockService struct {
	saveFn         func(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error)
	getFn          func(ctx context.Context, userID, appID string) (*ApplicationResponse, error)
	listFn         func(ctx context.Context, userID string) ([]*ApplicationResponse, error)
	updateFn       func(ctx context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error)
	deleteFn       func(ctx context.Context, userID, appID string) error
	updateStatusFn func(ctx context.Context, userID, appID, newStatus string) (*ApplicationResponse, error)
	createNoteFn   func(ctx context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error)
	listNotesFn    func(ctx context.Context, userID, appID string) ([]*NoteResponse, error)
	deleteNoteFn   func(ctx context.Context, userID, noteID string) error
}

func (m *mockService) SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error) {
	return m.saveFn(ctx, userID, req)
}

func (m *mockService) GetApplication(ctx context.Context, userID, appID string) (*ApplicationResponse, error) {
	return m.getFn(ctx, userID, appID)
}

func (m *mockService) ListApplications(ctx context.Context, userID string) ([]*ApplicationResponse, error) {
	return m.listFn(ctx, userID)
}

func (m *mockService) UpdateApplication(ctx context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error) {
	return m.updateFn(ctx, userID, appID, req)
}

func (m *mockService) DeleteApplication(ctx context.Context, userID, appID string) error {
	return m.deleteFn(ctx, userID, appID)
}

func (m *mockService) UpdateApplicationStatus(ctx context.Context, userID, appID, newStatus string) (*ApplicationResponse, error) {
	return m.updateStatusFn(ctx, userID, appID, newStatus)
}

func (m *mockService) CreateApplicationNote(ctx context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error) {
	return m.createNoteFn(ctx, userID, appID, req)
}

func (m *mockService) ListApplicationNotes(ctx context.Context, userID, appID string) ([]*NoteResponse, error) {
	return m.listNotesFn(ctx, userID, appID)
}

func (m *mockService) DeleteApplicationNote(ctx context.Context, userID, noteID string) error {
	return m.deleteNoteFn(ctx, userID, noteID)
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
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Save_Manual_MissingTitle(t *testing.T) {
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

// Helper to create a request with chi URL params and user context.
func newRequestWithChiParams(method, path string, body []byte, userID string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if userID != "" {
		req = req.WithContext(context.WithValue(context.Background(), middleware.UserIDKey, userID))
	}

	return req
}

func setupChiContext(r *http.Request, params map[string]string) *http.Request {
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
}

func TestHandler_List_Success(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, userID string) ([]*ApplicationResponse, error) {
			return []*ApplicationResponse{
				{ID: "1", UserID: userID, Title: "Job A", CompanyName: "Co A", Source: "remotive", Status: "Wishlist", CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z"},
				{ID: "2", UserID: userID, Title: "Job B", CompanyName: "Co B", Source: "manual", Status: "Applied", CreatedAt: "2025-01-02T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z"},
			}, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications", nil, "user-1")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var apps []*ApplicationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &apps); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(apps) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(apps))
	}
}

func TestHandler_List_Empty(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, userID string) ([]*ApplicationResponse, error) {
			return []*ApplicationResponse{}, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications", nil, "user-1")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var apps []*ApplicationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &apps); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(apps) != 0 {
		t.Fatalf("expected 0 applications, got %d", len(apps))
	}
}

func TestHandler_List_Unauthorized(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, userID string) ([]*ApplicationResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications", nil, "")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_GetOne_Success(t *testing.T) {
	svc := &mockService{
		getFn: func(_ context.Context, userID, appID string) (*ApplicationResponse, error) {
			return &ApplicationResponse{
				ID: appID, UserID: userID, Title: "Job A", CompanyName: "Co A",
				Source: "remotive", Status: "Wishlist",
				CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z",
			}, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications/123", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "123"})

	rr := httptest.NewRecorder()
	h.GetOne(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp ApplicationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "123" || resp.Title != "Job A" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_GetOne_NotFound(t *testing.T) {
	svc := &mockService{
		getFn: func(_ context.Context, userID, appID string) (*ApplicationResponse, error) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications/999", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()
	h.GetOne(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_GetOne_Unauthorized(t *testing.T) {
	svc := &mockService{
		getFn: func(_ context.Context, userID, appID string) (*ApplicationResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications/123", nil, "")
	rr := httptest.NewRecorder()
	h.GetOne(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error) {
			return &ApplicationResponse{
				ID: appID, UserID: userID, Title: "Updated Title", CompanyName: "Updated Co",
				Source: "remotive", Status: "Wishlist",
				CreatedAt: "2025-01-01T00:00:00Z", UpdatedAt: "2025-01-02T00:00:00Z",
			}, nil
		},
	}
	h := NewHandler(svc)

	body := UpdateApplicationRequest{Title: strPtr("Updated Title"), CompanyName: strPtr("Updated Co")}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPatch, "/applications/123", payload, "user-1")
	req = setupChiContext(req, map[string]string{"id": "123"})

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp ApplicationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Title != "Updated Title" || resp.CompanyName != "Updated Co" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		},
	}
	h := NewHandler(svc)

	body := UpdateApplicationRequest{Title: strPtr("Updated Title")}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPatch, "/applications/999", payload, "user-1")
	req = setupChiContext(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodPatch, "/applications/123", []byte(`{invalid}`), "user-1")
	req = setupChiContext(req, map[string]string{"id": "123"})

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Update_Unauthorized(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	body := UpdateApplicationRequest{Title: strPtr("Updated Title")}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPatch, "/applications/123", payload, "")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, userID, appID string) error {
			return nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodDelete, "/applications/123", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "123"})

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, userID, appID string) error {
			return httputil.NotFoundError{Message: "application not found"}
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodDelete, "/applications/999", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "999"})

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_Delete_Unauthorized(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, userID, appID string) error {
			return nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodDelete, "/applications/123", nil, "")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_CreateNote_Success(t *testing.T) {
	svc := &mockService{
		createNoteFn: func(_ context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error) {
			return &NoteResponse{
				ID:            "note-1",
				ApplicationID: appID,
				UserID:        userID,
				Body:          req.Body,
				CreatedAt:     "2025-01-01T00:00:00Z",
			}, nil
		},
	}
	h := NewHandler(svc)

	body := CreateNoteRequest{Body: "Follow up next week"}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPost, "/applications/app-1/notes", payload, "user-1")
	req = setupChiContext(req, map[string]string{"id": "app-1"})

	rr := httptest.NewRecorder()
	h.CreateNote(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp NoteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ApplicationID != "app-1" || resp.Body != "Follow up next week" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_CreateNote_Unauthorized(t *testing.T) {
	svc := &mockService{
		createNoteFn: func(_ context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	body := CreateNoteRequest{Body: "Follow up next week"}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPost, "/applications/app-1/notes", payload, "")
	req = setupChiContext(req, map[string]string{"id": "app-1"})

	rr := httptest.NewRecorder()
	h.CreateNote(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_CreateNote_InvalidJSON(t *testing.T) {
	svc := &mockService{
		createNoteFn: func(_ context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodPost, "/applications/app-1/notes", []byte(`{invalid}`), "user-1")
	req = setupChiContext(req, map[string]string{"id": "app-1"})

	rr := httptest.NewRecorder()
	h.CreateNote(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_CreateNote_EmptyBody(t *testing.T) {
	svc := &mockService{
		createNoteFn: func(_ context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error) {
			return nil, httputil.ValidationError{Message: "body is required"}
		},
	}
	h := NewHandler(svc)

	body := CreateNoteRequest{Body: ""}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPost, "/applications/app-1/notes", payload, "user-1")
	req = setupChiContext(req, map[string]string{"id": "app-1"})

	rr := httptest.NewRecorder()
	h.CreateNote(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "body is required") {
		t.Fatalf("expected validation error about missing body, got: %s", rr.Body.String())
	}
}

func TestHandler_CreateNote_ApplicationNotFound(t *testing.T) {
	svc := &mockService{
		createNoteFn: func(_ context.Context, userID, appID string, req CreateNoteRequest) (*NoteResponse, error) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		},
	}
	h := NewHandler(svc)

	body := CreateNoteRequest{Body: "Follow up next week"}
	payload, _ := json.Marshal(body)

	req := newRequestWithChiParams(http.MethodPost, "/applications/missing/notes", payload, "user-1")
	req = setupChiContext(req, map[string]string{"id": "missing"})

	rr := httptest.NewRecorder()
	h.CreateNote(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_ListNotes_Success(t *testing.T) {
	svc := &mockService{
		listNotesFn: func(_ context.Context, userID, appID string) ([]*NoteResponse, error) {
			return []*NoteResponse{
				{ID: "note-1", ApplicationID: appID, UserID: userID, Body: "First note", CreatedAt: "2025-01-02T00:00:00Z"},
				{ID: "note-2", ApplicationID: appID, UserID: userID, Body: "Second note", CreatedAt: "2025-01-01T00:00:00Z"},
			}, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications/app-1/notes", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "app-1"})

	rr := httptest.NewRecorder()
	h.ListNotes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var notes []*NoteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &notes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
}

func TestHandler_ListNotes_Unauthorized(t *testing.T) {
	svc := &mockService{
		listNotesFn: func(_ context.Context, userID, appID string) ([]*NoteResponse, error) {
			return nil, nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications/app-1/notes", nil, "")
	req = setupChiContext(req, map[string]string{"id": "app-1"})

	rr := httptest.NewRecorder()
	h.ListNotes(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_ListNotes_ApplicationNotFound(t *testing.T) {
	svc := &mockService{
		listNotesFn: func(_ context.Context, userID, appID string) ([]*NoteResponse, error) {
			return nil, httputil.NotFoundError{Message: "application not found"}
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodGet, "/applications/missing/notes", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "missing"})

	rr := httptest.NewRecorder()
	h.ListNotes(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_DeleteNote_Success(t *testing.T) {
	svc := &mockService{
		deleteNoteFn: func(_ context.Context, userID, noteID string) error {
			return nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodDelete, "/notes/note-1", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "note-1"})

	rr := httptest.NewRecorder()
	h.DeleteNote(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_DeleteNote_Unauthorized(t *testing.T) {
	svc := &mockService{
		deleteNoteFn: func(_ context.Context, userID, noteID string) error {
			return nil
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodDelete, "/notes/note-1", nil, "")
	req = setupChiContext(req, map[string]string{"id": "note-1"})

	rr := httptest.NewRecorder()
	h.DeleteNote(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_DeleteNote_NotFound(t *testing.T) {
	svc := &mockService{
		deleteNoteFn: func(_ context.Context, userID, noteID string) error {
			return httputil.NotFoundError{Message: "note not found"}
		},
	}
	h := NewHandler(svc)

	req := newRequestWithChiParams(http.MethodDelete, "/notes/missing", nil, "user-1")
	req = setupChiContext(req, map[string]string{"id": "missing"})

	rr := httptest.NewRecorder()
	h.DeleteNote(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
