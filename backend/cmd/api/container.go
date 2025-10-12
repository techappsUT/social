// ============================================================================
// FILE: backend/cmd/api/container.go
// FIXED: NewInMemoryCacheService and DeleteUserUseCase with TokenService
// ============================================================================
package main

import (
	"database/sql"
	"fmt"

	appAuth "github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/common"
	appTeam "github.com/techappsUT/social-queue/internal/application/team"
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

	// User Use Cases
	CreateUserUC *appUser.CreateUserUseCase
	LoginUC      *appAuth.LoginUseCase
	UpdateUserUC *appUser.UpdateUserUseCase
	GetUserUC    *appUser.GetUserUseCase
	DeleteUserUC *appUser.DeleteUserUseCase

	// Team Use Cases
	CreateTeamUC *appTeam.CreateTeamUseCase
	GetTeamUC    *appTeam.GetTeamUseCase
	UpdateTeamUC *appTeam.UpdateTeamUseCase
	DeleteTeamUC *appTeam.DeleteTeamUseCase
	ListTeamsUC  *appTeam.ListTeamsUseCase

	// Handlers
	AuthHandler *handlers.AuthHandlerV2
	TeamHandler *handlers.TeamHandler

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware

	// Services (exposed for other use)
	TokenService common.TokenService
	EmailService common.EmailService
	Logger       common.Logger
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

	emailConfig := services.EmailConfig{
		Provider:    c.Config.Email.Provider,
		APIKey:      c.Config.Email.APIKey,
		FromAddress: c.Config.Email.FromAddress,
		FromName:    c.Config.Email.FromName,
	}
	c.EmailService = services.NewEmailService(emailConfig)

	c.Logger = services.NewLogger()

	// Initialize middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.TokenService)

	return nil
}

// initializeUseCases sets up all application use cases
func (c *Container) initializeUseCases() error {
	// Repositories
	userRepo := persistence.NewUserRepository(c.DB)
	teamRepo := persistence.NewTeamRepository(c.DB)
	memberRepo := persistence.NewTeamMemberRepository(c.DB)

	// Domain Services
	userService := userDomain.NewService(userRepo)

	// Cache Service
	// FIXED: Changed from NewInMemoryCache() to NewInMemoryCacheService()
	cacheService := services.NewInMemoryCacheService()

	// ========================================================================
	// USER USE CASES
	// ========================================================================

	c.CreateUserUC = appUser.NewCreateUserUseCase(
		userRepo,
		userService,
		c.TokenService,
		c.EmailService,
		c.Logger,
	)

	c.LoginUC = appAuth.NewLoginUseCase(
		userRepo,
		userService,
		c.TokenService,
		cacheService,
		c.Logger,
	)

	c.UpdateUserUC = appUser.NewUpdateUserUseCase(
		userRepo,
		c.Logger,
	)

	c.GetUserUC = appUser.NewGetUserUseCase(
		userRepo,
		c.Logger,
	)

	// FIXED: Added c.TokenService as second parameter
	c.DeleteUserUC = appUser.NewDeleteUserUseCase(
		userRepo,
		c.TokenService,
		c.Logger,
	)

	// ========================================================================
	// TEAM USE CASES
	// ========================================================================

	c.CreateTeamUC = appTeam.NewCreateTeamUseCase(
		teamRepo,
		memberRepo,
		userRepo,
		c.Logger,
	)

	c.GetTeamUC = appTeam.NewGetTeamUseCase(
		teamRepo,
		memberRepo,
		userRepo,
		c.Logger,
	)

	c.UpdateTeamUC = appTeam.NewUpdateTeamUseCase(
		teamRepo,
		memberRepo,
		userRepo,
		c.Logger,
	)

	c.DeleteTeamUC = appTeam.NewDeleteTeamUseCase(
		teamRepo,
		memberRepo,
		c.Logger,
	)

	c.ListTeamsUC = appTeam.NewListTeamsUseCase(
		teamRepo,
		memberRepo,
		userRepo,
		c.Logger,
	)

	return nil
}

// initializeHandlers sets up all HTTP handlers
func (c *Container) initializeHandlers() error {
	// User/Auth Handler
	c.AuthHandler = handlers.NewAuthHandlerV2(
		c.CreateUserUC,
		c.LoginUC,
		c.UpdateUserUC,
		c.GetUserUC,
		c.DeleteUserUC,
	)

	// Team Handler
	c.TeamHandler = handlers.NewTeamHandler(
		c.CreateTeamUC,
		c.GetTeamUC,
		c.UpdateTeamUC,
		c.DeleteTeamUC,
		c.ListTeamsUC,
	)

	return nil
}
