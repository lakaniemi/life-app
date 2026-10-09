package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/auth"
	"github.com/lakaniemi/life-app/api/internal/db"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, pool *pgxpool.Pool, verifier auth.TokenVerifier) {
	queries := db.New(pool)
	authService := auth.NewService(logger, queries, verifier)

	mux.Handle("GET /health", handleHealth(logger, pool))

	mux.Handle("POST /auth/nonce", handleCreateNonce(logger, authService))
	mux.Handle("POST /auth/google", handleGoogleLogin(logger, authService))
	mux.Handle("POST /auth/logout", requireAuth(logger, authService, handleLogout(logger, authService)))

	mux.Handle("GET /me", requireAuth(logger, authService, handleGetMe(logger, queries)))
	mux.Handle("PATCH /me", requireAuth(logger, authService, handlePatchMe(logger, queries)))
}
