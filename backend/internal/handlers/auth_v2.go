// path: backend/internal/handlers/auth_v2.go
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

// AuthHandlerV2 is the Clean Architecture handler
type AuthHandlerV2 struct {
	createUserUC *user.CreateUserUseCase
	loginUC      *auth.LoginUseCase
	updateUserUC *user.UpdateUserUseCase // 🆕 NEW
	getUserUC    *user.GetUserUseCase    // 🆕 NEW
	deleteUserUC *user.DeleteUserUseCase // 🆕 NEW
}

// NewAuthHandlerV2 creates a new auth handler with use cases
func NewAuthHandlerV2(
	createUserUC *user.CreateUserUseCase,
	loginUC *auth.LoginUseCase,
	updateUserUC *user.UpdateUserUseCase, // 🆕 NEW
	getUserUC *user.GetUserUseCase, // 🆕 NEW
	deleteUserUC *user.DeleteUserUseCase, // 🆕 NEW
) *AuthHandlerV2 {
	return &AuthHandlerV2{
		createUserUC: createUserUC,
		loginUC:      loginUC,
		updateUserUC: updateUserUC, // 🆕 NEW
		getUserUC:    getUserUC,    // 🆕 NEW
		deleteUserUC: deleteUserUC, // 🆕 NEW
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandlerV2) RegisterRoutes(r chi.Router) {
	// Public routes
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		// Add auth middleware here if not already applied globally

		// 🆕 NEW: User management endpoints
		r.Get("/users/{id}", h.GetUser)       // Get user by ID
		r.Put("/users/{id}", h.UpdateUser)    // Update user profile
		r.Delete("/users/{id}", h.DeleteUser) // Delete user account
	})
}

// ============================================================================
// EXISTING HANDLERS (Keep these)
// ============================================================================

// Signup handles user registration
func (h *AuthHandlerV2) Signup(w http.ResponseWriter, r *http.Request) {
	var input user.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.createUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		h.respondError(w, status, err.Error())
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
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	// Don't send refresh token in response body
	output.RefreshToken = ""

	h.respondJSON(w, http.StatusCreated, output)
}

// Login handles user authentication
func (h *AuthHandlerV2) Login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := h.loginUC.Execute(r.Context(), input)
	if err != nil {
		// Always return 401 for login failures (security)
		h.respondError(w, http.StatusUnauthorized, "Invalid credentials")
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
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	// Don't send refresh token in response body
	output.RefreshToken = ""

	h.respondJSON(w, http.StatusOK, output)
}

// ============================================================================
// 🆕 NEW HANDLERS
// ============================================================================

// GetUser retrieves a user by ID
func (h *AuthHandlerV2) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check authorization: users can only get their own profile (or admin can get any)
	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Allow if requesting own profile
	if requestUserID != userID {
		// Check if user is admin
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			h.respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
	}

	// Execute use case
	input := user.GetUserInput{UserID: userID}
	output, err := h.getUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		h.respondError(w, status, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, output)
}

// UpdateUser updates a user's profile
func (h *AuthHandlerV2) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check authorization: users can only update their own profile
	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if requestUserID != userID {
		// Check if user is admin
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			h.respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
	}

	// Parse request body
	var input user.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set user ID from URL
	input.UserID = userID

	// Execute use case
	output, err := h.updateUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		h.respondError(w, status, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, output)
}

// DeleteUser soft-deletes a user account
func (h *AuthHandlerV2) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check authorization: users can only delete their own account
	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if requestUserID != userID {
		// Check if user is admin
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			h.respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
	}

	// Parse optional request body for reason
	var reqBody struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&reqBody)

	// Execute use case
	input := user.DeleteUserInput{
		UserID: userID,
		Reason: reqBody.Reason,
	}
	output, err := h.deleteUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		h.respondError(w, status, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, output)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

func (h *AuthHandlerV2) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandlerV2) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

func (h *AuthHandlerV2) mapErrorToHTTPStatus(err error) int {
	// Map domain errors to HTTP status codes
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
