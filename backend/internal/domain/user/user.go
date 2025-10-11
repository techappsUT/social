// path: backend/internal/domain/user/user.go
// 🆕 NEW - Clean Architecture

package user

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents the core user entity in the domain layer.
// This is a pure domain object with no external dependencies.
type User struct {
	id            uuid.UUID
	email         string
	username      string
	passwordHash  string
	firstName     string
	lastName      string
	avatarURL     string
	role          Role
	status        Status
	emailVerified bool
	lastLoginAt   *time.Time
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// Role represents the user's role in the system
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
	RoleOwner Role = "owner"
)

// Status represents the user's account status
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
	StatusPending   Status = "pending"
)

// NewUser creates a new user entity with validation
func NewUser(email, username, password, firstName, lastName string) (*User, error) {
	// Validate email
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	// Validate username
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	// Validate password
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	// Validate names
	if strings.TrimSpace(firstName) == "" {
		return nil, ErrInvalidFirstName
	}
	if strings.TrimSpace(lastName) == "" {
		return nil, ErrInvalidLastName
	}

	now := time.Now().UTC()

	return &User{
		id:            uuid.New(),
		email:         strings.ToLower(strings.TrimSpace(email)),
		username:      strings.ToLower(strings.TrimSpace(username)),
		passwordHash:  hashedPassword,
		firstName:     strings.TrimSpace(firstName),
		lastName:      strings.TrimSpace(lastName),
		role:          RoleUser,
		status:        StatusPending,
		emailVerified: false,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// Reconstruct recreates a user entity from persistence layer
// This is used when loading from database
func Reconstruct(
	id uuid.UUID,
	email string,
	username string,
	passwordHash string,
	firstName string,
	lastName string,
	avatarURL string,
	role Role,
	status Status,
	emailVerified bool,
	lastLoginAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *User {
	return &User{
		id:            id,
		email:         email,
		username:      username,
		passwordHash:  passwordHash,
		firstName:     firstName,
		lastName:      lastName,
		avatarURL:     avatarURL,
		role:          role,
		status:        status,
		emailVerified: emailVerified,
		lastLoginAt:   lastLoginAt,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		deletedAt:     deletedAt,
	}
}

// Getters - Encapsulation of internal state

func (u *User) ID() uuid.UUID           { return u.id }
func (u *User) Email() string           { return u.email }
func (u *User) Username() string        { return u.username }
func (u *User) FirstName() string       { return u.firstName }
func (u *User) LastName() string        { return u.lastName }
func (u *User) FullName() string        { return u.firstName + " " + u.lastName }
func (u *User) AvatarURL() string       { return u.avatarURL }
func (u *User) Role() Role              { return u.role }
func (u *User) Status() Status          { return u.status }
func (u *User) IsEmailVerified() bool   { return u.emailVerified }
func (u *User) LastLoginAt() *time.Time { return u.lastLoginAt }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) DeletedAt() *time.Time   { return u.deletedAt }
func (u *User) PasswordHash() string    { return u.passwordHash }

// Business Logic Methods

// VerifyPassword checks if the provided password matches
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password))
	return err == nil
}

// ChangePassword updates the user's password
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	// Verify old password
	if !u.VerifyPassword(oldPassword) {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	u.passwordHash = hashedPassword
	u.updatedAt = time.Now().UTC()
	return nil
}

// ResetPassword sets a new password without checking the old one
// Used for password reset flows
func (u *User) ResetPassword(newPassword string) error {
	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	u.passwordHash = hashedPassword
	u.updatedAt = time.Now().UTC()
	return nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(firstName, lastName, avatarURL string) error {
	if strings.TrimSpace(firstName) == "" {
		return ErrInvalidFirstName
	}
	if strings.TrimSpace(lastName) == "" {
		return ErrInvalidLastName
	}

	u.firstName = strings.TrimSpace(firstName)
	u.lastName = strings.TrimSpace(lastName)
	u.avatarURL = strings.TrimSpace(avatarURL)
	u.updatedAt = time.Now().UTC()
	return nil
}

// ChangeEmail updates the user's email address
func (u *User) ChangeEmail(newEmail string) error {
	if err := validateEmail(newEmail); err != nil {
		return err
	}

	u.email = strings.ToLower(strings.TrimSpace(newEmail))
	u.emailVerified = false // Require re-verification
	u.updatedAt = time.Now().UTC()
	return nil
}

// VerifyEmail marks the email as verified
func (u *User) VerifyEmail() error {
	if u.emailVerified {
		return ErrEmailAlreadyVerified
	}

	u.emailVerified = true
	u.status = StatusActive
	u.updatedAt = time.Now().UTC()
	return nil
}

// RecordLogin updates the last login timestamp
func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.lastLoginAt = &now
	u.updatedAt = now
}

// Suspend suspends the user account
func (u *User) Suspend() error {
	if u.status == StatusSuspended {
		return ErrUserAlreadySuspended
	}

	u.status = StatusSuspended
	u.updatedAt = time.Now().UTC()
	return nil
}

// Activate activates a suspended or inactive user account
func (u *User) Activate() error {
	if u.status == StatusActive {
		return ErrUserAlreadyActive
	}

	u.status = StatusActive
	u.updatedAt = time.Now().UTC()
	return nil
}

// SoftDelete marks the user as deleted
func (u *User) SoftDelete() error {
	if u.deletedAt != nil {
		return ErrUserAlreadyDeleted
	}

	now := time.Now().UTC()
	u.deletedAt = &now
	u.status = StatusInactive
	u.updatedAt = now
	return nil
}

// Restore restores a soft-deleted user
func (u *User) Restore() error {
	if u.deletedAt == nil {
		return ErrUserNotDeleted
	}

	u.deletedAt = nil
	u.status = StatusActive
	u.updatedAt = time.Now().UTC()
	return nil
}

// ChangeRole updates the user's role
func (u *User) ChangeRole(newRole Role) error {
	if !isValidRole(newRole) {
		return ErrInvalidRole
	}

	u.role = newRole
	u.updatedAt = time.Now().UTC()
	return nil
}

// Domain Rules and Permissions

// CanManageTeam checks if the user can manage team settings
func (u *User) CanManageTeam() bool {
	return u.status == StatusActive &&
		u.emailVerified &&
		(u.role == RoleAdmin || u.role == RoleOwner)
}

// CanAccessPlatform checks if the user can access the platform
func (u *User) CanAccessPlatform() bool {
	return u.status == StatusActive &&
		u.emailVerified &&
		u.deletedAt == nil
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.status == StatusActive && u.deletedAt == nil
}

// IsAdmin checks if the user has admin privileges
func (u *User) IsAdmin() bool {
	return u.role == RoleAdmin || u.role == RoleOwner
}

// IsOwner checks if the user is an owner
func (u *User) IsOwner() bool {
	return u.role == RoleOwner
}

// Helper Functions

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrEmailRequired
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmailFormat
	}

	return nil
}

func validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return ErrUsernameRequired
	}

	if len(username) < 3 {
		return ErrUsernameTooShort
	}

	if len(username) > 30 {
		return ErrUsernameTooLong
	}

	// Check for valid characters (alphanumeric, underscore, dash)
	for _, ch := range username {
		if !isAlphanumeric(ch) && ch != '_' && ch != '-' {
			return ErrInvalidUsernameFormat
		}
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	// Could add more complex validation here (uppercase, lowercase, special chars, etc.)

	return nil
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	return string(hashedBytes), nil
}

func isAlphanumeric(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9')
}

func isValidRole(role Role) bool {
	switch role {
	case RoleUser, RoleAdmin, RoleOwner:
		return true
	default:
		return false
	}
}
