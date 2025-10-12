// path: backend/internal/application/team/helpers.go
// NEW FILE - Put helper functions here to avoid redeclaration
package team

import "strings"

// generateSlug creates a URL-friendly slug from a team name
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	return slug
}
