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

func handleCreateNonce(authService *auth.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		nonce, err := authService.NewNonce(r.Context())
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, nonceResponse{Nonce: nonce})
	}
}

type googleLoginRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}

type googleLoginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

func handleGoogleLogin(logger *slog.Logger, authService *auth.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		var req googleLoginRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		token, user, err := authService.LoginWithGoogle(ctx, req.IDToken)
		if errors.Is(err, auth.ErrInvalidIDToken) {
			// The reason stays in the logs: the response only says the token is
			// invalid. This also covers failing to fetch Google's keys, which
			// isn't the client's fault, but can't be told apart from a bad token
			// reliably.
			logger.WarnContext(ctx, "google login rejected", "err", err)
		}
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, googleLoginResponse{Token: token, User: newUserResponse(user)})
	}
}

func handleLogout(authService *auth.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		sessionID := sessionFrom(ctx).SessionID

		if err := authService.Logout(ctx, sessionID); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
