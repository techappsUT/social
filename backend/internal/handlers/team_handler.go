// ============================================================================
// FILE: backend/internal/handlers/team_handler.go
// UPDATED - Added 3 new member management methods
// ============================================================================
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/team"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/middleware"
)

type TeamHandler struct {
	createTeamUC       *team.CreateTeamUseCase
	getTeamUC          *team.GetTeamUseCase
	updateTeamUC       *team.UpdateTeamUseCase
	deleteTeamUC       *team.DeleteTeamUseCase
	listTeamsUC        *team.ListTeamsUseCase
	inviteMemberUC     *team.InviteMemberUseCase     // NEW
	removeMemberUC     *team.RemoveMemberUseCase     // NEW
	updateMemberRoleUC *team.UpdateMemberRoleUseCase // NEW
}

func NewTeamHandler(
	createTeamUC *team.CreateTeamUseCase,
	getTeamUC *team.GetTeamUseCase,
	updateTeamUC *team.UpdateTeamUseCase,
	deleteTeamUC *team.DeleteTeamUseCase,
	listTeamsUC *team.ListTeamsUseCase,
	inviteMemberUC *team.InviteMemberUseCase, // NEW
	removeMemberUC *team.RemoveMemberUseCase, // NEW
	updateMemberRoleUC *team.UpdateMemberRoleUseCase, // NEW
) *TeamHandler {
	return &TeamHandler{
		createTeamUC:       createTeamUC,
		getTeamUC:          getTeamUC,
		updateTeamUC:       updateTeamUC,
		deleteTeamUC:       deleteTeamUC,
		listTeamsUC:        listTeamsUC,
		inviteMemberUC:     inviteMemberUC,     // NEW
		removeMemberUC:     removeMemberUC,     // NEW
		updateMemberRoleUC: updateMemberRoleUC, // NEW
	}
}

// ============================================================================
// EXISTING HANDLERS (Keep these as-is)
// ============================================================================

// CreateTeam handles POST /api/v2/teams
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input team.CreateTeamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.OwnerID = userID

	output, err := h.createTeamUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondCreated(w, output)
}

// GetTeam handles GET /api/v2/teams/:id
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	input := team.GetTeamInput{
		TeamID: teamID,
		UserID: userID,
	}

	output, err := h.getTeamUC.Execute(r.Context(), input)
	if err != nil {
		if err.Error() == "access denied: user is not a team member" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusNotFound, "team not found")
		}
		return
	}

	respondSuccess(w, output)
}

// UpdateTeam handles PUT /api/v2/teams/:id
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	var input team.UpdateTeamInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.TeamID = teamID
	input.UserID = userID

	output, err := h.updateTeamUC.Execute(r.Context(), input)
	if err != nil {
		if err.Error() == "access denied: admin role required" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondSuccess(w, output)
}

// DeleteTeam handles DELETE /api/v2/teams/:id
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	input := team.DeleteTeamInput{
		TeamID: teamID,
		UserID: userID,
	}

	if err := h.deleteTeamUC.Execute(r.Context(), input); err != nil {
		if err.Error() == "access denied: only team owner can delete team" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondNoContent(w)
}

// ListTeams handles GET /api/v2/teams
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input := team.ListTeamsInput{
		UserID: userID,
	}

	output, err := h.listTeamsUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list teams")
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// NEW MEMBER MANAGEMENT HANDLERS
// ============================================================================

// InviteMember handles POST /api/v2/teams/:id/members
func (h *TeamHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	var input team.InviteMemberInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.TeamID = teamID
	input.InviterID = userID

	output, err := h.inviteMemberUC.Execute(r.Context(), input)
	if err != nil {
		// Handle different error types appropriately
		switch {
		case err.Error() == "access denied: only admins and owners can invite members":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "user is already a team member":
			respondError(w, http.StatusConflict, err.Error())
		case err.Error() == "team member limit reached":
			respondError(w, http.StatusForbidden, err.Error())
		default:
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondCreated(w, output)
}

// RemoveMember handles DELETE /api/v2/teams/:id/members/:userId
func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	userIDToRemoveStr := chi.URLParam(r, "userId")
	userIDToRemove, err := uuid.Parse(userIDToRemoveStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	input := team.RemoveMemberInput{
		TeamID:    teamID,
		UserID:    userIDToRemove,
		RemoverID: userID,
	}

	if err := h.removeMemberUC.Execute(r.Context(), input); err != nil {
		// Handle different error types appropriately
		switch {
		case err.Error() == "access denied: only admins and owners can remove members":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "cannot remove team owner, transfer ownership first":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "cannot remove the last admin":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "member not found":
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondNoContent(w)
}

// UpdateMemberRole handles PATCH /api/v2/teams/:id/members/:userId/role
func (h *TeamHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	memberUserIDStr := chi.URLParam(r, "userId")
	memberUserID, err := uuid.Parse(memberUserIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var requestBody struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := team.UpdateMemberRoleInput{
		TeamID:    teamID,
		UserID:    memberUserID,
		NewRole:   teamDomain.MemberRole(requestBody.Role),
		UpdaterID: userID,
	}

	output, err := h.updateMemberRoleUC.Execute(r.Context(), input)
	if err != nil {
		// Handle different error types appropriately
		switch {
		case err.Error() == "access denied: only team owner can change member roles":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "cannot change your own role":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "cannot demote the last owner":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "cannot demote the last admin":
			respondError(w, http.StatusForbidden, err.Error())
		case err.Error() == "member not found":
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondSuccess(w, output)
}
