package server

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
	"github.com/lakaniemi/life-app/api/internal/user"
)

// userResponse is the public view of a user. db.User isn't serialized
// directly, so internal fields like google_sub can't leak into responses.
type userResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func newUserResponse(userRow db.User) userResponse {
	return userResponse{ID: userRow.ID, Name: userRow.Name}
}

type meResponse struct {
	User     userResponse     `json:"user"`
	Families []familyResponse `json:"families"`
}

func handleGetMe(userService *user.Service, familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		userID := sessionFrom(ctx).UserID

		currentUser, err := userService.Get(ctx, userID)
		if err != nil {
			return err
		}
		families, err := familyService.ListForUser(ctx, userID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, meResponse{
			User:     newUserResponse(currentUser),
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

func handlePatchMe(userService *user.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		userID := sessionFrom(ctx).UserID
		var req patchMeRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		renamed, err := userService.Rename(ctx, userID, req.Name)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, patchMeResponse{User: newUserResponse(renamed)})
	}
}
