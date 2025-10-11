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

	// Legacy imports (keep for now)
	"github.com/techappsUT/social-queue/internal/auth"
	"github.com/techappsUT/social-queue/internal/handlers"
	appMiddleware "github.com/techappsUT/social-queue/internal/middleware"
	"github.com/techappsUT/social-queue/internal/social"
	"github.com/techappsUT/social-queue/internal/social/adapters"
	"github.com/techappsUT/social-queue/pkg/email"

	// Clean Architecture imports
	appAuth "github.com/techappsUT/social-queue/internal/application/auth"
	appUser "github.com/techappsUT/social-queue/internal/application/user"
	userDomain "github.com/techappsUT/social-queue/internal/domain/user"
	"github.com/techappsUT/social-queue/internal/infrastructure/persistence"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
)

// ============================================================================
// APPLICATION CONTAINER
// ============================================================================

type App struct {
	Config *Config

	// Databases
	DB    *gorm.DB
	SqlDB *sql.DB

	// Legacy Services
	AuthService   *auth.Service
	TokenService  *auth.TokenService
	EmailService  email.Service
	SocialService *social.Service

	// Legacy Handlers
	AuthHandler   *handlers.AuthHandler
	SocialHandler *handlers.SocialHandler

	// Clean Architecture Components
	CreateUserUC  *appUser.CreateUserUseCase
	LoginUC       *appAuth.LoginUseCase
	AuthHandlerV2 *handlers.AuthHandlerV2

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

	// Step 3: Initialize legacy services (keep for backward compatibility)
	log.Println("⚙️  Initializing legacy services...")
	app.initializeServices()
	log.Println("  ✓ Legacy services initialized")

	// Step 4: Initialize Clean Architecture (NEW)
	log.Println("🏗️  Initializing Clean Architecture...")
	if err := app.initializeCleanArchitecture(); err != nil {
		log.Printf("  ⚠️  Clean architecture init failed: %v", err)
		// Don't fail - legacy still works
	} else {
		log.Println("  ✓ Clean Architecture initialized")
	}

	// Step 5: Initialize handlers
	log.Println("🎯 Initializing handlers...")
	app.initializeHandlers()
	log.Println("  ✓ Handlers initialized")

	// Step 6: Setup HTTP router
	log.Println("🌐 Setting up HTTP router...")
	app.setupRouter()
	log.Println("  ✓ Router configured")

	// Step 7: Create HTTP server
	app.setupServer()

	log.Println("✅ Application initialized successfully")
	return app, nil
}

// ============================================================================
// CLEAN ARCHITECTURE INITIALIZATION (NEW)
// ============================================================================

func (app *App) initializeCleanArchitecture() error {
	// Infrastructure Layer
	userRepo := persistence.NewUserRepository(app.SqlDB)
	userService := userDomain.NewService(userRepo)

	// Infrastructure Services
	tokenService := services.NewJWTTokenService(
		app.Config.JWT.AccessSecret,
		app.Config.JWT.RefreshSecret,
	)

	emailConfig := EmailConfig{
		Provider:    app.Config.Email.Provider,
		APIKey:      app.Config.Email.APIKey,
		FromAddress: app.Config.Email.FromAddress,
		FromName:    app.Config.Email.FromName,
	}
	emailService := services.NewEmailService(emailConfig)
	cacheService := services.NewInMemoryCacheService()
	logger := services.NewLogger()

	// Application Layer (Use Cases)
	app.CreateUserUC = appUser.NewCreateUserUseCase(
		userRepo,
		userService,
		tokenService,
		emailService,
		logger,
	)

	app.LoginUC = appAuth.NewLoginUseCase(
		userRepo,
		userService,
		tokenService,
		cacheService,
		logger,
	)

	// Presentation Layer (Handlers)
	app.AuthHandlerV2 = handlers.NewAuthHandlerV2(
		app.CreateUserUC,
		app.LoginUC,
	)

	return nil
}

// ============================================================================
// DATABASE SETUP
// ============================================================================

func (app *App) setupDatabases() error {
	// GORM Connection (Legacy)
	gormDB, err := app.setupGORM()
	if err != nil {
		return fmt.Errorf("GORM setup failed: %w", err)
	}
	app.DB = gormDB

	// Standard SQL Connection (for SQLC and Clean Architecture)
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

	gormLogger := logger.Default
	if app.Config.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		CreateBatchSize:                          100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return db, nil
}

// ============================================================================
// LEGACY SERVICE INITIALIZATION (Keep for backward compatibility)
// ============================================================================

func (app *App) initializeServices() {
	// Auth Services (Legacy - GORM based)
	app.TokenService = auth.NewTokenService(
		app.Config.JWT.AccessSecret,
		app.Config.JWT.RefreshSecret,
		"socialqueue",
	)

	app.EmailService = email.NewMockEmailService(app.Config.BaseURL)
	app.AuthService = auth.NewService(app.DB, app.TokenService, app.EmailService)

	// Social Services
	if app.Config.Security.EncryptionKey != "" {
		encryption, err := social.NewTokenEncryption(app.Config.Security.EncryptionKey)
		if err != nil {
			log.Printf("⚠️  Failed to initialize encryption: %v", err)
			return
		}

		registry := app.setupSocialAdapters()
		limiter := social.NewRateLimiter()

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
	// Legacy handlers
	app.AuthHandler = handlers.NewAuthHandler(app.AuthService)
	app.SocialHandler = handlers.NewSocialHandler(app.SocialService)
	app.AuthMiddleware = appMiddleware.NewAuthMiddleware(app.TokenService)

	// Clean Architecture handlers are initialized in initializeCleanArchitecture()
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

	// API routes - Legacy (keep working)
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			app.AuthHandler.RegisterRoutes(r)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(app.AuthMiddleware.RequireAuth)

			r.Get("/me", app.handleMe)

			// Social routes
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

	// API V2 routes - Clean Architecture (NEW)
	r.Route("/api/v2", func(r chi.Router) {
		if app.AuthHandlerV2 != nil {
			app.AuthHandlerV2.RegisterRoutes(r)
		}

		// Add more V2 routes as you create more use cases
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users":[]}`))
}

// ============================================================================
// SERVER LIFECYCLE
// ============================================================================

func (app *App) setupServer() {
	app.Server = &http.Server{
		Addr:         ":" + app.Config.Server.Port,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (app *App) Start() {
	// Start server in goroutine
	go func() {
		log.Printf("📡 Server starting on port %s", app.Config.Server.Port)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server gracefully stopped")
}

func (app *App) Cleanup() {
	if app.DB != nil {
		if sqlDB, err := app.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	if app.SqlDB != nil {
		app.SqlDB.Close()
	}

	log.Println("🧹 Cleanup completed")
}
