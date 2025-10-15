// path: backend/cmd/api/router.go
// ✅ FIXED VERSION
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

// SetupRouter creates and configures the HTTP router ✅ Fixed: Capitalized
func SetupRouter(container *Container) *chi.Mux {
	r := chi.NewRouter()

	// ============================================================================
	// GLOBAL MIDDLEWARE
	// ============================================================================
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Configuration ✅ Fixed: Use container.Config.CORS.AllowedOrigins
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   container.Config.CORS.AllowedOrigins,
		AllowedMethods:   container.Config.CORS.AllowedMethods,
		AllowedHeaders:   container.Config.CORS.AllowedHeaders,
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
		// Auth routes (public: signup, login, etc.)
		routes.RegisterAuthRoutes(r, container.AuthHandler, container.AuthMiddleware)

		// User routes (protected) ✅ Fixed: Use AuthHandler, not UserHandler
		routes.RegisterUserRoutes(r, container.AuthHandler, container.AuthMiddleware)

		// Team routes (protected) ✅ Fixed: Add authMiddleware parameter
		if container.TeamHandler != nil {
			routes.RegisterTeamRoutes(r, container.TeamHandler, container.AuthMiddleware)
		}

		// Post routes (protected) ✅ Fixed: Add authMiddleware parameter
		if container.PostHandler != nil {
			routes.RegisterPostRoutes(r, container.PostHandler, container.AuthMiddleware)
		}

		// Social routes (protected) ✅ Fixed: Add authMiddleware parameter
		if container.SocialHandler != nil {
			routes.RegisterSocialRoutes(r, container.SocialHandler, container.AuthMiddleware)
		}
	})

	return r
}

// ============================================================================
// ROOT HANDLERS
// ============================================================================

func handleRoot(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message":  "SocialQueue API v2",
			"status":   "operational",
			"docs_url": "/api/v2/docs",
		}
		respondSuccess(w, response) // ✅ Fixed: Use lowercase helper
	}
}

func handleHealth(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database
		var dbHealthy bool
		var dbError string
		if err := container.DB.Ping(); err != nil {
			dbError = err.Error()
		} else {
			dbHealthy = true
		}

		// Check Redis
		var redisHealthy bool
		var redisError string
		if container.Redis != nil {
			if err := container.Redis.Ping(r.Context()).Err(); err != nil {
				redisError = err.Error()
			} else {
				redisHealthy = true
			}
		}

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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}

func handleVersion(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"version":    "2.0.0",
			"build_date": "2025-01-15",
			"go_version": "1.21+",
			"env":        container.Config.Environment,
		}
		respondSuccess(w, response) // ✅ Fixed: Use lowercase helper
	}
}

// ============================================================================
// RESPONSE HELPERS
// ============================================================================

func respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// func respondError(w http.ResponseWriter, status int, message string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"success": false,
// 		"error":   message,
// 	})
// }
