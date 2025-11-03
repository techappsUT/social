// ============================================================================
// FILE: backend/internal/handlers/social_handler.go
// ============================================================================
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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

	// Get state from query params (sent by frontend)
	state := r.URL.Query().Get("state")
	if state == "" {
		// Generate one if not provided (backward compatibility)
		state = uuid.New().String()
	}

	// Get the adapter
	adapter, ok := h.adapters[socialDomain.Platform(platform)]
	if !ok {
		respondError(w, http.StatusBadRequest, "unsupported platform")
		return
	}

	// Generate auth URL with the provided state
	authURL := adapter.GetAuthURL(state, []string{})

	respondSuccess(w, map[string]string{
		"authUrl": authURL,
		"state":   state, // Return the same state
	})
}

// OAuthCallback handles GET /api/v2/social/auth/:platform/callback
func (h *SocialHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	platform := chi.URLParam(r, "platform")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Get frontend URL from environment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}

	// Handle OAuth errors (user denied, etc.)
	if errorParam != "" {
		errorDesc := r.URL.Query().Get("error_description")
		redirectURL := fmt.Sprintf("%s/accounts?error=%s&description=%s",
			frontendURL, errorParam, url.QueryEscape(errorDesc))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	if code == "" {
		redirectURL := fmt.Sprintf("%s/accounts?error=missing_code", frontendURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to frontend with code and state
	// The frontend will then make an authenticated API call to complete the connection
	redirectURL := fmt.Sprintf("%s/accounts/callback?platform=%s&code=%s&state=%s",
		frontendURL, platform, url.QueryEscape(code), url.QueryEscape(state))

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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
	fmt.Printf("Disconnecting account: %s for user: %s\n", accountIDStr, userID)

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID format")
		return
	}

	input := appSocial.DisconnectAccountInput{
		AccountID: accountID,
		UserID:    userID,
	}

	if err := h.disconnectUC.Execute(r.Context(), input); err != nil {
		fmt.Printf("Disconnect error: %v\n", err)
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Account not found")
		} else if strings.Contains(err.Error(), "access denied") {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
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

// CompleteOAuthConnection handles POST /api/v2/social/auth/complete
func (h *SocialHandler) CompleteOAuthConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Platform string `json:"platform"`
		Code     string `json:"code"`
		State    string `json:"state"`
		TeamID   string `json:"teamId,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Remove logger, just respond with error
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Debug print without logger
	fmt.Printf("CompleteOAuth request: platform=%s, hasCode=%v, hasState=%v, teamId=%s\n",
		input.Platform, input.Code != "", input.State != "", input.TeamID)

	// Parse TeamID
	var teamID uuid.UUID
	if input.TeamID != "" {
		parsed, err := uuid.Parse(input.TeamID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid team ID format")
			return
		}
		teamID = parsed
	} else {
		respondError(w, http.StatusBadRequest, "team ID is required")
		return
	}

	// Prepare use case input
	ucInput := appSocial.ConnectAccountInput{
		UserID:   userID,
		TeamID:   teamID,
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
