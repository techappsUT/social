// path: backend/cmd/api/container.go
package main

import (
	"database/sql"
	"fmt"

	appAuth "github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/common"
	appUser "github.com/techappsUT/social-queue/internal/application/user"
	userDomain "github.com/techappsUT/social-queue/internal/domain/user"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/infrastructure/persistence"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// Container holds all application dependencies
type Container struct {
	// Configuration
	Config *Config

	// Database
	DB *sql.DB

	// Use Cases
	CreateUserUC *appUser.CreateUserUseCase
	LoginUC      *appAuth.LoginUseCase
	UpdateUserUC *appUser.UpdateUserUseCase
	GetUserUC    *appUser.GetUserUseCase
	DeleteUserUC *appUser.DeleteUserUseCase

	// Handlers
	AuthHandler *handlers.AuthHandlerV2

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware

	// Services (exposed for other use)
	TokenService common.TokenService
}

// NewContainer creates and initializes the dependency injection container
func NewContainer(config *Config, db *sql.DB) (*Container, error) {
	c := &Container{
		Config: config,
		DB:     db,
	}

	if err := c.initializeInfrastructure(); err != nil {
		return nil, fmt.Errorf("infrastructure initialization failed: %w", err)
	}

	if err := c.initializeUseCases(); err != nil {
		return nil, fmt.Errorf("use case initialization failed: %w", err)
	}

	if err := c.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("handler initialization failed: %w", err)
	}

	return c, nil
}

// initializeInfrastructure sets up repositories and services
func (c *Container) initializeInfrastructure() error {
	// Infrastructure Services
	c.TokenService = services.NewJWTTokenService(
		c.Config.JWT.AccessSecret,
		c.Config.JWT.RefreshSecret,
	)

	// Initialize middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.TokenService)

	return nil
}

// initializeUseCases sets up all application use cases
func (c *Container) initializeUseCases() error {
	// Repositories
	userRepo := persistence.NewUserRepository(c.DB)

	// Domain Services
	userService := userDomain.NewService(userRepo)

	// Infrastructure Services
	emailConfig := services.EmailConfig{
		Provider:    c.Config.Email.Provider,
		APIKey:      c.Config.Email.APIKey,
		FromAddress: c.Config.Email.FromAddress,
		FromName:    c.Config.Email.FromName,
	}
	emailService := services.NewEmailService(emailConfig)
	cacheService := services.NewInMemoryCacheService()
	logger := services.NewLogger()

	// User Use Cases
	c.CreateUserUC = appUser.NewCreateUserUseCase(
		userRepo,
		userService,
		c.TokenService,
		emailService,
		logger,
	)

	c.LoginUC = appAuth.NewLoginUseCase(
		userRepo,
		userService,
		c.TokenService,
		cacheService,
		logger,
	)

	c.UpdateUserUC = appUser.NewUpdateUserUseCase(
		userRepo,
		logger,
	)

	c.GetUserUC = appUser.NewGetUserUseCase(
		userRepo,
		logger,
	)

	c.DeleteUserUC = appUser.NewDeleteUserUseCase(
		userRepo,
		c.TokenService,
		logger,
	)

	// TODO: Add more use cases here as you create them
	// c.CreateTeamUC = appTeam.NewCreateTeamUseCase(...)
	// c.CreatePostUC = appPost.NewCreatePostUseCase(...)

	return nil
}

// initializeHandlers sets up HTTP handlers
func (c *Container) initializeHandlers() error {
	c.AuthHandler = handlers.NewAuthHandlerV2(
		c.CreateUserUC,
		c.LoginUC,
		c.UpdateUserUC,
		c.GetUserUC,
		c.DeleteUserUC,
	)

	// TODO: Add more handlers here
	// c.TeamHandler = handlers.NewTeamHandler(...)
	// c.PostHandler = handlers.NewPostHandler(...)

	return nil
}

// Cleanup releases all resources
func (c *Container) Cleanup() {
	if c.DB != nil {
		c.DB.Close()
	}
}
