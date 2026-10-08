package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireAuthRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// authorization returns the Authorization header to send, given a valid
		// token and an expired one.
		authorization func(validToken, expiredToken string) string
	}{
		{
			name:          "missing header",
			authorization: func(string, string) string { return "" },
		},
		{
			name:          "non-bearer scheme",
			authorization: func(valid, _ string) string { return "Basic " + valid },
		},
		{
			name:          "unknown token",
			authorization: func(string, string) string { return "Bearer " + newRandomToken() },
		},
		{
			name:          "expired session",
			authorization: func(_, expired string) string { return "Bearer " + expired },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)
			user, _, validToken := s.seedUser(t, "Alice")
			_, expiredToken := s.seedSession(t, user, time.Now().Add(-time.Minute))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil)
			if h := tt.authorization(validToken, expiredToken); h != "" {
				req.Header.Set("Authorization", h)
			}
			rec := httptest.NewRecorder()
			s.handler.ServeHTTP(rec, req)

			checkStatus(t, rec, http.StatusUnauthorized)
			if rec.Header().Get("WWW-Authenticate") == "" {
				t.Error("WWW-Authenticate header is missing")
			}
		})
	}
}

func TestRequireAuthAcceptsLowercaseScheme(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)
	_, _, token := s.seedUser(t, "Alice")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "bearer "+token)
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)

	checkStatus(t, rec, http.StatusOK)
}

func TestRequireAuthExtendsSessionAfterInterval(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)
	_, session, token := s.seedUser(t, "Alice")
	// A session last used two days ago that expires in an hour. The API
	// doesn't expose session timestamps, so this test reads the database.
	_, err := s.pool.Exec(t.Context(),
		`UPDATE sessions SET last_used_at = now() - interval '2 days', expires_at = now() + interval '1 hour' WHERE id = $1`,
		session.ID)
	if err != nil {
		t.Fatalf("backdate session: %v", err)
	}

	checkStatus(t, s.do(t, http.MethodGet, "/me", token, nil), http.StatusOK)

	var expiresAt time.Time
	if err := s.pool.QueryRow(t.Context(), `SELECT expires_at FROM sessions WHERE id = $1`, session.ID).Scan(&expiresAt); err != nil {
		t.Fatalf("read session: %v", err)
	}
	if minWant := time.Now().Add(sessionTTL - time.Hour); expiresAt.Before(minWant) {
		t.Errorf("expires_at = %v, want it extended to about %v from now", expiresAt, sessionTTL)
	}
}
