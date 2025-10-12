// path: backend/internal/middleware/rbac.go
package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// RequireRole checks if user has required role
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Check if user's role is in the allowed list
			hasRole := false
			for _, allowedRole := range allowedRoles {
				if userRole == allowedRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, `{"error":"Forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamMembership ensures user belongs to a team
func RequireTeamMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get team ID from context
		userTeamID, ok := GetTeamID(r.Context())
		if !ok || userTeamID == uuid.Nil {
			http.Error(w, `{"error":"Forbidden: no team membership"}`, http.StatusForbidden)
			return
		}

		// You can add additional logic here to verify team access
		// For example, extract team_id from URL and compare with user's team_id

		next.ServeHTTP(w, r)
	})
}

// Note: RequireAdmin is now in auth.go to avoid duplication
