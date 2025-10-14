// Complete fix for logout.go
package auth

import (
	"context"
	"fmt"

	"github.com/techappsUT/social-queue/internal/application/common"
)

type LogoutUseCase struct {
	tokenService common.TokenService
	logger       common.Logger
}

type LogoutInput struct {
	RefreshToken string `json:"refreshToken"`
}

type LogoutOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewLogoutUseCase(
	tokenService common.TokenService,
	logger common.Logger,
) *LogoutUseCase {
	return &LogoutUseCase{
		tokenService: tokenService,
		logger:       logger,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) (*LogoutOutput, error) {
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// Revoke the refresh token
	if err := uc.tokenService.RevokeRefreshToken(ctx, input.RefreshToken); err != nil {
		uc.logger.Error("Failed to revoke refresh token", "error", err)
		return nil, fmt.Errorf("failed to logout")
	}

	uc.logger.Info("User logged out successfully")

	return &LogoutOutput{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}
