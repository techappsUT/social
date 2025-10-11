package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type LoginUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	tokenService common.TokenService
	cacheService common.CacheService
	logger       common.Logger
}

type LoginInput struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginOutput struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresIn    int      `json:"expiresIn"`
}

type UserDTO struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	EmailVerified bool       `json:"emailVerified"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty"`
}

func NewLoginUseCase(
	userRepo user.Repository,
	userService *user.Service,
	tokenService common.TokenService,
	cacheService common.CacheService,
	logger common.Logger,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:     userRepo,
		userService:  userService,
		tokenService: tokenService,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Authenticate user
	authenticatedUser, err := uc.userService.AuthenticateUser(
		ctx,
		input.Identifier,
		input.Password,
	)
	if err != nil {
		uc.logger.Warn("Authentication failed", "identifier", input.Identifier)
		return nil, common.ErrInvalidCredentials
	}

	// Check if user can access platform
	if !authenticatedUser.CanAccessPlatform() {
		return nil, fmt.Errorf("account %s", authenticatedUser.Status())
	}

	// Generate tokens
	accessToken, err := uc.tokenService.GenerateAccessToken(
		authenticatedUser.ID().String(),
		authenticatedUser.Email(),
		string(authenticatedUser.Role()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(
		authenticatedUser.ID().String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// >>>>>> FIX HERE: Choose based on your domain method signature <<<<<<

	// If RecordLogin DOES NOT return error:
	authenticatedUser.RecordLogin("")

	// OR if RecordLogin RETURNS error:
	// if err := authenticatedUser.RecordLogin(""); err != nil {
	//     uc.logger.Warn("Failed to record login", "error", err)
	// }

	// Update user in database
	if err := uc.userRepo.Update(ctx, authenticatedUser); err != nil {
		uc.logger.Error("Failed to update last login", "error", err)
		// Continue - don't fail the login
	}

	// Cache user session
	uc.cacheUserSession(ctx, authenticatedUser)

	return &LoginOutput{
		User:         uc.mapToDTO(authenticatedUser),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}

func (uc *LoginUseCase) validateInput(input LoginInput) error {
	if input.Identifier == "" {
		return fmt.Errorf("email or username is required")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("invalid password")
	}
	return nil
}

func (uc *LoginUseCase) cacheUserSession(ctx context.Context, u *user.User) {
	key := fmt.Sprintf("session:%s", u.ID())
	value := fmt.Sprintf("%s:%s", u.Email(), u.Role())
	ttl := 1 * time.Hour

	if err := uc.cacheService.Set(ctx, key, value, ttl); err != nil {
		uc.logger.Warn("Failed to cache session", "error", err)
	}
}

func (uc *LoginUseCase) mapToDTO(u *user.User) *UserDTO {
	return &UserDTO{
		ID:            u.ID().String(),
		Email:         u.Email(),
		Username:      u.Username(),
		FirstName:     u.FirstName(),
		LastName:      u.LastName(),
		Role:          string(u.Role()),
		Status:        string(u.Status()),
		EmailVerified: u.IsEmailVerified(),
		LastLoginAt:   u.LastLoginAt(),
	}
}
