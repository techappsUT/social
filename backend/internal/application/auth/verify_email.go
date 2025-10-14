// path: backend/internal/application/auth/verify_email.go
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// VerifyEmailInput represents the input for email verification
type VerifyEmailInput struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmailOutput represents the output after email verification
type VerifyEmailOutput struct {
	Message       string    `json:"message"`
	EmailVerified bool      `json:"emailVerified"`
	VerifiedAt    time.Time `json:"verifiedAt"`
}

// VerifyEmailUseCase handles email verification
type VerifyEmailUseCase struct {
	userRepo     user.Repository
	tokenService common.TokenService
	cacheService common.CacheService
	logger       common.Logger
}

// NewVerifyEmailUseCase creates a new verify email use case
func NewVerifyEmailUseCase(
	userRepo user.Repository,
	tokenService common.TokenService,
	cacheService common.CacheService,
	logger common.Logger,
) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		userRepo:     userRepo,
		tokenService: tokenService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute verifies a user's email address
func (uc *VerifyEmailUseCase) Execute(ctx context.Context, input VerifyEmailInput) (*VerifyEmailOutput, error) {
	// 1. Get user ID from verification token cache
	userIDStr, err := uc.cacheService.Get(ctx, fmt.Sprintf("verify:%s", input.Token))
	if err != nil || userIDStr == "" {
		uc.logger.Warn(fmt.Sprintf("Invalid or expired verification token: %s", input.Token))
		return nil, fmt.Errorf("invalid or expired verification token")
	}

	// 2. Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// 3. Get user from database
	usr, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("User not found for verification: %s", userID))
		return nil, fmt.Errorf("user not found")
	}

	// 4. Check if already verified
	if usr.EmailVerified() {
		return &VerifyEmailOutput{
			Message:       "Email already verified",
			EmailVerified: true,
			VerifiedAt:    usr.EmailVerifiedAt(),
		}, nil
	}

	// 5. Mark email as verified (you'll need to add this method to your User domain)
	// For now, we'll update directly through repository
	err = uc.userRepo.MarkEmailVerified(ctx, userID)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to mark email as verified: %v", err))
		return nil, fmt.Errorf("failed to verify email")
	}

	// 6. Delete verification token from cache
	uc.cacheService.Delete(ctx, fmt.Sprintf("verify:%s", input.Token))

	uc.logger.Info(fmt.Sprintf("Email verified successfully for user: %s", usr.Email()))

	return &VerifyEmailOutput{
		Message:       "Email verified successfully",
		EmailVerified: true,
		VerifiedAt:    time.Now(),
	}, nil
}

// ResendVerificationInput represents the input for resending verification
type ResendVerificationInput struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendVerificationOutput represents the output after resending verification
type ResendVerificationOutput struct {
	Message string `json:"message"`
}

// ResendVerificationUseCase handles resending email verification
type ResendVerificationUseCase struct {
	userRepo     user.Repository
	emailService common.EmailService
	cacheService common.CacheService
	logger       common.Logger
}

// NewResendVerificationUseCase creates a new resend verification use case
func NewResendVerificationUseCase(
	userRepo user.Repository,
	emailService common.EmailService,
	cacheService common.CacheService,
	logger common.Logger,
) *ResendVerificationUseCase {
	return &ResendVerificationUseCase{
		userRepo:     userRepo,
		emailService: emailService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Execute resends the email verification link
func (uc *ResendVerificationUseCase) Execute(ctx context.Context, input ResendVerificationInput) (*ResendVerificationOutput, error) {
	// 1. Get user by email
	usr, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Don't reveal if email exists or not (security)
		uc.logger.Warn(fmt.Sprintf("Verification resend attempted for non-existent email: %s", input.Email))
		return &ResendVerificationOutput{
			Message: "If the email exists, a verification link has been sent",
		}, nil
	}

	// 2. Check if already verified
	if usr.EmailVerified() {
		return nil, fmt.Errorf("email already verified")
	}

	// 3. Check rate limiting (prevent spam)
	rateLimitKey := fmt.Sprintf("ratelimit:verify:%s", usr.ID().String())
	attempts, _ := uc.cacheService.Get(ctx, rateLimitKey)
	if attempts != "" {
		return nil, fmt.Errorf("verification email already sent, please wait before requesting again")
	}

	// 4. Generate new verification token
	token := uuid.New().String()

	// 5. Store token in cache (expires in 24 hours)
	err = uc.cacheService.Set(
		ctx,
		fmt.Sprintf("verify:%s", token),
		usr.ID().String(),
		24*time.Hour,
	)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to cache verification token: %v", err))
		return nil, fmt.Errorf("failed to generate verification token")
	}

	// 6. Set rate limit (1 email per 5 minutes)
	uc.cacheService.Set(ctx, rateLimitKey, "1", 5*time.Minute)

	// 7. Send verification email
	verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
	err = uc.emailService.SendVerificationEmail(usr.Email(), usr.FirstName(), verificationURL)
	if err != nil {
		uc.logger.Error(fmt.Sprintf("Failed to send verification email: %v", err))
		return nil, fmt.Errorf("failed to send verification email")
	}

	uc.logger.Info(fmt.Sprintf("Verification email resent to: %s", usr.Email()))

	return &ResendVerificationOutput{
		Message: "Verification email sent successfully",
	}, nil
}
