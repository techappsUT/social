// path: backend/internal/domain/post/errors.go
// 🆕 NEW - Clean Architecture

package post

import "errors"

// Post-related errors
var (
	// Post validation errors
	ErrPostNotFound  = errors.New("post not found")
	ErrInvalidTeamID = errors.New("invalid team ID")
	ErrInvalidUserID = errors.New("invalid user ID")

	// Content errors
	ErrEmptyContent              = errors.New("post content cannot be empty")
	ErrContentTooLong            = errors.New("post content exceeds maximum length")
	ErrContentTooLongForPlatform = errors.New("content too long for selected platform")
	ErrInvalidContent            = errors.New("invalid post content")
	ErrTooManyMediaFiles         = errors.New("too many media files attached")
	ErrMediaTypeMismatch         = errors.New("media types don't match media URLs")
	ErrInvalidMediaType          = errors.New("invalid media type")
	ErrMediaSizeTooLarge         = errors.New("media file size too large")
	ErrInstagramRequiresMedia    = errors.New("Instagram posts require at least one media file")

	// Platform errors
	ErrNoPlatformsSelected  = errors.New("no platforms selected for post")
	ErrInvalidPlatform      = errors.New("invalid platform selected")
	ErrPlatformNotConnected = errors.New("platform account not connected")
	ErrPlatformLimitReached = errors.New("platform rate limit reached")

	// Scheduling errors
	ErrScheduleTimeInPast      = errors.New("schedule time cannot be in the past")
	ErrScheduleTimeTooFar      = errors.New("schedule time too far in the future")
	ErrNotScheduled            = errors.New("post is not scheduled")
	ErrAlreadyScheduled        = errors.New("post is already scheduled")
	ErrCannotSchedulePublished = errors.New("cannot schedule a published post")
	ErrScheduleConflict        = errors.New("scheduling conflict with another post")

	// Status errors
	ErrInvalidStatus             = errors.New("invalid post status")
	ErrCannotEditPublished       = errors.New("cannot edit published post")
	ErrCannotEditWhilePublishing = errors.New("cannot edit post while publishing")
	ErrCannotCancelPublished     = errors.New("cannot cancel published post")
	ErrPostCanceled              = errors.New("post has been canceled")
	ErrAlreadyCanceled           = errors.New("post is already canceled")
	ErrPostAlreadyDeleted        = errors.New("post is already deleted")
	ErrPostNotDeleted            = errors.New("post is not deleted")

	// Queue errors
	ErrNotQueued      = errors.New("post is not in queue")
	ErrAlreadyInQueue = errors.New("post is already in queue")
	ErrQueueFull      = errors.New("post queue is full")
	ErrNotPublishing  = errors.New("post is not being published")

	// Approval errors
	ErrRequiresApproval     = errors.New("post requires approval before publishing")
	ErrApprovalNotRequired  = errors.New("post does not require approval")
	ErrAlreadyApproved      = errors.New("post is already approved")
	ErrNotApproved          = errors.New("post is not approved")
	ErrCannotApproveOwnPost = errors.New("cannot approve your own post")

	// Priority errors
	ErrInvalidPriority = errors.New("invalid post priority")

	// Publishing errors
	ErrPublishFailed       = errors.New("failed to publish post")
	ErrMaxRetriesExceeded  = errors.New("maximum retry attempts exceeded")
	ErrAccountSuspended    = errors.New("social account is suspended")
	ErrAccountDisconnected = errors.New("social account is disconnected")

	// Analytics errors
	ErrAnalyticsNotAvailable = errors.New("analytics not available for this post")
	ErrAnalyticsFetchFailed  = errors.New("failed to fetch analytics")

	// Limit errors
	ErrDailyLimitExceeded    = errors.New("daily post limit exceeded")
	ErrTeamQuotaExceeded     = errors.New("team post quota exceeded")
	ErrScheduleLimitExceeded = errors.New("scheduled post limit exceeded")
)

// ErrorCode represents a unique error code for API responses
type ErrorCode string

const (
	// Post errors (3000-3099)
	CodePostNotFound  ErrorCode = "POST_3001"
	CodeInvalidPost   ErrorCode = "POST_3002"
	CodeInvalidTeamID ErrorCode = "POST_3003"
	CodeInvalidUserID ErrorCode = "POST_3004"

	// Content errors (3100-3199)
	CodeEmptyContent      ErrorCode = "CONTENT_3101"
	CodeContentTooLong    ErrorCode = "CONTENT_3102"
	CodeInvalidContent    ErrorCode = "CONTENT_3103"
	CodeTooManyMediaFiles ErrorCode = "CONTENT_3104"
	CodeInvalidMediaType  ErrorCode = "CONTENT_3105"
	CodeMediaSizeTooLarge ErrorCode = "CONTENT_3106"

	// Platform errors (3200-3299)
	CodeNoPlatformsSelected  ErrorCode = "PLATFORM_3201"
	CodeInvalidPlatform      ErrorCode = "PLATFORM_3202"
	CodePlatformNotConnected ErrorCode = "PLATFORM_3203"
	CodePlatformLimitReached ErrorCode = "PLATFORM_3204"

	// Scheduling errors (3300-3399)
	CodeScheduleTimeInvalid ErrorCode = "SCHEDULE_3301"
	CodeNotScheduled        ErrorCode = "SCHEDULE_3302"
	CodeAlreadyScheduled    ErrorCode = "SCHEDULE_3303"
	CodeScheduleConflict    ErrorCode = "SCHEDULE_3304"

	// Status errors (3400-3499)
	CodeInvalidStatus ErrorCode = "STATUS_3401"
	CodeCannotEdit    ErrorCode = "STATUS_3402"
	CodeCannotCancel  ErrorCode = "STATUS_3403"
	CodePostCanceled  ErrorCode = "STATUS_3404"

	// Queue errors (3500-3599)
	CodeNotQueued      ErrorCode = "QUEUE_3501"
	CodeAlreadyInQueue ErrorCode = "QUEUE_3502"
	CodeQueueFull      ErrorCode = "QUEUE_3503"

	// Approval errors (3600-3699)
	CodeRequiresApproval ErrorCode = "APPROVAL_3601"
	CodeNotApproved      ErrorCode = "APPROVAL_3602"
	CodeAlreadyApproved  ErrorCode = "APPROVAL_3603"

	// Publishing errors (3700-3799)
	CodePublishFailed      ErrorCode = "PUBLISH_3701"
	CodeMaxRetriesExceeded ErrorCode = "PUBLISH_3702"
	CodeAccountIssue       ErrorCode = "PUBLISH_3703"

	// Limit errors (3800-3899)
	CodeDailyLimitExceeded ErrorCode = "LIMIT_3801"
	CodeQuotaExceeded      ErrorCode = "LIMIT_3802"

	// System errors (3900-3999)
	CodePostInternal  ErrorCode = "POST_3901"
	CodeDatabaseError ErrorCode = "POST_3902"
)

// ErrorMapping maps domain errors to error codes
var ErrorMapping = map[error]ErrorCode{
	ErrPostNotFound:          CodePostNotFound,
	ErrInvalidTeamID:         CodeInvalidTeamID,
	ErrInvalidUserID:         CodeInvalidUserID,
	ErrEmptyContent:          CodeEmptyContent,
	ErrContentTooLong:        CodeContentTooLong,
	ErrInvalidContent:        CodeInvalidContent,
	ErrTooManyMediaFiles:     CodeTooManyMediaFiles,
	ErrInvalidMediaType:      CodeInvalidMediaType,
	ErrMediaSizeTooLarge:     CodeMediaSizeTooLarge,
	ErrNoPlatformsSelected:   CodeNoPlatformsSelected,
	ErrInvalidPlatform:       CodeInvalidPlatform,
	ErrPlatformNotConnected:  CodePlatformNotConnected,
	ErrPlatformLimitReached:  CodePlatformLimitReached,
	ErrScheduleTimeInPast:    CodeScheduleTimeInvalid,
	ErrNotScheduled:          CodeNotScheduled,
	ErrAlreadyScheduled:      CodeAlreadyScheduled,
	ErrScheduleConflict:      CodeScheduleConflict,
	ErrInvalidStatus:         CodeInvalidStatus,
	ErrCannotEditPublished:   CodeCannotEdit,
	ErrCannotCancelPublished: CodeCannotCancel,
	ErrPostCanceled:          CodePostCanceled,
	ErrNotQueued:             CodeNotQueued,
	ErrAlreadyInQueue:        CodeAlreadyInQueue,
	ErrQueueFull:             CodeQueueFull,
	ErrRequiresApproval:      CodeRequiresApproval,
	ErrNotApproved:           CodeNotApproved,
	ErrAlreadyApproved:       CodeAlreadyApproved,
	ErrPublishFailed:         CodePublishFailed,
	ErrMaxRetriesExceeded:    CodeMaxRetriesExceeded,
	ErrAccountSuspended:      CodeAccountIssue,
	ErrDailyLimitExceeded:    CodeDailyLimitExceeded,
	ErrTeamQuotaExceeded:     CodeQuotaExceeded,
}

// GetErrorCode returns the error code for a given error
func GetErrorCode(err error) ErrorCode {
	if code, ok := ErrorMapping[err]; ok {
		return code
	}
	return CodePostInternal
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrPostNotFound)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrEmptyContent) ||
		errors.Is(err, ErrContentTooLong) ||
		errors.Is(err, ErrInvalidContent) ||
		errors.Is(err, ErrTooManyMediaFiles) ||
		errors.Is(err, ErrInvalidMediaType) ||
		errors.Is(err, ErrNoPlatformsSelected) ||
		errors.Is(err, ErrInvalidPlatform) ||
		errors.Is(err, ErrScheduleTimeInPast) ||
		errors.Is(err, ErrScheduleTimeTooFar) ||
		errors.Is(err, ErrInvalidPriority)
}

// IsStatusError checks if an error is related to post status
func IsStatusError(err error) bool {
	return errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrCannotEditPublished) ||
		errors.Is(err, ErrCannotEditWhilePublishing) ||
		errors.Is(err, ErrCannotCancelPublished) ||
		errors.Is(err, ErrPostCanceled) ||
		errors.Is(err, ErrAlreadyCanceled)
}

// IsSchedulingError checks if an error is related to scheduling
func IsSchedulingError(err error) bool {
	return errors.Is(err, ErrScheduleTimeInPast) ||
		errors.Is(err, ErrScheduleTimeTooFar) ||
		errors.Is(err, ErrNotScheduled) ||
		errors.Is(err, ErrAlreadyScheduled) ||
		errors.Is(err, ErrCannotSchedulePublished) ||
		errors.Is(err, ErrScheduleConflict)
}

// IsPublishingError checks if an error is related to publishing
func IsPublishingError(err error) bool {
	return errors.Is(err, ErrPublishFailed) ||
		errors.Is(err, ErrMaxRetriesExceeded) ||
		errors.Is(err, ErrAccountSuspended) ||
		errors.Is(err, ErrAccountDisconnected) ||
		errors.Is(err, ErrPlatformNotConnected) ||
		errors.Is(err, ErrPlatformLimitReached)
}

// IsLimitError checks if an error is a limit/quota error
func IsLimitError(err error) bool {
	return errors.Is(err, ErrDailyLimitExceeded) ||
		errors.Is(err, ErrTeamQuotaExceeded) ||
		errors.Is(err, ErrScheduleLimitExceeded) ||
		errors.Is(err, ErrQueueFull) ||
		errors.Is(err, ErrMediaSizeTooLarge) ||
		errors.Is(err, ErrTooManyMediaFiles)
}

// IsApprovalError checks if an error is related to approval workflow
func IsApprovalError(err error) bool {
	return errors.Is(err, ErrRequiresApproval) ||
		errors.Is(err, ErrApprovalNotRequired) ||
		errors.Is(err, ErrAlreadyApproved) ||
		errors.Is(err, ErrNotApproved) ||
		errors.Is(err, ErrCannotApproveOwnPost)
}
