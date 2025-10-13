// ============================================================================
// FILE: backend/internal/handlers/social_handler.go
// ============================================================================
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/adapters/social"
	appSocial "github.com/techappsUT/social-queue/internal/application/social"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
	"github.com/techappsUT/social-queue/internal/middleware"
)

type SocialHandler struct {
	connectAccountUC *appSocial.ConnectAccountUseCase
	disconnectUC     *appSocial.DisconnectAccountUseCase
	refreshTokensUC  *appSocial.RefreshTokensUseCase
	listAccountsUC   *appSocial.ListAccountsUseCase
	publishPostUC    *appSocial.PublishPostUseCase
	getAnalyticsUC   *appSocial.GetAnalyticsUseCase
	adapters         map[socialDomain.Platform]social.Adapter
}

func NewSocialHandler(
	connectAccountUC *appSocial.ConnectAccountUseCase,
	disconnectUC *appSocial.DisconnectAccountUseCase,
	refreshTokensUC *appSocial.RefreshTokensUseCase,
	listAccountsUC *appSocial.ListAccountsUseCase,
	publishPostUC *appSocial.PublishPostUseCase,
	getAnalyticsUC *appSocial.GetAnalyticsUseCase,
	adapters map[socialDomain.Platform]social.Adapter,
) *SocialHandler {
	return &SocialHandler{
		connectAccountUC: connectAccountUC,
		disconnectUC:     disconnectUC,
		refreshTokensUC:  refreshTokensUC,
		listAccountsUC:   listAccountsUC,
		publishPostUC:    publishPostUC,
		getAnalyticsUC:   getAnalyticsUC,
		adapters:         adapters,
	}
}

// GetOAuthURL handles GET /api/v2/social/auth/:platform
func (h *SocialHandler) GetOAuthURL(w http.ResponseWriter, r *http.Request) {
	platform := chi.URLParam(r, "platform")
	state := r.URL.Query().Get("state")

	if state == "" {
		state = uuid.New().String()
	}

	// Get adapter
	adapter, ok := h.adapters[socialDomain.Platform(platform)]
	if !ok {
		respondError(w, http.StatusBadRequest, "unsupported platform")
		return
	}

	// Generate auth URL
	authURL := adapter.GetAuthURL(state, []string{})

	respondSuccess(w, map[string]string{
		"authUrl": authURL,
		"state":   state,
	})
}

// OAuthCallback handles GET /api/v2/social/auth/:platform/callback
func (h *SocialHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	platform := chi.URLParam(r, "platform")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		respondError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	// Return code and state to frontend for connection
	respondSuccess(w, map[string]string{
		"code":     code,
		"state":    state,
		"platform": platform,
	})
}

// ConnectAccount handles POST /api/v2/social/accounts
func (h *SocialHandler) ConnectAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input struct {
		TeamID   uuid.UUID `json:"teamId"`
		Platform string    `json:"platform"`
		Code     string    `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ucInput := appSocial.ConnectAccountInput{
		TeamID:   input.TeamID,
		UserID:   userID,
		Platform: socialDomain.Platform(input.Platform),
		Code:     input.Code,
	}

	output, err := h.connectAccountUC.Execute(r.Context(), ucInput)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondCreated(w, output)
}

// ListAccounts handles GET /api/v2/teams/:teamId/social/accounts
func (h *SocialHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	input := appSocial.ListAccountsInput{
		TeamID: teamID,
		UserID: userID,
	}

	output, err := h.listAccountsUC.Execute(r.Context(), input)
	if err != nil {
		if err.Error() == "access denied: not a team member" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "failed to list accounts")
		}
		return
	}

	respondSuccess(w, output)
}

// DisconnectAccount handles DELETE /api/v2/social/accounts/:id
func (h *SocialHandler) DisconnectAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	input := appSocial.DisconnectAccountInput{
		AccountID: accountID,
		UserID:    userID,
	}

	if err := h.disconnectUC.Execute(r.Context(), input); err != nil {
		if err.Error() == "access denied: not a team member" {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondNoContent(w)
}

// RefreshTokens handles POST /api/v2/social/accounts/:id/refresh
func (h *SocialHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	input := appSocial.RefreshTokensInput{
		AccountID: accountID,
	}

	output, err := h.refreshTokensUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// PublishPost handles POST /api/v2/social/accounts/:id/publish
func (h *SocialHandler) PublishPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var body struct {
		Content   string   `json:"content"`
		MediaURLs []string `json:"mediaUrls"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := appSocial.PublishPostInput{
		AccountID: accountID,
		UserID:    userID,
		Content:   body.Content,
		MediaURLs: body.MediaURLs,
	}

	output, err := h.publishPostUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// GetAnalytics handles GET /api/v2/social/accounts/:id/posts/:postId/analytics
func (h *SocialHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	postID := chi.URLParam(r, "postId")
	if postID == "" {
		respondError(w, http.StatusBadRequest, "missing post ID")
		return
	}

	input := appSocial.GetAnalyticsInput{
		AccountID: accountID,
		PostID:    postID,
		UserID:    userID,
	}

	output, err := h.getAnalyticsUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}
