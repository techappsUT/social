package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/techappsUT/social-queue/internal/application/common"
)

// EmailConfig for the services package
type EmailConfig struct {
	Provider    string
	APIKey      string
	FromAddress string
	FromName    string
}

// EmailService implements common.EmailService
type EmailService struct {
	config      EmailConfig
	devMode     bool
	devCode     string
	frontendURL string
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig) common.EmailService {
	return &EmailService{
		config:      config,
		devMode:     os.Getenv("DEVELOPMENT_MODE") == "true",
		devCode:     getEnvOrDefault("DEV_EMAIL_VERIFICATION_CODE", "123456"),
		frontendURL: getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
	}
}

// SendVerificationEmail sends an email verification token
func (s *EmailService) SendVerificationEmail(ctx context.Context, email, token string) error {
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, token)

	if s.devMode {
		// For development, use simple code
		if s.devCode != "" {
			verificationLink = fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, s.devCode)
			log.Printf("🔧 DEV MODE - Email Verification:")
			log.Printf("  Email: %s", email)
			log.Printf("  Token: %s", s.devCode)
			log.Printf("  Link: %s", verificationLink)
			log.Printf("  You can POST to /api/v2/auth/verify-email with {\"token\": \"%s\"}", s.devCode)
			return nil
		}
	}

	log.Printf("📧 [%s] Sending verification email to %s", s.config.Provider, email)
	log.Printf("  Verification link: %s", verificationLink)
	// TODO: Implement actual email sending via SendGrid/SMTP
	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)

	if s.devMode {
		if s.devCode != "" {
			resetLink = fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, s.devCode)
			log.Printf("🔧 DEV MODE - Password Reset:")
			log.Printf("  Email: %s", email)
			log.Printf("  Token: %s", s.devCode)
			log.Printf("  Link: %s", resetLink)
			return nil
		}
	}

	log.Printf("📧 [%s] Sending password reset to %s with token %s", s.config.Provider, email, token)
	log.Printf("  Reset link: %s", resetLink)
	// TODO: Implement actual email sending
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	log.Printf("📧 [%s] Sending welcome email to %s (%s)", s.config.Provider, firstName, email)
	if s.devMode {
		log.Printf("  DEV MODE: Welcome email would be sent to %s", email)
	}
	// TODO: Implement actual email sending
	return nil
}

// SendInvitationEmail sends a team invitation email
func (s *EmailService) SendInvitationEmail(ctx context.Context, email, teamName, inviteToken string) error {
	log.Printf("📧 [%s] Sending invitation to %s for team %s", s.config.Provider, email, teamName)
	if s.devMode {
		inviteLink := fmt.Sprintf("%s/invite?token=%s", s.frontendURL, inviteToken)
		log.Printf("  DEV MODE: Invitation link: %s", inviteLink)
	}
	// TODO: Implement actual email sending
	return nil
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
