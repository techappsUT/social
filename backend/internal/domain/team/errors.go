// path: backend/internal/domain/team/errors.go
// 🆕 NEW - Clean Architecture

package team

import "errors"

// Team-related errors
var (
	// Team validation errors
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNameRequired  = errors.New("team name is required")
	ErrTeamNameTooShort  = errors.New("team name must be at least 2 characters")
	ErrTeamNameTooLong   = errors.New("team name must be at most 100 characters")

	// Slug errors
	ErrSlugRequired      = errors.New("slug is required")
	ErrSlugTooShort      = errors.New("slug must be at least 3 characters")
	ErrSlugTooLong       = errors.New("slug must be at most 50 characters")
	ErrInvalidSlugFormat = errors.New("slug can only contain letters, numbers, and dashes")
	ErrSlugAlreadyExists = errors.New("slug already exists")

	// Owner errors
	ErrInvalidOwner          = errors.New("invalid owner")
	ErrAlreadyOwner          = errors.New("user is already the owner")
	ErrCannotRemoveLastOwner = errors.New("cannot remove the last owner")
	ErrOwnerCannotLeave      = errors.New("owner cannot leave team, transfer ownership first")

	// Plan errors
	ErrInvalidPlan       = errors.New("invalid plan")
	ErrSamePlan          = errors.New("team is already on this plan")
	ErrCannotDowngrade   = errors.New("cannot downgrade to this plan")
	ErrPlanLimitExceeded = errors.New("plan limit exceeded")
	ErrTrialExpired      = errors.New("trial period has expired")

	// Status errors
	ErrTeamInactive         = errors.New("team is inactive")
	ErrTeamSuspended        = errors.New("team is suspended")
	ErrTeamDeleted          = errors.New("team is deleted")
	ErrTeamAlreadySuspended = errors.New("team is already suspended")
	ErrTeamAlreadyActive    = errors.New("team is already active")
	ErrTeamAlreadyDeleted   = errors.New("team is already deleted")
	ErrTeamNotDeleted       = errors.New("team is not deleted")

	// Limit errors
	ErrMemberLimitExceeded  = errors.New("member limit exceeded for this plan")
	ErrAccountLimitExceeded = errors.New("social account limit exceeded for this plan")
	ErrPostLimitExceeded    = errors.New("post limit exceeded for this plan")
	ErrStorageLimitExceeded = errors.New("storage limit exceeded for this plan")
	ErrAPILimitExceeded     = errors.New("API rate limit exceeded for this plan")
	ErrMediaSizeTooLarge    = errors.New("media file size exceeds limit")

	// Settings errors
	ErrInvalidTimezone   = errors.New("invalid timezone")
	ErrInvalidLanguage   = errors.New("invalid language")
	ErrInvalidDateFormat = errors.New("invalid date format")

	// Permission errors
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrFeatureNotAvailable     = errors.New("feature not available in current plan")
)

// Member-related errors
var (
	// Member validation errors
	ErrMemberNotFound      = errors.New("member not found")
	ErrMemberAlreadyExists = errors.New("member already exists")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidTeamID       = errors.New("invalid team ID")
	ErrInvalidMemberRole   = errors.New("invalid member role")

	// Member status errors
	ErrMemberAlreadyActive   = errors.New("member is already active")
	ErrMemberAlreadyInactive = errors.New("member is already inactive")
	ErrMemberAlreadyLeft     = errors.New("member has already left")
	ErrMemberNotInactive     = errors.New("member is not inactive")

	// Role errors
	ErrSameRole              = errors.New("member already has this role")
	ErrCannotDemoteOnlyOwner = errors.New("cannot demote the only owner")
	ErrCannotRemoveOwner     = errors.New("cannot remove owner from team")
	ErrOwnerRequired         = errors.New("team must have at least one owner")

	// Invitation errors
	ErrInvitationNotFound         = errors.New("invitation not found")
	ErrInvitationExpired          = errors.New("invitation has expired")
	ErrInvitationAlreadySent      = errors.New("invitation already sent to this email")
	ErrInvitationAlreadyAccepted  = errors.New("invitation already accepted")
	ErrInvitationAlreadyProcessed = errors.New("invitation already processed")
	ErrInvalidInvitationToken     = errors.New("invalid invitation token")
	ErrCannotInviteSelf           = errors.New("cannot invite yourself")
	ErrUserAlreadyMember          = errors.New("user is already a team member")

	// Permission errors for members
	ErrCannotManageTeam    = errors.New("cannot manage team settings")
	ErrCannotManageMembers = errors.New("cannot manage team members")
	ErrCannotCreatePosts   = errors.New("cannot create posts")
	ErrCannotEditPosts     = errors.New("cannot edit posts")
	ErrCannotDeletePosts   = errors.New("cannot delete posts")
	ErrCannotManageBilling = errors.New("cannot manage billing")
	ErrCannotViewAnalytics = errors.New("cannot view analytics")
)

// ErrorCode represents a unique error code for API responses
type ErrorCode string

const (
	// Team errors (2000-2099)
	CodeTeamNotFound      ErrorCode = "TEAM_2001"
	CodeTeamAlreadyExists ErrorCode = "TEAM_2002"
	CodeTeamNameRequired  ErrorCode = "TEAM_2003"
	CodeInvalidTeamName   ErrorCode = "TEAM_2004"
	CodeSlugAlreadyExists ErrorCode = "TEAM_2005"
	CodeInvalidSlug       ErrorCode = "TEAM_2006"

	// Plan errors (2100-2199)
	CodeInvalidPlan         ErrorCode = "PLAN_2101"
	CodePlanLimitExceeded   ErrorCode = "PLAN_2102"
	CodeCannotDowngrade     ErrorCode = "PLAN_2103"
	CodeTrialExpired        ErrorCode = "PLAN_2104"
	CodeFeatureNotAvailable ErrorCode = "PLAN_2105"

	// Member errors (2200-2299)
	CodeMemberNotFound      ErrorCode = "MEMBER_2201"
	CodeMemberAlreadyExists ErrorCode = "MEMBER_2202"
	CodeInvalidMemberRole   ErrorCode = "MEMBER_2203"
	CodeCannotRemoveOwner   ErrorCode = "MEMBER_2204"
	CodeMemberLimitExceeded ErrorCode = "MEMBER_2205"

	// Invitation errors (2300-2399)
	CodeInvitationNotFound     ErrorCode = "INVITE_2301"
	CodeInvitationExpired      ErrorCode = "INVITE_2302"
	CodeInvitationAlreadySent  ErrorCode = "INVITE_2303"
	CodeInvalidInvitationToken ErrorCode = "INVITE_2304"

	// Permission errors (2400-2499)
	CodeUnauthorized            ErrorCode = "PERM_2401"
	CodeInsufficientPermissions ErrorCode = "PERM_2402"

	// System errors (2900-2999)
	CodeTeamInternal  ErrorCode = "TEAM_2901"
	CodeDatabaseError ErrorCode = "TEAM_2902"
)

// ErrorMapping maps domain errors to error codes
var ErrorMapping = map[error]ErrorCode{
	ErrTeamNotFound:            CodeTeamNotFound,
	ErrTeamAlreadyExists:       CodeTeamAlreadyExists,
	ErrTeamNameRequired:        CodeTeamNameRequired,
	ErrSlugAlreadyExists:       CodeSlugAlreadyExists,
	ErrInvalidSlugFormat:       CodeInvalidSlug,
	ErrInvalidPlan:             CodeInvalidPlan,
	ErrPlanLimitExceeded:       CodePlanLimitExceeded,
	ErrCannotDowngrade:         CodeCannotDowngrade,
	ErrTrialExpired:            CodeTrialExpired,
	ErrFeatureNotAvailable:     CodeFeatureNotAvailable,
	ErrMemberNotFound:          CodeMemberNotFound,
	ErrMemberAlreadyExists:     CodeMemberAlreadyExists,
	ErrInvalidMemberRole:       CodeInvalidMemberRole,
	ErrCannotRemoveOwner:       CodeCannotRemoveOwner,
	ErrMemberLimitExceeded:     CodeMemberLimitExceeded,
	ErrInvitationNotFound:      CodeInvitationNotFound,
	ErrInvitationExpired:       CodeInvitationExpired,
	ErrInvitationAlreadySent:   CodeInvitationAlreadySent,
	ErrInvalidInvitationToken:  CodeInvalidInvitationToken,
	ErrUnauthorized:            CodeUnauthorized,
	ErrInsufficientPermissions: CodeInsufficientPermissions,
}

// GetErrorCode returns the error code for a given error
func GetErrorCode(err error) ErrorCode {
	if code, ok := ErrorMapping[err]; ok {
		return code
	}
	return CodeTeamInternal
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrTeamNotFound) ||
		errors.Is(err, ErrMemberNotFound) ||
		errors.Is(err, ErrInvitationNotFound)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrTeamNameRequired) ||
		errors.Is(err, ErrTeamNameTooShort) ||
		errors.Is(err, ErrTeamNameTooLong) ||
		errors.Is(err, ErrSlugRequired) ||
		errors.Is(err, ErrSlugTooShort) ||
		errors.Is(err, ErrSlugTooLong) ||
		errors.Is(err, ErrInvalidSlugFormat) ||
		errors.Is(err, ErrInvalidPlan) ||
		errors.Is(err, ErrInvalidMemberRole) ||
		errors.Is(err, ErrInvalidTimezone)
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInsufficientPermissions) ||
		errors.Is(err, ErrCannotManageTeam) ||
		errors.Is(err, ErrCannotManageMembers) ||
		errors.Is(err, ErrCannotCreatePosts) ||
		errors.Is(err, ErrCannotEditPosts) ||
		errors.Is(err, ErrCannotDeletePosts) ||
		errors.Is(err, ErrCannotManageBilling)
}

// IsLimitError checks if an error is a limit/quota error
func IsLimitError(err error) bool {
	return errors.Is(err, ErrMemberLimitExceeded) ||
		errors.Is(err, ErrAccountLimitExceeded) ||
		errors.Is(err, ErrPostLimitExceeded) ||
		errors.Is(err, ErrStorageLimitExceeded) ||
		errors.Is(err, ErrAPILimitExceeded) ||
		errors.Is(err, ErrPlanLimitExceeded) ||
		errors.Is(err, ErrMediaSizeTooLarge)
}

// IsDuplicateError checks if an error is a duplicate/conflict error
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrTeamAlreadyExists) ||
		errors.Is(err, ErrSlugAlreadyExists) ||
		errors.Is(err, ErrMemberAlreadyExists) ||
		errors.Is(err, ErrInvitationAlreadySent) ||
		errors.Is(err, ErrUserAlreadyMember)
}

// IsStatusError checks if an error is related to status
func IsStatusError(err error) bool {
	return errors.Is(err, ErrTeamInactive) ||
		errors.Is(err, ErrTeamSuspended) ||
		errors.Is(err, ErrTeamDeleted) ||
		errors.Is(err, ErrMemberAlreadyActive) ||
		errors.Is(err, ErrMemberAlreadyInactive) ||
		errors.Is(err, ErrMemberAlreadyLeft)
}
