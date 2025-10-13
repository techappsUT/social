// ============================================================================
// FILE: backend/cmd/api/container.go
// COMPLETE FIXED VERSION
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
	appPost "github.com/techappsUT/social-queue/internal/application/post"
	appSocial "github.com/techappsUT/social-queue/internal/application/social"
	"github.com/techappsUT/social-queue/internal/application/team"
	"github.com/techappsUT/social-queue/internal/application/user"
	"github.com/techappsUT/social-queue/internal/db"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/infrastructure/persistence"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// Container holds all application dependencies
type Container struct {
	Config *Config
	DB     *sql.DB

	// ========================================================================
	// INFRASTRUCTURE LAYER
	// ========================================================================
	TokenService      common.TokenService
	EmailService      common.EmailService
	Logger            common.Logger
	CacheService      common.CacheService
	EncryptionService *services.EncryptionService
	SocialAdapters    map[socialDomain.Platform]socialAdapter.Adapter

	// ========================================================================
	// REPOSITORIES
	// ========================================================================
	UserRepo   *persistence.UserRepository
	TeamRepo   *persistence.TeamRepository
	MemberRepo *persistence.TeamMemberRepository
	PostRepo   *persistence.PostRepository
	SocialRepo socialDomain.AccountRepository

	// ========================================================================
	// USER MODULE USE CASES
	// ========================================================================
	CreateUserUC *user.CreateUserUseCase
	LoginUC      *auth.LoginUseCase
	UpdateUserUC *user.UpdateUserUseCase
	GetUserUC    *user.GetUserUseCase
	DeleteUserUC *user.DeleteUserUseCase

	// ========================================================================
	// TEAM MODULE USE CASES
	// ========================================================================
	CreateTeamUC       *team.CreateTeamUseCase
	GetTeamUC          *team.GetTeamUseCase
	UpdateTeamUC       *team.UpdateTeamUseCase
	DeleteTeamUC       *team.DeleteTeamUseCase
	ListTeamsUC        *team.ListTeamsUseCase
	InviteMemberUC     *team.InviteMemberUseCase
	RemoveMemberUC     *team.RemoveMemberUseCase
	UpdateMemberRoleUC *team.UpdateMemberRoleUseCase

	// ========================================================================
	// POST MODULE USE CASES
	// ========================================================================
	CreateDraftUC  *appPost.CreateDraftUseCase
	SchedulePostUC *appPost.SchedulePostUseCase
	UpdatePostUC   *appPost.UpdatePostUseCase
	DeletePostUC   *appPost.DeletePostUseCase
	GetPostUC      *appPost.GetPostUseCase
	ListPostsUC    *appPost.ListPostsUseCase
	PublishNowUC   *appPost.PublishNowUseCase

	// ========================================================================
	// SOCIAL MODULE USE CASES
	// ========================================================================
	ConnectAccountUC    *appSocial.ConnectAccountUseCase
	DisconnectAccountUC *appSocial.DisconnectAccountUseCase
	RefreshTokensUC     *appSocial.RefreshTokensUseCase
	ListAccountsUC      *appSocial.ListAccountsUseCase
	PublishPostUC       *appSocial.PublishPostUseCase
	GetAnalyticsUC      *appSocial.GetAnalyticsUseCase

	// ========================================================================
	// HANDLERS
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

	// Repositories
	c.UserRepo = persistence.NewUserRepository(queries)
	c.TeamRepo = persistence.NewTeamRepository(queries)
	c.MemberRepo = persistence.NewTeamMemberRepository(queries)
	c.PostRepo = persistence.NewPostRepository(queries)

	// FIX: Social repository needs DB connection and encryption service
	if c.EncryptionService != nil {
		c.SocialRepo = persistence.NewSocialRepository(c.DB, c.EncryptionService)
		c.Logger.Info("Social repository initialized successfully")
	} else {
		c.Logger.Warn("Social repository not initialized - encryption service unavailable")
	}

	c.Logger.Info("Infrastructure layer initialized successfully")
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
	c.CreateUserUC = user.NewCreateUserUseCase(
		c.UserRepo,
		c.EmailService,
		c.Logger,
	)

	c.LoginUC = auth.NewLoginUseCase(
		c.UserRepo,
		c.TokenService,
		c.Logger,
	)

	c.UpdateUserUC = user.NewUpdateUserUseCase(
		c.UserRepo,
		c.Logger,
	)

	c.GetUserUC = user.NewGetUserUseCase(
		c.UserRepo,
		c.Logger,
	)

	c.DeleteUserUC = user.NewDeleteUserUseCase(
		c.UserRepo,
		c.Logger,
	)

	// ========================================================================
	// TEAM MODULE
	// ========================================================================
	c.CreateTeamUC = team.NewCreateTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.Logger,
	)

	c.GetTeamUC = team.NewGetTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.UpdateTeamUC = team.NewUpdateTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.DeleteTeamUC = team.NewDeleteTeamUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.ListTeamsUC = team.NewListTeamsUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.InviteMemberUC = team.NewInviteMemberUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.UserRepo,
		c.EmailService,
		c.Logger,
	)

	c.RemoveMemberUC = team.NewRemoveMemberUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.UpdateMemberRoleUC = team.NewUpdateMemberRoleUseCase(
		c.TeamRepo,
		c.MemberRepo,
		c.Logger,
	)

	// ========================================================================
	// POST MODULE
	// ========================================================================
	c.CreateDraftUC = appPost.NewCreateDraftUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.SchedulePostUC = appPost.NewSchedulePostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.UpdatePostUC = appPost.NewUpdatePostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.DeletePostUC = appPost.NewDeletePostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.GetPostUC = appPost.NewGetPostUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.ListPostsUC = appPost.NewListPostsUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	c.PublishNowUC = appPost.NewPublishNowUseCase(
		c.PostRepo,
		c.MemberRepo,
		c.Logger,
	)

	// ========================================================================
	// SOCIAL MODULE
	// ========================================================================
	socialRepo := c.SocialRepo
	if socialRepo != nil && len(c.SocialAdapters) > 0 {
		c.ConnectAccountUC = appSocial.NewConnectAccountUseCase(
			socialRepo,
			c.MemberRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.DisconnectAccountUC = appSocial.NewDisconnectAccountUseCase(
			socialRepo,
			c.MemberRepo,
			c.Logger,
		)

		c.RefreshTokensUC = appSocial.NewRefreshTokensUseCase(
			socialRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.ListAccountsUC = appSocial.NewListAccountsUseCase(
			socialRepo,
			c.MemberRepo,
			c.Logger,
		)

		c.PublishPostUC = appSocial.NewPublishPostUseCase(
			socialRepo,
			c.MemberRepo,
			c.SocialAdapters,
			c.Logger,
		)

		// FIX: Remove memberRepo from GetAnalyticsUseCase constructor
		c.GetAnalyticsUC = appSocial.NewGetAnalyticsUseCase(
			socialRepo,
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
