// ===========================================================================
// FILE: backend/internal/application/team/helpers.go
// FIXED: Slug generation with uniqueness check and random suffix
// ===========================================================================
package team

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

// generateSlug creates a URL-friendly slug from a team name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with dashes
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only alphanumeric and dashes)
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "")

	// Replace multiple consecutive dashes with single dash
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing dashes
	slug = strings.Trim(slug, "-")

	// If slug is empty after cleaning, use a default
	if slug == "" {
		slug = "team"
	}

	// Truncate if too long (max 50 chars, leave room for suffix)
	if len(slug) > 40 {
		slug = slug[:40]
	}

	return slug
}

// generateUniqueSlug generates a unique slug by checking existence and adding suffix
func generateUniqueSlug(baseSlug string, checkExists func(string) bool) string {
	// First, try the base slug
	if !checkExists(baseSlug) {
		return baseSlug
	}

	// If exists, try with incrementing numbers first (cleaner)
	for i := 1; i <= 10; i++ {
		candidate := baseSlug + "-" + string(rune('0'+i))
		if !checkExists(candidate) {
			return candidate
		}
	}

	// If still not unique, add random suffix
	randomSuffix := generateRandomString(6)
	return baseSlug + "-" + randomSuffix
}

// generateRandomString generates a random alphanumeric string
func generateRandomString(length int) string {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based suffix
		return "x"
	}
	return hex.EncodeToString(bytes)[:length]
}

// ===========================================================================
// FILE: backend/internal/application/team/create_team.go
// UPDATE: Use the new unique slug generation
// ===========================================================================

// Update the Execute method in CreateTeamUseCase:
