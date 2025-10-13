// ============================================================================
// FILE: backend/cmd/worker/main.go
// PURPOSE: Background worker binary for processing async jobs
// ============================================================================

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/infrastructure/persistence"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
)

// WorkerApp holds all worker dependencies
type WorkerApp struct {
	DB           *sql.DB
	Redis        *redis.Client
	Logger       common.Logger
	QueueService *services.WorkerQueueService
	Processors   []JobProcessor
}

// JobProcessor interface for all job processors
type JobProcessor interface {
	Name() string
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

func main() {
	log.Println("🔧 Starting SocialQueue Worker...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No .env file found, using environment variables")
	}

	// Initialize app
	app, err := NewWorkerApp()
	if err != nil {
		log.Fatalf("❌ Failed to initialize worker: %v", err)
	}
	defer app.Cleanup()

	// Start worker
	app.Start()
}

// NewWorkerApp initializes the worker application
func NewWorkerApp() (*WorkerApp, error) {
	// Initialize logger
	logger := services.NewConsoleLogger()

	// Connect to PostgreSQL
	db, err := connectDatabase()
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	logger.Info("✓ Connected to PostgreSQL")

	// Connect to Redis
	redisClient, err := connectRedis()
	if err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	logger.Info("✓ Connected to Redis")

	// Initialize services
	queueService := services.NewWorkerQueueService(redisClient, logger)
	queries := db.New(db)

	// Initialize repositories
	postRepo := persistence.NewPostRepository(db, queries)

	// Initialize job processors
	processors := []JobProcessor{
		NewPublishPostProcessor(postRepo, queueService, logger),
		NewFetchAnalyticsProcessor(postRepo, queueService, logger),
		NewCleanupProcessor(db, queueService, logger),
	}

	return &WorkerApp{
		DB:           db,
		Redis:        redisClient,
		Logger:       logger,
		QueueService: queueService,
		Processors:   processors,
	}, nil
}

// Start starts all job processors
func (app *WorkerApp) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start each processor in a goroutine
	for _, processor := range app.Processors {
		go func(p JobProcessor) {
			app.Logger.Info(fmt.Sprintf("▶️  Starting processor: %s", p.Name()))
			if err := p.Run(ctx); err != nil {
				app.Logger.Error(fmt.Sprintf("Processor %s failed: %v", p.Name(), err))
			}
		}(processor)
	}

	app.Logger.Info("✨ Worker started successfully")
	app.Logger.Info("📊 Active processors:")
	for _, p := range app.Processors {
		app.Logger.Info(fmt.Sprintf("   • %s", p.Name()))
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("🛑 Shutting down worker...")

	// Stop all processors gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	for _, processor := range app.Processors {
		if err := processor.Stop(shutdownCtx); err != nil {
			app.Logger.Error(fmt.Sprintf("Failed to stop processor %s: %v", processor.Name(), err))
		}
	}

	app.Logger.Info("✅ Worker stopped gracefully")
}

// Cleanup closes all connections
func (app *WorkerApp) Cleanup() {
	if app.Redis != nil {
		app.Redis.Close()
	}
	if app.DB != nil {
		app.DB.Close()
	}
}

// connectDatabase establishes PostgreSQL connection
func connectDatabase() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// connectRedis establishes Redis connection
func connectRedis() (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
