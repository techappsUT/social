// ============================================================================
// FILE: backend/cmd/api/router.go
// COMPLETE VERSION - Now includes Rate Limiting Middleware
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
	// GLOBAL MIDDLEWARE (Applied to all routes)
	// ============================================================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Structured Logging
	if container.Logger != nil {
		r.Use(appMiddleware.RequestLogger(container.Logger))
	}

	// CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   container.Config.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Global IP-based Rate Limiting (1000 requests per minute per IP)
	if container.RateLimiter != nil {
		r.Use(container.RateLimiter.RateLimitByIP(appMiddleware.DefaultRateLimitConfigs["ip"]))
	}

	// ============================================================================
	// PUBLIC ROUTES
	// ============================================================================

	r.Get("/", handleRoot(container))
	r.Get("/health", handleHealth(container))

	// ============================================================================
	// API V2 ROUTES (Clean Architecture)
	// ============================================================================

	r.Route("/api/v2", func(r chi.Router) {
		// ====================================================================
		// PUBLIC AUTHENTICATION ROUTES
		// ====================================================================
		r.Group(func(r chi.Router) {
			// Stricter rate limiting for auth endpoints (10 requests per minute)
			if container.RateLimiter != nil {
				r.Use(container.RateLimiter.RateLimitByIP(appMiddleware.DefaultRateLimitConfigs["auth"]))
			}

			r.Post("/auth/signup", container.AuthHandler.Signup)
			r.Post("/auth/login", container.AuthHandler.Login)
		})

		// ====================================================================
		// PROTECTED ROUTES (require authentication)
		// ====================================================================
		r.Group(func(r chi.Router) {
			r.Use(container.AuthMiddleware.RequireAuth)

			// User-based rate limiting (100 requests per minute per user)
			if container.RateLimiter != nil {
				r.Use(container.RateLimiter.RateLimitByUser(appMiddleware.DefaultRateLimitConfigs["user"]))
			}

			// Current user route
			r.Get("/me", handleMe)

			// ================================================================
			// USER MANAGEMENT ROUTES
			// ================================================================
			r.Get("/users/{id}", container.AuthHandler.GetUser)
			r.Put("/users/{id}", container.AuthHandler.UpdateUser)
			r.Delete("/users/{id}", container.AuthHandler.DeleteUser)

			// ================================================================
			// TEAM ROUTES
			// ================================================================
			r.Route("/teams", func(r chi.Router) {
				// Team CRUD
				r.Get("/", container.TeamHandler.ListTeams)
				r.Post("/", container.TeamHandler.CreateTeam)
				r.Get("/{id}", container.TeamHandler.GetTeam)
				r.Put("/{id}", container.TeamHandler.UpdateTeam)
				r.Delete("/{id}", container.TeamHandler.DeleteTeam)

				// Team member management
				r.Post("/{id}/members", container.TeamHandler.InviteMember)
				r.Delete("/{id}/members/{userId}", container.TeamHandler.RemoveMember)
				r.Patch("/{id}/members/{userId}/role", container.TeamHandler.UpdateMemberRole)

				// Team's posts
				r.Get("/{teamId}/posts", container.PostHandler.ListPosts)

				// Team's social accounts
				r.Get("/{teamId}/social/accounts", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.ListAccounts },
					handleSocialNotAvailable,
				))
			})

			// ================================================================
			// POST ROUTES
			// ================================================================
			r.Route("/posts", func(r chi.Router) {
				// Post CRUD
				r.Post("/", container.PostHandler.CreateDraft)
				r.Get("/{id}", container.PostHandler.GetPost)
				r.Put("/{id}", container.PostHandler.UpdatePost)
				r.Delete("/{id}", container.PostHandler.DeletePost)

				// Post actions
				r.Post("/{id}/schedule", container.PostHandler.SchedulePost)
				r.Post("/{id}/publish", container.PostHandler.PublishNow)
			})

			// ================================================================
			// SOCIAL OAUTH & ACCOUNT MANAGEMENT ROUTES
			// ================================================================
			r.Route("/social", func(r chi.Router) {
				// OAuth flow
				r.Get("/auth/{platform}", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.GetOAuthURL },
					handleSocialNotAvailable,
				))
				r.Get("/auth/{platform}/callback", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.OAuthCallback },
					handleSocialNotAvailable,
				))

				// Account management
				r.Post("/accounts", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.ConnectAccount },
					handleSocialNotAvailable,
				))
				r.Get("/accounts", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.ListAccounts },
					handleSocialNotAvailable,
				))
				r.Delete("/accounts/{id}", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.DisconnectAccount },
					handleSocialNotAvailable,
				))

				// Token management
				r.Post("/accounts/{id}/refresh", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.RefreshTokens },
					handleSocialNotAvailable,
				))

				// Analytics
				r.Get("/accounts/{id}/analytics", conditionalHandler(
					container.SocialHandler,
					func() http.HandlerFunc { return container.SocialHandler.GetAnalytics },
					handleSocialNotAvailable,
				))
			})
		})
	})

	return r
}

// ============================================================================
// HELPER HANDLERS
// ============================================================================

// conditionalHandler returns a handler or fallback if handler is nil
func conditionalHandler(handler interface{}, getHandler func() http.HandlerFunc, fallback http.HandlerFunc) http.HandlerFunc {
	if handler == nil {
		return fallback
	}
	return getHandler()
}

// handleSocialNotAvailable returns error when social features are not configured
func handleSocialNotAvailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   "Social features not available",
		"message": "Social OAuth features are not configured. Please set ENCRYPTION_KEY and OAuth credentials in environment variables.",
	})
}

// handleRoot returns API information
func handleRoot(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		socialEnabled := container.SocialHandler != nil

		response := map[string]interface{}{
			"name":    "SocialQueue API",
			"version": "2.0.0",
			"status":  "running",
			"features": map[string]bool{
				"authentication": true,
				"teams":          true,
				"posts":          true,
				"social_oauth":   socialEnabled,
				"rate_limiting":  container.RateLimiter != nil,
			},
			"rate_limits": map[string]interface{}{
				"ip":   "1000 requests per minute",
				"user": "100 requests per minute",
				"auth": "10 requests per minute",
			},
			"endpoints": map[string]string{
				"health": "/health",
				"api":    "/api/v2",
				"auth":   "/api/v2/auth",
				"teams":  "/api/v2/teams",
				"posts":  "/api/v2/posts",
				"social": "/api/v2/social",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleHealth returns health check status
func handleHealth(container *Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		dbHealthy := true
		dbError := ""
		if err := container.DB.Ping(); err != nil {
			dbHealthy = false
			dbError = err.Error()
		}

		// Check Redis connection
		redisHealthy := false
		redisError := ""
		if container.Redis != nil {
			if err := container.Redis.Ping(r.Context()).Err(); err != nil {
				redisError = err.Error()
			} else {
				redisHealthy = true
			}
		}

		// Check social features
		socialEnabled := container.SocialHandler != nil
		socialAdapterCount := len(container.SocialAdapters)

		status := "healthy"
		statusCode := http.StatusOK
		if !dbHealthy {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		response := map[string]interface{}{
			"status": status,
			"time":   time.Now().UTC(),
			"database": map[string]interface{}{
				"status": dbHealthy,
				"error":  dbError,
			},
			"redis": map[string]interface{}{
				"status": redisHealthy,
				"error":  redisError,
			},
			"features": map[string]interface{}{
				"social_oauth": map[string]interface{}{
					"enabled":  socialEnabled,
					"adapters": socialAdapterCount,
				},
				"rate_limiting": map[string]interface{}{
					"enabled": container.RateLimiter != nil,
					"backend": "redis",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}

// handleMe returns current authenticated user info
func handleMe(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userID, ok := appMiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	email, _ := appMiddleware.GetUserEmail(r.Context())
	role, _ := appMiddleware.GetUserRole(r.Context())

	response := map[string]interface{}{
		"userId": userID.String(),
		"email":  email,
		"role":   role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
