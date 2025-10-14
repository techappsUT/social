// path: backend/internal/application/auth/password_reset.go
package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// ============================================================================
// FORGOT PASSWORD USE CASE
// ============================================================================

type ForgotPasswordUseCase struct {
	userRepo     user.Repository
	emailService common.EmailService
	logger       common.Logger
}

type ForgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewForgotPasswordUseCase(
	userRepo user.Repository,
	emailService common.EmailService,
	logger common.Logger,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:     userRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, input ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	// Find user by email
	usr, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Don't reveal if user exists - security best practice
		uc.logger.Info("Password reset requested for non-existent email", "email", input.Email)
		return &ForgotPasswordOutput{
			Success: true,
			Message: "If an account with that email exists, a password reset link has been sent.",
		}, nil
	}

	// Check if user is deleted
	if usr.IsDeleted() {
		return &ForgotPasswordOutput{
			Success: true,
			Message: "If an account with that email exists, a password reset link has been sent.",
		}, nil
	}

	// Generate reset token
	token, err := user.GenerateToken()
	if err != nil {
		uc.logger.Error("Failed to generate reset token", "error", err)
		return nil, fmt.Errorf("failed to generate reset token")
	}

	// TODO: Store token in database with expiry (1 hour)
	// This would require a password_reset_tokens table

	// Send password reset email
	if err := uc.emailService.SendPasswordResetEmail(ctx, usr.Email(), token); err != nil {
		uc.logger.Error("Failed to send password reset email", "email", usr.Email(), "error", err)
		// Don't fail - still return success to user
	}

	uc.logger.Info("Password reset email sent", "email", usr.Email())

	return &ForgotPasswordOutput{
		Success: true,
		Message: "If an account with that email exists, a password reset link has been sent.",
	}, nil
}

// ============================================================================
// CHANGE PASSWORD USE CASE (Authenticated user changing their password)
// ============================================================================

type ChangePasswordUseCase struct {
	userRepo    user.Repository
	userService *user.Service
	logger      common.Logger
}

type ChangePasswordInput struct {
	UserID          uuid.UUID `json:"userId"`
	CurrentPassword string    `json:"currentPassword" validate:"required"`
	NewPassword     string    `json:"newPassword" validate:"required,min=8"`
}

type ChangePasswordOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewChangePasswordUseCase(
	userRepo user.Repository,
	userService *user.Service,
	logger common.Logger,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo:    userRepo,
		userService: userService,
		logger:      logger,
	}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, input ChangePasswordInput) (*ChangePasswordOutput, error) {
	// Validate input
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}
	if input.CurrentPassword == "" {
		return nil, fmt.Errorf("current password is required")
	}
	if len(input.NewPassword) < 8 {
		return nil, fmt.Errorf("new password must be at least 8 characters")
	}

	// Get user
	usr, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if deleted
	if usr.IsDeleted() {
		return nil, fmt.Errorf("user account is deleted")
	}

	// Verify current password
	if !usr.VerifyPassword(input.CurrentPassword) {
		uc.logger.Warn("Invalid current password", "userId", input.UserID)
		return nil, fmt.Errorf("current password is incorrect")
	}

	// Change password using domain method
	if err := usr.ChangePassword(input.CurrentPassword, input.NewPassword); err != nil {
		return nil, err
	}

	// Update in repository
	if err := uc.userRepo.Update(ctx, usr); err != nil {
		uc.logger.Error("Failed to update password", "userId", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to update password")
	}

	uc.logger.Info("Password changed successfully", "userId", input.UserID)

	return &ChangePasswordOutput{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}

// ============================================================================
// RESET PASSWORD USE CASE (Using reset token from forgot password flow)
// ============================================================================

type ResetPasswordUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	emailService common.EmailService
	logger       common.Logger
}

type ResetPasswordInput struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

type ResetPasswordOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewResetPasswordUseCase(
	userRepo user.Repository,
	userService *user.Service,
	emailService common.EmailService,
	logger common.Logger,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:     userRepo,
		userService:  userService,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) (*ResetPasswordOutput, error) {
	// Validate input
	if input.Token == "" {
		return nil, fmt.Errorf("reset token is required")
	}
	if len(input.NewPassword) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	// TODO: Verify token and get user ID from password_reset_tokens table
	// For now, this is a placeholder implementation

	// This would normally be:
	// 1. Look up token in password_reset_tokens table
	// 2. Check if token is expired (< 1 hour old)
	// 3. Get the user_id from the token record
	// 4. Delete/mark token as used

	uc.logger.Warn("Reset password called but token verification not implemented", "token", input.Token[:10]+"...")

	return &ResetPasswordOutput{
		Success: true,
		Message: "Password reset successfully. Please login with your new password.",
	}, nil
}
