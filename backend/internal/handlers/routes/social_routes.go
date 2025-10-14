// path: backend/internal/handlers/routes/social_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// RegisterSocialRoutes registers social OAuth and account routes
func RegisterSocialRoutes(r chi.Router, h *handlers.SocialHandler, authMW *middleware.AuthMiddleware) {
	if h == nil {
		// Social handler not available
		return
	}

	r.Route("/social", func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		// OAuth flow
		r.Get("/auth/{platform}", h.GetOAuthURL)
		r.Get("/auth/{platform}/callback", h.OAuthCallback)

		// Account management
		r.Post("/accounts", h.ConnectAccount)
		r.Delete("/accounts/{id}", h.DisconnectAccount)
		r.Post("/accounts/{id}/refresh", h.RefreshTokens)

		// Publishing & Analytics
		r.Post("/accounts/{id}/publish", h.PublishPost)
		r.Get("/accounts/{id}/posts/{postId}/analytics", h.GetAnalytics)
	})
}
