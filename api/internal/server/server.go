// Package server is the API's HTTP layer: routes, handlers, middleware and
// JSON encoding. Domain logic lives in other packages (e.g. internal/auth);
// handlers here decode requests, call it, and map its results to responses.
package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/auth"
)

// New returns the API's root handler with all routes and middleware applied.
func New(logger *slog.Logger, pool *pgxpool.Pool, verifier auth.TokenVerifier) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, pool, verifier)
	return logRequests(logger, mux)
}
