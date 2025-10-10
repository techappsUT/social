// path: backend/internal/social/adapters/twitter_adapter.go
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/techappsUT/social-queue/internal/social"
)

type TwitterAdapter struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewTwitterAdapter(clientID, clientSecret string) *TwitterAdapter {
	return &TwitterAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TwitterAdapter) GetPlatformName() social.PlatformType {
	return social.PlatformTwitter
}

func (t *TwitterAdapter) GetCapabilities() *social.PlatformCapabilities {
	return &social.PlatformCapabilities{
		SupportsImages:        true,
		SupportsVideos:        true,
		SupportsMultipleMedia: true,
		SupportsScheduling:    false,
		SupportsHashtags:      true,
		MaxTextLength:         280,
		MaxMediaCount:         4,
		MaxVideoSizeMB:        512,
	}
}

func (t *TwitterAdapter) AuthRedirect(ctx context.Context, state string, redirectURI string) (string, error) {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", t.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "tweet.read tweet.write users.read offline.access")
	params.Set("state", state)
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "plain")

	authURL := fmt.Sprintf("https://twitter.com/i/oauth2/authorize?%s", params.Encode())
	return authURL, nil
}

func (t *TwitterAdapter) HandleOAuthCallback(ctx context.Context, code string, redirectURI string) (*social.OAuthTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", t.clientID)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", "challenge")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.clientID, t.clientSecret)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("twitter oauth failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := t.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	return &social.OAuthTokenResponse{
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     tokenResp.RefreshToken,
		ExpiresIn:        tokenResp.ExpiresIn,
		ExpiresAt:        time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:        tokenResp.TokenType,
		Scope:            tokenResp.Scope,
		PlatformUserID:   userInfo.PlatformUserID, // FIXED: Use PlatformUserID instead of ID
		PlatformUsername: userInfo.Username,
	}, nil
}

func (t *TwitterAdapter) PostContent(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	payload := map[string]interface{}{
		"text": content.Text,
	}

	// Add media if present (simplified - actual implementation would upload media first)
	if len(content.MediaURLs) > 0 {
		// Media upload flow would go here
		// 1. Upload media using POST media/upload
		// 2. Get media_ids
		// 3. Attach to tweet
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/tweets", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &social.PostResult{
		PlatformPostID: result.Data.ID,
		URL:            fmt.Sprintf("https://twitter.com/%s/status/%s", token.PlatformUsername, result.Data.ID),
		PublishedAt:    time.Now(),
		Success:        true,
	}, nil
}

func (t *TwitterAdapter) RefreshTokenIfNeeded(ctx context.Context, token *social.PlatformToken) (*social.PlatformToken, error) {
	// Check if token needs refresh (refresh 5 minutes before expiry)
	if time.Until(token.ExpiresAt) > 5*time.Minute {
		return token, nil
	}

	data := url.Values{}
	data.Set("refresh_token", token.RefreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", t.clientID)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.clientID, t.clientSecret)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	// Update token
	token.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		token.RefreshToken = tokenResp.RefreshToken
	}
	token.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return token, nil
}

func (t *TwitterAdapter) GetAccountInfo(ctx context.Context, token *social.PlatformToken) (*social.AccountInfo, error) {
	return t.getUserInfo(ctx, token.AccessToken)
}

func (t *TwitterAdapter) ValidateToken(ctx context.Context, token *social.PlatformToken) (bool, error) {
	_, err := t.getUserInfo(ctx, token.AccessToken)
	return err == nil, err
}

func (t *TwitterAdapter) RevokeToken(ctx context.Context, token *social.PlatformToken) error {
	data := url.Values{}
	data.Set("token", token.AccessToken)
	data.Set("client_id", t.clientID)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitter.com/2/oauth2/revoke", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.clientID, t.clientSecret)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke token: status %d", resp.StatusCode)
	}

	return nil
}

func (t *TwitterAdapter) GetRateLimits(ctx context.Context, token *social.PlatformToken) (*social.RateLimitInfo, error) {
	// Twitter returns rate limit info in response headers
	// This is a simplified version
	return &social.RateLimitInfo{
		Limit:     300,
		Remaining: 299,
		ResetAt:   time.Now().Add(15 * time.Minute),
	}, nil
}

func (t *TwitterAdapter) getUserInfo(ctx context.Context, accessToken string) (*social.AccountInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitter.com/2/users/me?user.fields=profile_image_url,public_metrics,verified,description", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ID              string `json:"id"`
			Username        string `json:"username"`
			Name            string `json:"name"`
			ProfileImageURL string `json:"profile_image_url"`
			Verified        bool   `json:"verified"`
			Description     string `json:"description"`
			PublicMetrics   struct {
				FollowersCount int64 `json:"followers_count"`
				FollowingCount int64 `json:"following_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &social.AccountInfo{
		PlatformUserID:  result.Data.ID,
		Username:        result.Data.Username,
		DisplayName:     result.Data.Name,
		ProfileImageURL: result.Data.ProfileImageURL,
		FollowersCount:  result.Data.PublicMetrics.FollowersCount,
		FollowingCount:  result.Data.PublicMetrics.FollowingCount,
		IsVerified:      result.Data.Verified,
		Bio:             result.Data.Description,
	}, nil
}
