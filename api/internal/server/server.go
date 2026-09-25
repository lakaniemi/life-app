// Package server wires up the HTTP handlers and middleware for the API.
package server

import (
	"log/slog"
	"net/http"
)

// New returns the API's root handler with all routes and middleware applied.
// Dependencies are passed in explicitly; as the API grows (database, config)
// they get added as parameters here.
func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger)

	var handler http.Handler = mux
	handler = logRequests(logger, handler)
	return handler
}
