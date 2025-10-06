// path: backend/pkg/email/service.go

package email

import (
	"fmt"
	"log"
)

type Service interface {
	SendVerificationEmail(email, token string) error
	SendPasswordResetEmail(email, token string) error
}

type MockEmailService struct {
	baseURL string
}

func NewMockEmailService(baseURL string) *MockEmailService {
	return &MockEmailService{
		baseURL: baseURL,
	}
}

func (s *MockEmailService) SendVerificationEmail(email, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)
	log.Printf("📧 [MOCK] Verification email to %s: %s", email, verificationURL)
	// In production, integrate with SendGrid, AWS SES, or Mailgun
	return nil
}

func (s *MockEmailService) SendPasswordResetEmail(email, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)
	log.Printf("📧 [MOCK] Password reset email to %s: %s", email, resetURL)
	// In production, integrate with SendGrid, AWS SES, or Mailgun
	return nil
}
