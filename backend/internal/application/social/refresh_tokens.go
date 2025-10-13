// ============================================================================
// FILE: backend/internal/application/social/refresh_tokens.go
// ============================================================================
package social

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/adapters/social"
	"github.com/techappsUT/social-queue/internal/application/common"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
)

type RefreshTokensInput struct {
	AccountID uuid.UUID `json:"accountId" validate:"required"`
}

type RefreshTokensOutput struct {
	Account *SocialAccountDTO `json:"account"`
}

type RefreshTokensUseCase struct {
	socialRepo socialDomain.AccountRepository // FIXED
	adapters   map[socialDomain.Platform]social.Adapter
	logger     common.Logger
}

func NewRefreshTokensUseCase(
	socialRepo socialDomain.AccountRepository, // FIXED
	adapters map[socialDomain.Platform]social.Adapter,
	logger common.Logger,
) *RefreshTokensUseCase {
	return &RefreshTokensUseCase{
		socialRepo: socialRepo,
		adapters:   adapters,
		logger:     logger,
	}
}

func (uc *RefreshTokensUseCase) Execute(ctx context.Context, input RefreshTokensInput) (*RefreshTokensOutput, error) {
	// 1. Get account
	account, err := uc.socialRepo.FindByID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	// 2. Check if refresh needed
	if account.ExpiresAt() != nil && time.Until(*account.ExpiresAt()) > 24*time.Hour {
		// Token still valid for more than 24 hours
		return &RefreshTokensOutput{Account: MapAccountToDTO(account)}, nil
	}

	// 3. Get adapter
	adapter, ok := uc.adapters[account.Platform()]
	if !ok {
		return nil, fmt.Errorf("unsupported platform")
	}

	// 4. Refresh token
	credentials := account.Credentials()
	newToken, err := adapter.RefreshToken(ctx, credentials.RefreshToken)
	if err != nil {
		uc.logger.Error("Token refresh failed", "accountId", input.AccountID, "error", err)
		// Mark account as needing reconnection
		account.MarkExpired()
		uc.socialRepo.Update(ctx, account)
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// 5. Update account with new credentials
	newCredentials := socialDomain.Credentials{
		AccessToken:    newToken.AccessToken,
		RefreshToken:   newToken.RefreshToken,
		ExpiresAt:      newToken.ExpiresAt,
		Scope:          newToken.Scopes,
		PlatformUserID: credentials.PlatformUserID,
	}

	if err := account.RefreshCredentials(newCredentials); err != nil {
		return nil, fmt.Errorf("failed to update credentials: %w", err)
	}

	// 6. Save to repository
	if err := uc.socialRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account")
	}

	uc.logger.Info("Tokens refreshed", "accountId", input.AccountID)

	return &RefreshTokensOutput{Account: MapAccountToDTO(account)}, nil
}
