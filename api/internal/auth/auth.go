// Package auth implements sign-in and sessions: login nonces, exchanging a
// Google ID token for a session, and authenticating session tokens. It knows
// nothing about HTTP; internal/server maps its errors to responses. See
// docs/AUTH.md for the design and its reasons.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/lakaniemi/life-app/api/internal/auth/google"
	"github.com/lakaniemi/life-app/api/internal/db"
)

const (
	// SessionTTL is how long a session lasts after its last use (sliding).
	SessionTTL = 90 * 24 * time.Hour
	// Expiry is extended at most this often, so a busy session costs one
	// extra write per day, not one per request.
	sessionTouchInterval = 24 * time.Hour
	// Long enough for the user to finish Google sign-in.
	nonceTTL = 10 * time.Minute
)

// Errors that callers distinguish with errors.Is. Anything else is an
// internal failure.
var (
	ErrInvalidIDToken = errors.New("invalid Google ID token")
	ErrInvalidNonce   = errors.New("nonce is missing, unknown, used or expired")
	ErrInvalidSession = errors.New("invalid or expired session")
)

// TokenVerifier verifies Google ID tokens. *google.Verifier implements it;
// tests use a fake.
type TokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (google.Identity, error)
}

// Service performs sign-in and session operations against the database.
type Service struct {
	queries  *db.Queries
	verifier TokenVerifier
}

// NewService returns a Service that stores state through queries and
// verifies Google ID tokens with verifier.
func NewService(queries *db.Queries, verifier TokenVerifier) *Service {
	return &Service{queries: queries, verifier: verifier}
}
