package httputil

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with the given status code and value.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ValidationError represents a validation error (400 Bad Request).
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// ConflictError represents a conflict error (409 Conflict).
type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

// NotFoundError represents a not-found error (404).
type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}
