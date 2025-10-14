// path: backend/internal/application/auth/logout.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
)

// LogoutInput represents the input for logout
type LogoutInput struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"-"` // Set from context
}

// LogoutOutput represents the output after logout
type LogoutOutput struct {
	Message string `json:"message"`
}

// LogoutUseCase handles user logout
type LogoutUseCase struct {
	tokenService common.TokenService
	cacheService common.CacheService
	logger       common.Logger
}

// NewLogoutUseCase creates a new logout use case
func NewLogoutUseCase(
	tokenService common.TokenService,
	cacheService common.CacheService,
	logger common.Logger,
) *LogoutUseCase {
	return &LogoutUseCase{
		tokenService: tokenService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute logs out the user by blacklisting tokens
func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) (*LogoutOutput, error) {
	// 1. Blacklist refresh token (30 days)
	if input.RefreshToken != "" {
		ttl := 30 * 24 * time.Hour
		err := uc.cacheService.Set(
			ctx,
			fmt.Sprintf("blacklist:refresh:%s", input.RefreshToken),
			"1",
			ttl,
		)
		if err != nil {
			uc.logger.Warn(fmt.Sprintf("Failed to blacklist refresh token: %v", err))
		}

		// Remove refresh token from cache
		uc.cacheService.Delete(ctx, fmt.Sprintf("refresh:%s", input.RefreshToken))
	}

	// 2. Blacklist access token (15 minutes - until it expires naturally)
	if input.AccessToken != "" {
		ttl := 15 * time.Minute
		err := uc.cacheService.Set(
			ctx,
			fmt.Sprintf("blacklist:access:%s", input.AccessToken),
			"1",
			ttl,
		)
		if err != nil {
			uc.logger.Warn(fmt.Sprintf("Failed to blacklist access token: %v", err))
		}
	}

	// 3. Clear user session from cache
	if input.UserID != "" {
		uc.cacheService.Delete(ctx, fmt.Sprintf("session:%s", input.UserID))
	}

	uc.logger.Info(fmt.Sprintf("User logged out successfully: %s", input.UserID))

	return &LogoutOutput{
		Message: "Logged out successfully",
	}, nil
}
