// path: backend/internal/social/adapters/facebook_adapter.go
package adapters

import (
	"context"
	"time"

	"github.com/techappsUT/social-queue/internal/social"
)

type FacebookAdapter struct {
	appID     string
	appSecret string
}

func NewFacebookAdapter(appID, appSecret string) *FacebookAdapter {
	return &FacebookAdapter{
		appID:     appID,
		appSecret: appSecret,
	}
}

func (f *FacebookAdapter) GetPlatformName() social.PlatformType {
	return social.PlatformFacebook
}

func (f *FacebookAdapter) GetCapabilities() *social.PlatformCapabilities {
	return &social.PlatformCapabilities{
		SupportsImages:        true,
		SupportsVideos:        true,
		SupportsMultipleMedia: true,
		SupportsScheduling:    true,
		SupportsHashtags:      true,
		MaxTextLength:         63206,
		MaxMediaCount:         10,
		MaxVideoSizeMB:        4096,
	}
}

func (f *FacebookAdapter) AuthRedirect(ctx context.Context, state string, redirectURI string) (string, error) {
	// TODO: Implement Facebook OAuth
	return "", nil
}

func (f *FacebookAdapter) HandleOAuthCallback(ctx context.Context, code string, redirectURI string) (*social.OAuthTokenResponse, error) {
	// TODO: Implement
	return nil, nil
}

func (f *FacebookAdapter) PostContent(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	// TODO: Implement
	return nil, nil
}

func (f *FacebookAdapter) RefreshTokenIfNeeded(ctx context.Context, token *social.PlatformToken) (*social.PlatformToken, error) {
	// TODO: Implement
	return token, nil
}

func (f *FacebookAdapter) GetAccountInfo(ctx context.Context, token *social.PlatformToken) (*social.AccountInfo, error) {
	// TODO: Implement
	return nil, nil
}

func (f *FacebookAdapter) ValidateToken(ctx context.Context, token *social.PlatformToken) (bool, error) {
	// TODO: Implement
	return false, nil
}

func (f *FacebookAdapter) RevokeToken(ctx context.Context, token *social.PlatformToken) error {
	// TODO: Implement
	return nil
}

func (f *FacebookAdapter) GetRateLimits(ctx context.Context, token *social.PlatformToken) (*social.RateLimitInfo, error) {
	return &social.RateLimitInfo{
		Limit:     200,
		Remaining: 199,
		ResetAt:   time.Now().Add(1 * time.Hour),
	}, nil
}
