package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/lakaniemi/life-app/api/internal/auth/google"
	"github.com/lakaniemi/life-app/api/internal/db"
)

// Postgres error code for unique_violation.
const pgUniqueViolation = "23505"

// LoginWithGoogle exchanges a Google ID token for a new session, creating the
// user on first login. It returns the session token and the user.
//
// Errors wrap ErrInvalidIDToken or ErrInvalidNonce when the caller is at
// fault; for ErrInvalidIDToken the wrapped error says why.
func (s *Service) LoginWithGoogle(ctx context.Context, idToken string) (string, db.User, error) {
	identity, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return "", db.User{}, fmt.Errorf("%w: %w", ErrInvalidIDToken, err)
	}

	// Checked after the token, so a forged token can't burn a valid nonce.
	if identity.Nonce == "" {
		return "", db.User{}, ErrInvalidNonce
	}
	_, err = s.queries.ConsumeNonce(ctx, identity.Nonce)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", db.User{}, ErrInvalidNonce
	}
	if err != nil {
		return "", db.User{}, fmt.Errorf("consume nonce: %w", err)
	}

	user, err := s.findOrCreateUser(ctx, identity)
	if err != nil {
		return "", db.User{}, err
	}

	// Cleanup only: expired sessions are rejected anyway, and the next login
	// tries again.
	if err := s.queries.DeleteExpiredSessionsForUser(ctx, user.ID); err != nil {
		s.logger.WarnContext(ctx, "delete expired sessions", "err", err, "user_id", user.ID)
	}
	token, err := s.createSession(ctx, user)
	if err != nil {
		return "", db.User{}, err
	}
	return token, user, nil
}

// findOrCreateUser returns the user with identity's Google sub, creating them
// on first login. An existing user's name is never overwritten: after the
// first login it's theirs to edit.
func (s *Service) findOrCreateUser(ctx context.Context, identity google.Identity) (db.User, error) {
	user, err := s.queries.GetUserByGoogleSub(ctx, identity.Subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, fmt.Errorf("get user: %w", err)
	}

	user, err = s.queries.CreateUser(ctx, db.CreateUserParams{GoogleSub: identity.Subject, Name: identity.Name})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		// A concurrent first login created the user between our get and
		// create; theirs won, so use it.
		user, err = s.queries.GetUserByGoogleSub(ctx, identity.Subject)
	}
	if err != nil {
		return db.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}
