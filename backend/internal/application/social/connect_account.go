// ============================================================================
// FILE: backend/internal/application/social/connect_account.go
// FIXED VERSION - Uses correct domain interface and methods
// ============================================================================
package social

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/adapters/social"
	"github.com/techappsUT/social-queue/internal/application/common"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type ConnectAccountInput struct {
	TeamID   uuid.UUID             `json:"teamId" validate:"required"`
	UserID   uuid.UUID             `json:"userId" validate:"required"`
	Platform socialDomain.Platform `json:"platform" validate:"required"`
	Code     string                `json:"code" validate:"required"` // OAuth authorization code
}

type ConnectAccountOutput struct {
	Account *SocialAccountDTO `json:"account"`
}

type ConnectAccountUseCase struct {
	socialRepo socialDomain.AccountRepository // FIXED: Use AccountRepository
	memberRepo team.MemberRepository
	adapters   map[socialDomain.Platform]social.Adapter
	logger     common.Logger
}

func NewConnectAccountUseCase(
	socialRepo socialDomain.AccountRepository, // FIXED
	memberRepo team.MemberRepository,
	adapters map[socialDomain.Platform]social.Adapter,
	logger common.Logger,
) *ConnectAccountUseCase {
	return &ConnectAccountUseCase{
		socialRepo: socialRepo,
		memberRepo: memberRepo,
		adapters:   adapters,
		logger:     logger,
	}
}

func (uc *ConnectAccountUseCase) Execute(ctx context.Context, input ConnectAccountInput) (*ConnectAccountOutput, error) {
	// 1. Verify user is team member
	_, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	// 2. Get platform adapter
	adapter, ok := uc.adapters[input.Platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", input.Platform)
	}

	// 3. Exchange authorization code for access token
	token, err := adapter.ExchangeCode(ctx, input.Code)
	if err != nil {
		uc.logger.Error("OAuth code exchange failed",
			"platform", input.Platform,
			"error", err)
		return nil, fmt.Errorf("failed to connect account: %w", err)
	}

	// 4. Validate token immediately
	valid, err := adapter.ValidateToken(ctx, token)
	if err != nil || !valid {
		uc.logger.Error("Token validation failed",
			"platform", input.Platform,
			"error", err)
		return nil, fmt.Errorf("received invalid token from %s", input.Platform)
	}

	// 5. Check if account already connected to this team
	existing, err := uc.socialRepo.FindByTeamAndPlatform(ctx, input.TeamID, input.Platform)
	if err == nil && len(existing) > 0 {
		// Check if same platform user
		for _, acc := range existing {
			// FIXED: Access PlatformUserID from Credentials
			if acc.Credentials().PlatformUserID == token.PlatformUserID {
				return nil, fmt.Errorf("this %s account is already connected", input.Platform)
			}
		}
	}

	// 6. Create domain entity
	account, err := socialDomain.NewAccount(input.TeamID, input.UserID, input.Platform, socialDomain.AccountTypePersonal)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// 7. Build credentials from token
	credentials := socialDomain.Credentials{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresAt:      token.ExpiresAt,
		Scope:          token.Scopes,
		PlatformUserID: token.PlatformUserID, // Store platform user ID
	}

	// 8. Create profile info (simplified - in production, call platform API)
	profile := socialDomain.ProfileInfo{
		Username:    fmt.Sprintf("user_%s", token.PlatformUserID[:8]),
		DisplayName: fmt.Sprintf("User %s", token.PlatformUserID[:8]),
		ProfileURL:  "",
		AvatarURL:   "",
	}

	// 9. Connect the account (activates it)
	if err := account.Connect(credentials, profile); err != nil {
		return nil, fmt.Errorf("failed to activate account: %w", err)
	}

	// 10. Save to repository (with encrypted tokens)
	if err := uc.socialRepo.Create(ctx, account); err != nil {
		uc.logger.Error("Failed to save social account",
			"teamId", input.TeamID,
			"platform", input.Platform,
			"error", err)
		return nil, fmt.Errorf("failed to save account")
	}

	uc.logger.Info("Social account connected",
		"teamId", input.TeamID,
		"platform", input.Platform,
		"accountId", account.ID())

	return &ConnectAccountOutput{
		Account: MapAccountToDTO(account),
	}, nil
}
