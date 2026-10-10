package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/auth"
	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, pool *pgxpool.Pool, verifier auth.TokenVerifier) {
	queries := db.New(pool)
	authService := auth.NewService(logger, queries, verifier)
	familyService := family.NewService(pool)

	mux.Handle("GET /health", handleHealth(logger, pool))

	mux.Handle("POST /auth/nonce", handleCreateNonce(logger, authService))
	mux.Handle("POST /auth/google", handleGoogleLogin(logger, authService))
	mux.Handle("POST /auth/logout", requireAuth(logger, authService, handleLogout(logger, authService)))

	mux.Handle("GET /me", requireAuth(logger, authService, handleGetMe(logger, queries)))
	mux.Handle("PATCH /me", requireAuth(logger, authService, handlePatchMe(logger, queries)))

	mux.Handle("POST /families", requireAuth(logger, authService, handleCreateFamily(logger, familyService)))
	mux.Handle("GET /families", requireAuth(logger, authService, handleListFamilies(logger, queries)))
	mux.Handle("GET /families/{familyID}",
		requireAuth(logger, authService, requireFamilyMembership(logger, queries, handleGetFamily(logger, queries))))
	mux.Handle("PATCH /families/{familyID}",
		requireAuth(logger, authService, requireFamilyMembership(logger, queries, handlePatchFamily(logger, queries))))
	mux.Handle("DELETE /families/{familyID}",
		requireAuth(logger, authService, requireFamilyMembership(logger, queries, handleDeleteFamily(logger, queries))))
	mux.Handle("PATCH /families/{familyID}/members/{userID}",
		requireAuth(logger, authService, requireFamilyMembership(logger, queries, handlePatchFamilyMember(logger, familyService))))
	mux.Handle("DELETE /families/{familyID}/members/{userID}",
		requireAuth(logger, authService, requireFamilyMembership(logger, queries, handleDeleteFamilyMember(logger, familyService))))
}
