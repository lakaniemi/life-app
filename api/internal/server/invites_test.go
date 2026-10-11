package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/lakaniemi/life-app/api/internal/family"
)

func TestCreateInvite(t *testing.T) {
	t.Parallel()

	t.Run("admin gets a token that previews the family", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodPost, fixture.familyPath()+"/invites", "Alice", nil)

		checkStatus(t, rec, http.StatusCreated)
		invite := decode[createInviteResponse](t, rec).Invite
		if until := time.Until(invite.ExpiresAt); until < 23*time.Hour || until > 25*time.Hour {
			t.Errorf("expiresAt = %v, want about 24h from now", invite.ExpiresAt)
		}
		rec = fixture.do(t, http.MethodGet, "/invites/"+invite.Token, "Mallory", nil)
		checkStatus(t, rec, http.StatusOK)
		if got := decode[previewInviteResponse](t, rec).Family.ID; got != fixture.family.ID {
			t.Errorf("previewed family = %v, want %v", got, fixture.family.ID)
		}
	})

	t.Run("keeps the family's earlier active invites", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)
		earlier := fixture.server.seedInvite(t, fixture.family.ID, time.Now().Add(time.Hour))

		rec := fixture.do(t, http.MethodPost, fixture.familyPath()+"/invites", "Alice", nil)
		checkStatus(t, rec, http.StatusCreated)

		rec = fixture.do(t, http.MethodGet, "/invites/"+earlier, "Mallory", nil)
		checkStatus(t, rec, http.StatusOK)
	})

	t.Run("member is forbidden", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)

		rec := fixture.do(t, http.MethodPost, fixture.familyPath()+"/invites", "Bob", nil)

		checkStatus(t, rec, http.StatusForbidden)
	})
}

func TestPreviewInvite(t *testing.T) {
	t.Parallel()
	fixture := newFamilyFixture(t, family.RoleMember)
	active := fixture.server.seedInvite(t, fixture.family.ID, time.Now().Add(time.Hour))
	expired := fixture.server.seedInvite(t, fixture.family.ID, time.Now().Add(-time.Minute))

	t.Run("shows the family to a non-member", func(t *testing.T) {
		t.Parallel()
		rec := fixture.do(t, http.MethodGet, "/invites/"+active, "Mallory", nil)

		checkStatus(t, rec, http.StatusOK)
		want := previewInviteResponse{Family: invitedFamilyResponse{ID: fixture.family.ID, Name: "Smiths"}, AlreadyMember: false}
		checkDiff(t, "body", want, decode[previewInviteResponse](t, rec))
	})

	t.Run("tells a member they already belong", func(t *testing.T) {
		t.Parallel()
		rec := fixture.do(t, http.MethodGet, "/invites/"+active, "Bob", nil)

		checkStatus(t, rec, http.StatusOK)
		if !decode[previewInviteResponse](t, rec).AlreadyMember {
			t.Error("alreadyMember = false, want true")
		}
	})

	for name, token := range map[string]string{"unknown": "no-such-token", "expired": expired} {
		t.Run(name+" token is not found", func(t *testing.T) {
			t.Parallel()
			rec := fixture.do(t, http.MethodGet, "/invites/"+token, "Mallory", nil)

			checkStatus(t, rec, http.StatusNotFound)
		})
	}
}

func TestAcceptInvite(t *testing.T) {
	t.Parallel()

	t.Run("joins as member, and the token works again", func(t *testing.T) {
		t.Parallel()
		// Bob isn't a member here, so both he and Mallory can join.
		fixture := newFamilyFixture(t, "")
		token := fixture.server.seedInvite(t, fixture.family.ID, time.Now().Add(time.Hour))

		for _, name := range []string{"Mallory", "Bob"} {
			rec := fixture.do(t, http.MethodPost, "/invites/"+token+"/accept", name, nil)

			checkStatus(t, rec, http.StatusNoContent)
		}
		want := map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember, "Mallory": family.RoleMember}
		checkDiff(t, "members", want, fixture.memberRoles(t, "Alice"))
	})

	t.Run("leaves an existing member's role unchanged", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)
		token := fixture.server.seedInvite(t, fixture.family.ID, time.Now().Add(time.Hour))

		rec := fixture.do(t, http.MethodPost, "/invites/"+token+"/accept", "Alice", nil)

		checkStatus(t, rec, http.StatusNoContent)
		want := map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}
		checkDiff(t, "members", want, fixture.memberRoles(t, "Alice"))
	})

	t.Run("unknown or expired token is not found", func(t *testing.T) {
		t.Parallel()
		fixture := newFamilyFixture(t, family.RoleMember)
		expired := fixture.server.seedInvite(t, fixture.family.ID, time.Now().Add(-time.Minute))

		for _, token := range []string{"no-such-token", expired} {
			rec := fixture.do(t, http.MethodPost, "/invites/"+token+"/accept", "Mallory", nil)

			checkStatus(t, rec, http.StatusNotFound)
		}
		want := map[string]string{"Alice": family.RoleAdmin, "Bob": family.RoleMember}
		checkDiff(t, "members", want, fixture.memberRoles(t, "Alice"))
	})
}
