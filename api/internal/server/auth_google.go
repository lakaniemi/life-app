package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/googleauth"
)

// Postgres error code for unique_violation.
const pgUniqueViolation = "23505"

// tokenVerifier verifies Google ID tokens. *googleauth.Verifier implements it;
// tests use a fake.
type tokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (googleauth.Identity, error)
}

type googleLoginRequest struct {
	IDToken string `json:"idToken"`
}

type googleLoginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

// handleGoogleLogin exchanges a Google ID token for a session. See
// docs/AUTH.md for the flow.
func handleGoogleLogin(logger *slog.Logger, queries *db.Queries, verifier tokenVerifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req googleLoginRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeError(w, r, logger, http.StatusBadRequest, "invalid_body", err.Error())
			return
		}
		if req.IDToken == "" {
			writeError(w, r, logger, http.StatusBadRequest, "invalid_body", "idToken is required")
			return
		}

		identity, err := verifier.Verify(r.Context(), req.IDToken)
		if err != nil {
			// The reason stays in the logs. This also covers failing to fetch
			// Google's keys, which isn't the client's fault, but can't be told
			// apart from a bad token reliably.
			logger.WarnContext(r.Context(), "google login: invalid id token", "err", err)
			writeError(w, r, logger, http.StatusUnauthorized, "invalid_id_token", "invalid Google ID token")
			return
		}

		// Checked after the token, so a forged token can't burn a valid nonce.
		if identity.Nonce == "" {
			writeError(w, r, logger, http.StatusUnauthorized, "invalid_nonce", "ID token has no nonce")
			return
		}
		_, err = queries.ConsumeNonce(r.Context(), identity.Nonce)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, r, logger, http.StatusUnauthorized, "invalid_nonce", "nonce is unknown, used or expired")
			return
		}
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}

		user, err := findOrCreateUser(r.Context(), queries, identity)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}

		if err := queries.DeleteExpiredSessionsForUser(r.Context(), user.ID); err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		token := newRandomToken()
		_, err = queries.CreateSession(r.Context(), db.CreateSessionParams{
			UserID:    user.ID,
			TokenHash: hashToken(token),
			ExpiresAt: time.Now().Add(sessionTTL),
		})
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}

		writeJSON(w, r, logger, http.StatusOK, googleLoginResponse{Token: token, User: newUserResponse(user)})
	})
}

// findOrCreateUser returns the user with identity's Google sub, creating them
// on first login. An existing user's name is never overwritten: after the
// first login it's theirs to edit.
func findOrCreateUser(ctx context.Context, queries *db.Queries, identity googleauth.Identity) (db.User, error) {
	user, err := queries.GetUserByGoogleSub(ctx, identity.Subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, fmt.Errorf("get user: %w", err)
	}

	user, err = queries.CreateUser(ctx, db.CreateUserParams{GoogleSub: identity.Subject, Name: identity.Name})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		// A concurrent first login created the user between our get and
		// create; theirs won, so use it.
		user, err = queries.GetUserByGoogleSub(ctx, identity.Subject)
	}
	if err != nil {
		return db.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}
