package applications

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"jobscout/internal/httputil"
	"jobscout/internal/middleware"
)

// ApplicationService defines the operations needed by the handler.
type ApplicationService interface {
	SaveApplication(ctx context.Context, userID string, req SaveApplicationRequest) (*ApplicationResponse, error)
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
