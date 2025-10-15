// ============================================================================
// FILE: backend/cmd/api/container.go
// ✅ COMPLETE FIXED VERSION - All errors resolved
// ============================================================================
package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"

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
	Redis  *redis.Client

	// Infrastructure Services
	TokenService      common.TokenService
	EmailService      common.EmailService
	CacheService      common.CacheService
	Logger            common.Logger
	WorkerQueue       *services.WorkerQueueService
	EncryptionService *services.EncryptionService
	Queries           *db.Queries // ← ADD THIS LINE

	// Repositories
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

	// Use Cases - Auth (ALL auth use cases)
	LoginUC              *auth.LoginUseCase
	RefreshTokenUC       *auth.RefreshTokenUseCase
	LogoutUC             *auth.LogoutUseCase
	VerifyEmailUC        *auth.VerifyEmailUseCase
	ResendVerificationUC *auth.ResendVerificationUseCase
	ForgotPasswordUC     *auth.ForgotPasswordUseCase
	ResetPasswordUC      *auth.ResetPasswordUseCase
	ChangePasswordUC     *auth.ChangePasswordUseCase

	// Use Cases - User
	CreateUserUC *userUC.CreateUserUseCase
	UpdateUserUC *userUC.UpdateUserUseCase
	GetUserUC    *userUC.GetUserUseCase
	DeleteUserUC *userUC.DeleteUserUseCase

	// Use Cases - Team
	CreateTeamUC       *teamUC.CreateTeamUseCase
	GetTeamUC          *teamUC.GetTeamUseCase
	UpdateTeamUC       *teamUC.UpdateTeamUseCase
	DeleteTeamUC       *teamUC.DeleteTeamUseCase
	ListTeamsUC        *teamUC.ListTeamsUseCase
	InviteMemberUC     *teamUC.InviteMemberUseCase
	RemoveMemberUC     *teamUC.RemoveMemberUseCase
	UpdateMemberRoleUC *teamUC.UpdateMemberRoleUseCase

	// Use Cases - Post
	CreateDraftUC  *postUC.CreateDraftUseCase
	SchedulePostUC *postUC.SchedulePostUseCase
	UpdatePostUC   *postUC.UpdatePostUseCase
	DeletePostUC   *postUC.DeletePostUseCase
	GetPostUC      *postUC.GetPostUseCase
	ListPostsUC    *postUC.ListPostsUseCase
	PublishNowUC   *postUC.PublishNowUseCase

	// Use Cases - Social
	ConnectAccountUC    *socialUC.ConnectAccountUseCase
	DisconnectAccountUC *socialUC.DisconnectAccountUseCase
	RefreshTokensUC     *socialUC.RefreshTokensUseCase
	ListAccountsUC      *socialUC.ListAccountsUseCase
	PublishPostUC       *socialUC.PublishPostUseCase
	GetAnalyticsUC      *socialUC.GetAnalyticsUseCase

	// HTTP Handlers
	AuthHandler   *handlers.AuthHandler // ✅ FIXED: Changed from AuthHandlerV2
	TeamHandler   *handlers.TeamHandler
	PostHandler   *handlers.PostHandler
	SocialHandler *handlers.SocialHandler

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware
	RateLimiter    *middleware.RateLimiter
}

// ============================================================================
// CONTAINER INITIALIZATION
// ============================================================================

func NewContainer(config *Config, db *sql.DB) (*Container, error) {
	container := &Container{
		Config: config,
		DB:     db,
	}

	// Initialize in dependency order
	if err := container.initializeInfrastructure(); err != nil {
		return nil, fmt.Errorf("infrastructure initialization failed: %w", err)
	}

	if err := container.initializeDomainServices(); err != nil {
		return nil, fmt.Errorf("domain services initialization failed: %w", err)
	}

	if err := container.initializeSocialAdapters(); err != nil {
		return nil, fmt.Errorf("social adapters initialization failed: %w", err)
	}

	if err := container.initializeUseCases(); err != nil {
		return nil, fmt.Errorf("use cases initialization failed: %w", err)
	}

	if err := container.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("handlers initialization failed: %w", err)
	}

	return container, nil
}

// ============================================================================
// INFRASTRUCTURE LAYER
// ============================================================================

func (c *Container) initializeInfrastructure() error {
	// Logger (required by other components)
	c.Logger = services.NewLogger()

	// ========================================================================
	// TOKEN & AUTH SERVICES
	// ========================================================================
	c.TokenService = services.NewJWTTokenService(
		c.Config.JWT.AccessSecret,
		c.Config.JWT.RefreshSecret,
	)
	c.Logger.Info("Token service initialized successfully")

	// ========================================================================
	// EMAIL SERVICE
	// ========================================================================
	emailConfig := services.EmailConfig{
		Provider:    c.Config.Email.Provider,
		APIKey:      c.Config.Email.APIKey,
		FromAddress: c.Config.Email.FromAddress,
		FromName:    c.Config.Email.FromName,
	}
	c.EmailService = services.NewEmailService(emailConfig)
	c.Logger.Info("Email service initialized successfully")

	// ========================================================================
	// REDIS CLIENT & CACHE SERVICE
	// ========================================================================
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := 6379
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			redisPort = p
		}
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			redisDB = d
		}
	}

	// Initialize Redis Cache Service
	cacheService, err := services.NewRedisCacheService(
		redisHost,
		redisPort,
		redisPassword,
		redisDB,
		c.Logger,
	)
	if err != nil {
		c.Logger.Warn(fmt.Sprintf("Failed to initialize Redis cache, falling back to in-memory: %v", err))
		c.CacheService = services.NewInMemoryCacheService()
	} else {
		c.CacheService = cacheService
		c.Logger.Info("✅ Redis cache service initialized successfully")

		// Store Redis client for worker queue and rate limiting
		c.Redis = cacheService.(*services.RedisCacheService).Client()
	}

	// ========================================================================
	// RATE LIMITER (NEW)
	// ========================================================================
	if c.Redis != nil {
		c.RateLimiter = middleware.NewRateLimiter(c.Redis, c.Logger)
		c.Logger.Info("✅ Rate limiter initialized successfully")
	} else {
		c.Logger.Warn("Rate limiter not initialized - Redis unavailable")
	}

	// ========================================================================
	// WORKER QUEUE SERVICE
	// ========================================================================
	if c.Redis != nil {
		c.WorkerQueue = services.NewWorkerQueueService(c.Redis, c.Logger)
		c.Logger.Info("✅ Worker queue service initialized successfully")
	} else {
		c.Logger.Warn("Worker queue not initialized - Redis unavailable")
	}

	// ========================================================================
	// ENCRYPTION SERVICE (for Social OAuth)
	// ========================================================================
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

	// ========================================================================
	// SQLC QUERIES - Initialize BEFORE repositories
	// ========================================================================
	// queries := db.New(c.DB)
	c.Queries = db.New(c.DB) // ← CHANGE FROM: queries := db.New(c.DB)
	c.Logger.Info("✅ SQLC queries initialized")

	// ========================================================================
	// REPOSITORIES
	// ========================================================================
	// ✅ FIX: Pass both c.DB and queries to NewUserRepository
	c.UserRepo = persistence.NewUserRepository(c.DB, c.Queries)
	c.TeamRepo = persistence.NewTeamRepository(c.DB)
	c.MemberRepo = persistence.NewTeamMemberRepository(c.DB)
	c.PostRepo = persistence.NewPostRepository(c.DB, c.Queries)

	// Social Repository (requires encryption service)
	if c.EncryptionService != nil {
		c.SocialRepo = persistence.NewSocialRepository(c.Queries, c.EncryptionService)
		c.Logger.Info("Social repository initialized successfully")
	} else {
		c.Logger.Warn("Social repository not initialized - encryption service unavailable")
	}

	c.Logger.Info("✅ Infrastructure layer initialized successfully")
	return nil
}

// initializeDomainServices sets up domain services
func (c *Container) initializeDomainServices() error {
	// User Domain Service
	c.UserService = userDomain.NewService(c.UserRepo)

	// Team Domain Service
	c.TeamService = teamDomain.NewService(c.TeamRepo, c.MemberRepo)

	c.Logger.Info("✅ Domain services initialized successfully")
	return nil
}

// initializeSocialAdapters sets up platform-specific OAuth adapters
func (c *Container) initializeSocialAdapters() error {
	c.SocialAdapters = make(map[socialDomain.Platform]socialAdapter.Adapter)

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	// Twitter Adapter
	twitterClientID := os.Getenv("TWITTER_CLIENT_ID")
	twitterClientSecret := os.Getenv("TWITTER_CLIENT_SECRET")
	if twitterClientID != "" && twitterClientSecret != "" {
		twitterAdapter := twitter.NewTwitterAdapter(
			twitterClientID,
			twitterClientSecret,
			fmt.Sprintf("%s/api/v2/social/auth/twitter/callback", baseURL),
		)
		c.SocialAdapters[socialDomain.PlatformTwitter] = twitterAdapter
		c.Logger.Info("Twitter adapter initialized")
	}

	// LinkedIn Adapter
	linkedinClientID := os.Getenv("LINKEDIN_CLIENT_ID")
	linkedinClientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	if linkedinClientID != "" && linkedinClientSecret != "" {
		linkedinAdapter := linkedin.NewLinkedInAdapter(
			linkedinClientID,
			linkedinClientSecret,
			fmt.Sprintf("%s/api/v2/social/auth/linkedin/callback", baseURL),
		)
		c.SocialAdapters[socialDomain.PlatformLinkedIn] = linkedinAdapter
		c.Logger.Info("LinkedIn adapter initialized")
	}

	// Facebook Adapter
	facebookAppID := os.Getenv("FACEBOOK_APP_ID")
	facebookAppSecret := os.Getenv("FACEBOOK_APP_SECRET")
	if facebookAppID != "" && facebookAppSecret != "" {
		facebookAdapter := facebook.NewFacebookAdapter(
			facebookAppID,
			facebookAppSecret,
			fmt.Sprintf("%s/api/v2/social/auth/facebook/callback", baseURL),
		)
		c.SocialAdapters[socialDomain.PlatformFacebook] = facebookAdapter
		c.Logger.Info("Facebook adapter initialized")
	}

	if len(c.SocialAdapters) > 0 {
		c.Logger.Info(fmt.Sprintf("✅ %d social adapters initialized", len(c.SocialAdapters)))
	} else {
		c.Logger.Warn("No social adapters initialized - OAuth credentials not provided")
	}

	return nil
}

// ============================================================================
// USE CASES (APPLICATION LAYER)
// ============================================================================

func (c *Container) initializeUseCases() error {
	// ========================================================================
	// AUTH USE CASES - Initialize ALL auth use cases
	// ========================================================================
	c.LoginUC = auth.NewLoginUseCase(
		c.UserRepo,
		c.UserService,
		c.TokenService,
		c.CacheService,
		c.Logger,
	)

	// ✅ NEW: Initialize remaining auth use cases
	// Note: You'll need to create these use case files if they don't exist yet
	// For now, setting them to nil to allow compilation

	// Uncomment these when the use cases are implemented:

	c.RefreshTokenUC = auth.NewRefreshTokenUseCase(
		c.UserRepo,
		c.TokenService,
		c.CacheService, // ← CRITICAL: This was missing!
		c.Logger,
	)

	c.LogoutUC = auth.NewLogoutUseCase(
		c.TokenService,
		c.Logger,
	)

	c.VerifyEmailUC = auth.NewVerifyEmailUseCase(
		c.UserRepo,
		c.Queries, // ← Add this
		c.Logger,
	)

	c.ResendVerificationUC = auth.NewResendVerificationUseCase(
		c.UserRepo,
		c.Queries, // ← Add this
		c.EmailService,
		c.Logger,
	)

	c.ForgotPasswordUC = auth.NewForgotPasswordUseCase(
		c.UserRepo,
		c.Queries, // ← Add this
		c.EmailService,
		c.Logger,
	)

	c.ResetPasswordUC = auth.NewResetPasswordUseCase(
		c.UserRepo,
		c.Queries,      // ← Add this
		c.UserService,  // ← Add this
		c.EmailService, // ← Add this
		c.Logger,
	)

	c.ChangePasswordUC = auth.NewChangePasswordUseCase(
		c.UserRepo,
		c.UserService, // ← Add this
		c.Logger,
	)

	// ========================================================================
	// USER USE CASES
	// ========================================================================
	c.CreateUserUC = userUC.NewCreateUserUseCase(
		c.UserRepo,
		c.UserService,
		c.TokenService,
		c.EmailService,
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
	// TEAM USE CASES
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
		c.UserRepo,
		c.Logger,
	)

	// ========================================================================
	// POST USE CASES
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
	// SOCIAL USE CASES (if available)
	// ========================================================================
	if c.SocialRepo != nil && c.EncryptionService != nil && len(c.SocialAdapters) > 0 {
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

		c.Logger.Info("✅ Social use cases initialized successfully")
	} else {
		c.Logger.Warn("Social use cases not initialized - missing encryption service or adapters")
	}

	c.Logger.Info("✅ Use cases initialized successfully")
	return nil
}

// initializeHandlers sets up HTTP handlers
func (c *Container) initializeHandlers() error {
	// ✅ FIX: Use full AuthHandler with ALL 12 parameters
	c.AuthHandler = handlers.NewAuthHandler(
		c.CreateUserUC,
		c.GetUserUC,
		c.UpdateUserUC,
		c.DeleteUserUC,
		c.LoginUC,
		c.RefreshTokenUC,       // May be nil if not implemented yet
		c.LogoutUC,             // May be nil if not implemented yet
		c.VerifyEmailUC,        // May be nil if not implemented yet
		c.ResendVerificationUC, // May be nil if not implemented yet
		c.ForgotPasswordUC,     // May be nil if not implemented yet
		c.ResetPasswordUC,      // May be nil if not implemented yet
		c.ChangePasswordUC,     // May be nil if not implemented yet
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

	// Social Handler (if social use cases available)
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
		c.Logger.Info("✅ Social handler initialized successfully")
	} else {
		c.Logger.Warn("Social handler not initialized - social features unavailable")
	}

	// Auth Middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.TokenService)

	c.Logger.Info("✅ Handlers initialized successfully")
	return nil
}

// Cleanup closes all open connections
func (c *Container) Cleanup() error {
	// Close Redis connection
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to close Redis connection: %v", err))
			return err
		}
		c.Logger.Info("Redis connection closed")
	}

	// Close database connection
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			c.Logger.Error(fmt.Sprintf("Failed to close database connection: %v", err))
			return err
		}
		c.Logger.Info("Database connection closed")
	}

	return nil
}
