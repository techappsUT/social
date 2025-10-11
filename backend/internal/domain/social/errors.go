// path: backend/internal/domain/social/errors.go
// 🆕 NEW - Clean Architecture

package social

import "errors"

// Account-related errors
var (
	// Account validation errors
	ErrAccountNotFound    = errors.New("social account not found")
	ErrInvalidTeamID      = errors.New("invalid team ID")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrInvalidPlatform    = errors.New("invalid platform")
	ErrInvalidAccountType = errors.New("invalid account type")

	// Connection errors
	ErrAccountNotConnected     = errors.New("social account not connected")
	ErrAccountAlreadyConnected = errors.New("social account already connected")
	ErrAccountNotActive        = errors.New("social account not active")
	ErrAccountExpired          = errors.New("social account credentials expired")
	ErrAccountRevoked          = errors.New("social account access revoked")
	ErrAccountSuspended        = errors.New("social account suspended")
	ErrAccountRateLimited      = errors.New("social account rate limited")
	ErrAccountAlreadyExpired   = errors.New("account already marked as expired")
	ErrReconnectRequired       = errors.New("account requires reconnection")

	// Deletion errors
	ErrAccountAlreadyDeleted = errors.New("social account already deleted")
	ErrAccountNotDeleted     = errors.New("social account not deleted")

	// Platform-specific errors
	ErrInstagramRequiresBusiness = errors.New("Instagram requires a business account")
	ErrFacebookRequiresPage      = errors.New("Facebook requires a page or business account")
	ErrLinkedInGroupNotSupported = errors.New("LinkedIn groups are not supported")
	ErrPlatformNotSupported      = errors.New("platform not supported")
	ErrPlatformNotConfigured     = errors.New("platform not configured")

	// OAuth errors
	ErrInvalidAuthCode     = errors.New("invalid authorization code")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenExchangeFailed = errors.New("token exchange failed")
	ErrTokenRefreshFailed  = errors.New("token refresh failed")
	ErrInvalidScope        = errors.New("invalid OAuth scope")
	ErrMissingScope        = errors.New("required scope missing")

	// Rate limiting errors
	ErrDailyLimitExceeded  = errors.New("daily post limit exceeded")
	ErrHourlyLimitExceeded = errors.New("hourly post limit exceeded")
	ErrRateLimitExceeded   = errors.New("platform rate limit exceeded")

	// Publishing errors
	ErrPublishFailed     = errors.New("failed to publish to platform")
	ErrMediaUploadFailed = errors.New("failed to upload media")
	ErrInvalidMediaType  = errors.New("invalid media type for platform")
	ErrMediaSizeTooLarge = errors.New("media size exceeds platform limit")
	ErrContentTooLong    = errors.New("content exceeds platform character limit")
	ErrTooManyHashtags   = errors.New("too many hashtags for platform")
	ErrTooManyMentions   = errors.New("too many mentions for platform")
	ErrTooManyMediaFiles = errors.New("too many media files for platform")
	ErrPostNotFound      = errors.New("platform post not found")

	// Analytics errors
	ErrAnalyticsNotAvailable = errors.New("analytics not available for this account")
	ErrAnalyticsFetchFailed  = errors.New("failed to fetch analytics from platform")

	// Webhook errors
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
	ErrWebhookEventNotFound    = errors.New("webhook event not found")
	ErrWebhookProcessingFailed = errors.New("webhook processing failed")

	// Duplicate errors
	ErrAccountAlreadyExists  = errors.New("social account already exists for this team")
	ErrDuplicatePlatformUser = errors.New("this platform account is already connected")

	// Permission errors
	ErrUnauthorized            = errors.New("unauthorized to perform this action")
	ErrInsufficientPermissions = errors.New("insufficient permissions on platform")
)

// ErrorCode represents a unique error code for API responses
type ErrorCode string

const (
	// Account errors (4000-4099)
	CodeAccountNotFound         ErrorCode = "SOCIAL_4001"
	CodeInvalidAccount          ErrorCode = "SOCIAL_4002"
	CodeAccountNotConnected     ErrorCode = "SOCIAL_4003"
	CodeAccountAlreadyConnected ErrorCode = "SOCIAL_4004"
	CodeAccountExpired          ErrorCode = "SOCIAL_4005"
	CodeAccountRevoked          ErrorCode = "SOCIAL_4006"
	CodeAccountSuspended        ErrorCode = "SOCIAL_4007"
	CodeAccountRateLimited      ErrorCode = "SOCIAL_4008"

	// Platform errors (4100-4199)
	CodeInvalidPlatform       ErrorCode = "PLATFORM_4101"
	CodePlatformNotSupported  ErrorCode = "PLATFORM_4102"
	CodePlatformNotConfigured ErrorCode = "PLATFORM_4103"
	CodePlatformRequirement   ErrorCode = "PLATFORM_4104"

	// OAuth errors (4200-4299)
	CodeInvalidAuthCode     ErrorCode = "OAUTH_4201"
	CodeInvalidRefreshToken ErrorCode = "OAUTH_4202"
	CodeTokenExchangeFailed ErrorCode = "OAUTH_4203"
	CodeTokenRefreshFailed  ErrorCode = "OAUTH_4204"
	CodeInvalidScope        ErrorCode = "OAUTH_4205"

	// Publishing errors (4300-4399)
	CodePublishFailed     ErrorCode = "PUBLISH_4301"
	CodeMediaUploadFailed ErrorCode = "PUBLISH_4302"
	CodeInvalidContent    ErrorCode = "PUBLISH_4303"
	CodeContentTooLong    ErrorCode = "PUBLISH_4304"
	CodeMediaSizeTooLarge ErrorCode = "PUBLISH_4305"

	// Rate limit errors (4400-4499)
	CodeRateLimitExceeded   ErrorCode = "RATE_4401"
	CodeDailyLimitExceeded  ErrorCode = "RATE_4402"
	CodeHourlyLimitExceeded ErrorCode = "RATE_4403"

	// Analytics errors (4500-4599)
	CodeAnalyticsNotAvailable ErrorCode = "ANALYTICS_4501"
	CodeAnalyticsFetchFailed  ErrorCode = "ANALYTICS_4502"

	// Webhook errors (4600-4699)
	CodeInvalidWebhookSignature ErrorCode = "WEBHOOK_4601"
	CodeWebhookProcessingFailed ErrorCode = "WEBHOOK_4602"

	// System errors (4900-4999)
	CodeSocialInternal ErrorCode = "SOCIAL_4901"
	CodeDatabaseError  ErrorCode = "SOCIAL_4902"
)

// ErrorMapping maps domain errors to error codes
var ErrorMapping = map[error]ErrorCode{
	ErrAccountNotFound:           CodeAccountNotFound,
	ErrAccountNotConnected:       CodeAccountNotConnected,
	ErrAccountAlreadyConnected:   CodeAccountAlreadyConnected,
	ErrAccountExpired:            CodeAccountExpired,
	ErrAccountRevoked:            CodeAccountRevoked,
	ErrAccountSuspended:          CodeAccountSuspended,
	ErrAccountRateLimited:        CodeAccountRateLimited,
	ErrInvalidPlatform:           CodeInvalidPlatform,
	ErrPlatformNotSupported:      CodePlatformNotSupported,
	ErrPlatformNotConfigured:     CodePlatformNotConfigured,
	ErrInstagramRequiresBusiness: CodePlatformRequirement,
	ErrFacebookRequiresPage:      CodePlatformRequirement,
	ErrInvalidAuthCode:           CodeInvalidAuthCode,
	ErrInvalidRefreshToken:       CodeInvalidRefreshToken,
	ErrTokenExchangeFailed:       CodeTokenExchangeFailed,
	ErrTokenRefreshFailed:        CodeTokenRefreshFailed,
	ErrInvalidScope:              CodeInvalidScope,
	ErrPublishFailed:             CodePublishFailed,
	ErrMediaUploadFailed:         CodeMediaUploadFailed,
	ErrContentTooLong:            CodeContentTooLong,
	ErrMediaSizeTooLarge:         CodeMediaSizeTooLarge,
	ErrRateLimitExceeded:         CodeRateLimitExceeded,
	ErrDailyLimitExceeded:        CodeDailyLimitExceeded,
	ErrHourlyLimitExceeded:       CodeHourlyLimitExceeded,
	ErrAnalyticsNotAvailable:     CodeAnalyticsNotAvailable,
	ErrAnalyticsFetchFailed:      CodeAnalyticsFetchFailed,
	ErrInvalidWebhookSignature:   CodeInvalidWebhookSignature,
	ErrWebhookProcessingFailed:   CodeWebhookProcessingFailed,
}

// GetErrorCode returns the error code for a given error
func GetErrorCode(err error) ErrorCode {
	if code, ok := ErrorMapping[err]; ok {
		return code
	}
	return CodeSocialInternal
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrAccountNotFound) ||
		errors.Is(err, ErrPostNotFound) ||
		errors.Is(err, ErrWebhookEventNotFound)
}

// IsConnectionError checks if an error is related to account connection
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrAccountNotConnected) ||
		errors.Is(err, ErrAccountExpired) ||
		errors.Is(err, ErrAccountRevoked) ||
		errors.Is(err, ErrReconnectRequired) ||
		errors.Is(err, ErrAccountNotActive)
}

// IsOAuthError checks if an error is OAuth-related
func IsOAuthError(err error) bool {
	return errors.Is(err, ErrInvalidAuthCode) ||
		errors.Is(err, ErrInvalidRefreshToken) ||
		errors.Is(err, ErrTokenExchangeFailed) ||
		errors.Is(err, ErrTokenRefreshFailed) ||
		errors.Is(err, ErrInvalidScope) ||
		errors.Is(err, ErrMissingScope)
}

// IsRateLimitError checks if an error is rate limit related
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded) ||
		errors.Is(err, ErrDailyLimitExceeded) ||
		errors.Is(err, ErrHourlyLimitExceeded) ||
		errors.Is(err, ErrAccountRateLimited)
}

// IsPublishingError checks if an error is related to publishing
func IsPublishingError(err error) bool {
	return errors.Is(err, ErrPublishFailed) ||
		errors.Is(err, ErrMediaUploadFailed) ||
		errors.Is(err, ErrInvalidMediaType) ||
		errors.Is(err, ErrMediaSizeTooLarge) ||
		errors.Is(err, ErrContentTooLong) ||
		errors.Is(err, ErrTooManyHashtags) ||
		errors.Is(err, ErrTooManyMentions) ||
		errors.Is(err, ErrTooManyMediaFiles)
}

// IsPlatformError checks if an error is platform-specific
func IsPlatformError(err error) bool {
	return errors.Is(err, ErrInvalidPlatform) ||
		errors.Is(err, ErrPlatformNotSupported) ||
		errors.Is(err, ErrPlatformNotConfigured) ||
		errors.Is(err, ErrInstagramRequiresBusiness) ||
		errors.Is(err, ErrFacebookRequiresPage) ||
		errors.Is(err, ErrLinkedInGroupNotSupported)
}

// IsDuplicateError checks if an error is a duplicate/conflict error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrAccountAlreadyExists) ||
		errors.Is(err, ErrDuplicatePlatformUser) ||
		errors.Is(err, ErrAccountAlreadyConnected)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidTeamID) ||
		errors.Is(err, ErrInvalidUserID) ||
		errors.Is(err, ErrInvalidPlatform) ||
		errors.Is(err, ErrInvalidAccountType) ||
		errors.Is(err, ErrInvalidMediaType)
}
