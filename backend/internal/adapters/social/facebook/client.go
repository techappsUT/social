// ============================================================================
// FILE: backend/internal/adapters/social/facebook/client.go
// Complete Facebook OAuth + Graph API Publishing Implementation
// ============================================================================
package facebook

import (
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
	facebookAuthURL  = "https://www.facebook.com/v18.0/dialog/oauth"
	facebookTokenURL = "https://graph.facebook.com/v18.0/oauth/access_token"
	facebookGraphURL = "https://graph.facebook.com/v18.0"
	charLimit        = 63206
	maxRetries       = 3
)

type FacebookAdapter struct {
	appID       string
	appSecret   string
	redirectURI string
	httpClient  *http.Client
}

func NewFacebookAdapter(appID, appSecret, redirectURI string) *FacebookAdapter {
	return &FacebookAdapter{
		appID:       appID,
		appSecret:   appSecret,
		redirectURI: redirectURI,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthURL generates the OAuth authorization URL
func (f *FacebookAdapter) GetAuthURL(state string, scopes []string) string {
	if len(scopes) == 0 {
		scopes = []string{"pages_manage_posts", "pages_read_engagement", "pages_show_list"}
	}

	params := url.Values{}
	params.Set("client_id", f.appID)
	params.Set("redirect_uri", f.redirectURI)
	params.Set("scope", strings.Join(scopes, ","))
	params.Set("state", state)
	params.Set("response_type", "code")

	return fmt.Sprintf("%s?%s", facebookAuthURL, params.Encode())
}

// ExchangeCode exchanges authorization code for access token
func (f *FacebookAdapter) ExchangeCode(ctx context.Context, code string) (*social.Token, error) {
	params := url.Values{}
	params.Set("client_id", f.appID)
	params.Set("client_secret", f.appSecret)
	params.Set("redirect_uri", f.redirectURI)
	params.Set("code", code)

	endpoint := fmt.Sprintf("%s?%s", facebookTokenURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook oauth failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Exchange for long-lived token (60 days)
	longLivedToken, expiresAt, err := f.exchangeForLongLivedToken(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get long-lived token: %w", err)
	}

	return &social.Token{
		AccessToken: longLivedToken,
		ExpiresAt:   expiresAt,
		Scopes:      []string{}, // Facebook doesn't return scopes in response
	}, nil
}

// exchangeForLongLivedToken exchanges short-lived token for long-lived token
func (f *FacebookAdapter) exchangeForLongLivedToken(ctx context.Context, shortToken string) (string, *time.Time, error) {
	params := url.Values{}
	params.Set("grant_type", "fb_exchange_token")
	params.Set("client_id", f.appID)
	params.Set("client_secret", f.appSecret)
	params.Set("fb_exchange_token", shortToken)

	endpoint := fmt.Sprintf("%s?%s", facebookTokenURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("long-lived token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", nil, err
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, &expiresAt, nil
}

// RefreshToken - Facebook long-lived tokens last 60 days, no refresh mechanism
func (f *FacebookAdapter) RefreshToken(ctx context.Context, refreshToken string) (*social.Token, error) {
	return nil, fmt.Errorf("facebook tokens cannot be refreshed, user must re-authenticate")
}

// PublishPost creates a Facebook page post
func (f *FacebookAdapter) PublishPost(ctx context.Context, token *social.Token, content *social.PostContent) (*social.PublishResult, error) {
	// Validate content
	if len(content.Text) > charLimit {
		return nil, fmt.Errorf("post exceeds %d character limit", charLimit)
	}

	// Get user's pages
	pages, err := f.getUserPages(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no Facebook pages found")
	}

	// Use first page (in production, let user select)
	page := pages[0]

	// Build post payload
	payload := map[string]interface{}{
		"message": content.Text,
	}

	// Handle media
	if len(content.MediaURLs) > 0 {
		// For simplicity, using link post
		// Production should handle photo/video uploads
		payload["link"] = content.MediaURLs[0]
	}

	// Create post with retry logic
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := f.createPagePost(ctx, page.ID, page.AccessToken, payload)
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

// createPagePost makes the API call to create a page post
func (f *FacebookAdapter) createPagePost(ctx context.Context, pageID, pageAccessToken string, payload map[string]interface{}) (*social.PublishResult, error) {
	// Build form data
	data := url.Values{}
	for key, value := range payload {
		data.Set(key, fmt.Sprintf("%v", value))
	}

	endpoint := fmt.Sprintf("%s/%s/feed", facebookGraphURL, pageID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.Query().Set("access_token", pageAccessToken)

	// Add access token to URL
	req.URL.RawQuery = "access_token=" + url.QueryEscape(pageAccessToken)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("post creation failed (%d): %s", resp.StatusCode, string(body))
	}

	var postResp struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&postResp); err != nil {
		return nil, err
	}

	return &social.PublishResult{
		PlatformPostID: postResp.ID,
		URL:            fmt.Sprintf("https://www.facebook.com/%s", postResp.ID),
		PublishedAt:    time.Now(),
	}, nil
}

// getUserPages retrieves user's Facebook pages
func (f *FacebookAdapter) getUserPages(ctx context.Context, accessToken string) ([]FacebookPage, error) {
	endpoint := fmt.Sprintf("%s/me/accounts?access_token=%s", facebookGraphURL, url.QueryEscape(accessToken))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get pages (%d)", resp.StatusCode)
	}

	var pagesResp struct {
		Data []FacebookPage `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pagesResp); err != nil {
		return nil, err
	}

	return pagesResp.Data, nil
}

// ValidateToken checks if the access token is still valid
func (f *FacebookAdapter) ValidateToken(ctx context.Context, token *social.Token) (bool, error) {
	endpoint := fmt.Sprintf("%s/me?access_token=%s", facebookGraphURL, url.QueryEscape(token.AccessToken))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetPostAnalytics retrieves analytics for a Facebook post
func (f *FacebookAdapter) GetPostAnalytics(ctx context.Context, token *social.Token, postID string) (*social.Analytics, error) {
	endpoint := fmt.Sprintf("%s/%s?fields=likes.summary(true),comments.summary(true),shares&access_token=%s",
		facebookGraphURL, postID, url.QueryEscape(token.AccessToken))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get analytics (%d)", resp.StatusCode)
	}

	var analyticsResp struct {
		Likes struct {
			Summary struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		} `json:"likes"`
		Comments struct {
			Summary struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		} `json:"comments"`
		Shares struct {
			Count int `json:"count"`
		} `json:"shares"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&analyticsResp); err != nil {
		return nil, err
	}

	return &social.Analytics{
		Likes:    analyticsResp.Likes.Summary.TotalCount,
		Comments: analyticsResp.Comments.Summary.TotalCount,
		Shares:   analyticsResp.Shares.Count,
		Engagements: analyticsResp.Likes.Summary.TotalCount +
			analyticsResp.Comments.Summary.TotalCount +
			analyticsResp.Shares.Count,
	}, nil
}

type FacebookPage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccessToken string `json:"access_token"`
}
