// backend/internal/application/auth/verify_email.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// ============================================================================
// VERIFY EMAIL USE CASE
// ============================================================================

type VerifyEmailUseCase struct {
	userRepo user.Repository
	queries  *db.Queries
	logger   common.Logger
}

type VerifyEmailInput struct {
	Token string `json:"token" validate:"required"`
}

type VerifyEmailOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewVerifyEmailUseCase(
	userRepo user.Repository,
	queries *db.Queries,
	logger common.Logger,
) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		userRepo: userRepo,
		queries:  queries,
		logger:   logger,
	}
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, input VerifyEmailInput) (*VerifyEmailOutput, error) {
	// Development mode - find most recent unverified user
	if os.Getenv("DEVELOPMENT_MODE") == "true" {
		devCode := os.Getenv("DEV_EMAIL_VERIFICATION_CODE")
		if devCode != "" && input.Token == devCode {
			uc.logger.Info("DEV MODE: Email verification with dev code", "token", devCode)

			// Find the most recent unverified user
			// You'll need to add this query to your SQLC queries
			dbUser, err := uc.queries.GetMostRecentUnverifiedUser(ctx)
			if err != nil {
				return nil, fmt.Errorf("no unverified users found")
			}

			// Clear verification token and mark as verified
			if err := uc.queries.ClearVerificationToken(ctx, dbUser.ID); err != nil {
				return nil, fmt.Errorf("failed to verify email")
			}

			return &VerifyEmailOutput{
				Success: true,
				Message: "Email verified successfully (dev mode)",
			}, nil
		}
	}

	// 1. Get user by verification token
	dbUser, err := uc.queries.GetUserByVerificationToken(ctx, sql.NullString{
		String: input.Token,
		Valid:  true,
	})
	if err != nil {
		uc.logger.Warn("Invalid verification token attempt", "token", input.Token[:10]+"...")
		return nil, fmt.Errorf("invalid or expired verification token")
	}

	// 2. Get domain user
	usr, err := uc.userRepo.FindByID(ctx, dbUser.ID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 3. Check if already verified
	if usr.IsEmailVerified() {
		return &VerifyEmailOutput{
			Success: true,
			Message: "Email already verified",
		}, nil
	}

	// 4. Mark email as verified and clear token
	if err := uc.queries.ClearVerificationToken(ctx, dbUser.ID); err != nil {
		uc.logger.Error("Failed to clear verification token", "userId", dbUser.ID, "error", err)
		return nil, fmt.Errorf("failed to verify email")
	}

	uc.logger.Info("Email verified successfully", "userId", dbUser.ID)

	return &VerifyEmailOutput{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}

// ============================================================================
// RESEND VERIFICATION EMAIL USE CASE
// ============================================================================

type ResendVerificationUseCase struct {
	userRepo     user.Repository
	queries      *db.Queries
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
	queries *db.Queries,
	emailService common.EmailService,
	logger common.Logger,
) *ResendVerificationUseCase {
	return &ResendVerificationUseCase{
		userRepo:     userRepo,
		queries:      queries,
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

	// Store token in database (24 hours expiry)
	expiresAt := time.Now().Add(24 * time.Hour)
	err = uc.queries.SetVerificationToken(ctx, db.SetVerificationTokenParams{
		ID:                         usr.ID(),
		VerificationToken:          sql.NullString{String: token, Valid: true},
		VerificationTokenExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		uc.logger.Error("Failed to store verification token", "error", err)
		return nil, fmt.Errorf("failed to generate verification token")
	}

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
