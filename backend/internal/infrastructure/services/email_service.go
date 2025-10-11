package services

import (
	"context"
	"log"

	"github.com/techappsUT/social-queue/internal/application/common"
)

type EmailService struct {
	config EmailConfig
}

type EmailConfig struct {
	Provider    string
	APIKey      string
	FromAddress string
	FromName    string
}

func NewEmailService(config EmailConfig) common.EmailService {
	return &EmailService{config: config}
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, email, token string) error {
	log.Printf("📧 Sending verification email to %s", email)
	return nil
}

func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	log.Printf("📧 Sending password reset to %s", email)
	return nil
}

func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, firstName string) error {
	log.Printf("📧 Sending welcome email to %s", firstName)
	return nil
}

func (s *EmailService) SendInvitationEmail(ctx context.Context, email, teamName, inviteToken string) error {
	log.Printf("📧 Sending invitation to %s", email)
	return nil
}
