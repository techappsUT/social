// path: backend/cmd/server/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/techappsUT/social-queue/internal/api/handlers"
	"github.com/techappsUT/social-queue/internal/config"
	"github.com/techappsUT/social-queue/internal/social"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting Social Queue application...")

	// Load configuration
	cfg := config.Load()

	// Validate social configuration
	if err := validateSocialConfig(cfg); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connection established")

	// Initialize encryption
	if cfg.Security.EncryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY environment variable is required and must be 32 bytes")
	}

	encryption, err := social.NewTokenEncryption(cfg.Security.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}
	log.Println("✓ Token encryption initialized")

	// Setup social adapters
	log.Println("Registering social platform adapters...")
	registry := setupSocialAdapters(cfg)

	// Initialize rate limiter
	limiter := social.NewRateLimiter()
	log.Println("✓ Rate limiter initialized")

	// Create social service
	// Note: You'll need to implement DBQueries with your SQLC generated code
	// Example:
	// queries := sqlcgen.New(db)
	// For now, we'll pass nil and you'll replace it with actual implementation
	var queries social.DBQueries = nil // TODO: Replace with actual SQLC queries

	if queries == nil {
		log.Println("⚠ WARNING: Database queries not initialized. You need to integrate SQLC generated code.")
	}

	socialService := social.NewService(registry, queries, encryption, limiter)
	log.Println("✓ Social service initialized")

	// Setup HTTP router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Setup handlers
	socialAuthHandler := handlers.NewSocialAuthHandler(socialService)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API Routes
	r.Route("/api/social", func(r chi.Router) {
		// OAuth routes
		r.Get("/auth/{platform}/redirect", socialAuthHandler.InitiateOAuth)
		r.Get("/auth/{platform}/callback", socialAuthHandler.OAuthCallback)

		// Posting routes
		r.Post("/publish", socialAuthHandler.PublishContent)

		// TODO: Add more routes
		// r.Get("/accounts", socialAuthHandler.ListAccounts)
		// r.Delete("/accounts/{id}", socialAuthHandler.DisconnectAccount)
	})

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on http://%s", addr)
	log.Printf("📝 API Documentation: http://%s/api/social/auth/{platform}/redirect", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
