package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/family"
)

type inviteResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type createInviteResponse struct {
	Invite inviteResponse `json:"invite"`
}

func handleCreateInvite(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		actor := familyMembershipFrom(ctx)

		token, expiresAt, err := familyService.CreateInvite(ctx, actor)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusCreated, createInviteResponse{
			Invite: inviteResponse{Token: token, ExpiresAt: expiresAt},
		})
	}
}

// invitedFamilyResponse is a family as seen by someone who may not belong to
// it yet, so it has no role.
type invitedFamilyResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type previewInviteResponse struct {
	Family        invitedFamilyResponse `json:"family"`
	AlreadyMember bool                  `json:"alreadyMember"`
}

func handlePreviewInvite(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		userID := sessionFrom(ctx).UserID
		token := r.PathValue("token")

		preview, err := familyService.PreviewInvite(ctx, userID, token)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, previewInviteResponse{
			Family:        invitedFamilyResponse{ID: preview.ID, Name: preview.Name},
			AlreadyMember: preview.AlreadyMember,
		})
	}
}

func handleAcceptInvite(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		userID := sessionFrom(ctx).UserID
		token := r.PathValue("token")

		if err := familyService.AcceptInvite(ctx, userID, token); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
