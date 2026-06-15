package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"jobscout/internal/httputil"
)

// AuthService defines the operations that the auth handler needs from the service layer.
type AuthService interface {
	Register(ctx context.Context, email, password string) (*AuthResponse, error)
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
}

// Handler holds HTTP handlers for authentication endpoints.
type Handler struct {
	svc AuthService
}

// NewHandler creates a new handler.
func NewHandler(svc AuthService) *Handler {
	return &Handler{svc: svc}
}

// Register handles POST /auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.svc.Register(r.Context(), strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		httputil.WriteJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.svc.Login(r.Context(), strings.TrimSpace(req.Email), req.Password)
	if err != nil {
		httputil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
