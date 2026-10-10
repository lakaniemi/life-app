package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
		var req familyNameRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		created, err := familyService.Create(r.Context(), sessionFrom(r.Context()).UserID, req.Name)
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

func handleListFamilies(queries *db.Queries) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		rows, err := queries.ListFamiliesForUser(r.Context(), sessionFrom(r.Context()).UserID)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, listFamiliesResponse{Families: mapSlice(rows, newFamilyResponse)})
	}
}

type getFamilyResponse struct {
	Family  familyResponse   `json:"family"`
	Members []memberResponse `json:"members"`
}

func handleGetFamily(queries *db.Queries) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		membership := familyMembershipFrom(r.Context())

		familyRow, err := queries.GetFamily(r.Context(), membership.FamilyID)
		if errors.Is(err, pgx.ErrNoRows) {
			// Deleted since requireFamilyMembership checked.
			return family.ErrFamilyNotFound
		}
		if err != nil {
			return err
		}
		members, err := queries.ListFamilyMembers(r.Context(), membership.FamilyID)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, getFamilyResponse{
			Family:  familyResponse{ID: familyRow.ID, Name: familyRow.Name, Role: membership.Role},
			Members: mapSlice(members, newMemberResponse),
		})
	}
}

func handlePatchFamily(queries *db.Queries) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if err := checkFamilyAdmin(r); err != nil {
			return err
		}
		var req familyNameRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		membership := familyMembershipFrom(r.Context())
		familyRow, err := queries.UpdateFamilyName(r.Context(), db.UpdateFamilyNameParams{ID: membership.FamilyID, Name: req.Name})
		if errors.Is(err, pgx.ErrNoRows) {
			return family.ErrFamilyNotFound
		}
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, familyEnvelope{
			Family: familyResponse{ID: familyRow.ID, Name: familyRow.Name, Role: membership.Role},
		})
	}
}

// handleDeleteFamily deletes the family. Its memberships and all family-owned
// data go with it through ON DELETE CASCADE.
func handleDeleteFamily(queries *db.Queries) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if err := checkFamilyAdmin(r); err != nil {
			return err
		}
		if err := queries.DeleteFamily(r.Context(), familyMembershipFrom(r.Context()).FamilyID); err != nil {
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
		if err := checkFamilyAdmin(r); err != nil {
			return err
		}
		userID, err := memberIDFromPath(r)
		if err != nil {
			return err
		}
		var req patchMemberRequest
		if err := decodeAndValidateRequest(w, r, &req); err != nil {
			return err
		}

		member, err := familyService.ChangeRole(r.Context(), familyMembershipFrom(r.Context()).FamilyID, userID, req.Role)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, patchMemberResponse{Member: newMemberResponse(member)})
	}
}

// handleDeleteFamilyMember removes a member. Admins can remove anyone; any
// member can remove themselves, which is leaving the family.
func handleDeleteFamilyMember(familyService *family.Service) apiHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		userID, err := memberIDFromPath(r)
		if err != nil {
			return err
		}
		membership := familyMembershipFrom(r.Context())
		if userID != membership.UserID {
			if err := checkFamilyAdmin(r); err != nil {
				return err
			}
		}

		if err := familyService.RemoveMember(r.Context(), membership.FamilyID, userID); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
