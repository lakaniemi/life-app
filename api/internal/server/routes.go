package server

import (
	"log/slog"
	"net/http"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger) {
	mux.Handle("GET /health", handleHealth(logger))
}
