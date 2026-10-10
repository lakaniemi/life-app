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

// ChangeRole sets the role of the family's member userID and returns the
// updated member. It returns ErrFamilyNotFound or ErrFamilyMemberNotFound if
// either doesn't exist, and ErrSoleAdmin if the family would be left without
// an admin.
func (s *Service) ChangeRole(ctx context.Context, familyID, userID uuid.UUID, role string) (db.ListFamilyMembersRow, error) {
	var updated db.ListFamilyMembersRow
	err := s.withTx(ctx, func(q *db.Queries) error {
		members, err := lockMembers(ctx, q, familyID)
		if err != nil {
			return err
		}
		target, ok := findMember(members, userID)
		if !ok {
			return ErrFamilyMemberNotFound
		}
		if target.Role == RoleAdmin && role != RoleAdmin && countAdmins(members) == 1 {
			return ErrSoleAdmin
		}

		err = q.UpdateFamilyMemberRole(ctx, db.UpdateFamilyMemberRoleParams{FamilyID: familyID, UserID: userID, Role: role})
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

// RemoveMember removes the family's member userID. It returns
// ErrFamilyNotFound or ErrFamilyMemberNotFound if either doesn't exist,
// ErrSoleMember if they're the only member, and ErrSoleAdmin if the family
// would be left without an admin.
func (s *Service) RemoveMember(ctx context.Context, familyID, userID uuid.UUID) error {
	return s.withTx(ctx, func(q *db.Queries) error {
		members, err := lockMembers(ctx, q, familyID)
		if err != nil {
			return err
		}
		target, ok := findMember(members, userID)
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

		err = q.DeleteFamilyMember(ctx, db.DeleteFamilyMemberParams{FamilyID: familyID, UserID: userID})
		if err != nil {
			return fmt.Errorf("delete member: %w", err)
		}
		return nil
	})
}

// lockMembers locks the family, then reads its members. Call it inside a
// transaction; see the package doc for why the order matters.
func lockMembers(ctx context.Context, q *db.Queries, familyID uuid.UUID) ([]db.ListFamilyMembersRow, error) {
	_, err := q.LockFamily(ctx, familyID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Deleted since the caller's membership was checked.
		return nil, ErrFamilyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock family: %w", err)
	}

	members, err := q.ListFamilyMembers(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

func findMember(members []db.ListFamilyMembersRow, userID uuid.UUID) (db.ListFamilyMembersRow, bool) {
	i := slices.IndexFunc(members, func(m db.ListFamilyMembersRow) bool { return m.UserID == userID })
	if i < 0 {
		return db.ListFamilyMembersRow{}, false
	}
	return members[i], true
}

func countAdmins(members []db.ListFamilyMembersRow) int {
	n := 0
	for _, m := range members {
		if m.Role == RoleAdmin {
			n++
		}
	}
	return n
}
