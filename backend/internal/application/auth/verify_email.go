// path: backend/internal/application/auth/verify_email.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// ============================================================================
// VERIFY EMAIL USE CASE
// ============================================================================

type VerifyEmailUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	emailService common.EmailService
	logger       common.Logger
}

type VerifyEmailInput struct {
	Token string `json:"token" validate:"required"`
}

type VerifyEmailOutput struct {
	Success       bool       `json:"success"`
	Message       string     `json:"message"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"emailVerified"`
	VerifiedAt    *time.Time `json:"verifiedAt,omitempty"`
}

func NewVerifyEmailUseCase(
	userRepo user.Repository,
	userService *user.Service,
	emailService common.EmailService,
	logger common.Logger,
) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		userRepo:     userRepo,
		userService:  userService,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, input VerifyEmailInput) (*VerifyEmailOutput, error) {
	// Validate input
	if input.Token == "" {
		return nil, fmt.Errorf("verification token is required")
	}

	// TODO: In a real implementation, you would:
	// 1. Look up the token in an email_verification_tokens table
	// 2. Check if the token is expired
	// 3. Get the user_id associated with the token
	// 4. Mark the token as used

	// For now, this is a placeholder that searches all users
	// In production, you need a proper token-to-user mapping table

	uc.logger.Info("Email verification requested", "token", input.Token[:10]+"...")

	// TODO: Replace with actual token lookup
	// For now, return error since we don't have token table implemented
	return nil, fmt.Errorf("email verification token system not yet implemented - requires email_verification_tokens table")
}

// ============================================================================
// RESEND VERIFICATION EMAIL USE CASE
// ============================================================================

type ResendVerificationUseCase struct {
	userRepo     user.Repository
	emailService common.EmailService
	logger       common.Logger
}

type ResendVerificationInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendVerificationOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewResendVerificationUseCase(
	userRepo user.Repository,
	emailService common.EmailService,
	logger common.Logger,
) *ResendVerificationUseCase {
	return &ResendVerificationUseCase{
		userRepo:     userRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *ResendVerificationUseCase) Execute(ctx context.Context, input ResendVerificationInput) (*ResendVerificationOutput, error) {
	// Find user by email
	usr, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Don't reveal if user exists - security best practice
		uc.logger.Info("Verification email resend requested for non-existent email", "email", input.Email)
		return &ResendVerificationOutput{
			Success: true,
			Message: "If an account with that email exists and is unverified, a verification email has been sent.",
		}, nil
	}

	// Check if already verified
	if usr.IsEmailVerified() {
		return &ResendVerificationOutput{
			Success: true,
			Message: "This email is already verified. You can login now.",
		}, nil
	}

	// Generate new verification token
	token, err := user.GenerateToken()
	if err != nil {
		uc.logger.Error("Failed to generate verification token", "error", err)
		return nil, fmt.Errorf("failed to generate verification token")
	}

	// TODO: Store token in email_verification_tokens table with expiry (24 hours)

	// Send verification email
	if err := uc.emailService.SendVerificationEmail(ctx, usr.Email(), token); err != nil {
		uc.logger.Error("Failed to send verification email", "email", usr.Email(), "error", err)
		// Don't fail - still return success to user
	}

	uc.logger.Info("Verification email resent", "email", usr.Email())

	return &ResendVerificationOutput{
		Success: true,
		Message: "If an account with that email exists and is unverified, a verification email has been sent.",
	}, nil
}
