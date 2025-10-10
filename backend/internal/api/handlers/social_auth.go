// path: backend/internal/api/handlers/social_auth.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

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
func (h *SocialAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	platform := social.PlatformType(chi.URLParam(r, "platform"))
	userID := r.Context().Value("user_id").(int64)

	redirectURI := "https://yourdomain.com/api/social/auth/" + string(platform) + "/callback"

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

// GET /api/social/auth/{platform}/callback
func (h *SocialAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	platform := social.PlatformType(chi.URLParam(r, "platform"))
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	userID := r.Context().Value("user_id").(int64)

	redirectURI := "https://yourdomain.com/api/social/auth/" + string(platform) + "/callback"

	token, err := h.socialService.HandleOAuthCallback(r.Context(), platform, code, state, redirectURI, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "OAuth callback failed", err)
		return
	}

	// Redirect to frontend with success
	http.Redirect(w, r, "https://yourdomain.com/dashboard/integrations?success=true&platform="+string(platform)+"&token_id="+fmt.Sprint(token.ID), http.StatusTemporaryRedirect)
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
