// ============================================================================
// FILE: backend/cmd/api/container.go
// FULLY CORRECTED - All constructor signatures match actual use cases
// ============================================================================
package main

import (
	"database/sql"
	"fmt"
	"os"

	socialAdapter "github.com/techappsUT/social-queue/internal/adapters/social"
	"github.com/techappsUT/social-queue/internal/adapters/social/facebook"
	"github.com/techappsUT/social-queue/internal/adapters/social/linkedin"
	"github.com/techappsUT/social-queue/internal/adapters/social/twitter"
	"github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/common"
	postUC "github.com/techappsUT/social-queue/internal/application/post"
	socialUC "github.com/techappsUT/social-queue/internal/application/social"
	teamUC "github.com/techappsUT/social-queue/internal/application/team"
	userUC "github.com/techappsUT/social-queue/internal/application/user"
	"github.com/techappsUT/social-queue/internal/db"
	postDomain "github.com/techappsUT/social-queue/internal/domain/post"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
	userDomain "github.com/techappsUT/social-queue/internal/domain/user"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/infrastructure/persistence"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// ============================================================================
// DEPENDENCY INJECTION CONTAINER
// ============================================================================

type Container struct {
	// Configuration
	Config *Config
	DB     *sql.DB

	// Infrastructure Services
	TokenService      common.TokenService
	EmailService      common.EmailService
	CacheService      common.CacheService
	Logger            common.Logger
	EncryptionService *services.EncryptionService

	// Repositories (interfaces)
	UserRepo   userDomain.Repository
	TeamRepo   teamDomain.Repository
	MemberRepo teamDomain.MemberRepository
	PostRepo   postDomain.Repository
	SocialRepo socialDomain.AccountRepository

	// Domain Services
	UserService *userDomain.Service
	TeamService *teamDomain.Service

	// Social Platform Adapters
	SocialAdapters map[socialDomain.Platform]socialAdapter.Adapter

	// ========================================================================
	// USER MODULE USE CASES
	// ========================================================================
	CreateUserUC *userUC.CreateUserUseCase
	LoginUC      *auth.LoginUseCase
	UpdateUserUC *userUC.UpdateUserUseCase
	GetUserUC    *userUC.GetUserUseCase
	DeleteUserUC *userUC.DeleteUserUseCase

	// ========================================================================
	// TEAM MODULE USE CASES
	// ========================================================================
	CreateTeamUC       *teamUC.CreateTeamUseCase
	GetTeamUC          *teamUC.GetTeamUseCase
	UpdateTeamUC       *teamUC.UpdateTeamUseCase
	DeleteTeamUC       *teamUC.DeleteTeamUseCase
	ListTeamsUC        *teamUC.ListTeamsUseCase
	InviteMemberUC     *teamUC.InviteMemberUseCase
	RemoveMemberUC     *teamUC.RemoveMemberUseCase
	UpdateMemberRoleUC *teamUC.UpdateMemberRoleUseCase

	// ========================================================================
	// POST MODULE USE CASES
	// ========================================================================
	CreateDraftUC  *postUC.CreateDraftUseCase
	SchedulePostUC *postUC.SchedulePostUseCase
	UpdatePostUC   *postUC.UpdatePostUseCase
	DeletePostUC   *postUC.DeletePostUseCase
	GetPostUC      *postUC.GetPostUseCase
	ListPostsUC    *postUC.ListPostsUseCase
	PublishNowUC   *postUC.PublishNowUseCase

	// ========================================================================
	// SOCIAL MODULE USE CASES
	// ========================================================================
	ConnectAccountUC    *socialUC.ConnectAccountUseCase
	DisconnectAccountUC *socialUC.DisconnectAccountUseCase
	RefreshTokensUC     *socialUC.RefreshTokensUseCase
	ListAccountsUC      *socialUC.ListAccountsUseCase
	PublishPostUC       *socialUC.PublishPostUseCase
	GetAnalyticsUC      *socialUC.GetAnalyticsUseCase

	// ========================================================================
	// HTTP HANDLERS
	// ========================================================================
	AuthHandler   *handlers.AuthHandlerV2
	TeamHandler   *handlers.TeamHandler
	PostHandler   *handlers.PostHandler
	SocialHandler *handlers.SocialHandler

	// ========================================================================
	// MIDDLEWARE
	// ========================================================================
	AuthMiddleware *middleware.AuthMiddleware
}

// NewContainer creates and initializes the dependency injection container
func NewContainer(config *Config, database *sql.DB) (*Container, error) {
	c := &Container{
		Config: config,
		DB:     database,
	}

	if err := c.initializeInfrastructure(); err != nil {
		return nil, fmt.Errorf("infrastructure initialization failed: %w", err)
	}

	if err := c.initializeDomainServices(); err != nil {
		return nil, fmt.Errorf("domain services initialization failed: %w", err)
	}

	if err := c.initializeSocialAdapters(); err != nil {
		return nil, fmt.Errorf("social adapters initialization failed: %w", err)
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
	// Token Service
	c.TokenService = services.NewJWTTokenService(
		c.Config.JWT.AccessSecret,
		c.Config.JWT.RefreshSecret,
	)

	// Email Service
	emailConfig := services.EmailConfig{
		Provider:    c.Config.Email.Provider,
		APIKey:      c.Config.Email.APIKey,
		FromAddress: c.Config.Email.FromAddress,
		FromName:    c.Config.Email.FromName,
	}
	c.EmailService = services.NewEmailService(emailConfig)

	// Logger
	c.Logger = services.NewLogger()

	// Cache Service
	c.CacheService = services.NewInMemoryCacheService()

	// Encryption Service (for Social OAuth)
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		c.Logger.Warn("ENCRYPTION_KEY not set, social OAuth features will be limited")
	} else {
		var err error
		c.EncryptionService, err = services.NewEncryptionService(encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to create encryption service: %w", err)
		}
		c.Logger.Info("Encryption service initialized successfully")
	}

	// Initialize SQLC queries
	queries := db.New(c.DB)

	// ========================================================================
	// REPOSITORIES
	// ========================================================================
	c.UserRepo = persistence.NewUserRepository(c.DB)
	c.TeamRepo = persistence.NewTeamRepository(c.DB)
	c.MemberRepo = persistence.NewTeamMemberRepository(c.DB)
	c.PostRepo = persistence.NewPostRepository(c.DB, queries)

	// Social Repository
	if c.EncryptionService != nil {
		c.SocialRepo = persistence.NewSocialRepository(queries, c.EncryptionService)
		c.Logger.Info("Social repository initialized successfully")
	} else {
		c.Logger.Warn("Social repository not initialized - encryption service unavailable")
	}

	c.Logger.Info("Infrastructure layer initialized successfully")
	return nil
}

// initializeDomainServices sets up domain services
func (c *Container) initializeDomainServices() error {
	// User Domain Service
	c.UserService = userDomain.NewService(c.UserRepo)

	// Team Domain Service
	c.TeamService = teamDomain.NewService(c.TeamRepo, c.MemberRepo)

	c.Logger.Info("Domain services initialized successfully")
	return nil
}

// initializeSocialAdapters sets up platform adapters
func (c *Container) initializeSocialAdapters() error {
	c.SocialAdapters = make(map[socialDomain.Platform]socialAdapter.Adapter)

	// Twitter adapter
	twitterClientID := os.Getenv("TWITTER_CLIENT_ID")
	twitterClientSecret := os.Getenv("TWITTER_CLIENT_SECRET")
	if twitterClientID != "" && twitterClientSecret != "" {
		twitterAdapter := twitter.NewTwitterAdapter(twitterClientID, twitterClientSecret, "http://localhost:8000/api/v2/social/auth/twitter/callback")
		c.SocialAdapters[socialDomain.PlatformTwitter] = twitterAdapter
		c.Logger.Info("Twitter adapter initialized")
	}

	// LinkedIn adapter
	linkedinClientID := os.Getenv("LINKEDIN_CLIENT_ID")
	linkedinClientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	if linkedinClientID != "" && linkedinClientSecret != "" {
		linkedinAdapter := linkedin.NewLinkedInAdapter(linkedinClientID, linkedinClientSecret, "http://localhost:8000/api/v2/social/auth/linkedin/callback")
		c.SocialAdapters[socialDomain.PlatformLinkedIn] = linkedinAdapter
		c.Logger.Info("LinkedIn adapter initialized")
	}

	// Facebook adapter
	facebookAppID := os.Getenv("FACEBOOK_APP_ID")
	facebookAppSecret := os.Getenv("FACEBOOK_APP_SECRET")
	if facebookAppID != "" && facebookAppSecret != "" {
		facebookAdapter := facebook.NewFacebookAdapter(facebookAppID, facebookAppSecret, "http://localhost:8000/api/v2/social/auth/facebook/callback")
		c.SocialAdapters[socialDomain.PlatformFacebook] = facebookAdapter
		c.Logger.Info("Facebook adapter initialized")
	}

	if len(c.SocialAdapters) == 0 {
		c.Logger.Warn("No social adapters initialized - set environment variables for Twitter, LinkedIn, or Facebook")
	} else {
		c.Logger.Info("Social adapters initialized", "count", len(c.SocialAdapters))
	}

	return nil
}

// initializeUseCases sets up all application use cases
func (c *Container) initializeUseCases() error {
	// ========================================================================
	// USER MODULE
	// ========================================================================
	c.CreateUserUC = userUC.NewCreateUserUseCase(
		c.UserRepo,
		c.UserService,
		c.TokenService,
		c.EmailService,
		c.Logger,
	)

	c.LoginUC = auth.NewLoginUseCase(
		c.UserRepo,
		c.UserService,
		c.TokenService,
		c.CacheService,
		c.Logger,
	)

	c.UpdateUserUC = userUC.NewUpdateUserUseCase(
		c.UserRepo,
		c.Logger,
	)

	c.GetUserUC = userUC.NewGetUserUseCase(
		c.UserRepo,
		c.Logger,
	)

	c.DeleteUserUC = userUC.NewDeleteUserUseCase(
		c.UserRepo,
		c.TokenService,
		c.Logger,
	)

	// ========================================================================
	// TEAM MODULE
	// ========================================================================
	c.CreateTeamUC = teamUC.NewCreateTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.Logger,
	)

	c.GetTeamUC = teamUC.NewGetTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.Logger,
	)

	c.UpdateTeamUC = teamUC.NewUpdateTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.Logger,
	)

	c.DeleteTeamUC = teamUC.NewDeleteTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.ListTeamsUC = teamUC.NewListTeamsUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.Logger,
	)

	c.InviteMemberUC = teamUC.NewInviteMemberUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.EmailService,
		c.Logger,
	)

	c.RemoveMemberUC = teamUC.NewRemoveMemberUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.UpdateMemberRoleUC = teamUC.NewUpdateMemberRoleUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo, // ✅ Added missing UserRepo
		c.Logger,
	)

	// ========================================================================
	// POST MODULE
	// ========================================================================
	c.CreateDraftUC = postUC.NewCreateDraftUseCase(
		c.PostRepo,
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.SchedulePostUC = postUC.NewSchedulePostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.UpdatePostUC = postUC.NewUpdatePostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.DeletePostUC = postUC.NewDeletePostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.GetPostUC = postUC.NewGetPostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.ListPostsUC = postUC.NewListPostsUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.PublishNowUC = postUC.NewPublishNowUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	// ========================================================================
	// SOCIAL MODULE
	// ========================================================================
	if c.SocialRepo != nil && len(c.SocialAdapters) > 0 {
		c.ConnectAccountUC = socialUC.NewConnectAccountUseCase(
			c.SocialRepo,
			c.MemberRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.DisconnectAccountUC = socialUC.NewDisconnectAccountUseCase(
			c.SocialRepo,
			c.MemberRepo,
			c.Logger,
		)

		c.RefreshTokensUC = socialUC.NewRefreshTokensUseCase(
			c.SocialRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.ListAccountsUC = socialUC.NewListAccountsUseCase(
			c.SocialRepo,
			c.MemberRepo,
			c.Logger,
		)

		c.PublishPostUC = socialUC.NewPublishPostUseCase(
			c.SocialRepo,
			c.MemberRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.GetAnalyticsUC = socialUC.NewGetAnalyticsUseCase(
			c.SocialRepo,
			c.SocialAdapters,
			c.CacheService,
			c.Logger,
		)

		c.Logger.Info("Social use cases initialized successfully")
	} else {
		c.Logger.Warn("Social use cases not initialized - missing encryption service or adapters")
	}

	c.Logger.Info("Use cases initialized successfully")
	return nil
}

// initializeHandlers sets up HTTP handlers
func (c *Container) initializeHandlers() error {
	// Auth Handler
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
		c.InviteMemberUC,
		c.RemoveMemberUC,
		c.UpdateMemberRoleUC,
	)

	// Post Handler
	c.PostHandler = handlers.NewPostHandler(
		c.CreateDraftUC,
		c.SchedulePostUC,
		c.UpdatePostUC,
		c.DeletePostUC,
		c.GetPostUC,
		c.ListPostsUC,
		c.PublishNowUC,
	)

	// Social Handler
	if c.ConnectAccountUC != nil {
		c.SocialHandler = handlers.NewSocialHandler(
			c.ConnectAccountUC,
			c.DisconnectAccountUC,
			c.RefreshTokensUC,
			c.ListAccountsUC,
			c.PublishPostUC,
			c.GetAnalyticsUC,
			c.SocialAdapters,
		)
		c.Logger.Info("Social handler initialized successfully")
	} else {
		c.Logger.Warn("Social handler not initialized - social features unavailable")
	}

	// Auth Middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.TokenService)

	c.Logger.Info("Handlers initialized successfully")
	return nil
}
