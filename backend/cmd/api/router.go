// path: backend/cmd/api/router.go
package main

import (
	"encoding/json"
	"fmt"
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

				// Team member management
				r.Post("/{id}/members", container.TeamHandler.InviteMember)                    // Invite member
				r.Delete("/{id}/members/{userId}", container.TeamHandler.RemoveMember)         // Remove member
				r.Patch("/{id}/members/{userId}/role", container.TeamHandler.UpdateMemberRole) // Update role
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

// handleHealth performs health check including database ping
func handleHealth(c *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Test database connection
		dbStatus := "ok"
		if err := c.DB.PingContext(r.Context()); err != nil {
			dbStatus = "error"
		}

		response := map[string]interface{}{
			"status":      "ok",
			"database":    dbStatus,
			"service":     "socialqueue-api",
			"environment": c.Config.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		}

		if dbStatus == "error" {
			respondJSON(w, http.StatusServiceUnavailable, response)
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// handleMe returns current authenticated user information
func handleMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := appMiddleware.GetUserID(r.Context())
	email, _ := appMiddleware.GetUserEmail(r.Context())
	role, _ := appMiddleware.GetUserRole(r.Context())

	response := map[string]interface{}{
		"userId": userID.String(),
		"email":  email,
		"role":   role,
	}
	respondJSON(w, http.StatusOK, response)
}

// handleAdminUsers placeholder for admin users endpoint
func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"users": []interface{}{},
		"total": 0,
	}
	respondJSON(w, http.StatusOK, response)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}
