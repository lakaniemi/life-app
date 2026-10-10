package server

import (
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

func handleGetMe(queries *db.Queries) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		userID := sessionFrom(r.Context()).UserID
		user, err := queries.GetUserByID(r.Context(), userID)
		if err != nil {
			return err
		}
		families, err := queries.ListFamiliesForUser(r.Context(), userID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, meResponse{
			User:     newUserResponse(user),
			Families: mapSlice(families, newFamilyResponse),
		})
	}
}

type patchMeRequest struct {
	Name string `json:"name" mod:"trim" validate:"required,max=100"`
}

type patchMeResponse struct {
	User userResponse `json:"user"`
}

func handlePatchMe(queries *db.Queries) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req patchMeRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		user, err := queries.UpdateUserName(r.Context(), db.UpdateUserNameParams{
			ID:   sessionFrom(r.Context()).UserID,
			Name: req.Name,
		})
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, patchMeResponse{User: newUserResponse(user)})
	}
}
