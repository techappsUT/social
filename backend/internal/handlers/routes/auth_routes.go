// backend/internal/handlers/routes/auth_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// RegisterAuthRoutes sets up all authentication-related routes
func RegisterAuthRoutes(r chi.Router, h *handlers.AuthHandler, authMW *middleware.AuthMiddleware) {
	// ========================================================================
	// PUBLIC AUTH ROUTES (no authentication required)
	// ========================================================================
	r.Route("/auth", func(r chi.Router) {
		// User registration & authentication
		r.Post("/signup", h.Signup)
		r.Post("/login", h.Login)

		// Token management
		r.Post("/refresh", h.RefreshToken)
		r.Post("/logout", h.Logout)

		// Email verification
		r.Post("/verify-email", h.VerifyEmail)
		r.Post("/resend-verification", h.ResendVerification)

		// Password reset
		r.Post("/forgot-password", h.ForgotPassword)
		r.Post("/reset-password", h.ResetPassword)
	})

	// ========================================================================
	// PROTECTED AUTH ROUTES (require authentication)
	// ========================================================================
	r.Route("/auth", func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		// Change password (requires current password)
		r.Post("/change-password", h.ChangePassword)
	})
}
