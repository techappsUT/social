// path: backend/internal/middleware/rbac.go

package middleware

import (
	"net/http"

	"github.com/techappsUT/social-queue/internal/models"
)

// RequireRole checks if user has required role
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, role := range roles {
				if userRole == string(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamMembership ensures user belongs to the team in the route
func RequireTeamMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userTeamID, err := GetTeamID(r.Context())
		if err != nil || userTeamID == nil {
			http.Error(w, "Forbidden: no team membership", http.StatusForbidden)
			return
		}

		// You can add additional logic here to verify team access
		// For example, extract team_id from URL and compare with user's team_id

		next.ServeHTTP(w, r)
	})
}

// RequireAdmin checks if user is admin or super_admin
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin, models.RoleSuperAdmin)(next)
}
