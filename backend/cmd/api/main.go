// path: backend/cmd/api/main.go

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/techappsUT/social-queue/internal/auth"
	"github.com/techappsUT/social-queue/internal/handlers"
	appMiddleware "github.com/techappsUT/social-queue/internal/middleware"
	"github.com/techappsUT/social-queue/internal/models"
	"github.com/techappsUT/social-queue/pkg/email"
)

func main() {
	// Load environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "socialqueue")
	dbPassword := getEnv("DB_PASSWORD", "socialqueue_dev_password")
	dbName := getEnv("DB_NAME", "socialqueue_dev")

	accessSecret := getEnv("JWT_ACCESS_SECRET", "your-super-secret-access-key-change-in-production")
	refreshSecret := getEnv("JWT_REFRESH_SECRET", "your-super-secret-refresh-key-change-in-production")
	baseURL := getEnv("BASE_URL", "http://localhost:3000")

	// Database connection
	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPassword +
		" dbname=" + dbName + " port=" + dbPort + " sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.User{}, &models.Team{}, &models.RefreshToken{})

	// Initialize services
	tokenService := auth.NewTokenService(accessSecret, refreshSecret, "bufferclone")
	emailService := email.NewMockEmailService(baseURL)
	authService := auth.NewService(db, tokenService, emailService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := appMiddleware.NewAuthMiddleware(tokenService)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Group(func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	// Protected routes (requires authentication)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth)

		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			userID, _ := appMiddleware.GetUserID(r.Context())
			email, _ := appMiddleware.GetUserEmail(r.Context())
			role, _ := appMiddleware.GetUserRole(r.Context())

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"user_id":"` + userID.String() + `","email":"` + email + `","role":"` + role + `"}`))
		})

		// Admin only routes
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.RequireAdmin)

			r.Get("/api/admin/users", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Admin users list"))
			})
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK updated"))
	})

	// Start server
	// port := getEnv("PORT", "8080")
	port := "8000"
	log.Printf("🚀 Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
