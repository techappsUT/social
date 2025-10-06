// path: backend/internal/middleware/auth.go

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/techappsUT/social-queue/internal/auth"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
	TeamIDKey    contextKey = "team_id"
)

type AuthMiddleware struct {
	tokenService *auth.TokenService
}

func NewAuthMiddleware(tokenService *auth.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth validates JWT token and adds user info to context
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := m.tokenService.ValidateAccessToken(token)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				http.Error(w, "Token has expired", http.StatusUnauthorized)
			default:
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, TeamIDKey, claims.TeamID)

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
		ctx = context.WithValue(ctx, TeamIDKey, claims.TeamID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper functions to extract user info from context
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return uuid.Nil, auth.ErrInvalidToken
	}
	return uuid.Parse(userIDStr)
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

func GetTeamID(ctx context.Context) (*uuid.UUID, error) {
	teamIDStr, ok := ctx.Value(TeamIDKey).(string)
	if !ok || teamIDStr == "" {
		return nil, nil
	}
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return nil, err
	}
	return &teamID, nil
}
