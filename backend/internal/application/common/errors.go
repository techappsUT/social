// path: backend/internal/application/common/errors.go
package common

import "errors"

// ============================================================================
// APPLICATION ERRORS
// ============================================================================

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrUnauthorized       = errors.New("unauthorized")

	// Validation errors
	ErrInvalidInput    = errors.New("invalid input")
	ErrMissingRequired = errors.New("missing required field")

	// Resource errors
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")

	// Business logic errors
	ErrQuotaExceeded           = errors.New("quota exceeded")
	ErrRateLimitExceeded       = errors.New("rate limit exceeded")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)
