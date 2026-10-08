package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/googleauth"
)

// newNonce gets a nonce the way the app does.
func (s testServer) newNonce(t *testing.T) string {
	t.Helper()
	rec := s.do(t, http.MethodPost, "/auth/nonce", "", nil)
	checkStatus(t, rec, http.StatusOK)
	return decode[nonceResponse](t, rec).Nonce
}

// newGoogleToken registers an ID token that the fake Google verifier accepts.
func (s testServer) newGoogleToken(identity googleauth.Identity) string {
	token := "id-token-" + newRandomToken()
	s.googleTokens[token] = identity
	return token
}

func (s testServer) login(t *testing.T, idToken string) *httptest.ResponseRecorder {
	t.Helper()
	return s.do(t, http.MethodPost, "/auth/google", "", map[string]any{"idToken": idToken})
}

func TestCreateNonce(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)

	first, second := s.newNonce(t), s.newNonce(t)

	if first == "" || first == second {
		t.Errorf("nonces = %q, %q, want distinct non-empty values", first, second)
	}
}

func TestGoogleLogin(t *testing.T) {
	t.Parallel()

	t.Run("first login creates the user and a working session", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		idToken := s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice", Nonce: s.newNonce(t)})

		rec := s.login(t, idToken)

		checkStatus(t, rec, http.StatusOK)
		got := decode[googleLoginResponse](t, rec)
		if got.Token == "" || got.User.Name != "Alice" {
			t.Fatalf("response = %+v, want a token and user Alice", got)
		}
		me := decode[meResponse](t, s.do(t, http.MethodGet, "/me", got.Token, nil))
		if me.User != got.User {
			t.Errorf("GET /me user = %+v, want %+v", me.User, got.User)
		}
	})

	t.Run("later login finds the user and keeps their name", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		first := decode[googleLoginResponse](t, s.login(t,
			s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice", Nonce: s.newNonce(t)})))

		rec := s.login(t, s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice Renamed", Nonce: s.newNonce(t)}))

		checkStatus(t, rec, http.StatusOK)
		got := decode[googleLoginResponse](t, rec)
		if got.User != first.User {
			t.Errorf("user = %+v, want the existing %+v", got.User, first.User)
		}
		if got.Token == first.Token {
			t.Error("second login reused the first session token, want a new session")
		}
	})

	t.Run("missing id token", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)

		checkStatus(t, s.do(t, http.MethodPost, "/auth/google", "", map[string]any{}), http.StatusBadRequest)
	})
}

func TestGoogleLoginRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		// idToken returns the ID token to log in with.
		idToken func(t *testing.T, s testServer) string
	}{
		{
			name:    "token Google didn't issue",
			idToken: func(*testing.T, testServer) string { return "forged" },
		},
		{
			name: "token without nonce",
			idToken: func(_ *testing.T, s testServer) string {
				return s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice"})
			},
		},
		{
			name: "nonce we never issued",
			idToken: func(_ *testing.T, s testServer) string {
				return s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice", Nonce: newRandomToken()})
			},
		},
		{
			name: "replayed token",
			idToken: func(t *testing.T, s testServer) string {
				idToken := s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice", Nonce: s.newNonce(t)})
				checkStatus(t, s.login(t, idToken), http.StatusOK)
				return idToken
			},
		},
		{
			name: "expired nonce",
			idToken: func(t *testing.T, s testServer) string {
				nonce := newRandomToken()
				err := s.queries.CreateNonce(t.Context(), db.CreateNonceParams{Nonce: nonce, ExpiresAt: time.Now().Add(-time.Minute)})
				if err != nil {
					t.Fatalf("create nonce: %v", err)
				}
				return s.newGoogleToken(googleauth.Identity{Subject: "sub-1", Name: "Alice", Nonce: nonce})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer(t)

			rec := s.login(t, tt.idToken(t, s))

			checkStatus(t, rec, http.StatusUnauthorized)
		})
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)
	user, _, token := s.seedUser(t, "Alice")
	_, otherDeviceToken := s.seedSession(t, user, time.Now().Add(sessionTTL))

	checkStatus(t, s.do(t, http.MethodPost, "/auth/logout", token, nil), http.StatusNoContent)

	checkStatus(t, s.do(t, http.MethodGet, "/me", token, nil), http.StatusUnauthorized)
	checkStatus(t, s.do(t, http.MethodGet, "/me", otherDeviceToken, nil), http.StatusOK)
	checkStatus(t, s.do(t, http.MethodPost, "/auth/logout", "", nil), http.StatusUnauthorized)
}
