package server

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMe(t *testing.T) {
	t.Parallel()
	s := newTestServer(t)
	user, _, token := s.seedUser(t, "Alice")
	// Another user, to check /me returns the caller and not just anyone.
	s.seedUser(t, "Bob")

	rec := s.do(t, http.MethodGet, "/me", token, nil)

	checkStatus(t, rec, http.StatusOK)
	got := decode[map[string]any](t, rec)
	want := map[string]any{
		"user":     map[string]any{"id": user.ID.String(), "name": "Alice"},
		"families": []any{},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("body mismatch (-want +got):\n%s", diff)
	}
}

func TestPatchMe(t *testing.T) {
	t.Parallel()

	t.Run("updates the name", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		user, _, token := s.seedUser(t, "Alice")

		rec := s.do(t, http.MethodPatch, "/me", token, map[string]any{"name": "  Alicia  "})

		checkStatus(t, rec, http.StatusOK)
		want := patchMeResponse{User: userResponse{ID: user.ID, Name: "Alicia"}}
		if diff := cmp.Diff(want, decode[patchMeResponse](t, rec)); diff != "" {
			t.Errorf("body mismatch (-want +got):\n%s", diff)
		}
		me := decode[meResponse](t, s.do(t, http.MethodGet, "/me", token, nil))
		if me.User.Name != "Alicia" {
			t.Errorf("GET /me name = %q, want %q", me.User.Name, "Alicia")
		}
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)
		_, _, token := s.seedUser(t, "Alice")

		for _, body := range []any{
			map[string]any{"name": "   "},
			"not an object",
		} {
			checkStatus(t, s.do(t, http.MethodPatch, "/me", token, body), http.StatusBadRequest)
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		s := newTestServer(t)

		rec := s.do(t, http.MethodPatch, "/me", "", map[string]any{"name": "Mallory"})

		checkStatus(t, rec, http.StatusUnauthorized)
	})
}
