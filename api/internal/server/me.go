package server

import (
	"log/slog"
	"net/http"

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

type meResponse struct {
	User     userResponse     `json:"user"`
	Families []familyResponse `json:"families"`
}

func handleGetMe(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := sessionFrom(r.Context()).UserID
		user, err := queries.GetUserByID(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		families, err := queries.ListFamiliesForUser(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, meResponse{
			User:     newUserResponse(user),
			Families: mapSlice(families, newFamilyResponse),
		})
	})
}

type patchMeRequest struct {
	Name string `json:"name" mod:"trim" validate:"required,max=100"`
}

type patchMeResponse struct {
	User userResponse `json:"user"`
}

func handlePatchMe(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req patchMeRequest
		if !decodeAndValidateRequest(w, r, logger, &req) {
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
