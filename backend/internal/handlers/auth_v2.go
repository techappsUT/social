// path: backend/internal/handlers/auth_v2.go
// ✅ UPDATED - Use shared response helpers
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
	updateUserUC *user.UpdateUserUseCase
	getUserUC    *user.GetUserUseCase
	deleteUserUC *user.DeleteUserUseCase
}

// NewAuthHandlerV2 creates a new auth handler with use cases
func NewAuthHandlerV2(
	createUserUC *user.CreateUserUseCase,
	loginUC *auth.LoginUseCase,
	updateUserUC *user.UpdateUserUseCase,
	getUserUC *user.GetUserUseCase,
	deleteUserUC *user.DeleteUserUseCase,
) *AuthHandlerV2 {
	return &AuthHandlerV2{
		createUserUC: createUserUC,
		loginUC:      loginUC,
		updateUserUC: updateUserUC,
		getUserUC:    getUserUC,
		deleteUserUC: deleteUserUC,
	}
}

// ============================================================================
// AUTHENTICATION HANDLERS
// ============================================================================

// Signup handles user registration
func (h *AuthHandlerV2) Signup(w http.ResponseWriter, r *http.Request) {
	var input user.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body") // ✅ Use shared helper
		return
	}

	output, err := h.createUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		respondError(w, status, err.Error()) // ✅ Use shared helper
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

	respondCreated(w, output) // ✅ Use shared helper
}

// Login handles user authentication
func (h *AuthHandlerV2) Login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body") // ✅ Use shared helper
		return
	}

	output, err := h.loginUC.Execute(r.Context(), input)
	if err != nil {
		// Always return 401 for login failures (security)
		respondError(w, http.StatusUnauthorized, "Invalid credentials") // ✅ Use shared helper
		return
	}

	// Set refresh token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})

	// Don't send refresh token in response body
	output.RefreshToken = ""

	respondSuccess(w, output) // ✅ Use shared helper
}

// ============================================================================
// USER MANAGEMENT HANDLERS
// ============================================================================

// GetUser retrieves a user by ID
func (h *AuthHandlerV2) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID") // ✅ Use shared helper
		return
	}

	// Check authorization
	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized") // ✅ Use shared helper
		return
	}

	if requestUserID != userID {
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			respondError(w, http.StatusForbidden, "Forbidden") // ✅ Use shared helper
			return
		}
	}

	input := user.GetUserInput{UserID: userID}
	output, err := h.getUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		respondError(w, status, err.Error()) // ✅ Use shared helper
		return
	}

	respondSuccess(w, output) // ✅ Use shared helper
}

// UpdateUser updates user information
func (h *AuthHandlerV2) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID") // ✅ Use shared helper
		return
	}

	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized") // ✅ Use shared helper
		return
	}

	if requestUserID != userID {
		respondError(w, http.StatusForbidden, "Forbidden") // ✅ Use shared helper
		return
	}

	var input user.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body") // ✅ Use shared helper
		return
	}

	input.UserID = userID
	output, err := h.updateUserUC.Execute(r.Context(), input)
	if err != nil {
		status := h.mapErrorToHTTPStatus(err)
		respondError(w, status, err.Error()) // ✅ Use shared helper
		return
	}

	respondSuccess(w, output) // ✅ Use shared helper
}

// DeleteUser soft deletes a user account
func (h *AuthHandlerV2) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID") // ✅ Use shared helper
		return
	}

	requestUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized") // ✅ Use shared helper
		return
	}

	if requestUserID != userID {
		role, _ := middleware.GetUserRole(r.Context())
		if role != "admin" && role != "owner" {
			respondError(w, http.StatusForbidden, "Forbidden") // ✅ Use shared helper
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
		status := h.mapErrorToHTTPStatus(err)
		respondError(w, status, err.Error()) // ✅ Use shared helper
		return
	}

	respondSuccess(w, output) // ✅ Use shared helper
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// ✅ REMOVED: Duplicate respondJSON and respondError (use shared helpers)

func (h *AuthHandlerV2) mapErrorToHTTPStatus(err error) int {
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
