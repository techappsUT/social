// path: backend/internal/handlers/auth_handler.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/user"
	"github.com/techappsUT/social-queue/internal/middleware"
)

// AuthHandler handles all authentication-related endpoints
type AuthHandler struct {
	// User management
	createUserUC *user.CreateUserUseCase
	getUserUC    *user.GetUserUseCase
	updateUserUC *user.UpdateUserUseCase
	deleteUserUC *user.DeleteUserUseCase

	// Authentication
	loginUC              *auth.LoginUseCase
	refreshTokenUC       *auth.RefreshTokenUseCase
	logoutUC             *auth.LogoutUseCase
	verifyEmailUC        *auth.VerifyEmailUseCase
	resendVerificationUC *auth.ResendVerificationUseCase
	forgotPasswordUC     *auth.ForgotPasswordUseCase
	resetPasswordUC      *auth.ResetPasswordUseCase
	changePasswordUC     *auth.ChangePasswordUseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	createUserUC *user.CreateUserUseCase,
	getUserUC *user.GetUserUseCase,
	updateUserUC *user.UpdateUserUseCase,
	deleteUserUC *user.DeleteUserUseCase,
	loginUC *auth.LoginUseCase,
	refreshTokenUC *auth.RefreshTokenUseCase,
	logoutUC *auth.LogoutUseCase,
	verifyEmailUC *auth.VerifyEmailUseCase,
	resendVerificationUC *auth.ResendVerificationUseCase,
	forgotPasswordUC *auth.ForgotPasswordUseCase,
	resetPasswordUC *auth.ResetPasswordUseCase,
	changePasswordUC *auth.ChangePasswordUseCase,
) *AuthHandler {
	return &AuthHandler{
		createUserUC:         createUserUC,
		getUserUC:            getUserUC,
		updateUserUC:         updateUserUC,
		deleteUserUC:         deleteUserUC,
		loginUC:              loginUC,
		refreshTokenUC:       refreshTokenUC,
		logoutUC:             logoutUC,
		verifyEmailUC:        verifyEmailUC,
		resendVerificationUC: resendVerificationUC,
		forgotPasswordUC:     forgotPasswordUC,
		resetPasswordUC:      resetPasswordUC,
		changePasswordUC:     changePasswordUC,
	}
}

// ============================================================================
// PUBLIC AUTH ROUTES
// ============================================================================

// Signup handles POST /api/v2/auth/signup
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var input user.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.createUserUC.Execute(r.Context(), input)
	if err != nil {
		status := mapErrorToStatus(err)
		respondError(w, status, err.Error())
		return
	}

	// Set refresh token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	// Don't send refresh token in response body
	output.RefreshToken = ""

	respondCreated(w, output)
}

// Login handles POST /api/v2/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.loginUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Set refresh token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
	})

	// Don't send refresh token in response body
	output.RefreshToken = ""

	respondSuccess(w, output)
}

// RefreshToken handles POST /api/v2/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Try to get refresh token from cookie first
	var refreshToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		// Fallback to request body
		var input auth.RefreshTokenInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		refreshToken = input.RefreshToken
	}

	if refreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token required")
		return
	}

	input := auth.RefreshTokenInput{RefreshToken: refreshToken}
	output, err := h.refreshTokenUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Set new refresh token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
	})

	// Don't send refresh token in response body
	output.RefreshToken = ""

	respondSuccess(w, output)
}

// Logout handles POST /api/v2/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	// Get tokens
	var refreshToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	}

	accessToken := r.Header.Get("Authorization")
	if len(accessToken) > 7 {
		accessToken = accessToken[7:] // Remove "Bearer "
	}

	input := auth.LogoutInput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID.String(),
	}

	output, err := h.logoutUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // Delete cookie
	})

	respondSuccess(w, output)
}

// VerifyEmail handles POST /api/v2/auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input auth.VerifyEmailInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.verifyEmailUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ResendVerification handles POST /api/v2/auth/resend-verification
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var input auth.ResendVerificationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.resendVerificationUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ForgotPassword handles POST /api/v2/auth/forgot-password
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input auth.ForgotPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.forgotPasswordUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ResetPassword handles POST /api/v2/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input auth.ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.resetPasswordUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ChangePassword handles POST /api/v2/auth/change-password (authenticated)
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input auth.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	input.UserID = userID.String()

	output, err := h.changePasswordUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// USER MANAGEMENT ROUTES (authenticated)
// ============================================================================

// GetUser handles GET /api/v2/users/{id}
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Allow if requesting own profile or if admin
	if requestUserID != userID {
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
	}

	input := user.GetUserInput{UserID: userID}
	output, err := h.getUserUC.Execute(r.Context(), input)
	if err != nil {
		status := mapErrorToStatus(err)
		respondError(w, status, err.Error())
		return
	}

	respondSuccess(w, output)
}

// UpdateUser handles PUT /api/v2/users/{id}
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if requestUserID != userID {
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
	}

	var input user.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	input.UserID = userID

	output, err := h.updateUserUC.Execute(r.Context(), input)
	if err != nil {
		status := mapErrorToStatus(err)
		respondError(w, status, err.Error())
		return
	}

	respondSuccess(w, output)
}

// DeleteUser handles DELETE /api/v2/users/{id}
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if requestUserID != userID {
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&reqBody)

	input := user.DeleteUserInput{
		UserID: userID,
		Reason: reqBody.Reason,
	}
	output, err := h.deleteUserUC.Execute(r.Context(), input)
	if err != nil {
		status := mapErrorToStatus(err)
		respondError(w, status, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func mapErrorToStatus(err error) int {
	errMsg := err.Error()

	switch {
	case errMsg == "user not found":
		return http.StatusNotFound
	case errMsg == "email already registered":
		return http.StatusConflict
	case errMsg == "username already taken":
		return http.StatusConflict
	case errMsg == "validation failed" || errMsg == "user ID is required":
		return http.StatusBadRequest
	case errMsg == "cannot update inactive user":
		return http.StatusForbidden
	case errMsg == "user already deleted":
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}
