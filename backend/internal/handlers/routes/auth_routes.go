// path: backend/internal/handlers/routes/auth_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// RegisterAuthRoutes sets up all authentication-related routes
// Only includes IMPLEMENTED endpoints (no TODOs)
func RegisterAuthRoutes(r chi.Router, h *handlers.AuthHandlerV2, authMW *middleware.AuthMiddleware) {
	// ========================================================================
	// PUBLIC AUTH ROUTES (no authentication required)
	// ========================================================================
	r.Route("/auth", func(r chi.Router) {
		// ✅ Implemented endpoints
		r.Post("/signup", h.Signup)
		r.Post("/login", h.Login)

		// TODO: Implement these later
		// r.Post("/verify-email", h.VerifyEmail)
		// r.Post("/resend-verification", h.ResendVerification)
		// r.Post("/forgot-password", h.ForgotPassword)
		// r.Post("/reset-password", h.ResetPassword)
		// r.Post("/refresh", h.RefreshToken)
		// r.Post("/logout", h.Logout)
		// r.Post("/change-password", h.ChangePassword)
	})
}
