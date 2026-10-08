package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// authSession is what requireAuth puts in the request context.
type authSession struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// ctxKey is unexported, so no other package can read or overwrite the value
// stored under it.
type ctxKey struct{}

// sessionFrom returns the session requireAuth stored. Only call it from
// handlers wrapped in requireAuth.
func sessionFrom(ctx context.Context) authSession {
	s, ok := ctx.Value(ctxKey{}).(authSession)
	if !ok {
		panic("sessionFrom: no session in context; is the handler wrapped in requireAuth?")
	}
	return s
}

// requireAuth rejects requests without a valid session with 401, and passes
// the rest to next with the session in the context.
func requireAuth(logger *slog.Logger, queries *db.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, r, logger, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}

		session, err := queries.GetSessionByTokenHash(r.Context(), hashToken(token))
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, r, logger, http.StatusUnauthorized, "unauthorized", "invalid or expired session")
			return
		}
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}

		if time.Since(session.LastUsedAt) > sessionTouchInterval {
			err := queries.TouchSession(r.Context(), db.TouchSessionParams{
				ID:        session.ID,
				ExpiresAt: time.Now().Add(sessionTTL),
			})
			if err != nil {
				writeInternalError(w, r, logger, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), ctxKey{}, authSession{UserID: session.UserID, SessionID: session.ID})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// bearerToken extracts the token from "Authorization: Bearer <token>". The
// scheme is case-insensitive (RFC 9110 section 11.1).
func bearerToken(r *http.Request) (string, bool) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}
