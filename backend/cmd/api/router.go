// ============================================================================
// FILE: backend/cmd/api/router.go
// FIXED: Removed TeamHandler methods that don't exist yet
// (InviteMember, RemoveMember, UpdateMemberRole)
// ============================================================================
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	appMiddleware "github.com/techappsUT/social-queue/internal/middleware"
)

// setupRouter creates and configures the HTTP router
func setupRouter(container *Container) *chi.Mux {
	r := chi.NewRouter()

	// ============================================================================
	// GLOBAL MIDDLEWARE
	// ============================================================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.DefaultLogger)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   container.Config.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ============================================================================
	// PUBLIC ROUTES
	// ============================================================================

	r.Get("/", handleRoot(container))
	r.Get("/health", handleHealth(container))

	// ============================================================================
	// API V2 ROUTES (Clean Architecture)
	// ============================================================================

	r.Route("/api/v2", func(r chi.Router) {
		// Public authentication routes
		r.Post("/auth/signup", container.AuthHandler.Signup)
		r.Post("/auth/login", container.AuthHandler.Login)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(container.AuthMiddleware.RequireAuth)

			// Current user route
			r.Get("/me", handleMe)

			// User management routes
			r.Get("/users/{id}", container.AuthHandler.GetUser)
			r.Put("/users/{id}", container.AuthHandler.UpdateUser)
			r.Delete("/users/{id}", container.AuthHandler.DeleteUser)

			// Team routes
			r.Route("/teams", func(r chi.Router) {
				r.Get("/", container.TeamHandler.ListTeams)         // List user's teams
				r.Post("/", container.TeamHandler.CreateTeam)       // Create team
				r.Get("/{id}", container.TeamHandler.GetTeam)       // Get team details
				r.Put("/{id}", container.TeamHandler.UpdateTeam)    // Update team
				r.Delete("/{id}", container.TeamHandler.DeleteTeam) // Delete team

				// REMOVED: Team member management methods - these use cases haven't been implemented yet
				// TODO: Implement these in future iteration:
				// - POST /{id}/members - InviteMember
				// - DELETE /{id}/members/{userId} - RemoveMember
				// - PATCH /{id}/members/{userId}/role - UpdateMemberRole
			})

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.RequireAdmin)
				r.Get("/admin/users", handleAdminUsers)
			})
		})
	})

	return r
}

// ============================================================================
// HANDLER FUNCTIONS
// ============================================================================

// handleRoot returns API information
func handleRoot(c *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message":     "SocialQueue API",
			"version":     "2.0.0",
			"environment": c.Config.Environment,
			"docs":        "/api/v2/docs",
			"endpoints": map[string]string{
				"health": "/health",
				"signup": "/api/v2/auth/signup",
				"login":  "/api/v2/auth/login",
				"teams":  "/api/v2/teams",
			},
		}
		respondJSON(w, http.StatusOK, response)
	}
}

// handleHealth returns system health status
func handleHealth(c *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Test database connection
		ctx := r.Context()
		if err := c.DB.PingContext(ctx); err != nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status":   "unhealthy",
				"database": "down",
				"error":    err.Error(),
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":      "healthy",
			"database":    "up",
			"environment": c.Config.Environment,
			"timestamp":   time.Now().Unix(),
		})
	}
}

// handleMe returns the current authenticated user's info
func handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := appMiddleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	email, _ := appMiddleware.GetUserEmail(r.Context())
	role, _ := appMiddleware.GetUserRole(r.Context())

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"userId": userID.String(),
		"email":  email,
		"role":   role,
	})
}

// handleAdminUsers is a placeholder for admin user management
func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Admin users endpoint - to be implemented",
	})
}

// ============================================================================
// RESPONSE HELPERS
// ============================================================================

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	})
}
