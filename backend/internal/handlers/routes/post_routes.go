// path: backend/internal/handlers/routes/post_routes.go
package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// RegisterPostRoutes registers post-related routes
func RegisterPostRoutes(r chi.Router, h *handlers.PostHandler, authMW *middleware.AuthMiddleware) {
	r.Route("/posts", func(r chi.Router) {
		r.Use(authMW.RequireAuth)

		// Post CRUD
		r.Post("/", h.CreateDraft)
		r.Get("/", h.ListPosts)
		r.Get("/{id}", h.GetPost)
		r.Put("/{id}", h.UpdatePost)
		r.Delete("/{id}", h.DeletePost)

		// Post actions
		r.Post("/{id}/schedule", h.SchedulePost)
		r.Post("/{id}/publish", h.PublishNow)
	})
}
