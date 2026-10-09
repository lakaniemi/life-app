package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lakaniemi/life-app/api/internal/dbtest"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		databaseUp bool
		wantStatus int
		wantBody   string
	}{
		{
			name:       "database up",
			databaseUp: true,
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok","database":"ok"}` + "\n",
		},
		{
			name:       "database down still responds ok",
			databaseUp: false,
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"degraded","database":"unavailable"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := dbtest.New(t)
			if !tt.databaseUp {
				// A closed pool fails every ping, like an unreachable database.
				pool.Close()
			}
			handler := New(slog.New(slog.DiscardHandler), pool, fakeVerifier{})
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
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
