package family

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// inviteTTL is long enough for a link shared over chat to be opened later
// the same day.
const inviteTTL = 24 * time.Hour

// CreateInvite creates an invite to the actor's family and returns its token
// and expiry. Only an admin may. The token isn't stored, so this is the only
// time it's available.
func (s *Service) CreateInvite(ctx context.Context, actor db.FamilyMember) (string, time.Time, error) {
	token := newInviteToken()
	var created db.FamilyInvite
	err := s.withTx(ctx, func(txQueries *db.Queries) error {
		members, err := lockMembers(ctx, txQueries, actor.FamilyID)
		if err != nil {
			return err
		}
		if err := checkAdmin(members, actor); err != nil {
			return err
		}

		created, err = txQueries.CreateInvite(ctx, db.CreateInviteParams{
			TokenHash: hashInviteToken(token),
			FamilyID:  actor.FamilyID,
			ExpiresAt: time.Now().Add(inviteTTL),
		})
		if err != nil {
			return fmt.Errorf("create invite: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", time.Time{}, err
	}

	// Cleaning up here keeps the table bounded without a background job. It
	// runs after the commit because a failed statement aborts a Postgres
	// transaction, and a failure isn't the caller's problem: the next invite
	// tries again, and expired invites are rejected anyway.
	if err := s.queries.DeleteExpiredFamilyInvites(ctx, actor.FamilyID); err != nil {
		s.logger.WarnContext(ctx, "delete expired invites", "err", err)
	}
	return token, created.ExpiresAt, nil
}

// PreviewInvite returns the family the invite is for, and whether the user is
// already a member, so they can confirm before accepting.
func (s *Service) PreviewInvite(ctx context.Context, userID uuid.UUID, token string) (db.PreviewInviteRow, error) {
	preview, err := s.queries.PreviewInvite(ctx, db.PreviewInviteParams{TokenHash: hashInviteToken(token), UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.PreviewInviteRow{}, ErrInviteNotFound
	}
	if err != nil {
		return db.PreviewInviteRow{}, fmt.Errorf("preview invite: %w", err)
	}
	return preview, nil
}

// AcceptInvite adds the user to the invite's family as a member. A user who
// is already a member keeps their role, so accepting twice is harmless.
func (s *Service) AcceptInvite(ctx context.Context, userID uuid.UUID, token string) error {
	return s.withTx(ctx, func(txQueries *db.Queries) error {
		familyID, err := txQueries.GetInviteFamilyID(ctx, hashInviteToken(token))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}
		if err != nil {
			return fmt.Errorf("get invite: %w", err)
		}
		// Joining checks no rule, but membership changes all happen under the
		// lock. It also catches the family being deleted since the invite was
		// read, which would otherwise fail the insert's foreign key check.
		_, err = txQueries.LockFamily(ctx, familyID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}
		if err != nil {
			return fmt.Errorf("lock family: %w", err)
		}

		err = txQueries.AddFamilyMemberIfAbsent(ctx, db.AddFamilyMemberIfAbsentParams{FamilyID: familyID, UserID: userID, Role: RoleMember})
		if err != nil {
			return fmt.Errorf("add member: %w", err)
		}
		return nil
	})
}

// newInviteToken returns 32 random bytes (256 bits), base64url-encoded, so
// it's safe in a URL.
func newInviteToken() string {
	b := make([]byte, 32)
	// crypto/rand.Read never returns an error; it crashes the program instead
	// if the OS can't provide randomness.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// hashInviteToken returns what family_invites stores instead of the token. A
// fast hash is enough because the token is random, with nothing to guess.
func hashInviteToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
