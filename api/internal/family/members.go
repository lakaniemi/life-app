package family

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// ChangeRole sets the role of the member targetID in the actor's family and
// returns the updated member. Only an admin may. It returns
// ErrFamilyMemberNotFound if targetID isn't a member, and ErrSoleAdmin if the
// family would be left without an admin.
func (s *Service) ChangeRole(ctx context.Context, actor db.FamilyMember, targetID uuid.UUID, role string) (db.ListFamilyMembersRow, error) {
	var updated db.ListFamilyMembersRow
	err := s.withTx(ctx, func(txQueries *db.Queries) error {
		members, err := lockMembers(ctx, txQueries, actor.FamilyID)
		if err != nil {
			return err
		}
		if err := checkAdmin(members, actor); err != nil {
			return err
		}
		target, ok := findMember(members, targetID)
		if !ok {
			return ErrFamilyMemberNotFound
		}
		if target.Role == RoleAdmin && role != RoleAdmin && countAdmins(members) == 1 {
			return ErrSoleAdmin
		}

		err = txQueries.UpdateFamilyMemberRole(ctx, db.UpdateFamilyMemberRoleParams{FamilyID: actor.FamilyID, UserID: targetID, Role: role})
		if err != nil {
			return fmt.Errorf("update role: %w", err)
		}
		target.Role = role
		updated = target
		return nil
	})
	if err != nil {
		return db.ListFamilyMembersRow{}, err
	}
	return updated, nil
}

// RemoveMember removes the member targetID from the actor's family. Admins can
// remove anyone; any member can remove themselves, which is leaving the
// family. It returns ErrFamilyMemberNotFound if targetID isn't a member,
// ErrSoleMember if they're the only member, and ErrSoleAdmin if the family
// would be left without an admin.
func (s *Service) RemoveMember(ctx context.Context, actor db.FamilyMember, targetID uuid.UUID) error {
	return s.withTx(ctx, func(txQueries *db.Queries) error {
		members, err := lockMembers(ctx, txQueries, actor.FamilyID)
		if err != nil {
			return err
		}
		if targetID != actor.UserID {
			if err := checkAdmin(members, actor); err != nil {
				return err
			}
		}
		target, ok := findMember(members, targetID)
		if !ok {
			return ErrFamilyMemberNotFound
		}
		// Checked first: the only member is also the only admin, and deleting
		// the family is the way out.
		if len(members) == 1 {
			return ErrSoleMember
		}
		if target.Role == RoleAdmin && countAdmins(members) == 1 {
			return ErrSoleAdmin
		}

		err = txQueries.DeleteFamilyMember(ctx, db.DeleteFamilyMemberParams{FamilyID: actor.FamilyID, UserID: targetID})
		if err != nil {
			return fmt.Errorf("delete member: %w", err)
		}
		return nil
	})
}

// lockMembers locks the family, then reads its members. Call it inside a
// transaction; see the package doc for why the order matters.
func lockMembers(ctx context.Context, txQueries *db.Queries, familyID uuid.UUID) ([]db.ListFamilyMembersRow, error) {
	_, err := txQueries.LockFamily(ctx, familyID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Deleted since the actor's membership was read.
		return nil, ErrFamilyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock family: %w", err)
	}

	members, err := txQueries.ListFamilyMembers(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

// checkAdmin returns ErrForbidden unless the actor is an admin in members,
// which lockMembers read. An actor who is no longer a member gets
// ErrFamilyNotFound, like any outsider.
func checkAdmin(members []db.ListFamilyMembersRow, actor db.FamilyMember) error {
	current, ok := findMember(members, actor.UserID)
	if !ok {
		return ErrFamilyNotFound
	}
	if current.Role != RoleAdmin {
		return ErrForbidden
	}
	return nil
}

func findMember(members []db.ListFamilyMembersRow, userID uuid.UUID) (db.ListFamilyMembersRow, bool) {
	i := slices.IndexFunc(members, func(member db.ListFamilyMembersRow) bool { return member.UserID == userID })
	if i < 0 {
		return db.ListFamilyMembersRow{}, false
	}
	return members[i], true
}

func countAdmins(members []db.ListFamilyMembersRow) int {
	n := 0
	for _, member := range members {
		if member.Role == RoleAdmin {
			n++
		}
	}
	return n
}
