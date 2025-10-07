// path: backend/internal/dto/auth.go

package dto

// ============================================================================
// REQUEST DTOs (Accept both camelCase and snake_case via JSON tags)
// ============================================================================

type SignupRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required"` // camelCase for frontend
	LastName  string `json:"lastName" validate:"required"`  // camelCase for frontend
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"` // camelCase
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"` // camelCase, also read from cookie
}

// ============================================================================
// RESPONSE DTOs (camelCase for frontend consistency)
// ============================================================================

type AuthResponse struct {
	AccessToken  string    `json:"accessToken"`  // camelCase
	RefreshToken string    `json:"refreshToken"` // camelCase
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	FirstName     string  `json:"firstName"` // camelCase
	LastName      string  `json:"lastName"`  // camelCase
	FullName      string  `json:"fullName"`  // camelCase - computed field
	Role          string  `json:"role"`
	TeamID        *string `json:"teamId"`        // camelCase
	EmailVerified bool    `json:"emailVerified"` // camelCase
	AvatarURL     *string `json:"avatarUrl"`     // camelCase
}

type MessageResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"` // Validation errors
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// Helper to create UserInfo from User model
func NewUserInfo(id, email, firstName, lastName, role string, teamID *string, emailVerified bool) *UserInfo {
	fullName := firstName
	if lastName != "" {
		fullName = firstName + " " + lastName
	}

	return &UserInfo{
		ID:            id,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		FullName:      fullName,
		Role:          role,
		TeamID:        teamID,
		EmailVerified: emailVerified,
	}
}
