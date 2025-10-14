// path: backend/internal/application/auth/refresh_token.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// RefreshTokenInput represents the input for refreshing tokens
type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// RefreshTokenOutput represents the output after refreshing tokens
type RefreshTokenOutput struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	userRepo     user.Repository
	tokenService common.TokenService
	cacheService common.CacheService
	logger       common.Logger
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(
	userRepo user.Repository,
	tokenService common.TokenService,
	cacheService common.CacheService,
	logger common.Logger,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:     userRepo,
		tokenService: tokenService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute refreshes the access token using a refresh token
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	// 1. Validate refresh token
	claims, err := uc.tokenService.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		uc.logger.Warn(fmt.Sprintf("Invalid refresh token attempt: %v", err))
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// 2. Check if token is blacklisted (revoked)
	blacklisted, err := uc.cacheService.Get(ctx, fmt.Sprintf("blacklist:refresh:%s", input.RefreshToken))
	if err == nil && blacklisted != "" {
		uc.logger.Warn(fmt.Sprintf("Attempt to use blacklisted refresh token for user: %s", claims.UserID))
		return nil, fmt.Errorf("token has been revoked")
	}

	// 3. Get user from database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	usr, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("User not found for refresh token: %s", claims.UserID))
		return nil, fmt.Errorf("user not found")
	}

	// 4. Check user is still active
	if usr.IsDeleted() {
		return nil, fmt.Errorf("user account is deleted")
	}

	// 5. Generate new access token
	newAccessToken, err := uc.tokenService.GenerateAccessToken(usr.ID().String(), usr.Email(), "user")
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to generate access token: %v", err))
		return nil, fmt.Errorf("failed to generate access token")
	}

	// 6. Generate new refresh token (rotate refresh tokens for security)
	newRefreshToken, err := uc.tokenService.GenerateRefreshToken(usr.ID().String())
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to generate refresh token: %v", err))
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// 7. Blacklist old refresh token
	ttl := 30 * 24 * time.Hour // 30 days
	err = uc.cacheService.Set(ctx, fmt.Sprintf("blacklist:refresh:%s", input.RefreshToken), "1", ttl)
	if err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to blacklist old refresh token: %v", err))
	}

	// 8. Cache new refresh token
	err = uc.cacheService.Set(ctx, fmt.Sprintf("refresh:%s", newRefreshToken), usr.ID().String(), ttl)
	if err != nil {
		uc.logger.Warn(fmt.Sprintf("Failed to cache new refresh token: %v", err))
	}

	uc.logger.Info(fmt.Sprintf("Token refreshed successfully for user: %s", usr.Email()))

	return &RefreshTokenOutput{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(15 * time.Minute), // Access token expires in 15 min
	}, nil
}
