// path: backend/internal/social/service.go
package social

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBQueries interface - implement this based on your SQLC generated code
type DBQueries interface {
	CreateSocialToken(ctx context.Context, params CreateSocialTokenParams) (SocialToken, error)
	GetSocialToken(ctx context.Context, id int64) (SocialToken, error)
	UpdateSocialToken(ctx context.Context, params UpdateSocialTokenParams) (SocialToken, error)
	InvalidateSocialToken(ctx context.Context, id int64) error
	GetExpiringSocialTokens(ctx context.Context, params GetExpiringSocialTokensParams) ([]SocialToken, error)
}

// Database model structs - these should match your SQLC generated models
type SocialToken struct {
	ID               int64
	UserID           int64
	PlatformType     string
	PlatformUserID   string
	PlatformUsername sql.NullString
	AccessToken      string
	RefreshToken     sql.NullString
	ExpiresAt        time.Time
	Scope            sql.NullString
	IsValid          sql.NullBool
	LastValidated    sql.NullTime
	Extra            map[string]interface{}
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateSocialTokenParams struct {
	UserID           int64
	PlatformType     string
	PlatformUserID   string
	PlatformUsername sql.NullString
	AccessToken      string
	RefreshToken     sql.NullString
	ExpiresAt        time.Time
	Scope            sql.NullString
	Extra            map[string]interface{}
}

type UpdateSocialTokenParams struct {
	ID           int64
	AccessToken  string
	RefreshToken sql.NullString
	ExpiresAt    time.Time
}

type GetExpiringSocialTokensParams struct {
	ExpiresAt time.Time
	Limit     int32
}

type Service struct {
	registry   *AdapterRegistry
	queries    DBQueries
	encryption *TokenEncryption
	limiter    *RateLimiter
}

func NewService(
	registry *AdapterRegistry,
	queries DBQueries,
	encryption *TokenEncryption,
	limiter *RateLimiter,
) *Service {
	return &Service{
		registry:   registry,
		queries:    queries,
		encryption: encryption,
		limiter:    limiter,
	}
}

// InitiateOAuth starts the OAuth flow for a platform
func (s *Service) InitiateOAuth(ctx context.Context, platform PlatformType, userID int64, redirectURI string) (string, string, error) {
	adapter, err := s.registry.Get(platform)
	if err != nil {
		return "", "", err
	}

	// Generate state token (in production, store in Redis with expiry)
	state := fmt.Sprintf("%d:%s:%d", userID, platform, time.Now().Unix())

	authURL, err := adapter.AuthRedirect(ctx, state, redirectURI)
	if err != nil {
		return "", "", err
	}

	return authURL, state, nil
}

// HandleOAuthCallback processes the OAuth callback
func (s *Service) HandleOAuthCallback(ctx context.Context, platform PlatformType, code string, state string, redirectURI string, userID int64) (*PlatformToken, error) {
	adapter, err := s.registry.Get(platform)
	if err != nil {
		return nil, err
	}

	// Validate state (in production, verify from Redis)

	// Exchange code for tokens
	tokenResp, err := adapter.HandleOAuthCallback(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}

	// Create platform token
	token := &PlatformToken{
		UserID:           userID,
		PlatformType:     platform,
		PlatformUserID:   tokenResp.PlatformUserID,
		PlatformUsername: tokenResp.PlatformUsername,
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		ExpiresAt:        tokenResp.ExpiresAt,
		Scope:            tokenResp.Scope,
		IsValid:          true,
		Extra:            tokenResp.Extra,
	}

	// Encrypt tokens
	if err := s.encryption.EncryptToken(token); err != nil {
		return nil, err
	}

	// Store in database
	result, err := s.queries.CreateSocialToken(ctx, CreateSocialTokenParams{
		UserID:           token.UserID,
		PlatformType:     string(token.PlatformType),
		PlatformUserID:   token.PlatformUserID,
		PlatformUsername: sql.NullString{String: token.PlatformUsername, Valid: token.PlatformUsername != ""},
		AccessToken:      token.AccessToken,
		RefreshToken:     sql.NullString{String: token.RefreshToken, Valid: token.RefreshToken != ""},
		ExpiresAt:        token.ExpiresAt,
		Scope:            sql.NullString{String: token.Scope, Valid: token.Scope != ""},
		Extra:            token.Extra,
	})
	if err != nil {
		return nil, err
	}

	token.ID = result.ID
	token.CreatedAt = result.CreatedAt
	token.UpdatedAt = result.UpdatedAt

	return token, nil
}

// PublishContent publishes content to a platform
func (s *Service) PublishContent(ctx context.Context, tokenID int64, content *PostContent) (*PostResult, error) {
	// Get token from database
	dbToken, err := s.queries.GetSocialToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	// Convert to platform token
	token := s.dbTokenToPlatformToken(dbToken)

	// Decrypt tokens
	if err := s.encryption.DecryptToken(token); err != nil {
		return nil, err
	}

	// Get adapter
	adapter, err := s.registry.Get(token.PlatformType)
	if err != nil {
		return nil, err
	}

	// Check rate limits
	if err := s.limiter.Wait(ctx, token.PlatformType, token.PlatformUserID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Refresh token if needed
	token, err = adapter.RefreshTokenIfNeeded(ctx, token)
	if err != nil {
		return nil, err
	}

	// Update token in DB if refreshed
	if err := s.updateTokenIfChanged(ctx, dbToken, token); err != nil {
		return nil, err
	}

	// Post content
	result, err := adapter.PostContent(ctx, token, content)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RefreshExpiredTokens refreshes tokens that are about to expire
func (s *Service) RefreshExpiredTokens(ctx context.Context, expiryThreshold time.Time) error {
	tokens, err := s.queries.GetExpiringSocialTokens(ctx, GetExpiringSocialTokensParams{
		ExpiresAt: expiryThreshold,
		Limit:     100,
	})
	if err != nil {
		return err
	}

	for _, dbToken := range tokens {
		token := s.dbTokenToPlatformToken(dbToken)

		// Decrypt
		if err := s.encryption.DecryptToken(token); err != nil {
			continue
		}

		// Get adapter
		adapter, err := s.registry.Get(token.PlatformType)
		if err != nil {
			continue
		}

		// Refresh
		refreshed, err := adapter.RefreshTokenIfNeeded(ctx, token)
		if err != nil {
			// Invalidate token if refresh fails
			s.queries.InvalidateSocialToken(ctx, token.ID)
			continue
		}

		// Update in database
		if err := s.updateTokenIfChanged(ctx, dbToken, refreshed); err != nil {
			continue
		}
	}

	return nil
}

func (s *Service) dbTokenToPlatformToken(dbToken SocialToken) *PlatformToken {
	return &PlatformToken{
		ID:               dbToken.ID,
		UserID:           dbToken.UserID,
		PlatformType:     PlatformType(dbToken.PlatformType),
		PlatformUserID:   dbToken.PlatformUserID,
		PlatformUsername: dbToken.PlatformUsername.String,
		AccessToken:      dbToken.AccessToken,
		RefreshToken:     dbToken.RefreshToken.String,
		ExpiresAt:        dbToken.ExpiresAt,
		Scope:            dbToken.Scope.String,
		IsValid:          dbToken.IsValid.Bool,
		LastValidated:    dbToken.LastValidated.Time,
		Extra:            dbToken.Extra,
		CreatedAt:        dbToken.CreatedAt,
		UpdatedAt:        dbToken.UpdatedAt,
	}
}

func (s *Service) updateTokenIfChanged(ctx context.Context, old SocialToken, new *PlatformToken) error {
	if old.AccessToken == new.AccessToken {
		return nil
	}

	// Encrypt new tokens
	if err := s.encryption.EncryptToken(new); err != nil {
		return err
	}

	_, err := s.queries.UpdateSocialToken(ctx, UpdateSocialTokenParams{
		ID:           new.ID,
		AccessToken:  new.AccessToken,
		RefreshToken: sql.NullString{String: new.RefreshToken, Valid: new.RefreshToken != ""},
		ExpiresAt:    new.ExpiresAt,
	})

	return err
}
