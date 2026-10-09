package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/lakaniemi/life-app/api/internal/auth"
)

type nonceResponse struct {
	Nonce string `json:"nonce"`
}

func handleCreateNonce(logger *slog.Logger, authService *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := authService.NewNonce(r.Context())
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, nonceResponse{Nonce: nonce})
	})
}

type googleLoginRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}

type googleLoginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

func handleGoogleLogin(logger *slog.Logger, authService *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req googleLoginRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeError(w, r, logger, http.StatusBadRequest, "invalid_body", err.Error())
			return
		}
		if !validateRequest(w, r, logger, req) {
			return
		}

		token, user, err := authService.LoginWithGoogle(r.Context(), req.IDToken)
		switch {
		case errors.Is(err, auth.ErrInvalidIDToken):
			// The reason stays in the logs. This also covers failing to fetch
			// Google's keys, which isn't the client's fault, but can't be told
			// apart from a bad token reliably.
			logger.WarnContext(r.Context(), "google login rejected", "err", err)
			writeError(w, r, logger, http.StatusUnauthorized, "invalid_id_token", "invalid Google ID token")
			return
		case errors.Is(err, auth.ErrInvalidNonce):
			writeError(w, r, logger, http.StatusUnauthorized, "invalid_nonce", err.Error())
			return
		case err != nil:
			writeInternalError(w, r, logger, err)
			return
		}

		writeJSON(w, r, logger, http.StatusOK, googleLoginResponse{Token: token, User: newUserResponse(user)})
	})
}

func handleLogout(logger *slog.Logger, authService *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := authService.Logout(r.Context(), sessionFrom(r.Context()).SessionID); err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
