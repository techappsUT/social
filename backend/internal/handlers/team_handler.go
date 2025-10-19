// ============================================================================
// FILE: backend/internal/handlers/team_handler.go
// UPDATED - Added 3 new member management methods
// ============================================================================
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/team"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/middleware"
)

type TeamHandler struct {
	createTeamUC            *team.CreateTeamUseCase
	getTeamUC               *team.GetTeamUseCase
	updateTeamUC            *team.UpdateTeamUseCase
	deleteTeamUC            *team.DeleteTeamUseCase
	listTeamsUC             *team.ListTeamsUseCase
	inviteMemberUC          *team.InviteMemberUseCase          // NEW
	removeMemberUC          *team.RemoveMemberUseCase          // NEW
	updateMemberRoleUC      *team.UpdateMemberRoleUseCase      // NEW
	acceptInvitationUC      *team.AcceptInvitationUseCase      // NEW
	getPendingInvitationsUC *team.GetPendingInvitationsUseCase // NEW
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
	acceptInvitationUC *team.AcceptInvitationUseCase, // NEW
	getPendingInvitationsUC *team.GetPendingInvitationsUseCase, // NEW
) *TeamHandler {
	return &TeamHandler{
		createTeamUC:            createTeamUC,
		getTeamUC:               getTeamUC,
		updateTeamUC:            updateTeamUC,
		deleteTeamUC:            deleteTeamUC,
		listTeamsUC:             listTeamsUC,
		inviteMemberUC:          inviteMemberUC,          // NEW
		removeMemberUC:          removeMemberUC,          // NEW
		updateMemberRoleUC:      updateMemberRoleUC,      // NEW
		acceptInvitationUC:      acceptInvitationUC,      // NEW
		getPendingInvitationsUC: getPendingInvitationsUC, // NEW
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
// func (h *TeamHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
// 	userID, ok := middleware.GetUserID(r.Context())
// 	if !ok {
// 		respondError(w, http.StatusUnauthorized, "unauthorized")
// 		return
// 	}

// 	teamIDStr := chi.URLParam(r, "id")
// 	teamID, err := uuid.Parse(teamIDStr)
// 	if err != nil {
// 		respondError(w, http.StatusBadRequest, "invalid team ID")
// 		return
// 	}

// 	var input team.InviteMemberInput
// 	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
// 		respondError(w, http.StatusBadRequest, "invalid request body")
// 		return
// 	}

// 	input.TeamID = teamID
// 	input.InviterID = userID

// 	output, err := h.inviteMemberUC.Execute(r.Context(), input)
// 	if err != nil {
// 		// Handle different error types appropriately
// 		switch {
// 		case err.Error() == "access denied: only admins and owners can invite members":
// 			respondError(w, http.StatusForbidden, err.Error())
// 		case err.Error() == "user is already a team member":
// 			respondError(w, http.StatusConflict, err.Error())
// 		case err.Error() == "team member limit reached":
// 			respondError(w, http.StatusForbidden, err.Error())
// 		default:
// 			respondError(w, http.StatusBadRequest, err.Error())
// 		}
// 		return
// 	}

// 	respondCreated(w, output)
// }

func (h *TeamHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		// ✅ Consistent error with code and message
		middleware.RespondUnauthorized(w, "Authentication required")
		return
	}

	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		// ✅ Clear error code
		middleware.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid team ID format")
		return
	}

	var requestBody struct {
		Email string `json:"email" validate:"required,email"`
		Role  string `json:"role" validate:"required,oneof=owner admin editor viewer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	// ✅ Validate with detailed field errors
	if err := middleware.ValidateStruct(requestBody); err != nil {
		fields := middleware.FormatValidationErrors(err)
		middleware.RespondValidationError(w, "Request validation failed", fields)
		return
	}

	input := team.InviteMemberInput{
		TeamID:    teamID,
		Email:     requestBody.Email,
		Role:      teamDomain.MemberRole(requestBody.Role),
		InviterID: userID,
	}

	output, err := h.inviteMemberUC.Execute(r.Context(), input)
	if err != nil {
		// ✅ Map domain errors to HTTP errors correctly
		switch {
		case strings.Contains(err.Error(), "access denied"):
			middleware.RespondForbidden(w, err.Error()) // ✅ 403
		case strings.Contains(err.Error(), "limit reached"):
			middleware.RespondError(w, http.StatusPaymentRequired, "limit_exceeded", err.Error()) // ✅ 402
		case strings.Contains(err.Error(), "not found"):
			middleware.RespondNotFound(w, "team") // ✅ 404
		case strings.Contains(err.Error(), "already exists"):
			middleware.RespondConflict(w, err.Error()) // ✅ 409
		default:
			middleware.RespondInternalError(w, "Failed to invite member") // ✅ 500, safe message
		}
		return
	}

	// ✅ Consistent success response
	middleware.RespondCreated(w, output)
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

func (h *TeamHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	// ✅ Validate auth first
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondUnauthorized(w, "Authentication required")
		return
	}

	// ✅ Validate URL params
	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid team ID format")
		return
	}

	memberUserIDStr := chi.URLParam(r, "userId")
	memberUserID, err := uuid.Parse(memberUserIDStr)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format")
		return
	}

	// ✅ Validate request body structure
	var requestBody struct {
		Role string `json:"role" validate:"required,oneof=owner admin editor viewer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	// ✅ Validate fields
	if err := middleware.ValidateStruct(requestBody); err != nil {
		fields := middleware.FormatValidationErrors(err)
		middleware.RespondValidationError(w, "Request validation failed", fields)
		return
	}

	// ✅ Only valid data reaches use case
	input := team.UpdateMemberRoleInput{
		TeamID:    teamID,
		UserID:    memberUserID,
		NewRole:   teamDomain.MemberRole(requestBody.Role),
		UpdaterID: userID,
	}

	output, err := h.updateMemberRoleUC.Execute(r.Context(), input)
	if err != nil {
		// Handle errors...
		return
	}

	middleware.RespondSuccess(w, output)
}

// ============================================================================
// NEW INVITATION HANDLERS
// ============================================================================

// AcceptInvitation handles POST /api/v2/teams/:id/accept
func (h *TeamHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	// 1. Get authenticated user
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondUnauthorized(w, "Authentication required")
		return
	}

	// 2. Validate team ID from URL
	teamIDStr := chi.URLParam(r, "id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid team ID format")
		return
	}

	// 3. Execute use case
	input := team.AcceptInvitationInput{
		TeamID: teamID,
		UserID: userID,
	}

	output, err := h.acceptInvitationUC.Execute(r.Context(), input)
	if err != nil {
		switch {
		case err.Error() == "invitation not found":
			middleware.RespondNotFound(w, "invitation")
		case err.Error() == "invitation already processed":
			middleware.RespondConflict(w, "Invitation has already been processed")
		default:
			middleware.RespondInternalError(w, "Failed to accept invitation")
		}
		return
	}

	// 4. Success response
	middleware.RespondSuccess(w, output)
}

// GetPendingInvitations handles GET /api/v2/invitations/pending
func (h *TeamHandler) GetPendingInvitations(w http.ResponseWriter, r *http.Request) {
	// 1. Get authenticated user
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondUnauthorized(w, "Authentication required")
		return
	}

	// 2. Execute use case
	input := team.GetPendingInvitationsInput{
		UserID: userID,
	}

	output, err := h.getPendingInvitationsUC.Execute(r.Context(), input)
	if err != nil {
		middleware.RespondInternalError(w, "Failed to get pending invitations")
		return
	}

	// 3. Success response
	middleware.RespondSuccess(w, output)
}
