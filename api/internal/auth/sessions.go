package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// Authenticate returns the session for token, extending its expiry if it
// hasn't been used for a while. An unknown or expired token returns
// ErrInvalidSession.
func (s *Service) Authenticate(ctx context.Context, token string) (db.Session, error) {
	session, err := s.queries.GetSessionByTokenHash(ctx, HashToken(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Session{}, ErrInvalidSession
	}
	if err != nil {
		return db.Session{}, fmt.Errorf("get session: %w", err)
	}

	if time.Since(session.LastUsedAt) > sessionTouchInterval {
		err := s.queries.TouchSession(ctx, db.TouchSessionParams{
			ID:        session.ID,
			ExpiresAt: time.Now().Add(SessionTTL),
		})
		// The session is valid either way. last_used_at stays old on failure,
		// so the next request retries the extension.
		if err != nil {
			s.logger.WarnContext(ctx, "extend session expiry", "err", err, "session_id", session.ID)
		}
	}
	return session, nil
}

// Logout deletes one session. The user's other sessions (e.g. on other
// devices) stay valid.
func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.queries.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// createSession stores a new session for user and returns its token. Only the
// token's hash is stored.
func (s *Service) createSession(ctx context.Context, user db.User) (string, error) {
	token := newToken()
	_, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: HashToken(token),
		ExpiresAt: time.Now().Add(SessionTTL),
	})
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}
