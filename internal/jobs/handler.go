package jobs

import (
	"net/http"
	"strconv"

	"jobscout/internal/httputil"
	"jobscout/internal/jobsource"
)

// Handler holds HTTP handlers for job search endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Search handles GET /jobs/search.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	filters := jobsource.SearchFilters{
		Keyword:                   r.URL.Query().Get("keyword"),
		Category:                  r.URL.Query().Get("category"),
		CompanyName:               r.URL.Query().Get("company_name"),
		JobType:                   r.URL.Query().Get("job_type"),
		CandidateRequiredLocation: r.URL.Query().Get("candidate_required_location"),
		SalaryText:                r.URL.Query().Get("salary_text"),
		PostedDate:                r.URL.Query().Get("posted_date"),
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be an integer between 1 and 100"})
			return
		}
		filters.Limit = limit
	}

	resp, err := h.svc.Search(r.Context(), filters)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
