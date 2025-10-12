// path: backend/internal/infrastructure/services/email_service.go
package services

import (
	"context"
	"log"

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
	config EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig) common.EmailService {
	return &EmailService{config: config}
}

// SendVerificationEmail sends an email verification token
func (s *EmailService) SendVerificationEmail(ctx context.Context, email, token string) error {
	log.Printf("📧 [%s] Sending verification email to %s with token %s", s.config.Provider, email, token)
	// TODO: Implement actual email sending via SendGrid/SMTP
	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	log.Printf("📧 [%s] Sending password reset to %s with token %s", s.config.Provider, email, token)
	// TODO: Implement actual email sending
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	log.Printf("📧 [%s] Sending welcome email to %s (%s)", s.config.Provider, firstName, email)
	// TODO: Implement actual email sending
	return nil
}

// SendInvitationEmail sends a team invitation email
func (s *EmailService) SendInvitationEmail(ctx context.Context, email, teamName, inviteToken string) error {
	log.Printf("📧 [%s] Sending invitation to %s for team %s", s.config.Provider, email, teamName)
	// TODO: Implement actual email sending
	return nil
}
