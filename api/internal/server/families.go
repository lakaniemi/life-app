package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/lakaniemi/life-app/api/internal/db"
	"github.com/lakaniemi/life-app/api/internal/family"
)

// writeFamilyError maps an error from internal/family to a response. Rule
// violations are 409: the request is valid, but the family's current state
// doesn't allow it.
func writeFamilyError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	switch {
	case errors.Is(err, family.ErrFamilyNotFound), errors.Is(err, family.ErrFamilyMemberNotFound):
		writeError(w, r, logger, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, family.ErrSoleAdmin):
		writeError(w, r, logger, http.StatusConflict, "sole_admin", err.Error())
	case errors.Is(err, family.ErrSoleMember):
		writeError(w, r, logger, http.StatusConflict, "sole_member", err.Error())
	default:
		writeInternalError(w, r, logger, err)
	}
}

// familyResponse is a family as seen by the caller: Role is the caller's role
// in it.
type familyResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role string    `json:"role"`
}

// newFamilyResponses returns a non-nil slice, so an empty list encodes as []
// rather than null.
func newFamilyResponses(rows []db.ListFamiliesForUserRow) []familyResponse {
	families := make([]familyResponse, len(rows))
	for i, f := range rows {
		families[i] = familyResponse{ID: f.ID, Name: f.Name, Role: f.Role}
	}
	return families
}

type memberResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

func newMemberResponse(m db.ListFamilyMembersRow) memberResponse {
	return memberResponse{ID: m.UserID, Name: m.Name, Role: m.Role, JoinedAt: m.JoinedAt}
}

// familyEnvelope is the body of responses that return a single family.
type familyEnvelope struct {
	Family familyResponse `json:"family"`
}

type familyNameRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// decodeFamilyName decodes, trims and validates a familyNameRequest. On false
// the response is written.
func decodeFamilyName(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (string, bool) {
	var req familyNameRequest
	if !decodeJSON(w, r, logger, &req) {
		return "", false
	}
	// Trimmed before validating, so a blank name fails "required".
	req.Name = strings.TrimSpace(req.Name)
	if !validateRequest(w, r, logger, req) {
		return "", false
	}
	return req.Name, true
}

func handleCreateFamily(logger *slog.Logger, families *family.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := decodeFamilyName(w, r, logger)
		if !ok {
			return
		}

		created, err := families.Create(r.Context(), sessionFrom(r.Context()).UserID, name)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusCreated, familyEnvelope{
			Family: familyResponse{ID: created.ID, Name: created.Name, Role: family.RoleAdmin},
		})
	})
}

type listFamiliesResponse struct {
	Families []familyResponse `json:"families"`
}

func handleListFamilies(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := queries.ListFamiliesForUser(r.Context(), sessionFrom(r.Context()).UserID)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, listFamiliesResponse{Families: newFamilyResponses(rows)})
	})
}

type getFamilyResponse struct {
	Family  familyResponse   `json:"family"`
	Members []memberResponse `json:"members"`
}

func handleGetFamily(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		membership := familyMembershipFrom(r.Context())

		f, err := queries.GetFamily(r.Context(), membership.FamilyID)
		if errors.Is(err, pgx.ErrNoRows) {
			// Deleted since requireFamilyMembership checked.
			writeError(w, r, logger, http.StatusNotFound, "not_found", "family not found")
			return
		}
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		rows, err := queries.ListFamilyMembers(r.Context(), membership.FamilyID)
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}

		members := make([]memberResponse, len(rows))
		for i, m := range rows {
			members[i] = newMemberResponse(m)
		}
		writeJSON(w, r, logger, http.StatusOK, getFamilyResponse{
			Family:  familyResponse{ID: f.ID, Name: f.Name, Role: membership.Role},
			Members: members,
		})
	})
}

func handlePatchFamily(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkFamilyAdmin(w, r, logger) {
			return
		}
		name, ok := decodeFamilyName(w, r, logger)
		if !ok {
			return
		}

		membership := familyMembershipFrom(r.Context())
		f, err := queries.UpdateFamilyName(r.Context(), db.UpdateFamilyNameParams{ID: membership.FamilyID, Name: name})
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, r, logger, http.StatusNotFound, "not_found", "family not found")
			return
		}
		if err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, familyEnvelope{
			Family: familyResponse{ID: f.ID, Name: f.Name, Role: membership.Role},
		})
	})
}

// handleDeleteFamily deletes the family. Its memberships and all family-owned
// data go with it through ON DELETE CASCADE.
func handleDeleteFamily(logger *slog.Logger, queries *db.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkFamilyAdmin(w, r, logger) {
			return
		}
		if err := queries.DeleteFamily(r.Context(), familyMembershipFrom(r.Context()).FamilyID); err != nil {
			writeInternalError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// memberIDFromPath parses the path's {userID}. A malformed ID can't name a
// member, so on false it has responded 404.
func memberIDFromPath(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (uuid.UUID, bool) {
	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		writeError(w, r, logger, http.StatusNotFound, "not_found", "family member not found")
		return uuid.UUID{}, false
	}
	return userID, true
}

type patchMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

type patchMemberResponse struct {
	Member memberResponse `json:"member"`
}

func handlePatchFamilyMember(logger *slog.Logger, families *family.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkFamilyAdmin(w, r, logger) {
			return
		}
		userID, ok := memberIDFromPath(w, r, logger)
		if !ok {
			return
		}
		var req patchMemberRequest
		if !decodeJSON(w, r, logger, &req) {
			return
		}
		if !validateRequest(w, r, logger, req) {
			return
		}

		member, err := families.ChangeRole(r.Context(), familyMembershipFrom(r.Context()).FamilyID, userID, req.Role)
		if err != nil {
			writeFamilyError(w, r, logger, err)
			return
		}
		writeJSON(w, r, logger, http.StatusOK, patchMemberResponse{Member: newMemberResponse(member)})
	})
}

// handleDeleteFamilyMember removes a member. Admins can remove anyone; any
// member can remove themselves, which is leaving the family.
func handleDeleteFamilyMember(logger *slog.Logger, families *family.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := memberIDFromPath(w, r, logger)
		if !ok {
			return
		}
		membership := familyMembershipFrom(r.Context())
		if userID != membership.UserID && !checkFamilyAdmin(w, r, logger) {
			return
		}

		if err := families.RemoveMember(r.Context(), membership.FamilyID, userID); err != nil {
			writeFamilyError(w, r, logger, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
