// path: backend/cmd/api/wire_v2.go
// CREATE THIS AS A NEW FILE - DON'T MODIFY main.go YET

package main

import (
	"database/sql"
	"log"

	"github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/user"
	userDomain "github.com/techappsUT/social-queue/internal/domain/user"
	"github.com/techappsUT/social-queue/internal/handlers"
	"github.com/techappsUT/social-queue/internal/infrastructure/persistence"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
)

// ApplicationV2 holds the new clean architecture setup
type ApplicationV2 struct {
	// Use Cases
	CreateUserUC *user.CreateUserUseCase
	LoginUC      *auth.LoginUseCase

	// Handler
	AuthHandler *handlers.AuthHandlerV2
}

// InitializeApplicationV2 sets up the clean architecture
func InitializeApplicationV2(sqlDB *sql.DB, config *Config) (*ApplicationV2, error) {
	log.Println("🏗️  Initializing Clean Architecture...")

	// ========================================
	// 1. Infrastructure Layer
	// ========================================

	// Repositories
	userRepo := persistence.NewUserRepository(sqlDB)

	// Domain Services
	userService := userDomain.NewService(userRepo)

	// Infrastructure Services
	tokenService := services.NewJWTTokenService(
		config.JWT.AccessSecret,
		config.JWT.RefreshSecret,
	)

	emailService := services.NewEmailService(config.Email)

	cacheService := services.NewInMemoryCacheService() // Start with in-memory

	logger := services.NewLogger()

	// ========================================
	// 2. Application Layer (Use Cases)
	// ========================================

	createUserUC := user.NewCreateUserUseCase(
		userRepo,
		userService,
		tokenService,
		emailService,
		logger,
	)

	loginUC := auth.NewLoginUseCase(
		userRepo,
		userService,
		tokenService,
		cacheService,
		logger,
	)

	// ========================================
	// 3. Presentation Layer (Handlers)
	// ========================================

	authHandler := handlers.NewAuthHandlerV2(
		createUserUC,
		loginUC,
	)

	log.Println("✅ Clean Architecture initialized")

	return &ApplicationV2{
		CreateUserUC: createUserUC,
		LoginUC:      loginUC,
		AuthHandler:  authHandler,
	}, nil
}

// ============================================================================
// INFRASTRUCTURE SERVICE IMPLEMENTATIONS (Create these files)
// ============================================================================

// path: backend/internal/infrastructure/services/token_service.go
/*
package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/techappsUT/social-queue/internal/application/common"
)

type JWTTokenService struct {
	accessSecret  string
	refreshSecret string
}

func NewJWTTokenService(accessSecret, refreshSecret string) common.TokenService {
	return &JWTTokenService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (s *JWTTokenService) GenerateAccessToken(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

func (s *JWTTokenService) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

func (s *JWTTokenService) ValidateAccessToken(tokenString string) (*common.TokenClaims, error) {
	// Implementation here
	return nil, nil
}

func (s *JWTTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	// Implementation here
	return nil
}
*/

// path: backend/internal/infrastructure/services/email_service.go
/*
package services

import (
	"context"
	"log"

	"github.com/techappsUT/social-queue/internal/application/common"
)

type EmailService struct {
	// Add SendGrid or SMTP config
}

func NewEmailService(config EmailConfig) common.EmailService {
	return &EmailService{}
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, email, token string) error {
	log.Printf("📧 Would send verification email to %s with token %s", email, token)
	return nil
}

func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	log.Printf("📧 Would send password reset email to %s", email)
	return nil
}

func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	log.Printf("📧 Would send welcome email to %s", firstName)
	return nil
}

func (s *EmailService) SendInvitationEmail(ctx context.Context, email, teamName, inviteToken string) error {
	log.Printf("📧 Would send invitation email to %s for team %s", email, teamName)
	return nil
}
*/

// path: backend/internal/infrastructure/services/cache_service.go
/*
package services

import (
	"context"
	"sync"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
)

type InMemoryCacheService struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryCacheService() common.CacheService {
	return &InMemoryCacheService{
		data: make(map[string]string),
	}
}

func (c *InMemoryCacheService) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.data[key]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (c *InMemoryCacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	// TODO: Implement TTL
	return nil
}

func (c *InMemoryCacheService) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

func (c *InMemoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.data[key]
	return exists, nil
}
*/
