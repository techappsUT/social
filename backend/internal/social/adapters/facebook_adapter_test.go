// path: backend/internal/social/adapters/facebook_adapter_test.go
package adapters

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/techappsUT/social-queue/internal/social"
)

func TestFacebookAdapter_GetPlatformName(t *testing.T) {
	adapter := NewFacebookAdapter("test_app_id", "test_secret")

	if adapter.GetPlatformName() != social.PlatformFacebook {
		t.Errorf("Expected platform Facebook, got %s", adapter.GetPlatformName())
	}
}

func TestFacebookAdapter_GetCapabilities(t *testing.T) {
	adapter := NewFacebookAdapter("test_app_id", "test_secret")
	caps := adapter.GetCapabilities()

	if !caps.SupportsImages {
		t.Error("Expected Facebook to support images")
	}

	if !caps.SupportsScheduling {
		t.Error("Expected Facebook to support scheduling")
	}

	if caps.MaxTextLength != 63206 {
		t.Errorf("Expected max text length 63206, got %d", caps.MaxTextLength)
	}
}

func TestFacebookAdapter_AuthRedirect(t *testing.T) {
	adapter := NewFacebookAdapter("test_app_id", "test_secret")

	ctx := context.Background()
	authURL, err := adapter.AuthRedirect(ctx, "test_state", "http://localhost/callback")

	if err != nil {
		t.Fatalf("AuthRedirect failed: %v", err)
	}

	if authURL == "" {
		t.Error("Expected non-empty auth URL")
	}

	// Verify URL contains required parameters
	if !contains(authURL, "client_id=test_app_id") {
		t.Error("Auth URL missing client_id")
	}

	if !contains(authURL, "state=test_state") {
		t.Error("Auth URL missing state")
	}
}

func TestFacebookAdapter_HandleOAuthCallback(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock token endpoint
		if contains(r.URL.Path, "/oauth/access_token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock_access_token",
				"token_type":   "bearer",
				"expires_in":   5184000,
			})
			return
		}

		// Mock user info endpoint
		if contains(r.URL.Path, "/me") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    "123456789",
				"name":  "Test User",
				"email": "test@example.com",
				"picture": map[string]interface{}{
					"data": map[string]interface{}{
						"url": "https://example.com/pic.jpg",
					},
				},
			})
			return
		}

		// Mock pages endpoint
		if contains(r.URL.Path, "/accounts") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{},
			})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	adapter := NewFacebookAdapter("test_app_id", "test_secret")
	adapter.graphAPIURL = server.URL

	ctx := context.Background()
	tokenResp, err := adapter.HandleOAuthCallback(ctx, "mock_code", "http://localhost/callback")

	if err != nil {
		t.Fatalf("HandleOAuthCallback failed: %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Error("Expected access token")
	}

	if tokenResp.PlatformUserID != "123456789" {
		t.Errorf("Expected user ID 123456789, got %s", tokenResp.PlatformUserID)
	}
}

func TestFacebookAdapter_ValidateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"is_valid":   true,
				"expires_at": time.Now().Add(24 * time.Hour).Unix(),
			},
		})
	}))
	defer server.Close()

	adapter := NewFacebookAdapter("test_app_id", "test_secret")
	adapter.graphAPIURL = server.URL

	token := &social.PlatformToken{
		AccessToken: "test_token",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	isValid, err := adapter.ValidateToken(context.Background(), token)

	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if !isValid {
		t.Error("Expected token to be valid")
	}
}

func TestFacebookWebhookHandler_VerifySignature(t *testing.T) {
	handler := NewFacebookWebhookHandler("test_secret", "test_verify_token")

	body := []byte(`{"test": "data"}`)

	// Generate valid signature
	mac := hmac.New(sha256.New, []byte("test_secret"))
	mac.Write(body)
	validSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !handler.verifySignature(body, validSignature) {
		t.Error("Expected signature to be valid")
	}

	invalidSignature := "sha256=invalid"
	if handler.verifySignature(body, invalidSignature) {
		t.Error("Expected signature to be invalid")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
