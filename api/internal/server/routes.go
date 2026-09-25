package server

import (
	"log/slog"
	"net/http"
)

// addRoutes registers every route in the API. Keeping them in one place
// makes the whole API surface visible at a glance.
func addRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.Handle("GET /health", handleHealth(logger))
}
