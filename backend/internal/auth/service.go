// path: backend/internal/auth/service.go

package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
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
func (s *Service) Signup(req dto.SignupRequest) (*models.User, error) {
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

	// Send verification email
	go s.emailService.SendVerificationEmail(user.Email, verificationToken)

	return &user, nil
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

	// Optional: Check if email is verified
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
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.db.Create(&dbRefreshToken).Error; err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         mapUserToUserInfo(&user),
	}, nil
}

// VerifyEmail verifies user's email address
func (s *Service) VerifyEmail(token string) error {
	var user models.User
	if err := s.db.Where("verification_token = ? AND verification_token_expires_at > ?",
		token, time.Now()).First(&user).Error; err != nil {
		return ErrInvalidToken
	}

	user.EmailVerified = true
	user.VerificationToken = nil
	user.VerificationTokenExpiresAt = nil

	return s.db.Save(&user).Error
}

// ForgotPassword generates a password reset token
func (s *Service) ForgotPassword(email string) error {
	var user models.User
	if err := s.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		// Don't reveal if user exists
		return nil
	}

	resetToken, err := GenerateSecureToken(32)
	if err != nil {
		return err
	}

	tokenExpiry := time.Now().Add(1 * time.Hour)
	user.ResetToken = &resetToken
	user.ResetTokenExpiresAt = &tokenExpiry

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	// Send reset email
	go s.emailService.SendPasswordResetEmail(user.Email, resetToken)

	return nil
}

// ResetPassword resets user password using reset token
func (s *Service) ResetPassword(token, newPassword string) error {
	var user models.User
	if err := s.db.Where("reset_token = ? AND reset_token_expires_at > ?",
		token, time.Now()).First(&user).Error; err != nil {
		return ErrInvalidToken
	}

	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHash
	user.ResetToken = nil
	user.ResetTokenExpiresAt = nil

	// Revoke all refresh tokens for security
	s.db.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Update("revoked", true)

	return s.db.Save(&user).Error
}

// RefreshAccessToken generates a new access token using refresh token
func (s *Service) RefreshAccessToken(refreshToken string) (*dto.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if refresh token exists and not revoked
	refreshTokenHash := hashToken(refreshToken)
	var dbToken models.RefreshToken
	if err := s.db.Where("token_hash = ? AND user_id = ? AND revoked = false AND expires_at > ?",
		refreshTokenHash, userID, time.Now()).First(&dbToken).Error; err != nil {
		return nil, ErrRefreshTokenRevoked
	}

	// Get user
	var user models.User
	if err := s.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Generate new access token
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, string(user.Role), user.TeamID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         mapUserToUserInfo(&user),
	}, nil
}

// RevokeRefreshToken revokes a refresh token
func (s *Service) RevokeRefreshToken(refreshToken string) error {
	refreshTokenHash := hashToken(refreshToken)
	return s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ?", refreshTokenHash).
		Update("revoked", true).Error
}

// Helper functions
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func mapUserToUserInfo(user *models.User) *dto.UserInfo {
	var teamID *string
	if user.TeamID != nil {
		teamIDStr := user.TeamID.String()
		teamID = &teamIDStr
	}

	return &dto.UserInfo{
		ID:            user.ID.String(),
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Role:          string(user.Role),
		TeamID:        teamID,
		EmailVerified: user.EmailVerified,
	}
}
