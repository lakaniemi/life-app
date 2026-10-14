package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakaniemi/life-app/api/internal/auth"
	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
	"github.com/lakaniemi/life-app/api/internal/user"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, pool *pgxpool.Pool, verifier auth.TokenVerifier, version string) {
	queries := db.New(pool)
	authService := auth.NewService(logger, queries, verifier)
	familyService := family.NewService(logger, pool)
	userService := user.NewService(queries)

	requireAuth := newRequireAuth(logger, authService)
	requireFamilyMembership := newRequireFamilyMembership(logger, familyService)

	mux.Handle("GET /health", handleHealth(logger, pool, version))

	mux.Handle("POST /auth/nonce", handleCreateNonce(authService))
	mux.Handle("POST /auth/google", handleGoogleLogin(logger, authService))
	mux.Handle("POST /auth/logout",
		withMiddlewares(requireAuth)(handleLogout(authService)))

	mux.Handle("GET /me",
		withMiddlewares(requireAuth)(handleGetMe(userService, familyService)))
	mux.Handle("PATCH /me",
		withMiddlewares(requireAuth)(handlePatchMe(userService)))

	mux.Handle("POST /families",
		withMiddlewares(requireAuth)(handleCreateFamily(familyService)))
	mux.Handle("GET /families",
		withMiddlewares(requireAuth)(handleListFamilies(familyService)))
	mux.Handle("GET /families/{familyID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleGetFamily(familyService)))
	mux.Handle("PATCH /families/{familyID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handlePatchFamily(familyService)))
	mux.Handle("DELETE /families/{familyID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleDeleteFamily(familyService)))
	mux.Handle("PATCH /families/{familyID}/members/{userID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handlePatchFamilyMember(familyService)))
	mux.Handle("DELETE /families/{familyID}/members/{userID}",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleDeleteFamilyMember(familyService)))
	mux.Handle("POST /families/{familyID}/invites",
		withMiddlewares(requireAuth, requireFamilyMembership)(handleCreateInvite(familyService)))

	// Invitees aren't members yet, so these aren't under /families/{familyID}.
	mux.Handle("GET /invites/{token}",
		withMiddlewares(requireAuth)(handlePreviewInvite(familyService)))
	mux.Handle("POST /invites/{token}/accept",
		withMiddlewares(requireAuth)(handleAcceptInvite(familyService)))
}
