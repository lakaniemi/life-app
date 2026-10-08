// Package server wires up the HTTP handlers and middleware for the API.
package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New returns the API's root handler with all routes and middleware applied.
func New(logger *slog.Logger, pool *pgxpool.Pool, verifier tokenVerifier) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, pool, verifier)

	var handler http.Handler = mux
	handler = logRequests(logger, handler)
	return handler
}
