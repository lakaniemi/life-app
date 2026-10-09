package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/auth"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap lets http.ResponseController reach the underlying writer, so
// features like flushing still work through this wrapper.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Handlers that never call WriteHeader implicitly send 200.
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		)
	})
}

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
func requireAuth(logger *slog.Logger, authService *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, r, logger, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}

		session, err := authService.Authenticate(r.Context(), token)
		if errors.Is(err, auth.ErrInvalidSession) {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			writeError(w, r, logger, http.StatusUnauthorized, "unauthorized", "invalid or expired session")
			return
		}
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
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
