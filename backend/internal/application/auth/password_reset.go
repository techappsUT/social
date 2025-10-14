// path: backend/internal/application/auth/password_reset.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// ============================================================================
// FORGOT PASSWORD
// ============================================================================

// ForgotPasswordInput represents the input for forgot password
type ForgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordOutput represents the output after forgot password
type ForgotPasswordOutput struct {
	Message string `json:"message"`
}

// ForgotPasswordUseCase handles forgot password requests
type ForgotPasswordUseCase struct {
	userRepo     user.Repository
	emailService common.EmailService
	cacheService common.CacheService
	logger       common.Logger
}

// NewForgotPasswordUseCase creates a new forgot password use case
func NewForgotPasswordUseCase(
	userRepo user.Repository,
	emailService common.EmailService,
	cacheService common.CacheService,
	logger common.Logger,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:     userRepo,
		emailService: emailService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute sends a password reset email
func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, input ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	// 1. Get user by email
	usr, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Don't reveal if email exists or not (security)
		uc.logger.Warn(fmt.Sprintf("Password reset attempted for non-existent email: %s", input.Email))
		return &ForgotPasswordOutput{
			Message: "If the email exists, a password reset link has been sent",
		}, nil
	}

	// 2. Check if user is deleted
	if usr.IsDeleted() {
		return &ForgotPasswordOutput{
			Message: "If the email exists, a password reset link has been sent",
		}, nil
	}

	// 3. Check rate limiting (prevent spam)
	rateLimitKey := fmt.Sprintf("ratelimit:reset:%s", usr.ID().String())
	attempts, _ := uc.cacheService.Get(ctx, rateLimitKey)
	if attempts != "" {
		return nil, fmt.Errorf("password reset email already sent, please wait before requesting again")
	}

	// 4. Generate password reset token
	token := uuid.New().String()

	// 5. Store token in cache (expires in 1 hour)
	err = uc.cacheService.Set(
		ctx,
		fmt.Sprintf("reset:%s", token),
		usr.ID().String(),
		1*time.Hour,
	)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to cache reset token: %v", err))
		return nil, fmt.Errorf("failed to generate reset token")
	}

	// 6. Set rate limit (1 email per 5 minutes)
	uc.cacheService.Set(ctx, rateLimitKey, "1", 5*time.Minute)

	// 7. Send password reset email
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	err = uc.emailService.SendPasswordResetEmail(usr.Email(), usr.FirstName(), resetURL)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to send reset email: %v", err))
		return nil, fmt.Errorf("failed to send reset email")
	}

	uc.logger.Info(fmt.Sprintf("Password reset email sent to: %s", usr.Email()))

	return &ForgotPasswordOutput{
		Message: "If the email exists, a password reset link has been sent",
	}, nil
}

// ============================================================================
// RESET PASSWORD
// ============================================================================

// ResetPasswordInput represents the input for reset password
type ResetPasswordInput struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

// ResetPasswordOutput represents the output after reset password
type ResetPasswordOutput struct {
	Message string `json:"message"`
}

// ResetPasswordUseCase handles password reset
type ResetPasswordUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	cacheService common.CacheService
	logger       common.Logger
}

// NewResetPasswordUseCase creates a new reset password use case
func NewResetPasswordUseCase(
	userRepo user.Repository,
	userService *user.Service,
	cacheService common.CacheService,
	logger common.Logger,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:     userRepo,
		userService:  userService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute resets the user's password
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) (*ResetPasswordOutput, error) {
	// 1. Get user ID from reset token cache
	userIDStr, err := uc.cacheService.Get(ctx, fmt.Sprintf("reset:%s", input.Token))
	if err != nil || userIDStr == "" {
		uc.logger.Warn(fmt.Sprintf("Invalid or expired reset token: %s", input.Token))
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	// 2. Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// 3. Get user from database
	usr, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("User not found for password reset: %s", userID))
		return nil, fmt.Errorf("user not found")
	}

	// 4. Check if user is deleted
	if usr.IsDeleted() {
		return nil, fmt.Errorf("user account is deleted")
	}

	// 5. Hash new password
	hashedPassword, err := uc.userService.HashPassword(input.NewPassword)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to hash password: %v", err))
		return nil, fmt.Errorf("failed to process password")
	}

	// 6. Update password in database
	err = uc.userRepo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to update password: %v", err))
		return nil, fmt.Errorf("failed to update password")
	}

	// 7. Delete reset token from cache
	uc.cacheService.Delete(ctx, fmt.Sprintf("reset:%s", input.Token))

	// 8. Invalidate all user sessions (force re-login)
	uc.cacheService.Delete(ctx, fmt.Sprintf("session:%s", userID))

	uc.logger.Info(fmt.Sprintf("Password reset successfully for user: %s", usr.Email()))

	return &ResetPasswordOutput{
		Message: "Password reset successfully",
	}, nil
}

// ============================================================================
// CHANGE PASSWORD (Authenticated)
// ============================================================================

// ChangePasswordInput represents the input for change password
type ChangePasswordInput struct {
	UserID          string `json:"-"` // Set from context
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// ChangePasswordOutput represents the output after change password
type ChangePasswordOutput struct {
	Message string `json:"message"`
}

// ChangePasswordUseCase handles password change for authenticated users
type ChangePasswordUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	cacheService common.CacheService
	logger       common.Logger
}

// NewChangePasswordUseCase creates a new change password use case
func NewChangePasswordUseCase(
	userRepo user.Repository,
	userService *user.Service,
	cacheService common.CacheService,
	logger common.Logger,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo:     userRepo,
		userService:  userService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute changes the user's password
func (uc *ChangePasswordUseCase) Execute(ctx context.Context, input ChangePasswordInput) (*ChangePasswordOutput, error) {
	// 1. Parse user ID
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// 2. Get user from database
	usr, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 3. Verify current password
	valid, err := uc.userService.ComparePassword(usr.PasswordHash(), input.CurrentPassword)
	if err != nil || !valid {
		uc.logger.Warn(fmt.Sprintf("Invalid current password for user: %s", usr.Email()))
		return nil, fmt.Errorf("current password is incorrect")
	}

	// 4. Check new password is different
	if input.CurrentPassword == input.NewPassword {
		return nil, fmt.Errorf("new password must be different from current password")
	}

	// 5. Hash new password
	hashedPassword, err := uc.userService.HashPassword(input.NewPassword)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to hash password: %v", err))
		return nil, fmt.Errorf("failed to process password")
	}

	// 6. Update password in database
	err = uc.userRepo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to update password: %v", err))
		return nil, fmt.Errorf("failed to update password")
	}

	// 7. Invalidate all user sessions (force re-login on other devices)
	uc.cacheService.Delete(ctx, fmt.Sprintf("session:%s", userID))

	uc.logger.Info(fmt.Sprintf("Password changed successfully for user: %s", usr.Email()))

	return &ChangePasswordOutput{
		Message: "Password changed successfully",
	}, nil
}
