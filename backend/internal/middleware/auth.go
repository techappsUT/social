// path: backend/internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
	TeamIDKey    contextKey = "team_id"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	tokenService common.TokenService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService common.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth validates JWT token and adds user info to context
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := m.tokenService.ValidateAccessToken(token)
		if err != nil {
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		if claims.TeamID != "" {
			ctx = context.WithValue(ctx, TeamIDKey, claims.TeamID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth validates token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		token := parts[1]
		claims, err := m.tokenService.ValidateAccessToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		if claims.TeamID != "" {
			ctx = context.WithValue(ctx, TeamIDKey, claims.TeamID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin ensures the user has admin role
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetUserRole(r.Context())
		if !ok || role != "admin" {
			http.Error(w, `{"error":"Admin access required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Helper functions to extract user info from context

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

// GetUserEmail extracts user email from context
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRole extracts user role from context
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// GetTeamID extracts team ID from context
func GetTeamID(ctx context.Context) (uuid.UUID, bool) {
	teamIDStr, ok := ctx.Value(TeamIDKey).(string)
	if !ok {
		return uuid.Nil, false
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return teamID, true
}
