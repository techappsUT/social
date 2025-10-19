// path: backend/internal/domain/team/team.go
// 🆕 NEW - Clean Architecture

package team

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Team represents a team/organization in the system
// Teams are the core multi-tenant entity - all resources belong to a team
type Team struct {
	id          uuid.UUID
	name        string
	slug        string // URL-friendly identifier
	description string
	avatarURL   string
	ownerID     uuid.UUID // User who owns the team
	plan        Plan
	status      Status
	settings    TeamSettings
	limits      TeamLimits
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// Plan represents the subscription plan for a team
type Plan string

const (
	PlanFree         Plan = "free"
	PlanStarter      Plan = "starter"
	PlanProfessional Plan = "professional"
	PlanEnterprise   Plan = "enterprise"
)

// Status represents the team's status
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
	StatusTrial     Status = "trial"
)

// TeamSettings holds team-specific settings
type TeamSettings struct {
	Timezone            string
	DefaultPostTime     string // e.g., "09:00"
	EnableNotifications bool
	EnableAnalytics     bool
	RequireApproval     bool // Require approval before posting
	AutoSchedule        bool
	Language            string
	DateFormat          string
}

// TeamLimits holds plan-based limitations
type TeamLimits struct {
	MaxMembers         int
	MaxSocialAccounts  int
	MaxScheduledPosts  int
	MaxPostsPerDay     int
	MaxMediaSize       int64 // in bytes
	AnalyticsRetention int   // in days
	APIRateLimit       int   // requests per hour
}

// NewTeam creates a new team with validation
func NewTeam(name, slug, description string, ownerID uuid.UUID) (*Team, error) {
	// Validate name
	if err := validateTeamName(name); err != nil {
		return nil, err
	}

	// Validate slug
	if err := validateSlug(slug); err != nil {
		return nil, err
	}

	if ownerID == uuid.Nil {
		return nil, ErrInvalidOwner
	}

	now := time.Now().UTC()

	// Set default limits for free plan
	limits := getDefaultLimits(PlanFree)

	// Default settings
	settings := TeamSettings{
		Timezone:            "UTC",
		DefaultPostTime:     "10:00",
		EnableNotifications: true,
		EnableAnalytics:     true,
		RequireApproval:     false,
		AutoSchedule:        false,
		Language:            "en",
		DateFormat:          "MM/DD/YYYY",
	}

	return &Team{
		id:          uuid.New(),
		name:        strings.TrimSpace(name),
		slug:        strings.ToLower(strings.TrimSpace(slug)),
		description: strings.TrimSpace(description),
		ownerID:     ownerID,
		plan:        PlanFree,
		status:      StatusTrial, // Start with trial
		settings:    settings,
		limits:      limits,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstruct recreates a team entity from persistence
func Reconstruct(
	id uuid.UUID,
	name string,
	slug string,
	description string,
	avatarURL string,
	ownerID uuid.UUID,
	plan Plan,
	status Status,
	settings TeamSettings,
	limits TeamLimits,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Team {
	return &Team{
		id:          id,
		name:        name,
		slug:        slug,
		description: description,
		avatarURL:   avatarURL,
		ownerID:     ownerID,
		plan:        plan,
		status:      status,
		settings:    settings,
		limits:      limits,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}
}

// Getters - Encapsulation of internal state

func (t *Team) ID() uuid.UUID          { return t.id }
func (t *Team) Name() string           { return t.name }
func (t *Team) Slug() string           { return t.slug }
func (t *Team) Description() string    { return t.description }
func (t *Team) AvatarURL() string      { return t.avatarURL }
func (t *Team) OwnerID() uuid.UUID     { return t.ownerID }
func (t *Team) Plan() Plan             { return t.plan }
func (t *Team) Status() Status         { return t.status }
func (t *Team) Settings() TeamSettings { return t.settings }
func (t *Team) Limits() TeamLimits     { return t.limits }
func (t *Team) CreatedAt() time.Time   { return t.createdAt }
func (t *Team) UpdatedAt() time.Time   { return t.updatedAt }
func (t *Team) DeletedAt() *time.Time  { return t.deletedAt }

// Business Logic Methods

// UpdateProfile updates team profile information
func (t *Team) UpdateProfile(name, description, avatarURL string) error {
	if err := validateTeamName(name); err != nil {
		return err
	}

	t.name = strings.TrimSpace(name)
	t.description = strings.TrimSpace(description)
	t.avatarURL = strings.TrimSpace(avatarURL)
	t.updatedAt = time.Now().UTC()
	return nil
}

// ChangeSlug updates the team's slug (URL identifier)
func (t *Team) ChangeSlug(newSlug string) error {
	if err := validateSlug(newSlug); err != nil {
		return err
	}

	t.slug = strings.ToLower(strings.TrimSpace(newSlug))
	t.updatedAt = time.Now().UTC()
	return nil
}

// TransferOwnership transfers team ownership to another user
func (t *Team) TransferOwnership(newOwnerID uuid.UUID) error {
	if newOwnerID == uuid.Nil {
		return ErrInvalidOwner
	}

	if t.ownerID == newOwnerID {
		return ErrAlreadyOwner
	}

	t.ownerID = newOwnerID
	t.updatedAt = time.Now().UTC()
	return nil
}

// UpgradePlan upgrades the team's subscription plan
func (t *Team) UpgradePlan(newPlan Plan) error {
	if !isValidPlan(newPlan) {
		return ErrInvalidPlan
	}

	if t.plan == newPlan {
		return ErrSamePlan
	}

	// Update limits based on new plan
	t.limits = getDefaultLimits(newPlan)
	t.plan = newPlan

	// Activate team if upgrading from trial
	if t.status == StatusTrial && newPlan != PlanFree {
		t.status = StatusActive
	}

	t.updatedAt = time.Now().UTC()
	return nil
}

// DowngradePlan downgrades the team's subscription plan
func (t *Team) DowngradePlan(newPlan Plan) error {
	if !isValidPlan(newPlan) {
		return ErrInvalidPlan
	}

	if t.plan == newPlan {
		return ErrSamePlan
	}

	// Check if downgrade is allowed (e.g., enterprise to professional)
	if !canDowngrade(t.plan, newPlan) {
		return ErrCannotDowngrade
	}

	t.limits = getDefaultLimits(newPlan)
	t.plan = newPlan
	t.updatedAt = time.Now().UTC()
	return nil
}

// UpdateSettings updates team settings
func (t *Team) UpdateSettings(settings TeamSettings) error {
	// Validate timezone
	if settings.Timezone == "" {
		return ErrInvalidTimezone
	}

	// Validate language
	if settings.Language == "" {
		settings.Language = "en"
	}

	t.settings = settings
	t.updatedAt = time.Now().UTC()
	return nil
}

// Suspend suspends the team (e.g., for non-payment)
func (t *Team) Suspend(reason string) error {
	if t.status == StatusSuspended {
		return ErrTeamAlreadySuspended
	}

	t.status = StatusSuspended
	t.updatedAt = time.Now().UTC()
	// Note: reason could be stored in an audit log
	return nil
}

// Activate activates a suspended or inactive team
func (t *Team) Activate() error {
	if t.status == StatusActive {
		return ErrTeamAlreadyActive
	}

	t.status = StatusActive
	t.updatedAt = time.Now().UTC()
	return nil
}

// SoftDelete marks the team as deleted
func (t *Team) SoftDelete() error {
	if t.deletedAt != nil {
		return ErrTeamAlreadyDeleted
	}

	now := time.Now().UTC()
	t.deletedAt = &now
	t.status = StatusInactive
	t.updatedAt = now
	return nil
}

// Restore restores a soft-deleted team
func (t *Team) Restore() error {
	if t.deletedAt == nil {
		return ErrTeamNotDeleted
	}

	t.deletedAt = nil
	t.status = StatusActive
	t.updatedAt = time.Now().UTC()
	return nil
}

// Business Rule Checks

// CanAddMember checks if the team can add more members
func (t *Team) CanAddMember(currentMemberCount int) bool {
	return currentMemberCount < t.limits.MaxMembers &&
		t.status == StatusActive &&
		t.deletedAt == nil
}

// CanAddSocialAccount checks if the team can add more social accounts
func (t *Team) CanAddSocialAccount(currentAccountCount int) bool {
	return currentAccountCount < t.limits.MaxSocialAccounts &&
		t.status == StatusActive &&
		t.deletedAt == nil
}

// CanSchedulePost checks if the team can schedule more posts
func (t *Team) CanSchedulePost(currentScheduledCount int) bool {
	return currentScheduledCount < t.limits.MaxScheduledPosts &&
		t.status == StatusActive &&
		t.deletedAt == nil
}

// CanPublishToday checks if the team can publish more posts today
func (t *Team) CanPublishToday(publishedTodayCount int) bool {
	return publishedTodayCount < t.limits.MaxPostsPerDay &&
		t.status == StatusActive &&
		t.deletedAt == nil
}

// CanUploadMedia checks if media size is within limits
func (t *Team) CanUploadMedia(sizeInBytes int64) bool {
	return sizeInBytes <= t.limits.MaxMediaSize &&
		t.status == StatusActive &&
		t.deletedAt == nil
}

// IsActive checks if the team is active
func (t *Team) IsActive() bool {
	return t.status == StatusActive && t.deletedAt == nil
}

// IsSuspended checks if the team is suspended
func (t *Team) IsSuspended() bool {
	return t.status == StatusSuspended
}

// IsOnTrial checks if the team is on trial
func (t *Team) IsOnTrial() bool {
	return t.status == StatusTrial
}

// IsPaid checks if the team is on a paid plan
func (t *Team) IsPaid() bool {
	return t.plan != PlanFree
}

// HasFeature checks if the team's plan includes a specific feature
func (t *Team) HasFeature(feature string) bool {
	switch feature {
	case "analytics":
		return t.plan != PlanFree
	case "team_members":
		return t.plan != PlanFree
	case "bulk_scheduling":
		return t.plan == PlanProfessional || t.plan == PlanEnterprise
	case "api_access":
		return t.plan == PlanEnterprise
	case "custom_branding":
		return t.plan == PlanEnterprise
	case "priority_support":
		return t.plan == PlanProfessional || t.plan == PlanEnterprise
	default:
		return false
	}
}

// GetTrialDaysRemaining calculates remaining trial days
func (t *Team) GetTrialDaysRemaining() int {
	if t.status != StatusTrial {
		return 0
	}

	trialDuration := 14 * 24 * time.Hour // 14 days trial
	elapsed := time.Since(t.createdAt)
	remaining := trialDuration - elapsed

	if remaining <= 0 {
		return 0
	}

	return int(remaining.Hours() / 24)
}

// Helper Functions

func validateTeamName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrTeamNameRequired
	}

	if len(name) < 2 {
		return ErrTeamNameTooShort
	}

	if len(name) > 100 {
		return ErrTeamNameTooLong
	}

	return nil
}

func validateSlug(slug string) error {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return ErrSlugRequired
	}

	if len(slug) < 3 {
		return ErrSlugTooShort
	}

	if len(slug) > 50 {
		return ErrSlugTooLong
	}

	// Check for valid characters (alphanumeric and dash)
	for _, ch := range slug {
		if !isAlphanumeric(ch) && ch != '-' {
			return ErrInvalidSlugFormat
		}
	}

	// Cannot start or end with dash
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return ErrInvalidSlugFormat
	}

	return nil
}

func isAlphanumeric(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9')
}

func isValidPlan(plan Plan) bool {
	switch plan {
	case PlanFree, PlanStarter, PlanProfessional, PlanEnterprise:
		return true
	default:
		return false
	}
}

func canDowngrade(currentPlan, newPlan Plan) bool {
	// Define downgrade paths
	switch currentPlan {
	case PlanEnterprise:
		return true // Can downgrade to any plan
	case PlanProfessional:
		return newPlan != PlanEnterprise
	case PlanStarter:
		return newPlan == PlanFree
	case PlanFree:
		return false // Cannot downgrade from free
	default:
		return false
	}
}

func getDefaultLimits(plan Plan) TeamLimits {
	switch plan {
	case PlanFree:
		return TeamLimits{
			MaxMembers:         3,
			MaxSocialAccounts:  3,
			MaxScheduledPosts:  10,
			MaxPostsPerDay:     5,
			MaxMediaSize:       5 * 1024 * 1024, // 5MB
			AnalyticsRetention: 7,               // 7 days
			APIRateLimit:       0,               // No API access
		}
	case PlanStarter:
		return TeamLimits{
			MaxMembers:         3,
			MaxSocialAccounts:  10,
			MaxScheduledPosts:  100,
			MaxPostsPerDay:     50,
			MaxMediaSize:       25 * 1024 * 1024, // 25MB
			AnalyticsRetention: 30,               // 30 days
			APIRateLimit:       100,              // 100 requests/hour
		}
	case PlanProfessional:
		return TeamLimits{
			MaxMembers:         10,
			MaxSocialAccounts:  25,
			MaxScheduledPosts:  500,
			MaxPostsPerDay:     200,
			MaxMediaSize:       100 * 1024 * 1024, // 100MB
			AnalyticsRetention: 90,                // 90 days
			APIRateLimit:       1000,              // 1000 requests/hour
		}
	case PlanEnterprise:
		return TeamLimits{
			MaxMembers:         -1,                // Unlimited
			MaxSocialAccounts:  -1,                // Unlimited
			MaxScheduledPosts:  -1,                // Unlimited
			MaxPostsPerDay:     -1,                // Unlimited
			MaxMediaSize:       500 * 1024 * 1024, // 500MB
			AnalyticsRetention: 365,               // 1 year
			APIRateLimit:       -1,                // Unlimited
		}
	default:
		return getDefaultLimits(PlanFree)
	}
}

// ===========================================================================
// FILE: backend/internal/domain/team/team.go
// ADD this method to your existing team.go file
// ===========================================================================

// Add this method after the NewTeam function

// ReconstructTeam recreates a team from persistence (repository)
// This is used when loading teams from the database
func ReconstructTeam(
	id uuid.UUID,
	name string,
	slug string,
	description string,
	ownerID uuid.UUID,
	settings TeamSettings,
) (*Team, error) {
	// Validate the data
	if err := validateTeamName(name); err != nil {
		return nil, err
	}

	if err := validateSlug(slug); err != nil {
		return nil, err
	}

	if ownerID == uuid.Nil {
		return nil, ErrInvalidOwner
	}

	// Get limits based on plan (default to free)
	limits := getDefaultLimits(PlanFree)

	return &Team{
		id:          id, // ✅ Use the ID from database
		name:        name,
		slug:        slug,
		description: description,
		ownerID:     ownerID,
		plan:        PlanFree,     // Default, should be stored in DB later
		status:      StatusActive, // Default
		settings:    settings,
		limits:      limits,
		createdAt:   time.Now().UTC(), // Ideally these should be parameters too
		updatedAt:   time.Now().UTC(),
	}, nil
}
