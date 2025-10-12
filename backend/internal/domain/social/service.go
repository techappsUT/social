// path: backend/internal/domain/social/service.go
// 🆕 NEW - Clean Architecture

package social

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
)

// Service provides domain-level business logic for social accounts
type Service struct {
	accountRepo AccountRepository
	tokenRepo   TokenRepository
	teamRepo    teamDomain.Repository
	platformReg PlatformRegistry
}

// NewService creates a new social domain service
func NewService(
	accountRepo AccountRepository,
	tokenRepo TokenRepository,
	teamRepo teamDomain.Repository,
	platformReg PlatformRegistry,
) *Service {
	return &Service{
		accountRepo: accountRepo,
		tokenRepo:   tokenRepo,
		teamRepo:    teamRepo,
		platformReg: platformReg,
	}
}

// ConnectAccount initiates the OAuth connection for a social account
func (s *Service) ConnectAccount(ctx context.Context, teamID, userID uuid.UUID, platform Platform, accountType AccountType) (*Account, string, error) {
	// Verify team exists and is active
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, "", fmt.Errorf("team not found: %w", err)
	}

	if !team.IsActive() {
		return nil, "", fmt.Errorf("team is not active")
	}

	// Check team's social account limit
	accountCount, err := s.accountRepo.CountByTeamID(ctx, teamID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to count accounts: %w", err)
	}

	if !team.CanAddSocialAccount(int(accountCount)) {
		return nil, "", fmt.Errorf("social account limit exceeded for team plan")
	}

	// Get platform adapter
	adapter, err := s.platformReg.Get(platform)
	if err != nil {
		return nil, "", ErrPlatformNotSupported
	}

	// Create account entity
	account, err := NewAccount(teamID, userID, platform, accountType)
	if err != nil {
		return nil, "", err
	}

	// Validate account for platform
	if err := account.ValidateForPlatform(); err != nil {
		return nil, "", err
	}

	// Save account in pending state
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, "", fmt.Errorf("failed to create account: %w", err)
	}

	// Generate OAuth URL with state parameter (account ID)
	authURL, err := adapter.GetAuthorizationURL(account.ID().String())
	if err != nil {
		// Clean up account if OAuth URL generation fails
		s.accountRepo.Delete(ctx, account.ID())
		return nil, "", fmt.Errorf("failed to generate auth URL: %w", err)
	}

	return account, authURL, nil
}

// CompleteConnection completes the OAuth flow and activates the account
func (s *Service) CompleteConnection(ctx context.Context, accountID uuid.UUID, authCode string) error {
	// Get account
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	// Get platform adapter
	adapter, err := s.platformReg.Get(account.Platform())
	if err != nil {
		return ErrPlatformNotSupported
	}

	// Exchange auth code for tokens
	credentials, err := adapter.ExchangeToken(ctx, authCode)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %w", err)
	}

	// Get profile info
	tempAccount := &Account{credentials: *credentials, platform: account.Platform()}
	profile, err := adapter.GetProfile(ctx, tempAccount)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	// Check if this platform user is already connected to another team account
	existing, err := s.accountRepo.FindByPlatformUserID(ctx, account.Platform(), credentials.PlatformUserID)
	if err == nil && existing != nil && existing.ID() != account.ID() {
		return ErrDuplicatePlatformUser
	}

	// Complete the connection
	if err := account.Connect(*credentials, *profile); err != nil {
		return err
	}

	// Save encrypted tokens
	if err := s.tokenRepo.SaveTokens(ctx, account.ID(), *credentials); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	// Update account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// DisconnectAccount disconnects a social account
func (s *Service) DisconnectAccount(ctx context.Context, accountID uuid.UUID, disconnectedBy uuid.UUID) error {
	// Get account
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	// Verify disconnector has permission (would check via team member repo)
	// For now, we'll assume this check is done at the use case level

	// Get platform adapter to revoke access
	adapter, err := s.platformReg.Get(account.Platform())
	if err == nil {
		// Try to revoke access on platform (non-critical if fails)
		adapter.RevokeAccess(ctx, account)
	}

	// Disconnect account
	if err := account.Disconnect(); err != nil {
		return err
	}

	// Delete tokens
	if err := s.tokenRepo.DeleteTokens(ctx, account.ID()); err != nil {
		// Log error but don't fail
	}

	// Update account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to disconnect account: %w", err)
	}

	return nil
}

// RefreshAccountToken refreshes an expired token
func (s *Service) RefreshAccountToken(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	if account.Status() != StatusExpired && !account.IsExpired() {
		return nil // Token still valid
	}

	// Get current credentials
	credentials, err := s.tokenRepo.GetTokens(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get tokens: %w", err)
	}

	if credentials.RefreshToken == "" {
		return ErrInvalidRefreshToken
	}

	// Get platform adapter
	adapter, err := s.platformReg.Get(account.Platform())
	if err != nil {
		return ErrPlatformNotSupported
	}

	// Refresh token
	newCredentials, err := adapter.RefreshToken(ctx, credentials.RefreshToken)
	if err != nil {
		// Mark account as needing reconnection if refresh fails
		account.status = StatusReconnectRequired
		s.accountRepo.Update(ctx, account)
		return ErrTokenRefreshFailed
	}

	// Update account with new credentials
	if err := account.RefreshCredentials(*newCredentials); err != nil {
		return err
	}

	// Save new tokens
	if err := s.tokenRepo.UpdateTokens(ctx, accountID, *newCredentials); err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	// Update account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// SyncAccountProfile syncs the account profile from the platform
func (s *Service) SyncAccountProfile(ctx context.Context, accountID uuid.UUID) error {
	// Get account
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return ErrAccountNotFound
	}

	if !account.IsActive() {
		return ErrAccountNotActive
	}

	// Get platform adapter
	adapter, err := s.platformReg.Get(account.Platform())
	if err != nil {
		return ErrPlatformNotSupported
	}

	// Get updated profile
	profile, err := adapter.GetProfile(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	// Update account profile
	account.UpdateProfile(*profile)

	// Get rate limits
	limits, err := adapter.GetRateLimits(ctx, account)
	if err == nil && limits != nil {
		account.UpdateRateLimits(*limits)
	}

	// Update account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// PublishToAccount publishes a post to a social account
func (s *Service) PublishToAccount(ctx context.Context, accountID uuid.UUID, request *PostRequest) (*PostResult, error) {
	// Get account
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	// Check if account can post
	if !account.CanPost() {
		if account.Status() == StatusRateLimited {
			return nil, ErrAccountRateLimited
		}
		if account.NeedsReconnection() {
			return nil, ErrReconnectRequired
		}
		return nil, ErrAccountNotActive
	}

	// Get platform adapter
	adapter, err := s.platformReg.Get(account.Platform())
	if err != nil {
		return nil, ErrPlatformNotSupported
	}

	// Validate content for platform
	if err := s.validateContentForPlatform(request, account); err != nil {
		return nil, err
	}

	// Publish post
	result, err := adapter.PublishPost(ctx, account, request)
	if err != nil {
		// Check if rate limited
		if platformErr, ok := err.(PlatformError); ok && platformErr.RetryAfter != nil {
			account.MarkRateLimited(*platformErr.RetryAfter)
			s.accountRepo.Update(ctx, account)
		}
		return nil, fmt.Errorf("failed to publish: %w", err)
	}

	// Update post count for rate limiting
	account.IncrementPostCount()

	// Check rate limit info from response
	if result.RateLimitInfo != nil {
		if result.RateLimitInfo.Remaining == 0 {
			account.MarkRateLimited(result.RateLimitInfo.ResetsAt)
		}
	}

	// Update account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		// Log error but don't fail the publish
	}

	return result, nil
}

// GetAccountAnalytics retrieves analytics for an account
func (s *Service) GetAccountAnalytics(ctx context.Context, accountID uuid.UUID, period time.Duration) (*AccountAnalytics, error) {
	// Get account
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	if !account.IsActive() {
		return nil, ErrAccountNotActive
	}

	// Get platform adapter
	adapter, err := s.platformReg.Get(account.Platform())
	if err != nil {
		return nil, ErrPlatformNotSupported
	}

	// Get analytics
	analytics, err := adapter.GetAccountAnalytics(ctx, account, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

// RefreshExpiredTokens refreshes all expired tokens
func (s *Service) RefreshExpiredTokens(ctx context.Context) (int, error) {
	// Find expired accounts
	accounts, err := s.accountRepo.FindExpiredAccounts(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired accounts: %w", err)
	}

	refreshed := 0
	for _, account := range accounts {
		if err := s.RefreshAccountToken(ctx, account.ID()); err == nil {
			refreshed++
		}
	}

	return refreshed, nil
}

// ResetRateLimits resets rate limits for accounts
func (s *Service) ResetRateLimits(ctx context.Context) (int, error) {
	// Find rate limited accounts
	accounts, err := s.accountRepo.FindRateLimitedAccounts(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find rate limited accounts: %w", err)
	}

	reset := 0
	now := time.Now()
	for _, account := range accounts {
		if account.rateLimits.ResetsAt.Before(now) {
			if err := account.ResetRateLimit(); err == nil {
				if err := s.accountRepo.Update(ctx, account); err == nil {
					reset++
				}
			}
		}
	}

	return reset, nil
}

// ValidateAccountHealth checks the health of an account
func (s *Service) ValidateAccountHealth(ctx context.Context, accountID uuid.UUID) (*AccountHealth, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	health := &AccountHealth{
		AccountID:      account.ID(),
		Status:         account.Status(),
		TokenValid:     !account.IsExpired(),
		RequiresAction: account.NeedsReconnection(),
		Issues:         []string{},
	}

	// Check various health indicators
	if account.IsExpired() {
		health.Issues = append(health.Issues, "Token expired")
	}
	if account.Status() == StatusRateLimited {
		health.Issues = append(health.Issues, "Rate limited")
		health.RateLimitStatus = fmt.Sprintf("Resets at %s", account.rateLimits.ResetsAt.Format(time.RFC3339))
	}
	if account.Status() == StatusRevoked {
		health.Issues = append(health.Issues, "Access revoked by platform")
	}
	if account.Status() == StatusSuspended {
		health.Issues = append(health.Issues, "Account suspended")
	}

	return health, nil
}

// Helper functions

func (s *Service) validateContentForPlatform(request *PostRequest, account *Account) error {
	caps := GetPlatformCapabilities(account.Platform())

	// Check text length
	if len(request.Text) > caps.MaxTextLength {
		return ErrContentTooLong
	}

	// Check media count
	if len(request.MediaURLs) > caps.MaxMediaFiles {
		return ErrTooManyMediaFiles
	}

	// Check hashtag limit
	if len(request.Hashtags) > account.rateLimits.HashtagLimit {
		return ErrTooManyHashtags
	}

	// Check mention limit
	if len(request.Mentions) > account.rateLimits.MentionLimit {
		return ErrTooManyMentions
	}

	// Platform-specific validations
	switch account.Platform() {
	case PlatformInstagram:
		if len(request.MediaURLs) == 0 && len(request.MediaIDs) == 0 {
			return fmt.Errorf("Instagram requires at least one media file")
		}
	}

	return nil
}

// Specifications for complex business rules

// ActiveAccountSpecification checks if an account is active
type ActiveAccountSpecification struct{}

func (s ActiveAccountSpecification) IsSatisfiedBy(account *Account) bool {
	return account.IsActive()
}

// PublishableAccountSpecification checks if an account can publish
type PublishableAccountSpecification struct{}

func (s PublishableAccountSpecification) IsSatisfiedBy(account *Account) bool {
	return account.CanPost()
}
