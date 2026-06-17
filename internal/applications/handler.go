package applications

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"jobscout/internal/httputil"
	"jobscout/internal/middleware"
)

// ApplicationService defines the operations needed by the handler.
type ApplicationService interface {
	SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error)
	GetApplication(ctx context.Context, userID, appID string) (*ApplicationResponse, error)
	ListApplications(ctx context.Context, userID string) ([]*ApplicationResponse, error)
	UpdateApplication(ctx context.Context, userID, appID string, req UpdateApplicationRequest) (*ApplicationResponse, error)
	DeleteApplication(ctx context.Context, userID, appID string) error
	UpdateApplicationStatus(ctx context.Context, userID, appID string, newStatus string) (*ApplicationResponse, error)
}

// Handler holds HTTP handlers for application endpoints.
type Handler struct {
	svc ApplicationService
}

// NewHandler creates a new handler.
func NewHandler(svc ApplicationService) *Handler {
	return &Handler{svc: svc}
}

// Save handles POST /applications.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req SaveApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.svc.SaveApplication(r.Context(), userID, req)
	if err != nil {
		var valErr httputil.ValidationError
		if errors.As(err, &valErr) {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": valErr.Message})
			return
		}
		var confErr httputil.ConflictError
		if errors.As(err, &confErr) {
			httputil.WriteJSON(w, http.StatusConflict, map[string]string{"error": confErr.Message})
			return
		}
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// List handles GET /applications.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	apps, err := h.svc.ListApplications(r.Context(), userID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, apps)
}

// GetOne handles GET /applications/{id}.
func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	appID := chi.URLParam(r, "id")
	if appID == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing application id"})
		return
	}

	resp, err := h.svc.GetApplication(r.Context(), userID, appID)
	if err != nil {
		var notFoundErr httputil.NotFoundError
		if errors.As(err, &notFoundErr) {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]string{"error": notFoundErr.Message})
			return
		}
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// Update handles PATCH /applications/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	appID := chi.URLParam(r, "id")
	if appID == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing application id"})
		return
	}

	var req UpdateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.svc.UpdateApplication(r.Context(), userID, appID, req)
	if err != nil {
		var valErr httputil.ValidationError
		if errors.As(err, &valErr) {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": valErr.Message})
			return
		}
		var notFoundErr httputil.NotFoundError
		if errors.As(err, &notFoundErr) {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]string{"error": notFoundErr.Message})
			return
		}
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /applications/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	appID := chi.URLParam(r, "id")
	if appID == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing application id"})
		return
	}

	err := h.svc.DeleteApplication(r.Context(), userID, appID)
	if err != nil {
		var notFoundErr httputil.NotFoundError
		if errors.As(err, &notFoundErr) {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]string{"error": notFoundErr.Message})
			return
		}
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateStatus handles PATCH /applications/{id}/status.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	appID := chi.URLParam(r, "id")
	if appID == "" {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing application id"})
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.svc.UpdateApplicationStatus(r.Context(), userID, appID, req.Status)
	if err != nil {
		var valErr httputil.ValidationError
		if errors.As(err, &valErr) {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": valErr.Message})
			return
		}
		var notFoundErr httputil.NotFoundError
		if errors.As(err, &notFoundErr) {
			httputil.WriteJSON(w, http.StatusNotFound, map[string]string{"error": notFoundErr.Message})
			return
		}
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
