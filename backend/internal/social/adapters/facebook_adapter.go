// path: backend/internal/social/adapters/facebook_adapter.go
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

// FacebookAdapter implements the SocialAdapter interface for Facebook
// IMPORTANT: Facebook requires different tokens for:
// - User profiles (personal posts, limited API access)
// - Pages (business posts, full Graph API access)
// - Instagram Business (requires Facebook Page connection)
type FacebookAdapter struct {
	appID       string
	appSecret   string
	httpClient  *http.Client
	apiVersion  string // Default: v19.0
	graphAPIURL string
}

// FacebookAccountType represents the type of Facebook account
type FacebookAccountType string

const (
	AccountTypeUser      FacebookAccountType = "user"
	AccountTypePage      FacebookAccountType = "page"
	AccountTypeInstagram FacebookAccountType = "instagram_business"
)

// NewFacebookAdapter creates a new Facebook adapter instance
func NewFacebookAdapter(appID, appSecret string) *FacebookAdapter {
	return &FacebookAdapter{
		appID:       appID,
		appSecret:   appSecret,
		apiVersion:  "v19.0", // Update to latest stable version
		graphAPIURL: "https://graph.facebook.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *FacebookAdapter) GetPlatformName() social.PlatformType {
	return social.PlatformFacebook
}

func (f *FacebookAdapter) GetCapabilities() *social.PlatformCapabilities {
	return &social.PlatformCapabilities{
		SupportsImages:        true,
		SupportsVideos:        true,
		SupportsMultipleMedia: true,  // Up to 10 images/videos in a single post
		SupportsScheduling:    true,  // Pages support scheduled_publish_time
		SupportsHashtags:      true,  // Instagram Business primarily
		MaxTextLength:         63206, // Facebook max character limit
		MaxMediaCount:         10,    // Max media items per post
		MaxVideoSizeMB:        4096,  // 4GB for videos
	}
}

// AuthRedirect generates the OAuth authorization URL
// Scope notes:
// - pages_show_list: Required to get user's pages
// - pages_read_engagement: Read page content
// - pages_manage_posts: Create/edit page posts
// - instagram_basic: Access Instagram business account
// - instagram_content_publish: Post to Instagram
// - business_management: Required for Instagram Business
func (f *FacebookAdapter) AuthRedirect(ctx context.Context, state string, redirectURI string) (string, error) {
	params := url.Values{}
	params.Set("client_id", f.appID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("response_type", "code")

	// Request comprehensive permissions
	// NOTE: Some permissions require App Review for production
	scopes := []string{
		"public_profile",            // Basic profile info (auto-approved)
		"email",                     // User email (auto-approved)
		"pages_show_list",           // List user's pages (requires review)
		"pages_read_engagement",     // Read page data (requires review)
		"pages_manage_posts",        // Publish to pages (requires review)
		"pages_read_user_content",   // Read page posts (requires review)
		"instagram_basic",           // Instagram account access (requires review)
		"instagram_content_publish", // Post to Instagram (requires review)
		"business_management",       // Business account access (requires review)
	}
	params.Set("scope", strings.Join(scopes, ","))

	// Enable re-authentication and re-request declined permissions
	params.Set("auth_type", "rerequest")

	authURL := fmt.Sprintf("https://www.facebook.com/%s/dialog/oauth?%s", f.apiVersion, params.Encode())
	return authURL, nil
}

// HandleOAuthCallback processes the OAuth callback and exchanges code for tokens
func (f *FacebookAdapter) HandleOAuthCallback(ctx context.Context, code string, redirectURI string) (*social.OAuthTokenResponse, error) {
	// Step 1: Exchange code for short-lived user access token
	params := url.Values{}
	params.Set("client_id", f.appID)
	params.Set("client_secret", f.appSecret)
	params.Set("redirect_uri", redirectURI)
	params.Set("code", code)

	tokenURL := fmt.Sprintf("%s/%s/oauth/access_token?%s", f.graphAPIURL, f.apiVersion, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook oauth failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"` // Short-lived: ~2 hours
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Step 2: Exchange short-lived token for long-lived token (60 days)
	longLivedToken, expiresIn, err := f.exchangeForLongLivedToken(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get long-lived token: %w", err)
	}

	// Step 3: Get user info
	userInfo, err := f.getUserInfo(ctx, longLivedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Step 4: Get user's pages (if permissions granted)
	pages, err := f.getUserPages(ctx, longLivedToken)
	if err != nil {
		// Log but don't fail - user might not have pages or permission denied
		fmt.Printf("Warning: Could not fetch user pages: %v\n", err)
		pages = []FacebookPage{}
	}

	// Store pages in Extra field for later use
	extra := map[string]interface{}{
		"account_type": string(AccountTypeUser),
		"pages":        pages,
	}

	// NOTE: Facebook doesn't provide refresh tokens - tokens are long-lived (60 days)
	// Apps must re-prompt users when tokens expire
	return &social.OAuthTokenResponse{
		AccessToken:      longLivedToken,
		RefreshToken:     "", // Facebook doesn't use refresh tokens
		ExpiresIn:        expiresIn,
		ExpiresAt:        time.Now().Add(time.Duration(expiresIn) * time.Second),
		TokenType:        tokenResp.TokenType,
		Scope:            "", // Would need to parse from granted permissions
		PlatformUserID:   userInfo.PlatformUserID,
		PlatformUsername: userInfo.Username,
		Extra:            extra,
	}, nil
}

// exchangeForLongLivedToken converts short-lived token to long-lived (60 days)
func (f *FacebookAdapter) exchangeForLongLivedToken(ctx context.Context, shortLivedToken string) (string, int64, error) {
	params := url.Values{}
	params.Set("grant_type", "fb_exchange_token")
	params.Set("client_id", f.appID)
	params.Set("client_secret", f.appSecret)
	params.Set("fb_exchange_token", shortLivedToken)

	tokenURL := fmt.Sprintf("%s/%s/oauth/access_token?%s", f.graphAPIURL, f.apiVersion, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return "", 0, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"` // ~5184000 seconds (60 days)
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, err
	}

	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

// FacebookPage represents a Facebook Page the user manages
type FacebookPage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	AccessToken string   `json:"access_token"` // Page access token (never expires!)
	Category    string   `json:"category"`
	Tasks       []string `json:"tasks"` // Permissions/tasks user can perform
}

// getUserPages fetches all pages the user manages
func (f *FacebookAdapter) getUserPages(ctx context.Context, accessToken string) ([]FacebookPage, error) {
	apiURL := fmt.Sprintf("%s/%s/me/accounts?fields=id,name,access_token,category,tasks",
		f.graphAPIURL, f.apiVersion)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch pages (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []FacebookPage `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// PostContent publishes content to Facebook
// IMPORTANT: Posting behavior depends on token type:
// - User token: Posts to user's profile (limited, not recommended for business)
// - Page token: Posts to Facebook Page (recommended, full features)
// - Instagram Business token: Posts to Instagram (requires Facebook Page connection)
func (f *FacebookAdapter) PostContent(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	// Determine account type from token metadata
	accountType := AccountTypeUser
	if token.Extra != nil {
		if at, ok := token.Extra["account_type"].(string); ok {
			accountType = FacebookAccountType(at)
		}
	}

	switch accountType {
	case AccountTypePage:
		return f.postToPage(ctx, token, content)
	case AccountTypeInstagram:
		return f.postToInstagram(ctx, token, content)
	default:
		return f.postToUserProfile(ctx, token, content)
	}
}

// postToUserProfile posts to user's personal Facebook profile
// NOTE: This has limited functionality and is deprecated for business use
func (f *FacebookAdapter) postToUserProfile(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	payload := map[string]interface{}{
		"message": content.Text,
	}

	// Add link if provided
	if content.Link != "" {
		payload["link"] = content.Link
	}

	// Note: User profile posts don't support multi-image uploads via API
	// Would require creating a photo album first

	return f.makePostRequest(ctx, token.AccessToken, "me/feed", payload)
}

// postToPage posts to a Facebook Page
// This is the recommended approach for business/brand posting
func (f *FacebookAdapter) postToPage(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	// Extract page ID from token metadata
	pageID, ok := token.Extra["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id not found in token metadata")
	}

	payload := map[string]interface{}{
		"message": content.Text,
	}

	// Add link if provided
	if content.Link != "" {
		payload["link"] = content.Link
	}

	// Handle scheduled posts
	if content.ScheduledAt != nil {
		// Page must have >1000 likes for scheduled posts
		payload["published"] = false
		payload["scheduled_publish_time"] = content.ScheduledAt.Unix()
	}

	// Handle media uploads
	if len(content.MediaURLs) > 0 {
		if len(content.MediaURLs) == 1 {
			// Single media post
			payload["url"] = content.MediaURLs[0]
		} else {
			// Multi-media post (carousel/album)
			// This requires creating attached_media with photo IDs
			// Implementation would need to:
			// 1. Upload each photo to page
			// 2. Get media IDs
			// 3. Attach to post
			return nil, fmt.Errorf("multi-media posts not yet implemented")
		}
	}

	endpoint := fmt.Sprintf("%s/feed", pageID)
	return f.makePostRequest(ctx, token.AccessToken, endpoint, payload)
}

// postToInstagram posts to Instagram Business account
// REQUIREMENTS:
// - Instagram Business account (not Creator account)
// - Instagram account connected to a Facebook Page
// - Specific image/video requirements (aspect ratio, size, etc.)
func (f *FacebookAdapter) postToInstagram(ctx context.Context, token *social.PlatformToken, content *social.PostContent) (*social.PostResult, error) {
	// Extract Instagram Business Account ID
	igAccountID, ok := token.Extra["instagram_business_account_id"].(string)
	if !ok || igAccountID == "" {
		return nil, fmt.Errorf("instagram_business_account_id not found in token metadata")
	}

	// Instagram requires a 2-step process:
	// Step 1: Create media container
	// Step 2: Publish the container

	if len(content.MediaURLs) == 0 {
		return nil, fmt.Errorf("instagram posts require at least one media item")
	}

	// Step 1: Create container
	containerPayload := map[string]interface{}{
		"image_url": content.MediaURLs[0], // Must be publicly accessible URL
		"caption":   content.Text,
	}

	containerEndpoint := fmt.Sprintf("%s/media", igAccountID)
	containerURL := fmt.Sprintf("%s/%s/%s", f.graphAPIURL, f.apiVersion, containerEndpoint)

	containerID, err := f.createInstagramContainer(ctx, token.AccessToken, containerURL, containerPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to create instagram container: %w", err)
	}

	// Step 2: Publish container
	publishPayload := map[string]interface{}{
		"creation_id": containerID,
	}

	publishEndpoint := fmt.Sprintf("%s/media_publish", igAccountID)
	return f.makePostRequest(ctx, token.AccessToken, publishEndpoint, publishPayload)
}

// createInstagramContainer creates a media container for Instagram
func (f *FacebookAdapter) createInstagramContainer(ctx context.Context, accessToken, apiURL string, payload map[string]interface{}) (string, error) {
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create container (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

// makePostRequest is a helper to make Graph API POST requests
func (f *FacebookAdapter) makePostRequest(ctx context.Context, accessToken, endpoint string, payload map[string]interface{}) (*social.PostResult, error) {
	apiURL := fmt.Sprintf("%s/%s/%s", f.graphAPIURL, f.apiVersion, endpoint)

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook post failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID     string `json:"id"`
		PostID string `json:"post_id"` // Sometimes returned instead of id
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	postID := result.ID
	if postID == "" {
		postID = result.PostID
	}

	return &social.PostResult{
		PlatformPostID: postID,
		URL:            fmt.Sprintf("https://www.facebook.com/%s", postID),
		PublishedAt:    time.Now(),
		Success:        true,
	}, nil
}

// RefreshTokenIfNeeded checks and refreshes token if needed
// NOTE: Facebook doesn't support traditional refresh tokens
// Long-lived tokens last 60 days and must be refreshed before expiry
func (f *FacebookAdapter) RefreshTokenIfNeeded(ctx context.Context, token *social.PlatformToken) (*social.PlatformToken, error) {
	// Check if token expires within 7 days
	if time.Until(token.ExpiresAt) > 7*24*time.Hour {
		return token, nil
	}

	// Refresh by exchanging current token for a new long-lived token
	newToken, expiresIn, err := f.exchangeForLongLivedToken(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update token
	token.AccessToken = newToken
	token.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	token.LastValidated = time.Now()

	return token, nil
}

// GetAccountInfo retrieves account information
func (f *FacebookAdapter) GetAccountInfo(ctx context.Context, token *social.PlatformToken) (*social.AccountInfo, error) {
	return f.getUserInfo(ctx, token.AccessToken)
}

// getUserInfo fetches user information from Facebook
func (f *FacebookAdapter) getUserInfo(ctx context.Context, accessToken string) (*social.AccountInfo, error) {
	apiURL := fmt.Sprintf("%s/%s/me?fields=id,name,email,picture.type(large)",
		f.graphAPIURL, f.apiVersion)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &social.AccountInfo{
		PlatformUserID:  result.ID,
		Username:        result.Email, // Facebook doesn't have public usernames
		DisplayName:     result.Name,
		ProfileImageURL: result.Picture.Data.URL,
		FollowersCount:  0, // Not available for user accounts via API
		FollowingCount:  0, // Not available for user accounts via API
		IsVerified:      false,
		Bio:             "",
	}, nil
}

// ValidateToken verifies if the token is still valid
func (f *FacebookAdapter) ValidateToken(ctx context.Context, token *social.PlatformToken) (bool, error) {
	// Use debug_token endpoint to check token validity
	apiURL := fmt.Sprintf("%s/%s/debug_token?input_token=%s&access_token=%s|%s",
		f.graphAPIURL, f.apiVersion, token.AccessToken, f.appID, f.appSecret)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var result struct {
		Data struct {
			IsValid   bool  `json:"is_valid"`
			ExpiresAt int64 `json:"expires_at"`
			Error     struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Data.IsValid, nil
}

// RevokeToken revokes the access token
func (f *FacebookAdapter) RevokeToken(ctx context.Context, token *social.PlatformToken) error {
	apiURL := fmt.Sprintf("%s/%s/me/permissions", f.graphAPIURL, f.apiVersion)

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to revoke token (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetRateLimits returns current rate limit status
// Facebook uses different rate limiting approaches:
// - User-level rate limits
// - App-level rate limits
// - Page-level rate limits
func (f *FacebookAdapter) GetRateLimits(ctx context.Context, token *social.PlatformToken) (*social.RateLimitInfo, error) {
	// Facebook doesn't expose rate limit info proactively
	// Rate limits are returned in response headers: X-Business-Use-Case-Usage, X-App-Usage, X-Page-Usage
	// For now, return conservative estimates
	return &social.RateLimitInfo{
		Limit:     200, // App-level: 200 calls per user per hour
		Remaining: 199,
		ResetAt:   time.Now().Add(1 * time.Hour),
	}, nil
}
