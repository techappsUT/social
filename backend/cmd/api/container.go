// ============================================================================
// FILE: backend/cmd/api/container.go
// COMPLETE VERSION - Includes User, Team, Post, AND Social modules
// ============================================================================
package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/techappsUT/social-queue/internal/adapters/social"
	"github.com/techappsUT/social-queue/internal/adapters/social/facebook"
	"github.com/techappsUT/social-queue/internal/adapters/social/linkedin"
	"github.com/techappsUT/social-queue/internal/adapters/social/twitter"
	appAuth "github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/common"
	appPost "github.com/techappsUT/social-queue/internal/application/post"
	appSocial "github.com/techappsUT/social-queue/internal/application/social"
	appTeam "github.com/techappsUT/social-queue/internal/application/team"
	appUser "github.com/techappsUT/social-queue/internal/application/user"
	"github.com/techappsUT/social-queue/internal/db"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
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

	// ========================================================================
	// SERVICES
	// ========================================================================
	TokenService      common.TokenService
	EmailService      common.EmailService
	Logger            common.Logger
	CacheService      common.CacheService
	EncryptionService *services.EncryptionService // NEW for Social

	// ========================================================================
	// SOCIAL ADAPTERS (NEW)
	// ========================================================================
	SocialAdapters map[socialDomain.Platform]social.Adapter

	// ========================================================================
	// USER USE CASES
	// ========================================================================
	CreateUserUC *appUser.CreateUserUseCase
	LoginUC      *appAuth.LoginUseCase
	UpdateUserUC *appUser.UpdateUserUseCase
	GetUserUC    *appUser.GetUserUseCase
	DeleteUserUC *appUser.DeleteUserUseCase

	// ========================================================================
	// TEAM USE CASES
	// ========================================================================
	CreateTeamUC       *appTeam.CreateTeamUseCase
	GetTeamUC          *appTeam.GetTeamUseCase
	UpdateTeamUC       *appTeam.UpdateTeamUseCase
	DeleteTeamUC       *appTeam.DeleteTeamUseCase
	ListTeamsUC        *appTeam.ListTeamsUseCase
	InviteMemberUC     *appTeam.InviteMemberUseCase
	RemoveMemberUC     *appTeam.RemoveMemberUseCase
	UpdateMemberRoleUC *appTeam.UpdateMemberRoleUseCase

	// ========================================================================
	// POST USE CASES
	// ========================================================================
	CreateDraftUC  *appPost.CreateDraftUseCase
	SchedulePostUC *appPost.SchedulePostUseCase
	UpdatePostUC   *appPost.UpdatePostUseCase
	DeletePostUC   *appPost.DeletePostUseCase
	GetPostUC      *appPost.GetPostUseCase
	ListPostsUC    *appPost.ListPostsUseCase
	PublishNowUC   *appPost.PublishNowUseCase

	// ========================================================================
	// SOCIAL USE CASES (NEW)
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
	SocialHandler *handlers.SocialHandler // NEW

	// ========================================================================
	// MIDDLEWARE
	// ========================================================================
	AuthMiddleware *middleware.AuthMiddleware
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

	// Encryption Service (NEW for Social module)
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		c.Logger.Warn("ENCRYPTION_KEY not set, social OAuth features will be limited")
		// Don't fail - allow app to start without social features
	} else {
		var err error
		c.EncryptionService, err = services.NewEncryptionService(encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to initialize encryption service: %w", err)
		}
		c.Logger.Info("Encryption service initialized successfully")
	}

	// Auth Middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.TokenService)

	return nil
}

// initializeSocialAdapters sets up social platform OAuth adapters (NEW)
func (c *Container) initializeSocialAdapters() error {
	c.SocialAdapters = make(map[socialDomain.Platform]social.Adapter)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000" // Default for development
	}

	// Twitter/X Adapter
	twitterClientID := os.Getenv("TWITTER_CLIENT_ID")
	twitterClientSecret := os.Getenv("TWITTER_CLIENT_SECRET")
	if twitterClientID != "" && twitterClientSecret != "" {
		redirectURI := baseURL + "/api/v2/social/auth/twitter/callback"
		c.SocialAdapters[socialDomain.PlatformTwitter] = twitter.NewTwitterAdapter(
			twitterClientID,
			twitterClientSecret,
			redirectURI,
		)
		c.Logger.Info("Twitter adapter initialized", "redirectURI", redirectURI)
	} else {
		c.Logger.Warn("Twitter OAuth credentials not configured")
	}

	// LinkedIn Adapter
	linkedinClientID := os.Getenv("LINKEDIN_CLIENT_ID")
	linkedinClientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	if linkedinClientID != "" && linkedinClientSecret != "" {
		redirectURI := baseURL + "/api/v2/social/auth/linkedin/callback"
		c.SocialAdapters[socialDomain.PlatformLinkedIn] = linkedin.NewLinkedInAdapter(
			linkedinClientID,
			linkedinClientSecret,
			redirectURI,
		)
		c.Logger.Info("LinkedIn adapter initialized", "redirectURI", redirectURI)
	} else {
		c.Logger.Warn("LinkedIn OAuth credentials not configured")
	}

	// Facebook Adapter
	facebookAppID := os.Getenv("FACEBOOK_APP_ID")
	facebookAppSecret := os.Getenv("FACEBOOK_APP_SECRET")
	if facebookAppID != "" && facebookAppSecret != "" {
		redirectURI := baseURL + "/api/v2/social/auth/facebook/callback"
		c.SocialAdapters[socialDomain.PlatformFacebook] = facebook.NewFacebookAdapter(
			facebookAppID,
			facebookAppSecret,
			redirectURI,
		)
		c.Logger.Info("Facebook adapter initialized", "redirectURI", redirectURI)
	} else {
		c.Logger.Warn("Facebook OAuth credentials not configured")
	}

	if len(c.SocialAdapters) == 0 {
		c.Logger.Warn("No social platform adapters configured - social features will be unavailable")
	} else {
		c.Logger.Info("Social adapters initialized", "count", len(c.SocialAdapters))
	}

	return nil
}

// initializeUseCases sets up all application use cases
func (c *Container) initializeUseCases() error {
	// ========================================================================
	// REPOSITORIES
	// ========================================================================
	userRepo := persistence.NewUserRepository(c.DB)
	teamRepo := persistence.NewTeamRepository(c.DB)
	memberRepo := persistence.NewTeamMemberRepository(c.DB)

	// Create SQLC queries instance
	queries := db.New(c.DB)
	postRepo := persistence.NewPostRepository(c.DB, queries)

	// Social Repository (only if encryption service is available)
	var socialRepo socialDomain.AccountRepository
	if c.EncryptionService != nil {
		socialRepo = persistence.NewSocialRepository(c.DB, c.EncryptionService)
		c.Logger.Info("Social repository initialized")
	}

	// ========================================================================
	// DOMAIN SERVICES
	// ========================================================================
	userService := userDomain.NewService(userRepo)

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
		c.CacheService,
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

	c.InviteMemberUC = appTeam.NewInviteMemberUseCase(
		teamRepo,
		memberRepo,
		userRepo,
		c.EmailService,
		c.Logger,
	)

	c.RemoveMemberUC = appTeam.NewRemoveMemberUseCase(
		teamRepo,
		memberRepo,
		c.Logger,
	)

	c.UpdateMemberRoleUC = appTeam.NewUpdateMemberRoleUseCase(
		teamRepo,
		memberRepo,
		userRepo,
		c.Logger,
	)

	// ========================================================================
	// POST USE CASES
	// ========================================================================
	c.CreateDraftUC = appPost.NewCreateDraftUseCase(
		postRepo,
		teamRepo,
		memberRepo,
		c.Logger,
	)

	c.SchedulePostUC = appPost.NewSchedulePostUseCase(
		postRepo,
		memberRepo,
		c.Logger,
	)

	c.UpdatePostUC = appPost.NewUpdatePostUseCase(
		postRepo,
		memberRepo,
		c.Logger,
	)

	c.DeletePostUC = appPost.NewDeletePostUseCase(
		postRepo,
		memberRepo,
		c.Logger,
	)

	c.GetPostUC = appPost.NewGetPostUseCase(
		postRepo,
		memberRepo,
		c.Logger,
	)

	c.ListPostsUC = appPost.NewListPostsUseCase(
		postRepo,
		memberRepo,
		c.Logger,
	)

	c.PublishNowUC = appPost.NewPublishNowUseCase(
		postRepo,
		memberRepo,
		c.Logger,
	)

	// ========================================================================
	// SOCIAL USE CASES (NEW)
	// ========================================================================
	if socialRepo != nil && len(c.SocialAdapters) > 0 {
		c.ConnectAccountUC = appSocial.NewConnectAccountUseCase(
			socialRepo,
			memberRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.DisconnectAccountUC = appSocial.NewDisconnectAccountUseCase(
			socialRepo,
			memberRepo,
			c.Logger,
		)

		c.RefreshTokensUC = appSocial.NewRefreshTokensUseCase(
			socialRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.ListAccountsUC = appSocial.NewListAccountsUseCase(
			socialRepo,
			memberRepo,
			c.Logger,
		)

		c.PublishPostUC = appSocial.NewPublishPostUseCase(
			socialRepo,
			memberRepo,
			c.SocialAdapters,
			c.Logger,
		)

		c.GetAnalyticsUC = appSocial.NewGetAnalyticsUseCase(
			socialRepo,
			memberRepo,
			c.SocialAdapters,
			c.CacheService,
			c.Logger,
		)

		c.Logger.Info("Social use cases initialized successfully")
	} else {
		c.Logger.Warn("Social use cases not initialized - missing encryption service or adapters")
	}

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

	// Social Handler (NEW)
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

	return nil
}
