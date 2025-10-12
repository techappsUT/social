// ============================================================================
// FILE: backend/cmd/api/main.go
// FIXED: Removed Container.Cleanup() call since Container doesn't have that method
// ============================================================================
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

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ============================================================================
// APPLICATION
// ============================================================================

// App represents the application
type App struct {
	Container *Container
	Server    *http.Server
}

// ============================================================================
// MAIN ENTRY POINT
// ============================================================================

func main() {
	log.Println("🚀 Starting SocialQueue API Server...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No .env file found, using environment variables")
	} else {
		log.Println("  ✓ Loaded .env file")
	}

	// Initialize and run application
	app, err := NewApp()
	if err != nil {
		log.Fatalf("❌ Failed to initialize application: %v", err)
	}

	// Start server with graceful shutdown
	app.Start()
}

// ============================================================================
// APPLICATION INITIALIZATION
// ============================================================================

// NewApp initializes the application
func NewApp() (*App, error) {
	// Step 1: Load configuration
	log.Println("📝 Loading configuration...")
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
	router := setupRouter(container)
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

// setupDatabase initializes and verifies the database connection
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

	log.Println("  ✓ Database connection verified")

	return db, nil
}

// ============================================================================
// SERVER LIFECYCLE
// ============================================================================

// Start starts the HTTP server with graceful shutdown
func (app *App) Start() {
	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://%s", app.Server.Addr)
		log.Printf("✨ Environment: %s", app.Container.Config.Environment)
		log.Println("")
		logAvailableEndpoints()
		log.Println("")

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

	// FIXED: Removed app.Cleanup() call since Container doesn't have Cleanup method
	// If cleanup is needed in the future, add a Cleanup() method to Container

	log.Println("✅ Server gracefully stopped")
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// logConfiguration logs important configuration details
func logConfiguration(config *Config) {
	log.Printf("  ℹ️  Environment: %s", config.Environment)
	log.Printf("  ℹ️  Server: %s:%s", config.Server.Host, config.Server.Port)
	log.Printf("  ℹ️  Database: %s@%s:%s/%s",
		config.Database.User,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
	)
}

// logAvailableEndpoints logs all available API endpoints
func logAvailableEndpoints() {
	log.Println("📍 Available endpoints:")
	log.Println("  GET  /health                  - Health check")
	log.Println("  GET  /                        - API info")
	log.Println("")
	log.Println("  📝 Authentication:")
	log.Println("  POST /api/v2/auth/signup      - User registration")
	log.Println("  POST /api/v2/auth/login       - User login")
	log.Println("")
	log.Println("  👤 User Management (Protected):")
	log.Println("  GET  /api/v2/users/{id}       - Get user profile")
	log.Println("  PUT  /api/v2/users/{id}       - Update user profile")
	log.Println("  DELETE /api/v2/users/{id}     - Delete user account")
	log.Println("  GET  /api/v2/me               - Get current user")
	log.Println("")
	log.Println("  🏢 Team Management (Protected):")
	log.Println("  GET  /api/v2/teams            - List teams")
	log.Println("  POST /api/v2/teams            - Create team")
	log.Println("  GET  /api/v2/teams/{id}       - Get team details")
	log.Println("  PUT  /api/v2/teams/{id}       - Update team")
	log.Println("  DELETE /api/v2/teams/{id}     - Delete team")
	log.Println("")
	log.Println("  🔧 Admin (Protected):")
	log.Println("  GET  /api/v2/admin/users      - List all users")
}
