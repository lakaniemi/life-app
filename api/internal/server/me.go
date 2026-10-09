package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/db"
)

// userResponse is the public view of a user. db.User isn't serialized
// directly, so internal fields like google_sub can't leak into responses.
type userResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func newUserResponse(u db.User) userResponse {
	return userResponse{ID: u.ID, Name: u.Name}
}

// familyResponse is filled in by the families endpoints.
type familyResponse struct{}

type meResponse struct {
	User     userResponse     `json:"user"`
	Families []familyResponse `json:"families"`
}

func handleGetMe(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := queries.GetUserByID(r.Context(), sessionFrom(r.Context()).UserID)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, meResponse{
			User: newUserResponse(user),
			// Non-nil, so it encodes as [] rather than null.
			Families: []familyResponse{},
		})
	})
}

type patchMeRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type patchMeResponse struct {
	User userResponse `json:"user"`
}

func handlePatchMe(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req patchMeRequest
		if !decodeJSON(w, r, logger, &req) {
			return
		}
		// Trimmed before validating, so a blank name fails "required".
		req.Name = strings.TrimSpace(req.Name)
		if !validateRequest(w, r, logger, req) {
			return
		}

		user, err := queries.UpdateUserName(r.Context(), db.UpdateUserNameParams{
			ID:   sessionFrom(r.Context()).UserID,
			Name: req.Name,
		})
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, patchMeResponse{User: newUserResponse(user)})
	})
}
