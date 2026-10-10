package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
)

// familyResponse is a family as seen by the caller: Role is the caller's role
// in it.
type familyResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

func newFamilyResponse(row db.ListFamiliesForUserRow) familyResponse {
	return familyResponse{ID: row.ID, Name: row.Name, Role: row.Role}
}

type memberResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

func newMemberResponse(member db.ListFamilyMembersRow) memberResponse {
	return memberResponse{ID: member.UserID, Name: member.Name, Role: member.Role, JoinedAt: member.JoinedAt}
}

// familyEnvelope is the body of responses that return a single family.
type familyEnvelope struct {
	Family familyResponse `json:"family"`
}

type familyNameRequest struct {
	Name string `json:"name" mod:"trim" validate:"required,max=100"`
}

func handleCreateFamily(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		userID := sessionFrom(ctx).UserID
		var req familyNameRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		created, err := familyService.Create(ctx, userID, req.Name)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusCreated, familyEnvelope{
			Family: familyResponse{ID: created.ID, Name: created.Name, Role: family.RoleAdmin},
		})
	}
}

type listFamiliesResponse struct {
	Families []familyResponse `json:"families"`
}

func handleListFamilies(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		userID := sessionFrom(ctx).UserID

		families, err := familyService.ListForUser(ctx, userID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, listFamiliesResponse{Families: mapSlice(families, newFamilyResponse)})
	}
}

type getFamilyResponse struct {
	Family  familyResponse   `json:"family"`
	Members []memberResponse `json:"members"`
}

func handleGetFamily(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		actor := familyMembershipFrom(ctx)

		familyRow, members, err := familyService.Get(ctx, actor)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, getFamilyResponse{
			Family:  familyResponse{ID: familyRow.ID, Name: familyRow.Name, Role: actor.Role},
			Members: mapSlice(members, newMemberResponse),
		})
	}
}

func handlePatchFamily(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		actor := familyMembershipFrom(ctx)
		var req familyNameRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		renamed, err := familyService.Rename(ctx, actor, req.Name)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, familyEnvelope{
			Family: familyResponse{ID: renamed.ID, Name: renamed.Name, Role: actor.Role},
		})
	}
}

func handleDeleteFamily(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		actor := familyMembershipFrom(ctx)

		if err := familyService.Delete(ctx, actor); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// memberIDFromPath parses the path's {userID}. A malformed ID can't name a
// member, so it's reported as not found.
func memberIDFromPath(r *http.Request) (uuid.UUID, error) {
	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		return uuid.UUID{}, family.ErrFamilyMemberNotFound
	}
	return userID, nil
}

type patchMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

type patchMemberResponse struct {
	Member memberResponse `json:"member"`
}

func handlePatchFamilyMember(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		actor := familyMembershipFrom(ctx)
		targetID, err := memberIDFromPath(r)
		if err != nil {
			return err
		}
		var req patchMemberRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		member, err := familyService.ChangeRole(ctx, actor, targetID, req.Role)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, patchMemberResponse{Member: newMemberResponse(member)})
	}
}

func handleDeleteFamilyMember(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		actor := familyMembershipFrom(ctx)
		targetID, err := memberIDFromPath(r)
		if err != nil {
			return err
		}

		if err := familyService.RemoveMember(ctx, actor, targetID); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
