// path: backend/cmd/api/main.go
// ✅ PROFESSIONAL FIXED VERSION - With .env loading
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

	"github.com/joho/godotenv" // ✅ CRITICAL: Load .env file
	_ "github.com/lib/pq"
)

// ============================================================================
// APPLICATION STRUCTURE
// ============================================================================

type App struct {
	Container *Container
	Server    *http.Server
}

// ============================================================================
// MAIN ENTRY POINT
// ============================================================================

func main() {
	log.Println("🚀 Starting SocialQueue API Server...")

	// ✅ CRITICAL: Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No .env file found, using environment variables")
	} else {
		log.Println("  ✓ Loaded .env file")
	}

	// Initialize application
	app, err := initializeApp()
	if err != nil {
		log.Fatalf("❌ Initialization failed: %v", err)
	}

	// Run with graceful shutdown
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// ============================================================================
// APPLICATION INITIALIZATION
// ============================================================================

func initializeApp() (*App, error) {
	// Step 1: Load configuration
	log.Println("⚙️  Loading configuration...")
	config := LoadConfig()
	logConfiguration(config)

	// Step 2: Setup database
	log.Println("🗄️  Connecting to database...")
	db, err := setupDatabase(config)
	if err != nil {
		return nil, fmt.Errorf("database setup failed: %w", err)
	}

	// Step 3: Initialize dependency injection container
	log.Println("🔧 Initializing dependencies...")
	container, err := NewContainer(config, db)
	if err != nil {
		return nil, fmt.Errorf("container initialization failed: %w", err)
	}
	log.Println("  ✓ Dependencies initialized")

	// Step 4: Setup HTTP router
	log.Println("🛣️  Setting up router...")
	router := SetupRouter(container)
	log.Println("  ✓ Router configured")

	// Step 5: Create HTTP server
	serverAddr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &App{
		Container: container,
		Server:    server,
	}, nil
}

// ============================================================================
// DATABASE SETUP
// ============================================================================

func setupDatabase(config *Config) (*sql.DB, error) {
	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	log.Printf("  Connecting to: %s@%s:%s/%s",
		config.Database.User,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
	)

	// Open connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	log.Println("  ✓ Database connection established")
	return db, nil
}

// ============================================================================
// SERVER LIFECYCLE
// ============================================================================

func (app *App) Run() error {
	// Channel to listen for errors
	serverErrors := make(chan error, 1)

	// Start HTTP server in goroutine
	go func() {
		log.Printf("🌐 Server listening on http://%s", app.Server.Addr)
		log.Printf("✨ Environment: %s", app.Container.Config.Environment)
		log.Println("")
		logAvailableEndpoints()
		log.Println("")
		serverErrors <- app.Server.ListenAndServe()
	}()

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive signal or error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("\n🛑 Received %v signal, starting graceful shutdown...", sig)

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server
		if err := app.Server.Shutdown(ctx); err != nil {
			log.Printf("❌ Graceful shutdown error: %v", err)
			return app.Server.Close()
		}

		log.Println("✅ Server stopped gracefully")
		return nil
	}
}

// ============================================================================
// LOGGING HELPERS
// ============================================================================

func logConfiguration(config *Config) {
	log.Printf("  Environment: %s", config.Environment)
	log.Printf("  Server: %s:%s", config.Server.Host, config.Server.Port)
	log.Printf("  Database: %s@%s:%s/%s",
		config.Database.User,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
	)
	log.Printf("  Email Provider: %s", config.Email.Provider)
}

func logAvailableEndpoints() {
	log.Println("📍 Available Endpoints:")
	log.Println("  GET  /health              - Health check")
	log.Println("  GET  /version             - API version")
	log.Println("  POST /api/v2/auth/signup  - User registration")
	log.Println("  POST /api/v2/auth/login   - User login")
	log.Println("  POST /api/v2/auth/refresh - Refresh token")
	log.Println("  GET  /api/v2/users/:id    - Get user (protected)")
	log.Println("  POST /api/v2/teams        - Create team (protected)")
	log.Println("  POST /api/v2/posts        - Create post (protected)")
}
