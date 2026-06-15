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