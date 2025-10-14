// ============================================================================
// FILE: backend/cmd/api/router.go
// ✅ COMPLETE FIXED VERSION - Works with updated container.go
// ============================================================================
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/techappsUT/social-queue/internal/handlers/routes"
)

// setupRouter creates and configures the HTTP router with organized route groups
func setupRouter(container *Container) *chi.Mux {
	r := chi.NewRouter()

	// ============================================================================
	// GLOBAL MIDDLEWARE
	// ============================================================================
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
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
	r.Get("/version", handleVersion(container))

	// ============================================================================
	// API V2 ROUTES (Clean Architecture)
	// ============================================================================
	r.Route("/api/v2", func(r chi.Router) {
		// ✅ This now works because container.AuthHandler is *handlers.AuthHandler
		routes.RegisterAuthRoutes(r, container.AuthHandler, container.AuthMiddleware)
		routes.RegisterUserRoutes(r, container.AuthHandler, container.AuthMiddleware)
		routes.RegisterTeamRoutes(r, container.TeamHandler, container.AuthMiddleware)
		routes.RegisterPostRoutes(r, container.PostHandler, container.AuthMiddleware)

		// Social routes (only if social features enabled)
		if container.SocialHandler != nil {
			routes.RegisterSocialRoutes(r, container.SocialHandler, container.AuthMiddleware)
		}

		// TODO: Add when implemented
		// routes.RegisterAnalyticsRoutes(r, container.AnalyticsHandler, container.AuthMiddleware)
		// routes.RegisterBillingRoutes(r, container.BillingHandler, container.AuthMiddleware)
		// routes.RegisterWebhookRoutes(r, container.WebhookHandler)
	})

	return r
}

// ============================================================================
// ROOT HANDLERS
// ============================================================================

func handleRoot(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		socialEnabled := container.SocialHandler != nil

		response := map[string]interface{}{
			"name":    "SocialQueue API",
			"version": "2.0.0",
			"status":  "running",
			"docs":    "/api/v2/docs",
			"features": map[string]bool{
				"authentication": true,
				"teams":          true,
				"posts":          true,
				"social_oauth":   socialEnabled,
				"analytics":      false, // TODO
				"billing":        false, // TODO
			},
			"endpoints": map[string]string{
				"health":    "/health",
				"version":   "/version",
				"api":       "/api/v2",
				"auth":      "/api/v2/auth",
				"users":     "/api/v2/users",
				"teams":     "/api/v2/teams",
				"posts":     "/api/v2/posts",
				"social":    "/api/v2/social",
				"analytics": "/api/v2/analytics",
				"billing":   "/api/v2/billing",
			},
		}
		respondJSON(w, http.StatusOK, response)
	}
}

func handleHealth(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		dbHealthy := true
		dbError := ""
		if err := container.DB.Ping(); err != nil {
			dbHealthy = false
			dbError = err.Error()
		}

		// Check Redis connection (if available)
		redisHealthy := true
		redisError := ""
		if container.Redis != nil {
			if err := container.Redis.Ping(r.Context()).Err(); err != nil {
				redisHealthy = false
				redisError = err.Error()
			}
		}

		// Determine overall status
		status := "healthy"
		statusCode := http.StatusOK
		if !dbHealthy || !redisHealthy {
			status = "degraded"
			if !dbHealthy {
				status = "unhealthy"
				statusCode = http.StatusServiceUnavailable
			}
		}

		response := map[string]interface{}{
			"status":    status,
			"timestamp": time.Now().UTC(),
			"services": map[string]interface{}{
				"database": map[string]interface{}{
					"status": dbHealthy,
					"error":  dbError,
				},
				"redis": map[string]interface{}{
					"status": redisHealthy,
					"error":  redisError,
				},
			},
			"features": map[string]interface{}{
				"social_oauth": map[string]interface{}{
					"enabled":  container.SocialHandler != nil,
					"adapters": len(container.SocialAdapters),
				},
			},
		}

		respondJSON(w, statusCode, response)
	}
}

func handleVersion(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"version":    "2.0.0",
			"build_date": "2025-01-13", // TODO: Inject from build
			"go_version": "1.21+",
			"env":        container.Config.Environment,
		}
		respondJSON(w, http.StatusOK, response)
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
