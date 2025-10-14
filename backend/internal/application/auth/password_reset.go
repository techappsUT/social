// backend/internal/application/auth/password_reset.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// ============================================================================
// FORGOT PASSWORD USE CASE
// ============================================================================

type ForgotPasswordUseCase struct {
	userRepo     user.Repository
	queries      *db.Queries
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
	queries *db.Queries,
	emailService common.EmailService,
	logger common.Logger,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:     userRepo,
		queries:      queries,
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

	// Store token in database (1 hour expiry)
	expiresAt := time.Now().Add(1 * time.Hour)
	err = uc.queries.SetResetToken(ctx, db.SetResetTokenParams{
		ID:                  usr.ID(),
		ResetToken:          sql.NullString{String: token, Valid: true},
		ResetTokenExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		uc.logger.Error("Failed to store reset token", "error", err)
		return nil, fmt.Errorf("failed to generate reset token")
	}

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
// RESET PASSWORD USE CASE
// ============================================================================

type ResetPasswordUseCase struct {
	userRepo     user.Repository
	queries      *db.Queries
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
	queries *db.Queries,
	userService *user.Service,
	emailService common.EmailService,
	logger common.Logger,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:     userRepo,
		queries:      queries,
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

	// 1. Get user by reset token
	dbUser, err := uc.queries.GetUserByResetToken(ctx, sql.NullString{
		String: input.Token,
		Valid:  true,
	})
	if err != nil {
		uc.logger.Warn("Invalid password reset token attempt", "token", input.Token[:10]+"...")
		return nil, fmt.Errorf("invalid or expired reset token")
	}

	// 2. Get domain user
	usr, err := uc.userRepo.FindByID(ctx, dbUser.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 3. Check if deleted
	if usr.IsDeleted() {
		return nil, fmt.Errorf("user account is deleted")
	}

	// 4. Reset password using domain method
	if err := usr.ResetPassword(input.NewPassword); err != nil {
		return nil, err
	}

	// 5. Update user in database
	if err := uc.userRepo.Update(ctx, usr); err != nil {
		uc.logger.Error("Failed to update password", "userId", usr.ID(), "error", err)
		return nil, fmt.Errorf("failed to reset password")
	}

	// 6. Clear reset token
	if err := uc.queries.ClearResetToken(ctx, dbUser.ID); err != nil {
		uc.logger.Warn("Failed to clear reset token", "error", err)
	}

	// 7. Revoke all refresh tokens for security
	if err := uc.queries.RevokeAllUserTokens(ctx, usr.ID()); err != nil {
		uc.logger.Warn("Failed to revoke refresh tokens", "error", err)
	}

	// 8. Send confirmation email (optional)
	go func() {
		// Add SendPasswordChangedEmail to your email service interface
		uc.logger.Info("Password reset successful", "userId", usr.ID())
	}()

	uc.logger.Info("Password reset successfully", "userId", usr.ID())

	return &ResetPasswordOutput{
		Success: true,
		Message: "Password reset successfully. Please login with your new password.",
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
