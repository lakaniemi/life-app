package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/db"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, pool *pgxpool.Pool) {
	queries := db.New(pool)

	mux.Handle("GET /health", handleHealth(logger, pool))

	mux.Handle("GET /me", requireAuth(logger, queries, handleGetMe(logger, queries)))
	mux.Handle("PATCH /me", requireAuth(logger, queries, handlePatchMe(logger, queries)))
}
