// path: backend/internal/domain/user/service.go
// 🆕 NEW - Clean Architecture

package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// AuthenticateUser authenticates a user by email/username and password
func (s *Service) AuthenticateUser(ctx context.Context, identifier, password string) (*User, error) {
	// Find user by email or username
	user, err := s.repo.FindByEmailOrUsername(ctx, identifier)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user can access platform
	if !user.CanAccessPlatform() {
		if user.status == StatusSuspended {
			return nil, ErrUserSuspended
		}
		if !user.emailVerified {
			return nil, ErrEmailNotVerified
		}
		return nil, ErrUserInactive
	}

	// Verify password
	if !user.VerifyPassword(password) {
		return nil, ErrInvalidCredentials
	}

	// Record login
	user.RecordLogin("")
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

// ChangeUserRole changes a user's role
func (s *Service) ChangeUserRole(ctx context.Context, userID uuid.UUID, newRole Role, changedByID uuid.UUID) error {
	// Fetch the user performing the change
	changedBy, err := s.repo.FindByID(ctx, changedByID)
	if err != nil {
		return ErrUnauthorized
	}

	// Check if the user has permission to change roles
	if !changedBy.IsAdmin() {
		return ErrUnauthorized
	}

	// Fetch target user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Prevent demoting the last owner
	if user.IsOwner() && newRole != RoleOwner {
		count, err := s.repo.CountByRole(ctx, RoleOwner)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if count <= 1 {
			return ErrCannotDemoteLastOwner
		}
	}

	// Change role using domain logic
	if err := user.ChangeRole(newRole); err != nil {
		return err
	}

	// Update role in repository
	if err := s.repo.UpdateRole(ctx, user.ID(), user.Role()); err != nil {
		return fmt.Errorf("failed to change role: %w", err)
	}

	return nil
}

// DeleteUser performs a soft delete on a user
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Fetch user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Prevent deleting the last owner
	if user.IsOwner() {
		count, err := s.repo.CountByRole(ctx, RoleOwner)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if count <= 1 {
			return ErrCannotDeleteLastOwner
		}
	}

	// Soft delete user using domain logic
	if err := user.SoftDelete(); err != nil {
		return err
	}

	// Update user in repository
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// RestoreUser restores a soft-deleted user
func (s *Service) RestoreUser(ctx context.Context, userID uuid.UUID) error {
	// Fetch user (including deleted ones)
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Restore user using domain logic
	if err := user.Restore(); err != nil {
		return err
	}

	// Update user in repository
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	return nil
}

// CleanupInactiveUsers finds and handles inactive users
func (s *Service) CleanupInactiveUsers(ctx context.Context, inactiveDuration time.Duration) (int, error) {
	// Find inactive users
	users, err := s.repo.FindInactiveSince(ctx, time.Now().Add(-inactiveDuration), 0, 1000)
	if err != nil {
		return 0, fmt.Errorf("failed to find inactive users: %w", err)
	}

	count := 0
	for _, user := range users {
		// Suspend inactive users
		if err := user.Suspend(); err == nil {
			if err := s.repo.UpdateStatus(ctx, user.ID(), user.Status()); err == nil {
				count++
			}
		}
	}

	return count, nil
}

// CleanupUnverifiedUsers removes old unverified accounts
func (s *Service) CleanupUnverifiedUsers(ctx context.Context, maxAge time.Duration) (int, error) {
	// Find old unverified users
	users, err := s.repo.FindUnverifiedOlderThan(ctx, maxAge)
	if err != nil {
		return 0, fmt.Errorf("failed to find unverified users: %w", err)
	}

	count := 0
	for _, user := range users {
		// Hard delete old unverified accounts
		if err := s.repo.HardDelete(ctx, user.ID()); err == nil {
			count++
		}
	}

	return count, nil
}

// Token generation utilities (for email verification, password reset, etc.)

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

// TokenService handles token-based operations (email verification, password reset)
// This would typically be in the application layer but keeping here for completeness
type TokenService interface {
	// GenerateEmailVerificationToken generates a token for email verification
	GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error)

	// VerifyEmailToken verifies an email verification token
	VerifyEmailToken(ctx context.Context, token string) (uuid.UUID, error)

	// GeneratePasswordResetToken generates a token for password reset
	GeneratePasswordResetToken(ctx context.Context, email string) (string, error)

	// VerifyPasswordResetToken verifies a password reset token
	VerifyPasswordResetToken(ctx context.Context, token string) (uuid.UUID, error)

	// InvalidateToken invalidates a token
	InvalidateToken(ctx context.Context, token string) error
}

// Specifications for complex queries (Domain-Driven Design pattern)

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
