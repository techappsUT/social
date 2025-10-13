// ============================================================================
// FILE: backend/internal/adapters/social/twitter/client.go
// FIXED VERSION - Populates PlatformUserID in Token
// ============================================================================
package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/techappsUT/social-queue/internal/adapters/social"
)

const (
	twitterAuthURL  = "https://twitter.com/i/oauth2/authorize"
	twitterTokenURL = "https://api.twitter.com/2/oauth2/token"
	twitterAPIURL   = "https://api.twitter.com/2"
	maxRetries      = 3
	charLimit       = 280
)

type TwitterAdapter struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

func NewTwitterAdapter(clientID, clientSecret, redirectURI string) *TwitterAdapter {
	return &TwitterAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthURL generates the OAuth authorization URL with PKCE
func (t *TwitterAdapter) GetAuthURL(state string, scopes []string) string {
	if len(scopes) == 0 {
		scopes = []string{"tweet.read", "tweet.write", "users.read", "offline.access"}
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", t.clientID)
	params.Set("redirect_uri", t.redirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)
	params.Set("code_challenge", "challenge")
	params.Set("code_challenge_method", "plain")

	return fmt.Sprintf("%s?%s", twitterAuthURL, params.Encode())
}

// ExchangeCode exchanges authorization code for access token
func (t *TwitterAdapter) ExchangeCode(ctx context.Context, code string) (*social.Token, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", t.clientID)
	data.Set("redirect_uri", t.redirectURI)
	data.Set("code_verifier", "challenge")

	req, err := http.NewRequestWithContext(ctx, "POST", twitterTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.clientID, t.clientSecret)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("twitter oauth failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Get user ID from Twitter API
	platformUserID, err := t.getUserID(ctx, tokenResp.AccessToken)
	if err != nil {
		// Fallback to a placeholder if user ID fetch fails
		platformUserID = "twitter_" + tokenResp.AccessToken[:10]
	}

	return &social.Token{
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      &expiresAt,
		Scopes:         strings.Split(tokenResp.Scope, " "),
		PlatformUserID: platformUserID, // ADDED
	}, nil
}

// getUserID fetches the authenticated user's ID
func (t *TwitterAdapter) getUserID(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", twitterAPIURL+"/users/me", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user info (%d)", resp.StatusCode)
	}

	var userResp struct {
		Data struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return "", err
	}

	return userResp.Data.ID, nil
}

// RefreshToken refreshes an expired access token
func (t *TwitterAdapter) RefreshToken(ctx context.Context, refreshToken string) (*social.Token, error) {
	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", t.clientID)

	req, err := http.NewRequestWithContext(ctx, "POST", twitterTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.clientID, t.clientSecret)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	platformUserID, _ := t.getUserID(ctx, tokenResp.AccessToken)

	return &social.Token{
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		ExpiresAt:      &expiresAt,
		Scopes:         strings.Split(tokenResp.Scope, " "),
		PlatformUserID: platformUserID, // ADDED
	}, nil
}

// PublishPost publishes a tweet
func (t *TwitterAdapter) PublishPost(ctx context.Context, token *social.Token, content *social.PostContent) (*social.PublishResult, error) {
	if len(content.Text) > charLimit {
		return nil, fmt.Errorf("tweet exceeds %d character limit", charLimit)
	}

	payload := map[string]interface{}{
		"text": content.Text,
	}

	if len(content.MediaURLs) > 0 {
		mediaIDs, err := t.uploadMedia(ctx, token.AccessToken, content.MediaURLs)
		if err != nil {
			return nil, fmt.Errorf("media upload failed: %w", err)
		}
		payload["media"] = map[string]interface{}{
			"media_ids": mediaIDs,
		}
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := t.createTweet(ctx, token.AccessToken, payload)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if strings.Contains(err.Error(), "429") {
			time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
			continue
		}
		break
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (t *TwitterAdapter) createTweet(ctx context.Context, accessToken string, payload map[string]interface{}) (*social.PublishResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", twitterAPIURL+"/tweets", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tweet creation failed (%d): %s", resp.StatusCode, string(body))
	}

	var tweetResp struct {
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tweetResp); err != nil {
		return nil, err
	}

	return &social.PublishResult{
		PlatformPostID: tweetResp.Data.ID,
		URL:            fmt.Sprintf("https://twitter.com/i/status/%s", tweetResp.Data.ID),
		PublishedAt:    time.Now(),
	}, nil
}

func (t *TwitterAdapter) uploadMedia(ctx context.Context, accessToken string, mediaURLs []string) ([]string, error) {
	// Placeholder - implement actual media upload
	return []string{}, nil
}

func (t *TwitterAdapter) ValidateToken(ctx context.Context, token *social.Token) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", twitterAPIURL+"/users/me", nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (t *TwitterAdapter) GetPostAnalytics(ctx context.Context, token *social.Token, postID string) (*social.Analytics, error) {
	endpoint := fmt.Sprintf("%s/tweets/%s?tweet.fields=public_metrics", twitterAPIURL, postID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get analytics (%d)", resp.StatusCode)
	}

	var analyticsResp struct {
		Data struct {
			PublicMetrics struct {
				Likes       int `json:"like_count"`
				Retweets    int `json:"retweet_count"`
				Replies     int `json:"reply_count"`
				Impressions int `json:"impression_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&analyticsResp); err != nil {
		return nil, err
	}

	return &social.Analytics{
		Impressions: analyticsResp.Data.PublicMetrics.Impressions,
		Engagements: analyticsResp.Data.PublicMetrics.Likes +
			analyticsResp.Data.PublicMetrics.Retweets +
			analyticsResp.Data.PublicMetrics.Replies,
		Likes:    analyticsResp.Data.PublicMetrics.Likes,
		Shares:   analyticsResp.Data.PublicMetrics.Retweets,
		Comments: analyticsResp.Data.PublicMetrics.Replies,
	}, nil
}
