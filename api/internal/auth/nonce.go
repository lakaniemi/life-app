package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// NewNonce issues a single-use nonce for the app to pass to Google sign-in,
// which embeds it in the ID token.
func (s *Service) NewNonce(ctx context.Context) (string, error) {
	// Cleaning up here keeps the table bounded without a background job.
	if err := s.queries.DeleteExpiredNonces(ctx); err != nil {
		return "", fmt.Errorf("delete expired nonces: %w", err)
	}

	nonce := newToken()
	err := s.queries.CreateNonce(ctx, db.CreateNonceParams{Nonce: nonce, ExpiresAt: time.Now().Add(nonceTTL)})
	if err != nil {
		return "", fmt.Errorf("create nonce: %w", err)
	}
	return nonce, nil
}
