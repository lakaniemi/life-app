// Package family implements families and the rules for their membership. It
// knows nothing about HTTP, and it doesn't decide who may act: internal/server
// checks the caller's role. This package keeps every family valid, and reports
// rule violations as sentinel errors.
//
// # Membership rules
//
//  1. Every family has at least one admin, so the only admin can't leave or be
//     demoted. They promote someone else first.
//  2. The only member can't leave. They delete the family instead.
//
// # Concurrency
//
// The rules are check-then-write: count the admins, then delete a row. Postgres
// locks only the rows a statement writes, so concurrent requests that write
// different rows don't wait for each other. Without more locking, two admins
// leaving at once would both count two admins, both delete themselves, and
// leave the family with none.
//
// So every change that checks a family-wide rule first locks the family's row
// in families (SELECT … FOR UPDATE, see Queries.LockFamily) and holds the lock
// until its transaction ends. That row works as a per-family mutex:
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
//   - Lock the family before reading anything a family-wide rule depends on.
//   - Read it in a separate statement after the lock.
package family

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	ErrSoleAdmin            = errors.New("a family must keep at least one admin")
	ErrSoleMember           = errors.New("the family's only member can't leave; delete the family instead")
)

// Service creates families and changes their membership.
type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewService returns a Service that stores families in pool's database.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, queries: db.New(pool)}
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
