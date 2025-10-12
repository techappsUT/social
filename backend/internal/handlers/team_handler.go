// ============================================================================
// FILE: backend/internal/handlers/team_handler.go
// CORRECTLY FIXED - middleware.GetUserID returns (uuid.UUID, bool) not error
// ============================================================================
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/team"
	"github.com/techappsUT/social-queue/internal/middleware"
)

type TeamHandler struct {
	createTeamUC *team.CreateTeamUseCase
	getTeamUC    *team.GetTeamUseCase
	updateTeamUC *team.UpdateTeamUseCase
	deleteTeamUC *team.DeleteTeamUseCase
	listTeamsUC  *team.ListTeamsUseCase
}

func NewTeamHandler(
	createTeamUC *team.CreateTeamUseCase,
	getTeamUC *team.GetTeamUseCase,
	updateTeamUC *team.UpdateTeamUseCase,
	deleteTeamUC *team.DeleteTeamUseCase,
	listTeamsUC *team.ListTeamsUseCase,
) *TeamHandler {
	return &TeamHandler{
		createTeamUC: createTeamUC,
		getTeamUC:    getTeamUC,
		updateTeamUC: updateTeamUC,
		deleteTeamUC: deleteTeamUC,
		listTeamsUC:  listTeamsUC,
	}
}

// CreateTeam handles POST /api/v2/teams
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	// FIXED: GetUserID returns (uuid.UUID, bool) not (uuid.UUID, error)
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
	// FIXED: GetUserID returns (uuid.UUID, bool) not (uuid.UUID, error)
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
	// FIXED: GetUserID returns (uuid.UUID, bool) not (uuid.UUID, error)
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
	// FIXED: GetUserID returns (uuid.UUID, bool) not (uuid.UUID, error)
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
	// FIXED: GetUserID returns (uuid.UUID, bool) not (uuid.UUID, error)
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
