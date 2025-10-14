// path: backend/internal/common/validation/validators.go
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ============================================================================
// VALIDATION ERRORS
// ============================================================================

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ============================================================================
// COMMON VALIDATORS
// ============================================================================

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	slugRegex  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	urlRegex   = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)

// ValidateEmail checks if email is valid
func ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}
	if len(email) > 255 {
		return &ValidationError{Field: "email", Message: "email too long (max 255 characters)"}
	}
	return nil
}

// ValidatePassword checks password requirements
func ValidatePassword(password string) error {
	if password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	if len(password) > 72 {
		return &ValidationError{Field: "password", Message: "password too long (max 72 characters)"}
	}

	// Check for complexity (at least one letter and one number)
	hasLetter := false
	hasNumber := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		return &ValidationError{Field: "password", Message: "password must contain at least one letter and one number"}
	}

	return nil
}

// ValidateSlug checks if slug is valid
func ValidateSlug(slug string) error {
	if slug == "" {
		return &ValidationError{Field: "slug", Message: "slug is required"}
	}
	if !slugRegex.MatchString(slug) {
		return &ValidationError{Field: "slug", Message: "slug must be lowercase alphanumeric with hyphens"}
	}
	if len(slug) < 3 {
		return &ValidationError{Field: "slug", Message: "slug too short (min 3 characters)"}
	}
	if len(slug) > 50 {
		return &ValidationError{Field: "slug", Message: "slug too long (max 50 characters)"}
	}
	return nil
}

// ValidateUUID checks if string is valid UUID
func ValidateUUID(id string, fieldName string) error {
	if id == "" {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", fieldName)}
	}
	if _, err := uuid.Parse(id); err != nil {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("invalid %s format", fieldName)}
	}
	return nil
}

// ValidateURL checks if URL is valid
func ValidateURL(url string, fieldName string) error {
	if url == "" {
		return nil // URL is optional
	}
	if !urlRegex.MatchString(url) {
		return &ValidationError{Field: fieldName, Message: "invalid URL format"}
	}
	if len(url) > 2048 {
		return &ValidationError{Field: fieldName, Message: "URL too long (max 2048 characters)"}
	}
	return nil
}

// ValidateStringLength checks string length constraints
func ValidateStringLength(value string, fieldName string, min, max int) error {
	length := utf8.RuneCountInString(value)

	if min > 0 && length < min {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be at least %d characters", fieldName, min),
		}
	}

	if max > 0 && length > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must not exceed %d characters", fieldName, max),
		}
	}

	return nil
}

// ValidateRequired checks if value is not empty
func ValidateRequired(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("%s is required", fieldName)}
	}
	return nil
}

// ValidateFutureDate checks if date is in the future
func ValidateFutureDate(date time.Time, fieldName string) error {
	if date.Before(time.Now().UTC()) {
		return &ValidationError{Field: fieldName, Message: fmt.Sprintf("%s must be in the future", fieldName)}
	}
	return nil
}

// ValidateEnum checks if value is in allowed list
func ValidateEnum(value string, fieldName string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ValidationError{
		Field:   fieldName,
		Message: fmt.Sprintf("%s must be one of: %s", fieldName, strings.Join(allowed, ", ")),
	}
}

// ============================================================================
// PLATFORM-SPECIFIC VALIDATORS
// ============================================================================

// ValidatePostContent checks post content for platform constraints
func ValidatePostContent(content string, platform string) error {
	if content == "" {
		return &ValidationError{Field: "content", Message: "content is required"}
	}

	// Platform-specific character limits
	var maxLength int
	switch platform {
	case "twitter":
		maxLength = 280
	case "linkedin":
		maxLength = 3000
	case "facebook":
		maxLength = 63206
	case "instagram":
		maxLength = 2200
	default:
		maxLength = 2200 // Default limit
	}

	length := utf8.RuneCountInString(content)
	if length > maxLength {
		return &ValidationError{
			Field:   "content",
			Message: fmt.Sprintf("content exceeds %s limit of %d characters", platform, maxLength),
		}
	}

	return nil
}

// ValidateHashtags checks hashtag format and count
func ValidateHashtags(hashtags []string) error {
	if len(hashtags) > 30 {
		return &ValidationError{Field: "hashtags", Message: "too many hashtags (max 30)"}
	}

	hashtagRegex := regexp.MustCompile(`^#[a-zA-Z0-9_]+$`)
	for _, tag := range hashtags {
		if !hashtagRegex.MatchString(tag) {
			return &ValidationError{
				Field:   "hashtags",
				Message: fmt.Sprintf("invalid hashtag format: %s", tag),
			}
		}
		if len(tag) > 50 {
			return &ValidationError{
				Field:   "hashtags",
				Message: fmt.Sprintf("hashtag too long: %s (max 50 characters)", tag),
			}
		}
	}

	return nil
}

// ============================================================================
// COMPOSITE VALIDATORS
// ============================================================================

// Validator is a function that validates and returns an error
type Validator func() error

// ValidateAll runs multiple validators and collects errors
func ValidateAll(validators ...Validator) error {
	var errors ValidationErrors

	for _, validator := range validators {
		if err := validator(); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errors = append(errors, *ve)
			} else if ves, ok := err.(ValidationErrors); ok {
				errors = append(errors, ves...)
			} else {
				// Unknown error type
				errors = append(errors, ValidationError{
					Field:   "unknown",
					Message: err.Error(),
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
