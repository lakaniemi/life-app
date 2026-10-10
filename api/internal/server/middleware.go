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
	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
)

// middleware wraps a handler with behavior that runs around it. Middleware
// constructors take their dependencies and return one, so routes.go binds
// the dependencies once and each route only lists which middlewares apply.
type middleware func(http.Handler) http.Handler

// withMiddlewares applies mws to a handler, with the first one listed running
// first (outermost).
func withMiddlewares(mws ...middleware) middleware {
	return func(h http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

// Unwrap lets http.ResponseController reach the underlying writer, so
// features like flushing still work through this wrapper.
func (rec *statusRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}

func newLogRequests(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
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
}

// authSession is what requireAuth puts in the request context.
type authSession struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

// Context keys are unexported types, so no other package can read or
// overwrite the values stored under them.
type (
	loggerKey           struct{}
	sessionKey          struct{}
	familyMembershipKey struct{}
)

// newAttachLogger returns attachLogger, which puts logger in the request
// context for apiHandler.ServeHTTP. Everything else gets the logger as an
// argument.
func newAttachLogger(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loggerKey{}, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// loggerFrom returns the logger attachLogger stored, which server.New
// applies to every request.
func loggerFrom(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok {
		panic("loggerFrom: no logger in context; is the handler served through server.New?")
	}
	return logger
}

// sessionFrom returns the session requireAuth stored. Only call it from
// handlers wrapped in requireAuth.
func sessionFrom(ctx context.Context) authSession {
	s, ok := ctx.Value(sessionKey{}).(authSession)
	if !ok {
		panic("sessionFrom: no session in context; is the handler wrapped in requireAuth?")
	}
	return s
}

// newRequireAuth returns requireAuth, which rejects requests without a valid
// session with 401, and passes the rest on with the session in the context.
func newRequireAuth(logger *slog.Logger, authService *auth.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				w.Header().Set("WWW-Authenticate", "Bearer")
				writeErrorResponse(w, r, logger, newAPIError(http.StatusUnauthorized, "unauthorized", "missing bearer token"))
				return
			}

			session, err := authService.Authenticate(r.Context(), token)
			if errors.Is(err, auth.ErrInvalidSession) {
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				writeErrorResponse(w, r, logger, newAPIError(http.StatusUnauthorized, "unauthorized", "invalid or expired session"))
				return
			}
			if err != nil {
				writeErrorResponse(w, r, logger, err)
				return
			}

			ctx := context.WithValue(r.Context(), sessionKey{}, authSession{UserID: session.UserID, SessionID: session.ID})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// familyMembershipFrom returns the caller's membership that
// requireFamilyMembership stored: the actor to pass to family.Service. Only
// call it from handlers wrapped in requireFamilyMembership.
func familyMembershipFrom(ctx context.Context) db.FamilyMember {
	membership, ok := ctx.Value(familyMembershipKey{}).(db.FamilyMember)
	if !ok {
		panic("familyMembershipFrom: no membership in context; is the handler wrapped in requireFamilyMembership?")
	}
	return membership
}

// newRequireFamilyMembership returns requireFamilyMembership, which passes the
// request on, with the caller's membership in the context, only if the caller
// belongs to the family in the path's {familyID}. Otherwise it responds 404
// rather than 403, so outsiders can't learn whether a family exists. It must
// run after requireAuth.
//
// Every route under /families/{familyID} goes through it. It only checks
// membership; family.Service checks the role where an action needs one.
func newRequireFamilyMembership(logger *slog.Logger, familyService *family.Service) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := sessionFrom(ctx).UserID
			familyID, err := uuid.Parse(r.PathValue("familyID"))
			if err != nil {
				writeErrorResponse(w, r, logger, family.ErrFamilyNotFound)
				return
			}

			membership, err := familyService.Membership(ctx, familyID, userID)
			if err != nil {
				writeErrorResponse(w, r, logger, err)
				return
			}

			ctx = context.WithValue(ctx, familyMembershipKey{}, membership)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
