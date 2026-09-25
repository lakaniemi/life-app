package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	// Discard logs so test output stays readable.
	handler := New(slog.New(slog.DiscardHandler))

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "GET returns ok",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}` + "\n",
		},
		{
			name:       "POST is not allowed",
			method:     http.MethodPost,
			path:       "/health",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "unknown path is not found",
			method:     http.MethodGet,
			path:       "/nope",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody == "" {
				return
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}
			body, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
