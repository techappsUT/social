// path: backend/internal/domain/social/account.go
// 🆕 NEW - Clean Architecture

package social

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Account represents a connected social media account
type Account struct {
	id          uuid.UUID
	teamID      uuid.UUID
	userID      uuid.UUID // User who connected the account
	platform    Platform
	accountType AccountType
	username    string
	displayName string
	profileURL  string
	avatarURL   string
	credentials Credentials
	metadata    AccountMetadata
	status      Status
	rateLimits  RateLimits
	lastSyncAt  *time.Time
	connectedAt time.Time
	expiresAt   *time.Time
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// Platform represents a social media platform
type Platform string

const (
	PlatformTwitter   Platform = "twitter"
	PlatformFacebook  Platform = "facebook"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformTikTok    Platform = "tiktok"
	PlatformPinterest Platform = "pinterest"
	PlatformYouTube   Platform = "youtube"
)

// AccountType represents the type of social account
type AccountType string

const (
	AccountTypePersonal AccountType = "personal"
	AccountTypeBusiness AccountType = "business"
	AccountTypePage     AccountType = "page"
	AccountTypeGroup    AccountType = "group"
	AccountTypeChannel  AccountType = "channel"
)

// Status represents the account connection status
type Status string

const (
	StatusActive            Status = "active"
	StatusInactive          Status = "inactive"
	StatusExpired           Status = "expired"
	StatusRevoked           Status = "revoked"
	StatusRateLimited       Status = "rate_limited"
	StatusSuspended         Status = "suspended"
	StatusReconnectRequired Status = "reconnect_required"
)

// Credentials holds OAuth tokens and secrets
type Credentials struct {
	AccessToken          string
	RefreshToken         string
	TokenSecret          string // For OAuth 1.0a (Twitter)
	ExpiresAt            *time.Time
	Scope                []string
	PlatformUserID       string
	PlatformAccountID    string // For platforms with multiple accounts
	EncryptedAt          time.Time
	EncryptionKeyVersion int
}

// AccountMetadata holds platform-specific metadata
type AccountMetadata struct {
	FollowersCount   int
	FollowingCount   int
	PostsCount       int
	Verified         bool
	BusinessCategory string
	Location         string
	Website          string
	Bio              string
	PlatformFeatures []string // Platform-specific features enabled
	CustomFields     map[string]interface{}
}

// RateLimits holds platform rate limit information
type RateLimits struct {
	PostsPerHour     int
	PostsPerDay      int
	MediaPerPost     int
	CharacterLimit   int
	HashtagLimit     int
	MentionLimit     int
	CurrentHourPosts int
	CurrentDayPosts  int
	ResetsAt         time.Time
	CustomLimits     map[string]int
}

// NewAccount creates a new social media account connection
func NewAccount(teamID, userID uuid.UUID, platform Platform, accountType AccountType) (*Account, error) {
	if teamID == uuid.Nil {
		return nil, ErrInvalidTeamID
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if !isValidPlatform(platform) {
		return nil, ErrInvalidPlatform
	}
	if !isValidAccountType(accountType) {
		return nil, ErrInvalidAccountType
	}

	now := time.Now().UTC()

	return &Account{
		id:          uuid.New(),
		teamID:      teamID,
		userID:      userID,
		platform:    platform,
		accountType: accountType,
		status:      StatusInactive, // Will be activated after OAuth
		rateLimits:  getDefaultRateLimits(platform),
		metadata: AccountMetadata{
			PlatformFeatures: []string{},
			CustomFields:     make(map[string]interface{}),
		},
		connectedAt: now,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstruct recreates an account from persistence
func Reconstruct(
	id uuid.UUID,
	teamID uuid.UUID,
	userID uuid.UUID,
	platform Platform,
	accountType AccountType,
	username string,
	displayName string,
	profileURL string,
	avatarURL string,
	credentials Credentials,
	metadata AccountMetadata,
	status Status,
	rateLimits RateLimits,
	lastSyncAt *time.Time,
	connectedAt time.Time,
	expiresAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Account {
	return &Account{
		id:          id,
		teamID:      teamID,
		userID:      userID,
		platform:    platform,
		accountType: accountType,
		username:    username,
		displayName: displayName,
		profileURL:  profileURL,
		avatarURL:   avatarURL,
		credentials: credentials,
		metadata:    metadata,
		status:      status,
		rateLimits:  rateLimits,
		lastSyncAt:  lastSyncAt,
		connectedAt: connectedAt,
		expiresAt:   expiresAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}
}

// Getters
func (a *Account) ID() uuid.UUID             { return a.id }
func (a *Account) TeamID() uuid.UUID         { return a.teamID }
func (a *Account) UserID() uuid.UUID         { return a.userID }
func (a *Account) Platform() Platform        { return a.platform }
func (a *Account) AccountType() AccountType  { return a.accountType }
func (a *Account) Username() string          { return a.username }
func (a *Account) DisplayName() string       { return a.displayName }
func (a *Account) ProfileURL() string        { return a.profileURL }
func (a *Account) AvatarURL() string         { return a.avatarURL }
func (a *Account) Credentials() Credentials  { return a.credentials }
func (a *Account) Metadata() AccountMetadata { return a.metadata }
func (a *Account) Status() Status            { return a.status }
func (a *Account) RateLimits() RateLimits    { return a.rateLimits }
func (a *Account) LastSyncAt() *time.Time    { return a.lastSyncAt }
func (a *Account) ConnectedAt() time.Time    { return a.connectedAt }
func (a *Account) ExpiresAt() *time.Time     { return a.expiresAt }
func (a *Account) CreatedAt() time.Time      { return a.createdAt }
func (a *Account) UpdatedAt() time.Time      { return a.updatedAt }
func (a *Account) DeletedAt() *time.Time     { return a.deletedAt }

// Business Logic Methods

// Connect completes the OAuth connection with credentials
func (a *Account) Connect(credentials Credentials, profile ProfileInfo) error {
	if a.status == StatusActive {
		return ErrAccountAlreadyConnected
	}

	a.credentials = credentials
	a.username = profile.Username
	a.displayName = profile.DisplayName
	a.profileURL = profile.ProfileURL
	a.avatarURL = profile.AvatarURL
	a.metadata.FollowersCount = profile.FollowersCount
	a.metadata.FollowingCount = profile.FollowingCount
	a.metadata.PostsCount = profile.PostsCount
	a.metadata.Verified = profile.Verified
	a.metadata.Bio = profile.Bio

	// Set expiration if token has expiry
	if credentials.ExpiresAt != nil {
		a.expiresAt = credentials.ExpiresAt
	}

	a.status = StatusActive
	a.updatedAt = time.Now().UTC()
	return nil
}

// Disconnect disconnects the social account
func (a *Account) Disconnect() error {
	if a.status == StatusInactive {
		return ErrAccountNotConnected
	}

	a.status = StatusInactive
	// Clear sensitive credentials
	a.credentials = Credentials{}
	a.updatedAt = time.Now().UTC()
	return nil
}

// RefreshCredentials updates the account credentials after token refresh
func (a *Account) RefreshCredentials(credentials Credentials) error {
	if a.status == StatusInactive || a.status == StatusRevoked {
		return ErrAccountNotActive
	}

	a.credentials = credentials
	if credentials.ExpiresAt != nil {
		a.expiresAt = credentials.ExpiresAt
	}
	a.status = StatusActive // Reset status if was expired
	a.updatedAt = time.Now().UTC()
	return nil
}

// UpdateProfile updates account profile information
func (a *Account) UpdateProfile(profile ProfileInfo) {
	a.username = profile.Username
	a.displayName = profile.DisplayName
	a.profileURL = profile.ProfileURL
	a.avatarURL = profile.AvatarURL
	a.metadata.FollowersCount = profile.FollowersCount
	a.metadata.FollowingCount = profile.FollowingCount
	a.metadata.PostsCount = profile.PostsCount
	a.metadata.Verified = profile.Verified
	a.metadata.Bio = profile.Bio

	now := time.Now().UTC()
	a.lastSyncAt = &now
	a.updatedAt = now
}

// UpdateRateLimits updates the rate limit information
func (a *Account) UpdateRateLimits(limits RateLimits) {
	a.rateLimits = limits
	a.updatedAt = time.Now().UTC()
}

// MarkExpired marks the account credentials as expired
func (a *Account) MarkExpired() error {
	if a.status == StatusExpired {
		return ErrAccountAlreadyExpired
	}

	a.status = StatusExpired
	a.updatedAt = time.Now().UTC()
	return nil
}

// MarkRevoked marks the account as revoked by the platform
func (a *Account) MarkRevoked() error {
	a.status = StatusRevoked
	a.updatedAt = time.Now().UTC()
	return nil
}

// MarkRateLimited marks the account as rate limited
func (a *Account) MarkRateLimited(resetsAt time.Time) error {
	a.status = StatusRateLimited
	a.rateLimits.ResetsAt = resetsAt
	a.updatedAt = time.Now().UTC()
	return nil
}

// ResetRateLimit resets the rate limit status
func (a *Account) ResetRateLimit() error {
	if a.status == StatusRateLimited {
		a.status = StatusActive
		a.rateLimits.CurrentHourPosts = 0
		a.rateLimits.CurrentDayPosts = 0
		a.updatedAt = time.Now().UTC()
	}
	return nil
}

// IncrementPostCount increments the post counters for rate limiting
func (a *Account) IncrementPostCount() {
	a.rateLimits.CurrentHourPosts++
	a.rateLimits.CurrentDayPosts++
	a.metadata.PostsCount++
	a.updatedAt = time.Now().UTC()
}

// SoftDelete soft deletes the account
func (a *Account) SoftDelete() error {
	if a.deletedAt != nil {
		return ErrAccountAlreadyDeleted
	}

	now := time.Now().UTC()
	a.deletedAt = &now
	a.status = StatusInactive
	a.updatedAt = now
	return nil
}

// Restore restores a soft-deleted account
func (a *Account) Restore() error {
	if a.deletedAt == nil {
		return ErrAccountNotDeleted
	}

	a.deletedAt = nil
	// Don't automatically reactivate - needs reconnection
	a.status = StatusReconnectRequired
	a.updatedAt = time.Now().UTC()
	return nil
}

// Business Rule Checks

// IsActive checks if the account is active and ready for posting
func (a *Account) IsActive() bool {
	return a.status == StatusActive &&
		a.deletedAt == nil &&
		!a.IsExpired()
}

// IsExpired checks if the account credentials are expired
func (a *Account) IsExpired() bool {
	if a.expiresAt == nil {
		return false
	}
	return a.expiresAt.Before(time.Now())
}

// NeedsReconnection checks if the account needs to be reconnected
func (a *Account) NeedsReconnection() bool {
	return a.status == StatusExpired ||
		a.status == StatusRevoked ||
		a.status == StatusReconnectRequired ||
		a.IsExpired()
}

// CanPost checks if the account can post right now
func (a *Account) CanPost() bool {
	if !a.IsActive() {
		return false
	}

	// Check rate limits
	if a.rateLimits.PostsPerHour > 0 && a.rateLimits.CurrentHourPosts >= a.rateLimits.PostsPerHour {
		return false
	}
	if a.rateLimits.PostsPerDay > 0 && a.rateLimits.CurrentDayPosts >= a.rateLimits.PostsPerDay {
		return false
	}

	return true
}

// GetRemainingPosts returns remaining posts for the current period
func (a *Account) GetRemainingPosts() (hourly int, daily int) {
	if a.rateLimits.PostsPerHour > 0 {
		hourly = a.rateLimits.PostsPerHour - a.rateLimits.CurrentHourPosts
		if hourly < 0 {
			hourly = 0
		}
	} else {
		hourly = -1 // Unlimited
	}

	if a.rateLimits.PostsPerDay > 0 {
		daily = a.rateLimits.PostsPerDay - a.rateLimits.CurrentDayPosts
		if daily < 0 {
			daily = 0
		}
	} else {
		daily = -1 // Unlimited
	}

	return hourly, daily
}

// ValidateForPlatform validates if the account is properly configured for its platform
func (a *Account) ValidateForPlatform() error {
	switch a.platform {
	case PlatformInstagram:
		if a.accountType != AccountTypeBusiness {
			return ErrInstagramRequiresBusiness
		}
	case PlatformFacebook:
		if a.accountType != AccountTypePage && a.accountType != AccountTypeBusiness {
			return ErrFacebookRequiresPage
		}
	case PlatformLinkedIn:
		if a.accountType == AccountTypeGroup {
			return ErrLinkedInGroupNotSupported
		}
	}
	return nil
}

// Helper Functions

func isValidPlatform(platform Platform) bool {
	switch platform {
	case PlatformTwitter, PlatformFacebook, PlatformLinkedIn,
		PlatformInstagram, PlatformTikTok, PlatformPinterest, PlatformYouTube:
		return true
	default:
		return false
	}
}

func isValidAccountType(accountType AccountType) bool {
	switch accountType {
	case AccountTypePersonal, AccountTypeBusiness, AccountTypePage,
		AccountTypeGroup, AccountTypeChannel:
		return true
	default:
		return false
	}
}

func getDefaultRateLimits(platform Platform) RateLimits {
	switch platform {
	case PlatformTwitter:
		return RateLimits{
			PostsPerHour:   50,
			PostsPerDay:    300,
			MediaPerPost:   4,
			CharacterLimit: 280,
			HashtagLimit:   30,
			MentionLimit:   50,
			CustomLimits:   make(map[string]int),
		}
	case PlatformFacebook:
		return RateLimits{
			PostsPerHour:   25,
			PostsPerDay:    200,
			MediaPerPost:   10,
			CharacterLimit: 63206,
			HashtagLimit:   30,
			MentionLimit:   50,
			CustomLimits:   make(map[string]int),
		}
	case PlatformLinkedIn:
		return RateLimits{
			PostsPerHour:   20,
			PostsPerDay:    100,
			MediaPerPost:   9,
			CharacterLimit: 3000,
			HashtagLimit:   30,
			MentionLimit:   30,
			CustomLimits:   make(map[string]int),
		}
	case PlatformInstagram:
		return RateLimits{
			PostsPerHour:   25,
			PostsPerDay:    100,
			MediaPerPost:   10,
			CharacterLimit: 2200,
			HashtagLimit:   30,
			MentionLimit:   20,
			CustomLimits:   make(map[string]int),
		}
	default:
		return RateLimits{
			PostsPerHour:   20,
			PostsPerDay:    100,
			MediaPerPost:   1,
			CharacterLimit: 5000,
			HashtagLimit:   10,
			MentionLimit:   10,
			CustomLimits:   make(map[string]int),
		}
	}
}

// ProfileInfo holds profile information from the platform
type ProfileInfo struct {
	Username       string
	DisplayName    string
	ProfileURL     string
	AvatarURL      string
	FollowersCount int
	FollowingCount int
	PostsCount     int
	Verified       bool
	Bio            string
}

// JSON serialization for storing complex fields

func (c Credentials) MarshalJSON() ([]byte, error) {
	// Never serialize tokens in plain text - this would be encrypted
	return json.Marshal(struct {
		Scope             []string   `json:"scope"`
		ExpiresAt         *time.Time `json:"expires_at,omitempty"`
		PlatformUserID    string     `json:"platform_user_id"`
		PlatformAccountID string     `json:"platform_account_id,omitempty"`
	}{
		Scope:             c.Scope,
		ExpiresAt:         c.ExpiresAt,
		PlatformUserID:    c.PlatformUserID,
		PlatformAccountID: c.PlatformAccountID,
	})
}
