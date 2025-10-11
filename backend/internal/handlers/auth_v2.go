// path: backend/internal/handlers/auth_v2.go
// CREATE THIS AS A NEW FILE - DON'T REPLACE auth.go YET

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/techappsUT/social-queue/internal/application/auth"
	"github.com/techappsUT/social-queue/internal/application/user"
)

// AuthHandlerV2 is the new handler that uses application layer
type AuthHandlerV2 struct {
	createUserUC *user.CreateUserUseCase
	loginUC      *auth.LoginUseCase
	// Add more use cases as you create them
}

// NewAuthHandlerV2 creates a new auth handler with use cases
func NewAuthHandlerV2(
	createUserUC *user.CreateUserUseCase,
	loginUC *auth.LoginUseCase,
) *AuthHandlerV2 {
	return &AuthHandlerV2{
		createUserUC: createUserUC,
		loginUC:      loginUC,
	}
}

// RegisterRoutes registers all auth routes
func (h *AuthHandlerV2) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.Signup)
	r.Post("/auth/login", h.Login)
	// Add other routes as you implement use cases
	// r.Post("/auth/refresh", h.RefreshToken)
	// r.Post("/auth/logout", h.Logout)
	// r.Post("/auth/verify-email", h.VerifyEmail)
}

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

// Helper methods
func (h *AuthHandlerV2) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandlerV2) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *AuthHandlerV2) mapErrorToHTTPStatus(err error) int {
	errStr := err.Error()
	switch {
	case contains(errStr, "already"):
		return http.StatusConflict
	case contains(errStr, "not found"):
		return http.StatusNotFound
	case contains(errStr, "invalid"):
		return http.StatusBadRequest
	case contains(errStr, "unauthorized"):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
