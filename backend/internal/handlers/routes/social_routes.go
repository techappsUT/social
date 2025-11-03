// backend/internal/handlers/routes/social_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// RegisterSocialRoutes registers social OAuth and account routes
func RegisterSocialRoutes(r chi.Router, h *handlers.SocialHandler, authMW *middleware.AuthMiddleware) {
	if h == nil {
		return
	}

	r.Route("/social", func(r chi.Router) {
		// PUBLIC OAuth routes
		r.Route("/auth", func(r chi.Router) {
			r.Get("/{platform}", h.GetOAuthURL)
			r.Get("/{platform}/callback", h.OAuthCallback)

			// Protected complete endpoint
			r.With(authMW.RequireAuth).Post("/complete", h.CompleteOAuthConnection)
		})

		// PROTECTED account management routes
		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAuth)

			// Account CRUD operations
			r.Route("/accounts", func(r chi.Router) {
				r.Post("/", h.ConnectAccount)
				r.Get("/", h.ListAccounts) // List all accounts for current user

				// Account-specific operations
				r.Route("/{id}", func(r chi.Router) {
					r.Delete("/", h.DisconnectAccount)
					r.Post("/refresh", h.RefreshTokens)
					r.Post("/publish", h.PublishPost)
					r.Get("/posts/{postId}/analytics", h.GetAnalytics)
				})
			})
		})
	})

	// Team-specific social account routes
	r.Route("/teams/{teamId}/social", func(r chi.Router) {
		r.Use(authMW.RequireAuth)
		r.Get("/accounts", h.ListAccounts)
	})
}

// // path: backend/internal/handlers/routes/social_routes.go
// package routes

// import (
// 	"github.com/go-chi/chi/v5"
// 	"github.com/techappsUT/social-queue/internal/handlers"
// 	"github.com/techappsUT/social-queue/internal/middleware"
// )

// // RegisterSocialRoutes registers social OAuth and account routes
// func RegisterSocialRoutes(r chi.Router, h *handlers.SocialHandler, authMW *middleware.AuthMiddleware) {
// 	if h == nil {
// 		// Social handler not available
// 		return
// 	}

// 	r.Route("/social", func(r chi.Router) {
// 		r.Use(authMW.RequireAuth)

// 		// OAuth flow
// 		r.Get("/auth/{platform}", h.GetOAuthURL)
// 		r.Get("/auth/{platform}/callback", h.OAuthCallback)

// 		// Account management
// 		r.Post("/accounts", h.ConnectAccount)
// 		r.Delete("/accounts/{id}", h.DisconnectAccount)
// 		r.Post("/accounts/{id}/refresh", h.RefreshTokens)

// 		// Publishing & Analytics
// 		r.Post("/accounts/{id}/publish", h.PublishPost)
// 		r.Get("/accounts/{id}/posts/{postId}/analytics", h.GetAnalytics)
// 	})
// }
