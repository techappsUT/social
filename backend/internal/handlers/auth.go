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

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/verify-email", h.VerifyEmail)
	r.Post("/auth/forgot-password", h.ForgotPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
	r.Post("/auth/refresh", h.RefreshToken)
	r.Post("/auth/logout", h.Logout)
}

// Signup godoc
// @Summary User signup
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.SignupRequest true "Signup request"
// @Success 201 {object} dto.MessageResponse
// @Router /auth/signup [post]
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

	user, err := h.authService.Signup(req)
	if err != nil {
		switch err {
		case auth.ErrUserAlreadyExists:
			respondError(w, http.StatusConflict, "User already exists")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	respondJSON(w, http.StatusCreated, dto.MessageResponse{
		Message: "User created successfully. Please check your email to verify your account.",
	})
	_ = user // user created successfully
}

// Login godoc
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Router /auth/login [post]
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

	// Set refresh token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    authResp.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	respondJSON(w, http.StatusOK, authResp)
}

// VerifyEmail godoc
// @Summary Verify email address
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verify email request"
// @Success 200 {object} dto.MessageResponse
// @Router /auth/verify-email [post]
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

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired verification token")
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Email verified successfully",
	})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} dto.MessageResponse
// @Router /auth/forgot-password [post]
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

	// Always return success to prevent user enumeration
	_ = h.authService.ForgotPassword(req.Email)

	respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} dto.MessageResponse
// @Router /auth/reset-password [post]
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

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid or expired reset token")
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Password reset successfully",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags auth
// @Produce json
// @Success 200 {object} dto.AuthResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Try to get refresh token from cookie first
	cookie, err := r.Cookie("refresh_token")
	var refreshToken string

	if err == nil {
		refreshToken = cookie.Value
	} else {
		// Fallback to JSON body
		var req dto.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		refreshToken = req.RefreshToken
	}

	authResp, err := h.authService.RefreshAccessToken(refreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	respondJSON(w, http.StatusOK, authResp)
}

// Logout godoc
// @Summary Logout user
// @Tags auth
// @Produce json
// @Success 200 {object} dto.MessageResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		_ = h.authService.RevokeRefreshToken(cookie.Value)
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete cookie
	})

	respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Logged out successfully",
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
