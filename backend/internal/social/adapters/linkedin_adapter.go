// path: backend/internal/social/adapters/linkedin_adapter.go
package adapters

import (
	"context"
	"time"

	"github.com/techappsUT/social-queue/internal/social"
)

type LinkedInAdapter struct {
	clientID     string
	clientSecret string
}

func NewLinkedInAdapter(clientID, clientSecret string) *LinkedInAdapter {
	return &LinkedInAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (l *LinkedInAdapter) GetPlatformName() social.PlatformType {
	return social.PlatformLinkedIn
}

func (l *LinkedInAdapter) GetCapabilities() *social.PlatformCapabilities {
	return &social.PlatformCapabilities{
		SupportsImages:        true,
		SupportsVideos:        true,
		SupportsMultipleMedia: true,
		SupportsScheduling:    false,
		SupportsHashtags:      true,
		MaxTextLength:         3000,
		MaxMediaCount:         9,
		MaxVideoSizeMB:        200,
	}
}

func (l *LinkedInAdapter) AuthRedirect(ctx context.Context, state string, redirectURI string) (string, error) {
	// TODO: Implement LinkedIn OAuth
	return "", nil
}

func (l *LinkedInAdapter) HandleOAuthCallback(ctx context.Context, code string, redirectURI string) (*social.OAuthTokenResponse, error) {
	// TODO: Implement
	return nil, nil
}

func (l *LinkedInAdapter) PostContent(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	// TODO: Implement
	return nil, nil
}

func (l *LinkedInAdapter) RefreshTokenIfNeeded(ctx context.Context, token *social.PlatformToken) (*social.PlatformToken, error) {
	// TODO: Implement
	return token, nil
}

func (l *LinkedInAdapter) GetAccountInfo(ctx context.Context, token *social.PlatformToken) (*social.AccountInfo, error) {
	// TODO: Implement
	return nil, nil
}

func (l *LinkedInAdapter) ValidateToken(ctx context.Context, token *social.PlatformToken) (bool, error) {
	// TODO: Implement
	return false, nil
}

func (l *LinkedInAdapter) RevokeToken(ctx context.Context, token *social.PlatformToken) error {
	// TODO: Implement
	return nil
}

func (l *LinkedInAdapter) GetRateLimits(ctx context.Context, token *social.PlatformToken) (*social.RateLimitInfo, error) {
	return &social.RateLimitInfo{
		Limit:     100,
		Remaining: 99,
		ResetAt:   time.Now().Add(24 * time.Hour),
	}, nil
}
