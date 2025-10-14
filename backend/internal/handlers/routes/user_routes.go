// path: backend/internal/handlers/routes/user_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// RegisterUserRoutes sets up all user management routes
// Only includes IMPLEMENTED endpoints (no TODOs)
func RegisterUserRoutes(r chi.Router, h *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	// ========================================================================
	// USER ROUTES (authentication required)
	// ========================================================================
	r.Route("/users", func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		// ✅ Implemented user profile operations
		r.Get("/{id}", h.GetUser)       // Get user by ID
		r.Put("/{id}", h.UpdateUser)    // Update user profile
		r.Delete("/{id}", h.DeleteUser) // Delete user account

		// TODO: Implement these later
		// r.Get("/{id}/preferences", h.GetUserPreferences)
		// r.Put("/{id}/preferences", h.UpdateUserPreferences)
		// r.Get("/{id}/notifications", h.GetUserNotifications)
		// r.Put("/{id}/notifications/{notifId}", h.MarkNotificationRead)
		// r.Get("/{id}/activity", h.GetUserActivity)
	})

	// ========================================================================
	// CURRENT USER ROUTES (/me shortcuts)
	// TODO: Implement these when you add "current user" methods
	// ========================================================================
	// r.Route("/me", func(r chi.Router) {
	// 	r.Use(authMW.RequireAuth)
	//
	// 	r.Get("/", h.GetCurrentUser)
	// 	r.Put("/", h.UpdateCurrentUser)
	// 	r.Get("/teams", h.GetCurrentUserTeams)
	// 	r.Get("/stats", h.GetCurrentUserStats)
	// })
}
