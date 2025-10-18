// path: backend/internal/domain/user/service.go
// ✅ COMPLETE FIXED VERSION
package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service provides domain-level business logic for users
// This is different from application services - it contains pure domain logic
type Service struct {
	repo Repository
}

// NewService creates a new user domain service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateUser creates a new user with all necessary validations
func (s *Service) CreateUser(ctx context.Context, email, username, password, firstName, lastName string) (*User, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Check if username already exists
	exists, err = s.repo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// Create new user entity
	user, err := NewUser(email, username, password, firstName, lastName)
	if err != nil {
		return nil, err
	}

	// Persist the user
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// ✅ FIX #5: Complete the AuthenticateUser method
// AuthenticateUser authenticates a user by email/username and password
func (s *Service) AuthenticateUser(ctx context.Context, identifier, password string) (*User, error) {
	// Find user by email or username
	user, err := s.repo.FindByEmailOrUsername(ctx, identifier)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user can access platform
	if !user.CanAccessPlatform() {
		// ✅ FIXED: Complete the if statement logic
		if user.Status() == StatusSuspended {
			return nil, ErrUserSuspended
		}
		if user.Status() == StatusInactive {
			return nil, ErrUserInactive
		}
		if user.IsDeleted() {
			return nil, ErrUserInactive
		}
		// Note: We removed emailVerified check from CanAccessPlatform
		// but you could add specific error here if needed:
		// if !user.IsEmailVerified() {
		//     return nil, ErrEmailNotVerified
		// }
		return nil, fmt.Errorf("cannot access platform")
	}

	// Verify password
	if !user.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	// Record login
	if err := user.RecordLogin(""); err != nil {
		// Log error but don't fail authentication
		// This is a non-critical operation
	}

	// Update last login in database
	if err := s.repo.UpdateLastLogin(ctx, user.ID(), *user.LastLoginAt()); err != nil {
		// Log error but don't fail authentication
		// This is a non-critical operation
	}

	return user, nil
}

// ChangeUserPassword changes a user's password
func (s *Service) ChangeUserPassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Change password using domain logic
	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		return err
	}

	// Update password in repository
	if err := s.repo.UpdatePassword(ctx, user.ID(), user.PasswordHash()); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ResetUserPassword resets a user's password (used with reset token)
func (s *Service) ResetUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Reset password using domain logic
	if err := user.ResetPassword(newPassword); err != nil {
		return err
	}

	// Update password in repository
	if err := s.repo.UpdatePassword(ctx, user.ID(), user.PasswordHash()); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// VerifyUserEmail marks a user's email as verified
func (s *Service) VerifyUserEmail(ctx context.Context, userID uuid.UUID) error {
	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify email using domain logic
	if err := user.VerifyEmail(); err != nil {
		return err
	}

	// Update verification status and user status
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// ChangeUserEmail changes a user's email address
func (s *Service) ChangeUserEmail(ctx context.Context, userID uuid.UUID, newEmail string) error {
	// Check if new email already exists
	exists, err := s.repo.ExistsByEmail(ctx, newEmail)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return ErrEmailAlreadyExists
	}

	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Change email using domain logic
	if err := user.ChangeEmail(newEmail); err != nil {
		return err
	}

	// Update user in repository
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to change email: %w", err)
	}

	return nil
}

// SuspendUser suspends a user account
func (s *Service) SuspendUser(ctx context.Context, userID uuid.UUID, reason string) error {
	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Suspend user using domain logic
	if err := user.Suspend(); err != nil {
		return err
	}

	// Update status in repository
	if err := s.repo.UpdateStatus(ctx, user.ID(), user.Status()); err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	// TODO: Could emit a domain event here for audit logging
	// e.g., events.Publish(UserSuspendedEvent{UserID: userID, Reason: reason})

	return nil
}

// ActivateUser activates a suspended or inactive user
func (s *Service) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Activate user using domain logic
	if err := user.Activate(); err != nil {
		return err
	}

	// Update status in repository
	if err := s.repo.UpdateStatus(ctx, user.ID(), user.Status()); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GenerateToken generates a secure random token
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateVerificationCode generates a 6-digit verification code
func GenerateVerificationCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	code := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", code%1000000), nil
}

// HashPassword hashes a plain text password
func (s *Service) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a plain text password with a hashed password
func (s *Service) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ============================================================================
// SPECIFICATIONS (Domain-Driven Design pattern)
// ============================================================================

// ActiveUserSpecification checks if a user is active
type ActiveUserSpecification struct{}

func (s ActiveUserSpecification) IsSatisfiedBy(user *User) bool {
	return user.IsActive()
}

// VerifiedUserSpecification checks if a user is verified
type VerifiedUserSpecification struct{}

func (s VerifiedUserSpecification) IsSatisfiedBy(user *User) bool {
	return user.IsEmailVerified()
}

// AdminUserSpecification checks if a user is an admin
type AdminUserSpecification struct{}

func (s AdminUserSpecification) IsSatisfiedBy(user *User) bool {
	return user.IsAdmin()
}

// Add this method to backend/internal/domain/user/service.go

// CreateUserWithToken creates a new user with a verification token
func (s *Service) CreateUserWithToken(
	ctx context.Context,
	email, username, password, firstName, lastName string,
	verificationToken string,
	tokenExpiry time.Time,
) (*User, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Check if username already exists
	exists, err = s.repo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// Create new user entity
	user, err := NewUser(email, username, password, firstName, lastName)
	if err != nil {
		return nil, err
	}

	// ✅ Set verification token BEFORE saving
	user.SetVerificationToken(verificationToken, tokenExpiry)

	// Persist the user (now includes verification token)
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
