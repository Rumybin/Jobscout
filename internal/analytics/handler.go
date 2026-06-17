package analytics

import (
	"context"
	"net/http"

	"jobscout/internal/httputil"
	"jobscout/internal/middleware"
)

// SummaryService defines the operations needed by the handler.
type SummaryService interface {
	Summary(ctx context.Context, userID string) (*SummaryResponse, error)
}

// Handler holds HTTP handlers for analytics endpoints.
type Handler struct {
	svc SummaryService
}

// NewHandler creates a new handler.
func NewHandler(svc SummaryService) *Handler {
	return &Handler{svc: svc}
}

// Summary handles GET /analytics/summary.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := h.svc.Summary(r.Context(), userID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
