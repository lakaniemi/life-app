package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
)

type seededUser struct {
	db.User
	token string
}

// familyFixture is a family with Alice as admin and Bob in bobRole, plus
// Mallory, who belongs to no family.
type familyFixture struct {
	server testServer
	family db.Family
	users  map[string]seededUser
}

// newFamilyFixture seeds a familyFixture on a fresh database. An empty bobRole
// leaves Bob out of the family, so Alice is its only member.
func newFamilyFixture(t *testing.T, bobRole string) familyFixture {
	t.Helper()
	server := newTestServer(t)
	users := map[string]seededUser{}
	for _, name := range []string{"Alice", "Bob", "Mallory"} {
		user, _, token := server.seedUser(t, name)
		users[name] = seededUser{User: user, token: token}
	}
	smiths := server.seedFamily(t, "Smiths")
	server.seedMember(t, smiths.ID, users["Alice"].User, family.RoleAdmin)
	if bobRole != "" {
		server.seedMember(t, smiths.ID, users["Bob"].User, bobRole)
	}
	return familyFixture{server: server, family: smiths, users: users}
}

// do sends a request as the named user.
func (fixture familyFixture) do(t *testing.T, method, path, as string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return fixture.server.do(t, method, path, fixture.users[as].token, body)
}

func (fixture familyFixture) familyPath() string {
	return "/families/" + fixture.family.ID.String()
}

func (fixture familyFixture) memberPath(name string) string {
	return fixture.familyPath() + "/members/" + fixture.users[name].ID.String()
}

// memberRoles returns the family's members as name → role, read through
// GET /families/{familyID} as the named member.
func (fixture familyFixture) memberRoles(t *testing.T, as string) map[string]string {
	t.Helper()
	rec := fixture.do(t, http.MethodGet, fixture.familyPath(), as, nil)
	checkStatus(t, rec, http.StatusOK)
	return rolesByName(decode[getFamilyResponse](t, rec).Members)
}

func rolesByName(members []memberResponse) map[string]string {
	roles := map[string]string{}
	for _, member := range members {
		roles[member.Name] = member.Role
	}
	return roles
}

func checkErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := decode[errorResponse](t, rec).Error.Code; got != want {
		t.Errorf("error code = %q, want %q", got, want)
	}
}

// TestFamilyRoutesAuthorization checks every family route's access control in
// one place. Add new family routes here.
func TestFamilyRoutesAuthorization(t *testing.T) {
	t.Parallel()
	fixture := newFamilyFixture(t, family.RoleMember)
	familyPath := fixture.familyPath()
	memberPath := fixture.memberPath("Bob")

	tests := []struct {
		method string
		path   string
		body   any
		// familyScoped routes are under /families/{familyID} and must hide the
		// family from non-members.
		familyScoped bool
	}{
		{http.MethodPost, "/families", map[string]any{"name": "X"}, false},
		{http.MethodGet, "/families", nil, false},
		{http.MethodGet, familyPath, nil, true},
		{http.MethodPatch, familyPath, map[string]any{"name": "X"}, true},
		{http.MethodDelete, familyPath, nil, true},
		{http.MethodPatch, memberPath, map[string]any{"role": "admin"}, true},
		{http.MethodDelete, memberPath, nil, true},
	}
	for _, tt := range tests {
		route := tt.method + " " + tt.path

		rec := fixture.server.do(t, tt.method, tt.path, "", tt.body)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without a token: status = %d, want 401", route, rec.Code)
		}

		if !tt.familyScoped {
			continue
		}
		rec = fixture.do(t, tt.method, tt.path, "Mallory", tt.body)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s as non-member: status = %d, want 404", route, rec.Code)
		}
		malformedPath := strings.Replace(tt.path, familyPath, "/families/not-a-uuid", 1)
		rec = fixture.do(t, tt.method, malformedPath, "Alice", tt.body)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s with malformed family ID: status = %d, want 404", route, rec.Code)
		}
	}

	// The requests above must not have changed anything.
	want := map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}
	if diff := cmp.Diff(want, fixture.memberRoles(t, "Alice")); diff != "" {
		t.Errorf("members mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateFamily(t *testing.T) {
	t.Parallel()

	t.Run("creates the family with the caller as admin", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		user, _, token := server.seedUser(t, "Alice")

		rec := server.do(t, http.MethodPost, "/families", token, map[string]any{"name": "  Smiths  "})

		checkStatus(t, rec, http.StatusCreated)
		created := decode[familyEnvelope](t, rec).Family
		if created.Name != "Smiths" || created.Role != family.RoleAdmin {
			t.Errorf("family = %+v, want name %q and role %q", created, "Smiths", family.RoleAdmin)
		}
		rec = server.do(t, http.MethodGet, "/families/"+created.ID.String(), token, nil)
		checkStatus(t, rec, http.StatusOK)
		members := decode[getFamilyResponse](t, rec).Members
		if len(members) != 1 || members[0].ID != user.ID || members[0].Role != family.RoleAdmin {
			t.Errorf("members = %+v, want only the creator as admin", members)
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
			{name: strings.Repeat("a", 101), wantRule: "max"},
		}
		for _, tt := range tests {
			rec := server.do(t, http.MethodPost, "/families", token, map[string]any{"name": tt.name})

			checkStatus(t, rec, http.StatusBadRequest)
			want := []fieldError{{Field: "name", Rule: tt.wantRule}}
			if diff := cmp.Diff(want, decode[errorResponse](t, rec).Error.Fields); diff != "" {
				t.Errorf("name %q: fields mismatch (-want +got):\n%s", tt.name, diff)
			}
		}
	})
}

func TestListFamilies(t *testing.T) {
	t.Parallel()

	t.Run("lists the caller's families with their role", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		alice, _, token := server.seedUser(t, "Alice")
		bob, _, _ := server.seedUser(t, "Bob")
		smiths := server.seedFamily(t, "Smiths")
		server.seedMember(t, smiths.ID, alice, family.RoleMember)
		joneses := server.seedFamily(t, "Joneses")
		server.seedMember(t, joneses.ID, alice, family.RoleAdmin)
		// Someone else's family, which must not be listed.
		browns := server.seedFamily(t, "Browns")
		server.seedMember(t, browns.ID, bob, family.RoleAdmin)

		rec := server.do(t, http.MethodGet, "/families", token, nil)

		checkStatus(t, rec, http.StatusOK)
		want := listFamiliesResponse{Families: []familyResponse{
			{ID: joneses.ID, Name: "Joneses", Role: family.RoleAdmin},
			{ID: smiths.ID, Name: "Smiths", Role: family.RoleMember},
		}}
		if diff := cmp.Diff(want, decode[listFamiliesResponse](t, rec)); diff != "" {
			t.Errorf("body mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("is an empty list without families", func(t *testing.T) {
		t.Parallel()
		server := newTestServer(t)
		_, _, token := server.seedUser(t, "Alice")

		rec := server.do(t, http.MethodGet, "/families", token, nil)

		checkStatus(t, rec, http.StatusOK)
		if diff := cmp.Diff(map[string]any{"families": []any{}}, decode[map[string]any](t, rec)); diff != "" {
			t.Errorf("body mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestGetFamily(t *testing.T) {
	t.Parallel()
	fixture := newFamilyFixture(t, family.RoleMember)

	rec := fixture.do(t, http.MethodGet, fixture.familyPath(), "Bob", nil)

	checkStatus(t, rec, http.StatusOK)
	got := decode[getFamilyResponse](t, rec)
	wantFamily := familyResponse{ID: fixture.family.ID, Name: "Smiths", Role: family.RoleMember}
	if diff := cmp.Diff(wantFamily, got.Family); diff != "" {
		t.Errorf("family mismatch (-want +got):\n%s", diff)
	}
	wantRoles := map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}
	if diff := cmp.Diff(wantRoles, rolesByName(got.Members)); diff != "" {
		t.Errorf("members mismatch (-want +got):\n%s", diff)
	}
}

func TestPatchFamily(t *testing.T) {
	t.Parallel()

	t.Run("admin renames the family", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodPatch, fixture.familyPath(), "Alice", map[string]any{"name": " Smithsons "})

		checkStatus(t, rec, http.StatusOK)
		want := familyEnvelope{Family: familyResponse{ID: fixture.family.ID, Name: "Smithsons", Role: family.RoleAdmin}}
		if diff := cmp.Diff(want, decode[familyEnvelope](t, rec)); diff != "" {
			t.Errorf("body mismatch (-want +got):\n%s", diff)
		}
		rec = fixture.do(t, http.MethodGet, fixture.familyPath(), "Bob", nil)
		if got := decode[getFamilyResponse](t, rec).Family.Name; got != "Smithsons" {
			t.Errorf("GET name = %q, want %q", got, "Smithsons")
		}
	})

	t.Run("member is forbidden", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodPatch, fixture.familyPath(), "Bob", map[string]any{"name": "Bobs"})

		checkStatus(t, rec, http.StatusForbidden)
		checkErrorCode(t, rec, "forbidden")
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodPatch, fixture.familyPath(), "Alice", map[string]any{"name": " "})

		checkStatus(t, rec, http.StatusBadRequest)
	})
}

func TestDeleteFamily(t *testing.T) {
	t.Parallel()

	t.Run("admin deletes the family for everyone", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodDelete, fixture.familyPath(), "Alice", nil)

		checkStatus(t, rec, http.StatusNoContent)
		checkStatus(t, fixture.do(t, http.MethodGet, fixture.familyPath(), "Alice", nil), http.StatusNotFound)
		rec = fixture.do(t, http.MethodGet, "/families", "Bob", nil)
		if families := decode[listFamiliesResponse](t, rec).Families; len(families) != 0 {
			t.Errorf("Bob's families = %+v, want none", families)
		}
	})

	t.Run("member is forbidden", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodDelete, fixture.familyPath(), "Bob", nil)

		checkStatus(t, rec, http.StatusForbidden)
		checkStatus(t, fixture.do(t, http.MethodGet, fixture.familyPath(), "Bob", nil), http.StatusOK)
	})
}

func TestPatchFamilyMember(t *testing.T) {
	t.Parallel()

	t.Run("admin promotes a member", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodPatch, fixture.memberPath("Bob"), "Alice", map[string]any{"role": "admin"})

		checkStatus(t, rec, http.StatusOK)
		member := decode[patchMemberResponse](t, rec).Member
		if member.ID != fixture.users["Bob"].ID || member.Name != "Bob" || member.Role != family.RoleAdmin {
			t.Errorf("member = %+v, want Bob as admin", member)
		}
		want := map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleAdmin}
		if diff := cmp.Diff(want, fixture.memberRoles(t, "Alice")); diff != "" {
			t.Errorf("members mismatch (-want +got):\n%s", diff)
		}
	})

	tests := []struct {
		name       string
		bobRole    string
		actor      string
		target     string
		role       string
		wantStatus int
		wantCode   string
		wantRoles  map[string]string
	}{
		{
			name:    "admin demotes themselves while another admin remains",
			bobRole: family.RoleAdmin, actor: "Alice", target: "Alice", role: "member",
			wantStatus: http.StatusOK,
			wantRoles:  map[string]string{"Alice": family.RoleMember, "Bob": family.RoleAdmin},
		},
		{
			name:    "sole admin can't demote themselves",
			bobRole: family.RoleMember, actor: "Alice", target: "Alice", role: "member",
			wantStatus: http.StatusConflict, wantCode: "sole_admin",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember},
		},
		{
			name:    "member can't change roles",
			bobRole: family.RoleMember, actor: "Bob", target: "Bob", role: "admin",
			wantStatus: http.StatusForbidden, wantCode: "forbidden",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember},
		},
		{
			name:    "non-member target is not found",
			bobRole: family.RoleMember, actor: "Alice", target: "Mallory", role: "admin",
			wantStatus: http.StatusNotFound, wantCode: "not_found",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember},
		},
		{
			name:    "unknown role is invalid",
			bobRole: family.RoleMember, actor: "Alice", target: "Bob", role: "owner",
			wantStatus: http.StatusBadRequest, wantCode: "invalid_body",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixture := newFamilyFixture(t, tt.bobRole)

			rec := fixture.do(t, http.MethodPatch, fixture.memberPath(tt.target), tt.actor, map[string]any{"role": tt.role})

			checkStatus(t, rec, tt.wantStatus)
			if tt.wantCode != "" {
				checkErrorCode(t, rec, tt.wantCode)
			}
			if diff := cmp.Diff(tt.wantRoles, fixture.memberRoles(t, "Bob")); diff != "" {
				t.Errorf("members mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDeleteFamilyMember(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		bobRole    string // "" leaves Bob out, so Alice is the only member
		actor      string
		target     string
		wantStatus int
		wantCode   string
		// wantRoles is the membership afterwards, read as wantRolesAs.
		wantRoles   map[string]string
		wantRolesAs string
	}{
		{
			name:    "admin removes a member",
			bobRole: family.RoleMember, actor: "Alice", target: "Bob",
			wantStatus: http.StatusNoContent,
			wantRoles:  map[string]string{"Alice": family.RoleAdmin}, wantRolesAs: "Alice",
		},
		{
			name:    "member leaves",
			bobRole: family.RoleMember, actor: "Bob", target: "Bob",
			wantStatus: http.StatusNoContent,
			wantRoles:  map[string]string{"Alice": family.RoleAdmin}, wantRolesAs: "Alice",
		},
		{
			name:    "admin leaves while another admin remains",
			bobRole: family.RoleAdmin, actor: "Alice", target: "Alice",
			wantStatus: http.StatusNoContent,
			wantRoles:  map[string]string{"Bob": family.RoleAdmin}, wantRolesAs: "Bob",
		},
		{
			name:    "member can't remove someone else",
			bobRole: family.RoleMember, actor: "Bob", target: "Alice",
			wantStatus: http.StatusForbidden, wantCode: "forbidden",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}, wantRolesAs: "Alice",
		},
		{
			name:    "sole admin can't leave while others remain",
			bobRole: family.RoleMember, actor: "Alice", target: "Alice",
			wantStatus: http.StatusConflict, wantCode: "sole_admin",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}, wantRolesAs: "Alice",
		},
		{
			name:    "sole member can't leave",
			bobRole: "", actor: "Alice", target: "Alice",
			wantStatus: http.StatusConflict, wantCode: "sole_member",
			wantRoles: map[string]string{"Alice": family.RoleAdmin}, wantRolesAs: "Alice",
		},
		{
			name:    "non-member target is not found",
			bobRole: family.RoleMember, actor: "Alice", target: "Mallory",
			wantStatus: http.StatusNotFound, wantCode: "not_found",
			wantRoles: map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}, wantRolesAs: "Alice",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fixture := newFamilyFixture(t, tt.bobRole)

			rec := fixture.do(t, http.MethodDelete, fixture.memberPath(tt.target), tt.actor, nil)

			checkStatus(t, rec, tt.wantStatus)
			if tt.wantCode != "" {
				checkErrorCode(t, rec, tt.wantCode)
			}
			if diff := cmp.Diff(tt.wantRoles, fixture.memberRoles(t, tt.wantRolesAs)); diff != "" {
				t.Errorf("members mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
