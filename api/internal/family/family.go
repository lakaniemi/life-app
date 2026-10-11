// Package family implements families: who belongs to them, who may change
// them, and the rules that keep every family valid. It knows nothing about
// HTTP, and reports rule violations as sentinel errors.
//
// # Authorization
//
// Methods that act on an existing family take the caller's membership (the
// actor), as returned by Membership, and act on the actor's family only.
// Being a member is the precondition for everything; outsiders get
// ErrFamilyNotFound, so a family's existence isn't revealed to them. Admin-only
// actions return ErrForbidden for other members, checked against the role
// read inside the transaction, after the lock (see Concurrency), not the
// actor's Role, which was read earlier and may be stale.
//
// # Membership rules
//
//  1. Every family has at least one admin, so the only admin can't leave or be
//     demoted. They promote someone else first.
//  2. The only member can't leave. They delete the family instead.
//
// # Invites
//
// Users join through an invite an admin creates: a random token, shown in the
// app as a QR code or shared as a link. It's multi-use until it expires, and
// there's no revoking or listing invites, so only its SHA-256 is stored.
// Accepting takes the user's ID instead of an actor, since they aren't a
// member yet. An unknown and an expired token both give ErrInviteNotFound, so
// callers can't tell which tokens ever existed.
//
// # Concurrency
//
// The rules are check-then-write: count the admins, then delete a row. Postgres
// locks only the rows a statement writes, so concurrent requests that write
// different rows don't wait for each other. Without more locking, two admins
// leaving at once would both count two admins, both delete themselves, and
// leave the family with none. Likewise, an admin demoted a moment ago could
// still finish an admin action.
//
// So every change that checks a family-wide rule or the actor's role first
// locks the family's row in families (SELECT … FOR UPDATE, see
// Queries.LockFamily) and holds the lock until its transaction ends. That row
// works as a per-family mutex:
//
//   - It's a single row, so every transaction takes the same lock and none can
//     deadlock. Locking each member row instead could, if two transactions
//     locked them in a different order.
//   - It also covers rows that don't exist yet. Inserting into family_members
//     makes Postgres check the foreign key, which takes a FOR KEY SHARE lock on
//     the referenced families row. That conflicts with FOR UPDATE, so a member
//     joining meanwhile waits too. Locks on the existing member rows wouldn't
//     stop that insert.
//   - The side effect: while the lock is held, inserts into any table that
//     references the family, and updates to the family row, wait for it. The
//     transactions here take milliseconds, so that's acceptable.
//
// What a transaction checks must be read after taking the lock, in a separate
// statement. Under Postgres's default Read Committed isolation, a statement sees
// the data as committed when that statement started. A single statement that
// waited for the lock and also read the members (say, in a CTE) would see them
// as they were before the previous lock holder committed, and the race would be
// back.
//
// Rules for code here:
//   - Keep transactions short, with no calls to other services inside them.
//   - Lock the family before reading anything a family-wide rule or a role
//     check depends on, and read it in a separate statement after the lock.
//     lockMembers does both.
//   - Check the actor's role with checkAdmin on the members read after the
//     lock, never with actor.Role.
package family

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// The values of family_members.role.
const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Errors that callers distinguish with errors.Is. Anything else is an
// internal failure.
var (
	ErrFamilyNotFound       = errors.New("family not found")
	ErrFamilyMemberNotFound = errors.New("family member not found")
	ErrForbidden            = errors.New("only a family admin can do this")
	ErrSoleAdmin            = errors.New("a family must keep at least one admin")
	ErrSoleMember           = errors.New("the family's only member can't leave; delete the family instead")
	ErrInviteNotFound       = errors.New("invite not found or expired")
)

// Service creates, reads and changes families and their membership.
type Service struct {
	logger  *slog.Logger
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewService returns a Service that stores families in pool's database.
// logger receives failures that don't fail the operation, such as cleanup
// errors.
func NewService(logger *slog.Logger, pool *pgxpool.Pool) *Service {
	return &Service{logger: logger, pool: pool, queries: db.New(pool)}
}

// Create creates a family with the user as its admin.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, name string) (db.Family, error) {
	var created db.Family
	err := s.withTx(ctx, func(txQueries *db.Queries) error {
		var err error
		created, err = txQueries.CreateFamily(ctx, name)
		if err != nil {
			return fmt.Errorf("create family: %w", err)
		}
		err = txQueries.AddFamilyMember(ctx, db.AddFamilyMemberParams{FamilyID: created.ID, UserID: userID, Role: RoleAdmin})
		if err != nil {
			return fmt.Errorf("add admin: %w", err)
		}
		return nil
	})
	if err != nil {
		return db.Family{}, err
	}
	return created, nil
}

// Membership returns the user's membership in the family, which is the actor
// for the methods that act on it. It returns ErrFamilyNotFound if the user
// isn't a member, whether or not the family exists.
func (s *Service) Membership(ctx context.Context, familyID, userID uuid.UUID) (db.FamilyMember, error) {
	membership, err := s.queries.GetFamilyMember(ctx, db.GetFamilyMemberParams{FamilyID: familyID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.FamilyMember{}, ErrFamilyNotFound
	}
	if err != nil {
		return db.FamilyMember{}, fmt.Errorf("get membership: %w", err)
	}
	return membership, nil
}

// ListForUser returns the families the user belongs to, with their role in
// each.
func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]db.ListFamiliesForUserRow, error) {
	families, err := s.queries.ListFamiliesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list families: %w", err)
	}
	return families, nil
}

// Get returns the actor's family and its members.
func (s *Service) Get(ctx context.Context, actor db.FamilyMember) (db.Family, []db.ListFamilyMembersRow, error) {
	familyRow, err := s.queries.GetFamily(ctx, actor.FamilyID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Deleted since the actor's membership was read.
		return db.Family{}, nil, ErrFamilyNotFound
	}
	if err != nil {
		return db.Family{}, nil, fmt.Errorf("get family: %w", err)
	}
	members, err := s.queries.ListFamilyMembers(ctx, actor.FamilyID)
	if err != nil {
		return db.Family{}, nil, fmt.Errorf("list members: %w", err)
	}
	return familyRow, members, nil
}

// Rename renames the actor's family. Only an admin may.
func (s *Service) Rename(ctx context.Context, actor db.FamilyMember, name string) (db.Family, error) {
	var renamed db.Family
	err := s.withTx(ctx, func(txQueries *db.Queries) error {
		members, err := lockMembers(ctx, txQueries, actor.FamilyID)
		if err != nil {
			return err
		}
		if err := checkAdmin(members, actor); err != nil {
			return err
		}

		renamed, err = txQueries.UpdateFamilyName(ctx, db.UpdateFamilyNameParams{ID: actor.FamilyID, Name: name})
		if err != nil {
			return fmt.Errorf("rename family: %w", err)
		}
		return nil
	})
	if err != nil {
		return db.Family{}, err
	}
	return renamed, nil
}

// Delete deletes the actor's family. Only an admin may. Its memberships and
// all family-owned data go with it through ON DELETE CASCADE.
func (s *Service) Delete(ctx context.Context, actor db.FamilyMember) error {
	return s.withTx(ctx, func(txQueries *db.Queries) error {
		members, err := lockMembers(ctx, txQueries, actor.FamilyID)
		if err != nil {
			return err
		}
		if err := checkAdmin(members, actor); err != nil {
			return err
		}

		if err := txQueries.DeleteFamily(ctx, actor.FamilyID); err != nil {
			return fmt.Errorf("delete family: %w", err)
		}
		return nil
	})
}

// withTx runs fn in a transaction, with txQueries bound to it. Inside fn, use
// txQueries, not s.queries, which would run outside the transaction. The
// transaction commits if fn returns nil and rolls back otherwise.
func (s *Service) withTx(ctx context.Context, fn func(txQueries *db.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rolls back on every path that doesn't reach Commit, including a panic in
	// fn. After a successful Commit, Rollback does nothing. Its error is
	// ignored: by then fn's or Commit's error is the one worth returning.
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
