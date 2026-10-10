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

	requireAuth := newRequireAuth(logger, authService)
	requireFamilyMembership := newRequireFamilyMembership(logger, queries)

	mux.Handle("GET /health", handleHealth(logger, pool))

	mux.Handle("POST /auth/nonce", handleCreateNonce(logger, authService))
	mux.Handle("POST /auth/google", handleGoogleLogin(logger, authService))
	mux.Handle("POST /auth/logout",
		withMiddlewares(requireAuth)(handleLogout(logger, authService)))

	mux.Handle("GET /me",
		withMiddlewares(requireAuth)(handleGetMe(logger, queries)))
	mux.Handle("PATCH /me",
		withMiddlewares(requireAuth)(handlePatchMe(logger, queries)))

	mux.Handle("POST /families",
		withMiddlewares(requireAuth)(handleCreateFamily(logger, familyService)))
	mux.Handle("GET /families",
		withMiddlewares(requireAuth)(handleListFamilies(logger, queries)))
	mux.Handle("GET /families/{familyID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleGetFamily(logger, queries)))
	mux.Handle("PATCH /families/{familyID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handlePatchFamily(logger, queries)))
	mux.Handle("DELETE /families/{familyID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleDeleteFamily(logger, queries)))
	mux.Handle("PATCH /families/{familyID}/members/{userID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handlePatchFamilyMember(logger, familyService)))
	mux.Handle("DELETE /families/{familyID}/members/{userID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleDeleteFamilyMember(logger, familyService)))
}
