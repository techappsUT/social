// path: backend/internal/handlers/social.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/social"
)

type SocialHandler struct {
	socialService *social.Service
}

func NewSocialHandler(socialService *social.Service) *SocialHandler {
	return &SocialHandler{
		socialService: socialService,
	}
}

func (h *SocialHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	if h.socialService == nil {
		respondError(w, http.StatusServiceUnavailable, "Social features not yet configured")
		return
	}

	platform := social.PlatformType(chi.URLParam(r, "platform"))
	userID := int64(1) // TODO: Get from auth context

	redirectURI := fmt.Sprintf("http://localhost:8000/api/social/auth/%s/callback", platform)

	authURL, state, err := h.socialService.InitiateOAuth(r.Context(), platform, userID, redirectURI)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initiate OAuth: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"authUrl": authURL,
		"state":   state,
	})
}

func (h *SocialHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	if h.socialService == nil {
		respondError(w, http.StatusServiceUnavailable, "Social features not yet configured")
		return
	}

	platform := social.PlatformType(chi.URLParam(r, "platform"))
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	userID := int64(1) // TODO: Get from auth context

	redirectURI := fmt.Sprintf("http://localhost:8000/api/social/auth/%s/callback", platform)

	token, err := h.socialService.HandleOAuthCallback(r.Context(), platform, code, state, redirectURI, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("OAuth callback failed: %v", err))
		return
	}

	frontendURL := fmt.Sprintf("http://localhost:3000/dashboard/integrations?success=true&platform=%s", platform)
	if token != nil {
		frontendURL += fmt.Sprintf("&token_id=%d", token.ID)
	}

	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

func (h *SocialHandler) PublishContent(w http.ResponseWriter, r *http.Request) {
	if h.socialService == nil {
		respondError(w, http.StatusServiceUnavailable, "Social features not yet configured")
		return
	}

	var req struct {
		TokenID   int64                  `json:"token_id"`
		Text      string                 `json:"text"`
		MediaURLs []string               `json:"media_urls,omitempty"`
		Options   map[string]interface{} `json:"options,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	content := &social.PostContent{
		Text:      req.Text,
		MediaURLs: req.MediaURLs,
		Options:   req.Options,
	}

	result, err := h.socialService.PublishContent(r.Context(), req.TokenID, content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to publish: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, result)
}
