// path: backend/internal/handlers/routes/team_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

func RegisterTeamRoutes(r chi.Router, h *handlers.TeamHandler, authMW *middleware.AuthMiddleware) {
	r.Route("/teams", func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		// Team CRUD
		r.Get("/", h.ListTeams)
		r.Post("/", h.CreateTeam)
		r.Get("/{id}", h.GetTeam)
		r.Put("/{id}", h.UpdateTeam)
		r.Delete("/{id}", h.DeleteTeam)

		// Team member management
		r.Post("/{id}/members", h.InviteMember)
		r.Delete("/{id}/members/{userId}", h.RemoveMember)
		r.Patch("/{id}/members/{userId}/role", h.UpdateMemberRole)

		// NEW: Accept invitation
		r.Post("/{id}/accept", h.AcceptInvitation)
	})

	// NEW: Invitations route (separate from /teams)
	r.Route("/invitations", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Get("/pending", h.GetPendingInvitations)
	})
}
