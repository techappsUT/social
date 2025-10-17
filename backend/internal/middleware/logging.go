package middleware

import (
	"net/http"
	"strings"
	// ... other imports
)

// extractIP extracts the real IP address from the request
func extractIP(r *http.Request) string {
	// Check X-Real-IP header first
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Check X-Forwarded-For header
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP if there are multiple
		parts := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}

	// Fall back to RemoteAddr
	// RemoteAddr can be in format "IP:port"
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// ... rest of your existing logging middleware code
