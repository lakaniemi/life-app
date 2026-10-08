package server

import (
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/db"
)

const maxNameLength = 100

// userResponse is the public view of a user. db.User isn't serialized
// directly, so internal fields like google_sub can't leak into responses.
type userResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func newUserResponse(u db.User) userResponse {
	return userResponse{ID: u.ID, Name: u.Name}
}

// familyResponse is filled in by the families endpoints (phase 3).
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
	Name string `json:"name"`
}

type patchMeResponse struct {
	User userResponse `json:"user"`
}

func handlePatchMe(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req patchMeRequest
		if err := decodeJSON(w, r, &req); err != nil {
			writeError(w, r, logger, http.StatusBadRequest, "invalid_body", err.Error())
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" || utf8.RuneCountInString(name) > maxNameLength {
			writeError(w, r, logger, http.StatusBadRequest, "invalid_name", "name must be 1-100 characters")
			return
		}

		user, err := queries.UpdateUserName(r.Context(), db.UpdateUserNameParams{
			ID:   sessionFrom(r.Context()).UserID,
			Name: name,
		})
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, patchMeResponse{User: newUserResponse(user)})
	})
}
