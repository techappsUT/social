// path: backend/internal/domain/user/errors.go
// 🆕 NEW - Clean Architecture

package user

import "errors"

// Domain errors for user entity
// These errors represent business rule violations and domain-specific error conditions

var (
	// Entity validation errors
	ErrInvalidUser       = errors.New("invalid user")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	// Email-related errors
	ErrEmailRequired        = errors.New("email is required")
	ErrInvalidEmailFormat   = errors.New("invalid email format")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrEmailAlreadyVerified = errors.New("email already verified")

	// Username-related errors
	ErrUsernameRequired      = errors.New("username is required")
	ErrUsernameTooShort      = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong       = errors.New("username must be at most 30 characters")
	ErrInvalidUsernameFormat = errors.New("username can only contain letters, numbers, underscores, and dashes")
	ErrUsernameAlreadyExists = errors.New("username already exists")

	// Password-related errors
	ErrPasswordRequired   = errors.New("password is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong    = errors.New("password must be at most 128 characters")
	ErrPasswordTooWeak    = errors.New("password is too weak")
	ErrInvalidOldPassword = errors.New("invalid old password")
	ErrPasswordMismatch   = errors.New("passwords do not match")
	ErrPasswordSameAsOld  = errors.New("new password cannot be the same as old password")

	// Name-related errors
	ErrInvalidFirstName = errors.New("first name is required")
	ErrInvalidLastName  = errors.New("last name is required")
	ErrNameTooLong      = errors.New("name is too long")

	// Authentication errors
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// Account status errors
	ErrUserInactive         = errors.New("user account is inactive")
	ErrUserSuspended        = errors.New("user account is suspended")
	ErrUserDeleted          = errors.New("user account is deleted")
	ErrUserAlreadySuspended = errors.New("user is already suspended")
	ErrUserAlreadyActive    = errors.New("user is already active")
	ErrUserAlreadyDeleted   = errors.New("user is already deleted")
	ErrUserNotDeleted       = errors.New("user is not deleted")

	// Role-related errors
	ErrInvalidRole           = errors.New("invalid role")
	ErrCannotChangeOwnRole   = errors.New("cannot change own role")
	ErrCannotDemoteLastOwner = errors.New("cannot demote the last owner")
	ErrCannotDeleteLastOwner = errors.New("cannot delete the last owner")

	// Token-related errors
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenAlreadyUsed = errors.New("token has already been used")
	ErrTokenNotFound    = errors.New("token not found")

	// Session-related errors
	ErrSessionExpired  = errors.New("session has expired")
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidSession  = errors.New("invalid session")

	// Rate limiting errors
	ErrTooManyAttempts = errors.New("too many attempts, please try again later")
	ErrAccountLocked   = errors.New("account is temporarily locked due to too many failed attempts")

	// Profile-related errors
	ErrInvalidAvatarURL    = errors.New("invalid avatar URL")
	ErrAvatarTooLarge      = errors.New("avatar file is too large")
	ErrInvalidAvatarFormat = errors.New("invalid avatar format")

	// Team-related errors (for user-team relationships)
	ErrUserNotInTeam      = errors.New("user is not a member of this team")
	ErrUserAlreadyInTeam  = errors.New("user is already a member of this team")
	ErrCannotLeaveAsOwner = errors.New("cannot leave team as the only owner")

	// Database/Repository errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrTransactionFailed  = errors.New("transaction failed")
	ErrOptimisticLock     = errors.New("optimistic lock error: data was modified by another process")

	// Validation errors
	ErrValidationFailed      = errors.New("validation failed")
	ErrInvalidInput          = errors.New("invalid input")
	ErrMissingRequiredFields = errors.New("missing required fields")

	// Generic errors
	ErrInternal       = errors.New("internal error")
	ErrNotImplemented = errors.New("not implemented")

	ErrAccountSuspended = errors.New("account is suspended")
)

// ErrorCode represents a unique error code for API responses
type ErrorCode string

const (
	// User errors (1000-1099)
	CodeUserNotFound      ErrorCode = "USER_1001"
	CodeUserAlreadyExists ErrorCode = "USER_1002"
	CodeInvalidUser       ErrorCode = "USER_1003"

	// Email errors (1100-1199)
	CodeEmailRequired        ErrorCode = "EMAIL_1101"
	CodeInvalidEmailFormat   ErrorCode = "EMAIL_1102"
	CodeEmailAlreadyExists   ErrorCode = "EMAIL_1103"
	CodeEmailNotVerified     ErrorCode = "EMAIL_1104"
	CodeEmailAlreadyVerified ErrorCode = "EMAIL_1105"

	// Username errors (1200-1299)
	CodeUsernameRequired      ErrorCode = "USERNAME_1201"
	CodeUsernameTooShort      ErrorCode = "USERNAME_1202"
	CodeUsernameTooLong       ErrorCode = "USERNAME_1203"
	CodeInvalidUsernameFormat ErrorCode = "USERNAME_1204"
	CodeUsernameAlreadyExists ErrorCode = "USERNAME_1205"

	// Password errors (1300-1399)
	CodePasswordRequired   ErrorCode = "PASSWORD_1301"
	CodePasswordTooShort   ErrorCode = "PASSWORD_1302"
	CodePasswordTooLong    ErrorCode = "PASSWORD_1303"
	CodePasswordTooWeak    ErrorCode = "PASSWORD_1304"
	CodeInvalidOldPassword ErrorCode = "PASSWORD_1305"
	CodePasswordMismatch   ErrorCode = "PASSWORD_1306"

	// Auth errors (1400-1499)
	CodeInvalidCredentials      ErrorCode = "AUTH_1401"
	CodeUnauthorized            ErrorCode = "AUTH_1402"
	CodeInsufficientPermissions ErrorCode = "AUTH_1403"
	CodeSessionExpired          ErrorCode = "AUTH_1404"
	CodeInvalidToken            ErrorCode = "AUTH_1405"
	CodeTokenExpired            ErrorCode = "AUTH_1406"

	// Account status errors (1500-1599)
	CodeUserInactive  ErrorCode = "STATUS_1501"
	CodeUserSuspended ErrorCode = "STATUS_1502"
	CodeUserDeleted   ErrorCode = "STATUS_1503"
	CodeAccountLocked ErrorCode = "STATUS_1504"

	// Rate limiting errors (1600-1699)
	CodeTooManyAttempts ErrorCode = "RATE_1601"

	// System errors (1900-1999)
	CodeInternal         ErrorCode = "SYSTEM_1901"
	CodeDatabaseError    ErrorCode = "SYSTEM_1902"
	CodeValidationFailed ErrorCode = "SYSTEM_1903"
)

// DomainError provides structured error information
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface
func (e DomainError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string, err error) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ErrorMapping maps domain errors to error codes
var ErrorMapping = map[error]ErrorCode{
	ErrUserNotFound:            CodeUserNotFound,
	ErrUserAlreadyExists:       CodeUserAlreadyExists,
	ErrInvalidUser:             CodeInvalidUser,
	ErrEmailRequired:           CodeEmailRequired,
	ErrInvalidEmailFormat:      CodeInvalidEmailFormat,
	ErrEmailAlreadyExists:      CodeEmailAlreadyExists,
	ErrEmailNotVerified:        CodeEmailNotVerified,
	ErrEmailAlreadyVerified:    CodeEmailAlreadyVerified,
	ErrUsernameRequired:        CodeUsernameRequired,
	ErrUsernameTooShort:        CodeUsernameTooShort,
	ErrUsernameTooLong:         CodeUsernameTooLong,
	ErrInvalidUsernameFormat:   CodeInvalidUsernameFormat,
	ErrUsernameAlreadyExists:   CodeUsernameAlreadyExists,
	ErrPasswordRequired:        CodePasswordRequired,
	ErrPasswordTooShort:        CodePasswordTooShort,
	ErrPasswordTooLong:         CodePasswordTooLong,
	ErrPasswordTooWeak:         CodePasswordTooWeak,
	ErrInvalidOldPassword:      CodeInvalidOldPassword,
	ErrPasswordMismatch:        CodePasswordMismatch,
	ErrInvalidCredentials:      CodeInvalidCredentials,
	ErrUnauthorized:            CodeUnauthorized,
	ErrInsufficientPermissions: CodeInsufficientPermissions,
	ErrUserInactive:            CodeUserInactive,
	ErrUserSuspended:           CodeUserSuspended,
	ErrUserDeleted:             CodeUserDeleted,
	ErrAccountLocked:           CodeAccountLocked,
	ErrTooManyAttempts:         CodeTooManyAttempts,
	ErrSessionExpired:          CodeSessionExpired,
	ErrInvalidToken:            CodeInvalidToken,
	ErrTokenExpired:            CodeTokenExpired,
	ErrInternal:                CodeInternal,
	ErrDatabaseConnection:      CodeDatabaseError,
	ErrValidationFailed:        CodeValidationFailed,
}

// GetErrorCode returns the error code for a given error
func GetErrorCode(err error) ErrorCode {
	if code, ok := ErrorMapping[err]; ok {
		return code
	}
	return CodeInternal
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrTokenNotFound) ||
		errors.Is(err, ErrSessionNotFound)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrMissingRequiredFields) ||
		errors.Is(err, ErrInvalidEmailFormat) ||
		errors.Is(err, ErrInvalidUsernameFormat) ||
		errors.Is(err, ErrPasswordTooShort) ||
		errors.Is(err, ErrPasswordTooLong) ||
		errors.Is(err, ErrUsernameTooShort) ||
		errors.Is(err, ErrUsernameTooLong)
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrTokenExpired)
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	return errors.Is(err, ErrInsufficientPermissions) ||
		errors.Is(err, ErrCannotChangeOwnRole) ||
		errors.Is(err, ErrCannotDemoteLastOwner) ||
		errors.Is(err, ErrCannotDeleteLastOwner)
}

// IsStatusError checks if an error is related to user status
func IsStatusError(err error) bool {
	return errors.Is(err, ErrUserInactive) ||
		errors.Is(err, ErrUserSuspended) ||
		errors.Is(err, ErrUserDeleted) ||
		errors.Is(err, ErrEmailNotVerified) ||
		errors.Is(err, ErrAccountLocked)
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrTooManyAttempts) ||
		errors.Is(err, ErrAccountLocked)
}

// IsDuplicateError checks if an error is a duplicate/conflict error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists) ||
		errors.Is(err, ErrEmailAlreadyExists) ||
		errors.Is(err, ErrUsernameAlreadyExists) ||
		errors.Is(err, ErrUserAlreadyInTeam)
}
