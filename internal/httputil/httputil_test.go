package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusCreated, map[string]string{"status": "created"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json, got %q", rec.Header().Get("Content-Type"))
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "created" {
		t.Fatalf("expected created status, got %q", body["status"])
	}
}

func TestErrorTypes(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{name: "validation", err: ValidationError{Message: "invalid request"}, want: "invalid request"},
		{name: "conflict", err: ConflictError{Message: "already exists"}, want: "already exists"},
		{name: "not found", err: NotFoundError{Message: "missing"}, want: "missing"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.err.Error())
			}
		})
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})
	}
}
