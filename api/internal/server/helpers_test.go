package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/dbtest"
)

type testServer struct {
	handler http.Handler
	queries *db.Queries
	pool    *pgxpool.Pool
}

// newTestServer returns the full API handler on a fresh database, plus
// queries for seeding and the pool for the rare direct database check.
func newTestServer(t *testing.T) testServer {
	t.Helper()
	pool := dbtest.New(t)
	return testServer{
		handler: New(slog.New(slog.DiscardHandler), pool),
		queries: db.New(pool),
		pool:    pool,
	}
}

// do sends a request through the handler. body, if non-nil, is encoded as
// JSON; token, if non-empty, is sent as a Bearer token.
func (s testServer) do(t *testing.T, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequestWithContext(t.Context(), method, path, reqBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	return rec
}

// seedUser creates a user and a session for them, returning both and the
// session's token.
func (s testServer) seedUser(t *testing.T, name string) (db.User, db.Session, string) {
	t.Helper()
	user, err := s.queries.CreateUser(t.Context(), db.CreateUserParams{GoogleSub: "sub-" + name, Name: name})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	session, token := s.seedSession(t, user, time.Now().Add(sessionTTL))
	return user, session, token
}

func (s testServer) seedSession(t *testing.T, user db.User, expiresAt time.Time) (db.Session, string) {
	t.Helper()
	token := newRandomToken()
	session, err := s.queries.CreateSession(t.Context(), db.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return session, token
}

// decode decodes a JSON response body into a T.
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
	return v
}

func checkStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, want, rec.Body.String())
	}
}
