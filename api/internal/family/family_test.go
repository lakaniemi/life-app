package family_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/dbtest"
	"github.com/lakaniemi/life-app/api/internal/family"
)

// TestAdminActionsCheckTheCurrentRole covers an admin demoted after their
// membership was read, which happens when the demotion lands between the
// middleware's read and the action's transaction. Endpoint tests can't time
// that, so this passes the stale membership to the service directly.
func TestAdminActionsCheckTheCurrentRole(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	pool := dbtest.New(t)
	queries := db.New(pool)
	familyService := family.NewService(slog.New(slog.DiscardHandler), pool)

	seedUser := func(name string) db.User {
		created, err := queries.CreateUser(ctx, db.CreateUserParams{GoogleSub: "sub-" + name, Name: name})
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		return created
	}
	alice, bob, carol := seedUser("Alice"), seedUser("Bob"), seedUser("Carol")
	smiths, err := familyService.Create(ctx, alice.ID, "Smiths")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	for _, member := range []struct {
		user db.User
		role string
	}{{bob, family.RoleAdmin}, {carol, family.RoleMember}} {
		err := queries.AddFamilyMember(ctx, db.AddFamilyMemberParams{FamilyID: smiths.ID, UserID: member.user.ID, Role: member.role})
		if err != nil {
			t.Fatalf("add member: %v", err)
		}
	}

	membership := func(userID uuid.UUID) db.FamilyMember {
		found, err := familyService.Membership(ctx, smiths.ID, userID)
		if err != nil {
			t.Fatalf("get membership: %v", err)
		}
		return found
	}
	staleBob := membership(bob.ID)
	if _, err := familyService.ChangeRole(ctx, membership(alice.ID), bob.ID, family.RoleMember); err != nil {
		t.Fatalf("demote Bob: %v", err)
	}

	actions := []struct {
		name string
		run  func() error
	}{
		{"rename", func() error {
			_, err := familyService.Rename(ctx, staleBob, "Bobs")
			return err
		}},
		{"change role", func() error {
			_, err := familyService.ChangeRole(ctx, staleBob, carol.ID, family.RoleAdmin)
			return err
		}},
		{"remove member", func() error { return familyService.RemoveMember(ctx, staleBob, carol.ID) }},
		{"delete", func() error { return familyService.Delete(ctx, staleBob) }},
		{"create invite", func() error {
			_, _, err := familyService.CreateInvite(ctx, staleBob)
			return err
		}},
	}
	for _, action := range actions {
		if err := action.run(); !errors.Is(err, family.ErrForbidden) {
			t.Errorf("%s as demoted admin: err = %v, want ErrForbidden", action.name, err)
		}
	}
}
