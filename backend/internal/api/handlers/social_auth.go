// path: backend/internal/api/handlers/social_auth.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/techappsUT/social-queue/internal/social"
	"github.com/techappsUT/social-queue/pkg/response"

	"github.com/go-chi/chi/v5"
)

type SocialAuthHandler struct {
	socialService *social.Service
}

func NewSocialAuthHandler(socialService *social.Service) *SocialAuthHandler {
	return &SocialAuthHandler{
		socialService: socialService,
	}
}

// GET /api/social/auth/{platform}/redirect
// func (h *SocialAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
// 	platform := social.PlatformType(chi.URLParam(r, "platform"))
// 	userID := r.Context().Value("user_id").(int64)

// 	redirectURI := "https://yourdomain.com/api/social/auth/" + string(platform) + "/callback"

// 	authURL, state, err := h.socialService.InitiateOAuth(r.Context(), platform, userID, redirectURI)
// 	if err != nil {
// 		response.Error(w, http.StatusInternalServerError, "Failed to initiate OAuth", err)
// 		return
// 	}

// 	response.JSON(w, http.StatusOK, map[string]interface{}{
// 		"auth_url": authURL,
// 		"state":    state,
// 	})
// }

// // GET /api/social/auth/{platform}/callback
// func (h *SocialAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
// 	platform := social.PlatformType(chi.URLParam(r, "platform"))
// 	code := r.URL.Query().Get("code")
// 	state := r.URL.Query().Get("state")
// 	userID := r.Context().Value("user_id").(int64)

// 	redirectURI := "https://yourdomain.com/api/social/auth/" + string(platform) + "/callback"

// 	token, err := h.socialService.HandleOAuthCallback(r.Context(), platform, code, state, redirectURI, userID)
// 	if err != nil {
// 		response.Error(w, http.StatusInternalServerError, "OAuth callback failed", err)
// 		return
// 	}

// 	// Redirect to frontend with success
// 	http.Redirect(w, r, "https://yourdomain.com/dashboard/integrations?success=true&platform="+string(platform)+"&token_id="+fmt.Sprint(token.ID), http.StatusTemporaryRedirect)
// }

func (h *SocialAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	platform := social.PlatformType(chi.URLParam(r, "platform"))
	userID := r.Context().Value("user_id").(int64)

	// Use environment variable for redirect URI
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
	}

	redirectURI := fmt.Sprintf("%s/api/v2/social/auth/%s/callback", baseURL, string(platform))

	authURL, state, err := h.socialService.InitiateOAuth(r.Context(), platform, userID, redirectURI)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to initiate OAuth", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"auth_url": authURL,
		"state":    state,
	})
}

func (h *SocialAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	platform := social.PlatformType(chi.URLParam(r, "platform"))

	// Check for errors from Facebook
	if errCode := r.URL.Query().Get("error_code"); errCode != "" {
		errMsg := r.URL.Query().Get("error_message")

		// Redirect to frontend with error
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}

		redirectURL := fmt.Sprintf("%s/dashboard/integrations?error=oauth_failed&platform=%s&message=%s",
			frontendURL, string(platform), url.QueryEscape(errMsg))

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	userID := r.Context().Value("user_id").(int64)

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
	}

	redirectURI := fmt.Sprintf("%s/api/v2/social/auth/%s/callback", baseURL, string(platform))

	token, err := h.socialService.HandleOAuthCallback(r.Context(), platform, code, state, redirectURI, userID)
	if err != nil {
		// Redirect to frontend with error
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}

		redirectURL := fmt.Sprintf("%s/dashboard/integrations?error=callback_failed&platform=%s",
			frontendURL, string(platform))

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to frontend with success
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	http.Redirect(w, r, fmt.Sprintf("%s/dashboard/integrations?success=true&platform=%s&token_id=%d",
		frontendURL, string(platform), token.ID), http.StatusTemporaryRedirect)
}

// POST /api/social/publish
func (h *SocialAuthHandler) PublishContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenID   int64                  `json:"token_id"`
		Text      string                 `json:"text"`
		MediaURLs []string               `json:"media_urls,omitempty"`
		Options   map[string]interface{} `json:"options,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	content := &social.PostContent{
		Text:      req.Text,
		MediaURLs: req.MediaURLs,
		Options:   req.Options,
	}

	result, err := h.socialService.PublishContent(r.Context(), req.TokenID, content)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to publish content", err)
		return
	}

	response.JSON(w, http.StatusOK, result)
}
