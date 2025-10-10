// path: backend/cmd/api/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/techappsUT/social-queue/internal/auth"
	"github.com/techappsUT/social-queue/internal/handlers"
	appMiddleware "github.com/techappsUT/social-queue/internal/middleware"
	"github.com/techappsUT/social-queue/internal/social"
	"github.com/techappsUT/social-queue/internal/social/adapters"
	"github.com/techappsUT/social-queue/pkg/email"
)

// ============================================================================
// APPLICATION CONTAINER
// ============================================================================

type App struct {
	Config *Config

	// Databases
	DB    *gorm.DB
	SqlDB *sql.DB

	// Services
	AuthService   *auth.Service
	TokenService  *auth.TokenService
	EmailService  email.Service
	SocialService *social.Service

	// Handlers
	AuthHandler   *handlers.AuthHandler
	SocialHandler *handlers.SocialHandler

	// Middleware
	AuthMiddleware *appMiddleware.AuthMiddleware

	// HTTP
	Router *chi.Mux
	Server *http.Server
}

// ============================================================================
// MAIN ENTRY POINT
// ============================================================================

func main() {
	log.Println("🚀 Starting SocialQueue API Server...")

	// Initialize application
	app, err := NewApp()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application: %v", err)
	}
	defer app.Cleanup()

	// Start server with graceful shutdown
	app.Start()
}

// ============================================================================
// APPLICATION INITIALIZATION
// ============================================================================

func NewApp() (*App, error) {
	app := &App{}

	// Step 1: Load configuration
	log.Println("📝 Loading configuration...")
	app.Config = LoadConfig()

	// Step 2: Setup databases
	log.Println("🗄️  Connecting to databases...")
	if err := app.setupDatabases(); err != nil {
		return nil, fmt.Errorf("database setup failed: %w", err)
	}
	log.Println("  ✓ Databases connected")

	// Step 3: Initialize services
	log.Println("⚙️  Initializing services...")
	app.initializeServices()
	log.Println("  ✓ Services initialized")

	// Step 4: Initialize handlers
	log.Println("🎯 Initializing handlers...")
	app.initializeHandlers()
	log.Println("  ✓ Handlers initialized")

	// Step 5: Setup HTTP router
	log.Println("🌐 Setting up HTTP router...")
	app.setupRouter()
	log.Println("  ✓ Router configured")

	// Step 6: Create HTTP server
	app.setupServer()

	log.Println("✅ Application initialized successfully")
	return app, nil
}

// ============================================================================
// DATABASE SETUP
// ============================================================================

func (app *App) setupDatabases() error {
	// GORM Connection (⚠️ LEGACY - for auth)
	gormDB, err := app.setupGORM()
	if err != nil {
		return fmt.Errorf("GORM setup failed: %w", err)
	}
	app.DB = gormDB

	// Standard SQL Connection (for future SQLC)
	sqlDB, err := app.setupSQL()
	if err != nil {
		return fmt.Errorf("SQL setup failed: %w", err)
	}
	app.SqlDB = sqlDB

	return nil
}

func (app *App) setupGORM() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		app.Config.Database.Host,
		app.Config.Database.User,
		app.Config.Database.Password,
		app.Config.Database.DBName,
		app.Config.Database.Port,
		app.Config.Database.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if app.Config.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		// 🔧 FIX: Disable auto-migration features that conflict with SQL migrations
		DisableForeignKeyConstraintWhenMigrating: true,
		// Don't create database if it doesn't exist
		CreateBatchSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// ❌ REMOVED: Auto-migrate - we use SQL migrations instead
	// The schema is already managed by golang-migrate
	// GORM will just use the existing tables

	log.Println("  ℹ️  Using existing database schema (managed by migrations)")

	return db, nil
}

func (app *App) setupSQL() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		app.Config.Database.Host,
		app.Config.Database.Port,
		app.Config.Database.User,
		app.Config.Database.Password,
		app.Config.Database.DBName,
		app.Config.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}

// ============================================================================
// SERVICE INITIALIZATION
// ============================================================================

func (app *App) initializeServices() {
	// Auth Services (⚠️ LEGACY - GORM based)
	app.TokenService = auth.NewTokenService(
		app.Config.JWT.AccessSecret,
		app.Config.JWT.RefreshSecret,
		"socialqueue",
	)

	app.EmailService = email.NewMockEmailService(app.Config.BaseURL)
	app.AuthService = auth.NewService(app.DB, app.TokenService, app.EmailService)

	// Social Services (🆕 NEW - for Clean Architecture)
	if app.Config.Security.EncryptionKey != "" {
		encryption, err := social.NewTokenEncryption(app.Config.Security.EncryptionKey)
		if err != nil {
			log.Printf("⚠️  Failed to initialize encryption: %v", err)
			log.Println("   Social media features will be disabled")
			return
		}

		registry := app.setupSocialAdapters()
		limiter := social.NewRateLimiter()

		// TODO: Replace nil with SQLC queries in Phase 2
		var queries social.DBQueries = nil
		app.SocialService = social.NewService(registry, queries, encryption, limiter)

		if len(registry.ListPlatforms()) > 0 {
			log.Printf("  ✓ Social service initialized with %d platforms", len(registry.ListPlatforms()))
		} else {
			log.Println("  ⚠️  Social service initialized but no platforms configured")
		}
	} else {
		log.Println("  ⚠️  ENCRYPTION_KEY not set - social features disabled")
	}
}

func (app *App) setupSocialAdapters() *social.AdapterRegistry {
	registry := social.NewAdapterRegistry()
	registered := 0

	// Twitter
	if app.Config.Social.Twitter.ClientID != "" && app.Config.Social.Twitter.ClientSecret != "" {
		adapter := adapters.NewTwitterAdapter(
			app.Config.Social.Twitter.ClientID,
			app.Config.Social.Twitter.ClientSecret,
		)
		if err := registry.Register(adapter); err != nil {
			log.Printf("  ⚠️  Twitter registration failed: %v", err)
		} else {
			log.Println("    ✓ Twitter")
			registered++
		}
	}

	// Facebook
	if app.Config.Social.Facebook.ClientID != "" && app.Config.Social.Facebook.ClientSecret != "" {
		adapter := adapters.NewFacebookAdapter(
			app.Config.Social.Facebook.ClientID,
			app.Config.Social.Facebook.ClientSecret,
		)
		if err := registry.Register(adapter); err != nil {
			log.Printf("  ⚠️  Facebook registration failed: %v", err)
		} else {
			log.Println("    ✓ Facebook")
			registered++
		}
	}

	// LinkedIn
	if app.Config.Social.LinkedIn.ClientID != "" && app.Config.Social.LinkedIn.ClientSecret != "" {
		adapter := adapters.NewLinkedInAdapter(
			app.Config.Social.LinkedIn.ClientID,
			app.Config.Social.LinkedIn.ClientSecret,
		)
		if err := registry.Register(adapter); err != nil {
			log.Printf("  ⚠️  LinkedIn registration failed: %v", err)
		} else {
			log.Println("    ✓ LinkedIn")
			registered++
		}
	}

	if registered == 0 {
		log.Println("  ⚠️  No social platforms configured")
	}

	return registry
}

// ============================================================================
// HANDLER INITIALIZATION
// ============================================================================

func (app *App) initializeHandlers() {
	app.AuthHandler = handlers.NewAuthHandler(app.AuthService)
	app.SocialHandler = handlers.NewSocialHandler(app.SocialService)
	app.AuthMiddleware = appMiddleware.NewAuthMiddleware(app.TokenService)
}

// ============================================================================
// HTTP ROUTER SETUP
// ============================================================================

func (app *App) setupRouter() {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.Config.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", app.handleHealth)
	r.Get("/", app.handleRoot)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			app.AuthHandler.RegisterRoutes(r)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(app.AuthMiddleware.RequireAuth)

			r.Get("/me", app.handleMe)

			// Social routes (if service available)
			if app.SocialService != nil {
				r.Route("/social", func(r chi.Router) {
					r.Get("/auth/{platform}/redirect", app.SocialHandler.InitiateOAuth)
					r.Get("/auth/{platform}/callback", app.SocialHandler.OAuthCallback)
					r.Post("/publish", app.SocialHandler.PublishContent)
				})
			}

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(appMiddleware.RequireAdmin)
				r.Get("/admin/users", app.handleAdminUsers)
			})
		})
	})

	app.Router = r
}

// ============================================================================
// HTTP HANDLERS
// ============================================================================

func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"socialqueue-api","environment":"%s"}`, app.Config.Environment)
}

func (app *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"SocialQueue API","version":"1.0.0","docs":"/api/docs"}`))
}

func (app *App) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := appMiddleware.GetUserID(r.Context())
	email, _ := appMiddleware.GetUserEmail(r.Context())
	role, _ := appMiddleware.GetUserRole(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"userId":"%s","email":"%s","role":"%s"}`, userID, email, role)
}

func (app *App) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Admin users endpoint - TODO","users":[]}`))
}

// ============================================================================
// SERVER SETUP & LIFECYCLE
// ============================================================================

func (app *App) setupServer() {
	app.Server = &http.Server{
		Addr:           fmt.Sprintf(":%s", app.Config.Server.Port),
		Handler:        app.Router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

func (app *App) Start() {
	// Graceful shutdown setup
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("\n🛑 Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.Server.Shutdown(ctx); err != nil {
			log.Printf("❌ Server forced to shutdown: %v", err)
		}

		close(done)
	}()

	// Print startup info
	app.printStartupInfo()

	// Start server
	log.Printf("🚀 Server listening on http://localhost:%s", app.Config.Server.Port)
	if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server failed to start: %v", err)
	}

	<-done
	log.Println("✅ Server stopped gracefully")
}

func (app *App) printStartupInfo() {
	baseURL := fmt.Sprintf("http://localhost:%s", app.Config.Server.Port)

	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("  Environment: %s", app.Config.Environment)
	log.Printf("  Base URL: %s", baseURL)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("  📍 Endpoints:")
	log.Printf("    • Health:  %s/health", baseURL)
	log.Printf("    • Auth:    %s/api/auth/*", baseURL)
	log.Printf("    • User:    %s/api/me", baseURL)

	if app.SocialService != nil {
		log.Printf("    • Social:  %s/api/social/*", baseURL)
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("")
}

func (app *App) Cleanup() {
	log.Println("🧹 Cleaning up resources...")

	if app.SqlDB != nil {
		app.SqlDB.Close()
	}

	log.Println("✅ Cleanup complete")
}
