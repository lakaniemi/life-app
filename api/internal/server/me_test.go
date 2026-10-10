package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/lakaniemi/life-app/api/internal/family"
)

func TestGetMe(t *testing.T) {
	t.Parallel()
	server := newTestServer(t)
	user, _, token := server.seedUser(t, "Alice")
	// Another user, to check /me returns the caller and not just anyone.
	server.seedUser(t, "Bob")

	rec := server.do(t, http.MethodGet, "/me", token, nil)

	checkStatus(t, rec, http.StatusOK)
	got := decode[map[string]any](t, rec)
	want := map[string]any{
		"user":     map[string]any{"id": user.ID.String(), "name": "Alice"},
		"families": []any{},
	}
	checkDiff(t, "body", want, got)
}

func TestGetMeListsFamilies(t *testing.T) {
	t.Parallel()
	server := newTestServer(t)
	user, _, token := server.seedUser(t, "Alice")
	smiths := server.seedFamily(t, "Smiths")
	server.seedMember(t, smiths.ID, user, family.RoleMember)

	rec := server.do(t, http.MethodGet, "/me", token, nil)

	checkStatus(t, rec, http.StatusOK)
	want := []familyResponse{{ID: smiths.ID, Name: "Smiths", Role: family.RoleMember}}
	checkDiff(t, "families", want, decode[meResponse](t, rec).Families)
}

func TestPatchMe(t *testing.T) {
	t.Parallel()

	t.Run("updates the name", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		user, _, token := server.seedUser(t, "Alice")

		rec := server.do(t, http.MethodPatch, "/me", token, map[string]any{"name": "  Alicia  "})

		checkStatus(t, rec, http.StatusOK)
		want := patchMeResponse{User: userResponse{ID: user.ID, Name: "Alicia"}}
		checkDiff(t, "body", want, decode[patchMeResponse](t, rec))
		me := decode[meResponse](t, server.do(t, http.MethodGet, "/me", token, nil))
		if me.User.Name != "Alicia" {
			t.Errorf("GET /me name = %q, want %q", me.User.Name, "Alicia")
		}
	})

	t.Run("rejects invalid names with the failing field", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		_, _, token := server.seedUser(t, "Alice")

		tests := []struct {
			name     string
			wantRule string
		}{
			{name: "   ", wantRule: "required"},
			{name: strings.Repeat("ä", 101), wantRule: "max"},
		}
		for _, tt := range tests {
			rec := server.do(t, http.MethodPatch, "/me", token, map[string]any{"name": tt.name})

			checkStatus(t, rec, http.StatusBadRequest)
			want := []fieldError{{Field: "name", Rule: tt.wantRule}}
			checkDiff(t, fmt.Sprintf("name %q: fields", tt.name), want, decode[errorResponse](t, rec).Error.Fields)
		}
	})

	t.Run("length counts characters, not bytes", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		_, _, token := server.seedUser(t, "Alice")

		// 100 characters, 200 bytes in UTF-8.
		rec := server.do(t, http.MethodPatch, "/me", token, map[string]any{"name": strings.Repeat("ä", 100)})

		checkStatus(t, rec, http.StatusOK)
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		_, _, token := server.seedUser(t, "Alice")

		checkStatus(t, server.do(t, http.MethodPatch, "/me", token, "not an object"), http.StatusBadRequest)
	})

	t.Run("rejects an oversize body", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		_, _, token := server.seedUser(t, "Alice")

		rec := server.do(t, http.MethodPatch, "/me", token, map[string]any{"name": strings.Repeat("a", maxRequestBodyBytes)})

		checkStatus(t, rec, http.StatusRequestEntityTooLarge)
		if got := decode[errorResponse](t, rec).Error.Code; got != "body_too_large" {
			t.Errorf("error code = %q, want %q", got, "body_too_large")
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)

		rec := server.do(t, http.MethodPatch, "/me", "", map[string]any{"name": "Mallory"})

		checkStatus(t, rec, http.StatusUnauthorized)
	})
}
