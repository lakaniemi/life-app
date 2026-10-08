package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/db"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, pool *pgxpool.Pool, verifier tokenVerifier) {
	queries := db.New(pool)

	mux.Handle("GET /health", handleHealth(logger, pool))

	mux.Handle("POST /auth/nonce", handleCreateNonce(logger, queries))
	mux.Handle("POST /auth/google", handleGoogleLogin(logger, queries, verifier))
	mux.Handle("POST /auth/logout", requireAuth(logger, queries, handleLogout(logger, queries)))

	mux.Handle("GET /me", requireAuth(logger, queries, handleGetMe(logger, queries)))
	mux.Handle("PATCH /me", requireAuth(logger, queries, handlePatchMe(logger, queries)))
}
