// path: backend/internal/auth/service.go

package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/techappsUT/social-queue/internal/dto"
	"github.com/techappsUT/social-queue/internal/models"
	"github.com/techappsUT/social-queue/pkg/email"
)

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailNotVerified    = errors.New("email not verified")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
)

type Service struct {
	db           *gorm.DB
	tokenService *TokenService
	emailService email.Service
}

func NewService(db *gorm.DB, tokenService *TokenService, emailService email.Service) *Service {
	return &Service{
		db:           db,
		tokenService: tokenService,
		emailService: emailService,
	}
}

// Signup creates a new user account
func (s *Service) Signup(req dto.SignupRequest) (*dto.MessageResponse, error) {
	// Check if user exists
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken, err := GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	tokenExpiry := time.Now().Add(24 * time.Hour)

	user := models.User{
		Email:                      req.Email,
		PasswordHash:               passwordHash,
		FirstName:                  req.FirstName,
		LastName:                   req.LastName,
		Role:                       models.RoleUser,
		VerificationToken:          &verificationToken,
		VerificationTokenExpiresAt: &tokenExpiry,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Send verification email (async)
	go s.emailService.SendVerificationEmail(user.Email, verificationToken)

	return &dto.MessageResponse{
		Message: "Account created successfully. Please check your email to verify your account.",
		Success: true,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Optional: Check if email is verified (uncomment if required)
	// if !user.EmailVerified {
	// 	return nil, ErrEmailNotVerified
	// }

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, string(user.Role), user.TeamID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Store refresh token hash in database
	refreshTokenHash := hashToken(refreshToken)
	dbRefreshToken := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
	}
	if err := s.db.Create(&dbRefreshToken).Error; err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// Build response with aligned format
	// teamIDStr := ""
	// if user.TeamID != nil {
	// 	teamIDStr = user.TeamID.String()
	// }

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.NewUserInfo(
			user.ID.String(),
			user.Email,
			user.FirstName,
			user.LastName,
			string(user.Role),
			func() *string {
				if user.TeamID != nil {
					s := user.TeamID.String()
					return &s
				}
				return nil
			}(),
			user.EmailVerified,
		),
	}, nil
}

// RefreshToken generates new access token using refresh token
func (s *Service) RefreshToken(refreshToken string) (*dto.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if refresh token exists and is not revoked
	tokenHash := hashToken(refreshToken)
	var dbToken models.RefreshToken
	if err := s.db.Where("token_hash = ? AND user_id = ? AND revoked = false AND expires_at > ?",
		tokenHash, claims.UserID, time.Now()).First(&dbToken).Error; err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", claims.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Generate new access token
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, string(user.Role), user.TeamID)
	if err != nil {
		return nil, err
	}

	// Optionally rotate refresh token (recommended for security)
	newRefreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	dbToken.Revoked = true
	s.db.Save(&dbToken)

	// Store new refresh token
	newTokenHash := hashToken(newRefreshToken)
	newDBToken := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	s.db.Create(&newDBToken)

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: dto.NewUserInfo(
			user.ID.String(),
			user.Email,
			user.FirstName,
			user.LastName,
			string(user.Role),
			func() *string {
				if user.TeamID != nil {
					s := user.TeamID.String()
					return &s
				}
				return nil
			}(),
			user.EmailVerified,
		),
	}, nil
}

// VerifyEmail verifies user's email address
func (s *Service) VerifyEmail(token string) (*dto.MessageResponse, error) {
	var user models.User
	if err := s.db.Where("verification_token = ? AND verification_token_expires_at > ?",
		token, time.Now()).First(&user).Error; err != nil {
		return nil, ErrInvalidToken
	}

	user.EmailVerified = true
	user.VerificationToken = nil
	user.VerificationTokenExpiresAt = nil

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &dto.MessageResponse{
		Message: "Email verified successfully!",
		Success: true,
	}, nil
}

// ResendVerification resends email verification link
func (s *Service) ResendVerification(email string) (*dto.MessageResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		// Don't reveal if user exists - security best practice
		return &dto.MessageResponse{
			Message: "If an account with that email exists and is unverified, a verification email has been sent.",
			Success: true,
		}, nil
	}

	// Check if already verified
	if user.EmailVerified {
		return &dto.MessageResponse{
			Message: "This email is already verified. You can login now.",
			Success: true,
		}, nil
	}

	// Generate new verification token
	verificationToken, err := GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	tokenExpiry := time.Now().Add(24 * time.Hour)
	user.VerificationToken = &verificationToken
	user.VerificationTokenExpiresAt = &tokenExpiry

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	// Send verification email (async)
	go s.emailService.SendVerificationEmail(user.Email, verificationToken)

	return &dto.MessageResponse{
		Message: "If an account with that email exists and is unverified, a verification email has been sent.",
		Success: true,
	}, nil
}

// ForgotPassword generates a password reset token
func (s *Service) ForgotPassword(email string) (*dto.MessageResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		// Don't reveal if user exists - security best practice
		return &dto.MessageResponse{
			Message: "If an account with that email exists, a password reset link has been sent.",
			Success: true,
		}, nil
	}

	resetToken, err := GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	tokenExpiry := time.Now().Add(1 * time.Hour)
	user.ResetToken = &resetToken
	user.ResetTokenExpiresAt = &tokenExpiry

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	// Send reset email (async)
	go s.emailService.SendPasswordResetEmail(user.Email, resetToken)

	return &dto.MessageResponse{
		Message: "If an account with that email exists, a password reset link has been sent.",
		Success: true,
	}, nil
}

// ResetPassword resets user password using reset token
func (s *Service) ResetPassword(token, newPassword string) (*dto.MessageResponse, error) {
	var user models.User
	if err := s.db.Where("reset_token = ? AND reset_token_expires_at > ?",
		token, time.Now()).First(&user).Error; err != nil {
		return nil, ErrInvalidToken
	}

	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = passwordHash
	user.ResetToken = nil
	user.ResetTokenExpiresAt = nil

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	// Revoke all refresh tokens for security
	s.db.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Update("revoked", true)

	return &dto.MessageResponse{
		Message: "Password reset successfully. Please login with your new password.",
		Success: true,
	}, nil
}

// Logout revokes the refresh token
func (s *Service) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil // Already logged out
	}

	tokenHash := hashToken(refreshToken)
	return s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked", true).Error
}

// Helper function to hash tokens for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
