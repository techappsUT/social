// path: backend/internal/handlers/auth.go

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/techappsUT/social-queue/internal/auth"
	"github.com/techappsUT/social-queue/internal/dto"
)

type AuthHandler struct {
	authService *auth.Service
	validate    *validator.Validate
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/verify-email", h.VerifyEmail)
	r.Post("/auth/resend-verification", h.ResendVerification)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
	r.Post("/auth/refresh", h.RefreshToken) // Aligned endpoint
	r.Post("/auth/logout", h.Logout)
}

// Signup handles user registration
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req dto.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.authService.Signup(req)
	if err != nil {
		switch err {
		case auth.ErrUserAlreadyExists:
			respondError(w, http.StatusConflict, "User already exists")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	authResp, err := h.authService.Login(req)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			respondError(w, http.StatusUnauthorized, "Invalid email or password")
		case auth.ErrEmailNotVerified:
			respondError(w, http.StatusForbidden, "Please verify your email first")
		default:
			respondError(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	// Set refresh token as HTTP-only cookie (more secure)
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    authResp.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	respondJSON(w, http.StatusOK, authResp)
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Try to get refresh token from cookie first
	cookie, err := r.Cookie("refreshToken")
	refreshToken := ""

	if err == nil {
		refreshToken = cookie.Value
	} else {
		// Fallback: Try to get from request body
		var req dto.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		respondError(w, http.StatusUnauthorized, "Refresh token required")
		return
	}

	authResp, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		switch err {
		case auth.ErrInvalidToken:
			respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		case auth.ErrUserNotFound:
			respondError(w, http.StatusUnauthorized, "User not found")
		default:
			respondError(w, http.StatusInternalServerError, "Token refresh failed")
		}
		return
	}

	// Update refresh token cookie with new token (token rotation)
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    authResp.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})

	respondJSON(w, http.StatusOK, authResp)
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.authService.VerifyEmail(req.Token)
	if err != nil {
		switch err {
		case auth.ErrInvalidToken:
			respondError(w, http.StatusBadRequest, "Invalid or expired verification token")
		default:
			respondError(w, http.StatusInternalServerError, "Email verification failed")
		}
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// ResendVerification handles resending verification email
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest // Reuse this DTO as it only needs email
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.authService.ResendVerification(req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to resend verification email")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// ForgotPassword handles password reset requests
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process password reset request")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// ResetPassword handles password reset with token
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		switch err {
		case auth.ErrInvalidToken:
			respondError(w, http.StatusBadRequest, "Invalid or expired reset token")
		default:
			respondError(w, http.StatusInternalServerError, "Password reset failed")
		}
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refreshToken")
	if err == nil && cookie.Value != "" {
		// Revoke the refresh token
		h.authService.Logout(cookie.Value)
	}

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete cookie
	})

	respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Logged out successfully",
		Success: true,
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, dto.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
