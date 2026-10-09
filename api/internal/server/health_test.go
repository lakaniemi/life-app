package server

import (
	"net/http"
	"testing"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		databaseUp bool
		wantBody   string
	}{
		{
			name:       "database up",
			databaseUp: true,
			wantBody:   `{"status":"ok","database":"ok"}` + "\n",
		},
		{
			name:       "database down still responds ok",
			databaseUp: false,
			wantBody:   `{"status":"degraded","database":"unavailable"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)
			if !tt.databaseUp {
				// A closed pool fails every ping, like an unreachable database.
				s.pool.Close()
			}

			rec := s.do(t, http.MethodGet, "/health", "", nil)

			checkStatus(t, rec, http.StatusOK)
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}
