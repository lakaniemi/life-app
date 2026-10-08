package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, pool *pgxpool.Pool) {
	mux.Handle("GET /health", handleHealth(logger, pool))
}
