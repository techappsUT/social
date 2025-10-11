// path: backend/internal/application/user/create_user.go
package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// CreateUserUseCase handles user registration
type CreateUserUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	tokenService common.TokenService
	emailService common.EmailService
	logger       common.Logger
}

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// CreateUserOutput represents the output after creating a user
type CreateUserOutput struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
}

// UserDTO represents user data transfer object
type UserDTO struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
}

// NewCreateUserUseCase creates a new instance
func NewCreateUserUseCase(
	userRepo user.Repository,
	userService *user.Service,
	tokenService common.TokenService,
	emailService common.EmailService,
	logger common.Logger,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:     userRepo,
		userService:  userService,
		tokenService: tokenService,
		emailService: emailService,
		logger:       logger,
	}
}

// Execute performs the user creation
func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create user through domain service
	newUser, err := uc.userService.CreateUser(
		ctx,
		input.Email,
		input.Username,
		input.Password,
		input.FirstName,
		input.LastName,
	)
	if err != nil {
		return nil, uc.mapDomainError(err)
	}

	// Generate tokens
	accessToken, err := uc.tokenService.GenerateAccessToken(
		newUser.ID().String(),
		newUser.Email(),
		string(newUser.Role()),
	)
	if err != nil {
		uc.logger.Error("Failed to generate access token", "error", err)
		return nil, fmt.Errorf("failed to generate tokens")
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(newUser.ID().String())
	if err != nil {
		uc.logger.Error("Failed to generate refresh token", "error", err)
		return nil, fmt.Errorf("failed to generate tokens")
	}

	// Send verification email (async)
	go func() {
		token := uc.generateVerificationToken(newUser.ID())
		if err := uc.emailService.SendVerificationEmail(context.Background(), newUser.Email(), token); err != nil {
			uc.logger.Error("Failed to send verification email", "error", err)
		}
	}()

	// Send welcome email (async)
	go func() {
		if err := uc.emailService.SendWelcomeEmail(context.Background(), newUser.Email(), newUser.FirstName()); err != nil {
			uc.logger.Error("Failed to send welcome email", "error", err)
		}
	}()

	return &CreateUserOutput{
		User:         uc.mapToDTO(newUser),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *CreateUserUseCase) validateInput(input CreateUserInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if input.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if input.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if input.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	return nil
}

func (uc *CreateUserUseCase) mapDomainError(err error) error {
	switch err {
	case user.ErrEmailAlreadyExists:
		return fmt.Errorf("email already registered")
	case user.ErrUsernameAlreadyExists:
		return fmt.Errorf("username already taken")
	default:
		return err
	}
}

func (uc *CreateUserUseCase) mapToDTO(u *user.User) *UserDTO {
	return &UserDTO{
		ID:            u.ID().String(),
		Email:         u.Email(),
		Username:      u.Username(),
		FirstName:     u.FirstName(),
		LastName:      u.LastName(),
		Role:          string(u.Role()),
		Status:        string(u.Status()),
		EmailVerified: u.IsEmailVerified(),
	}
}

func (uc *CreateUserUseCase) generateVerificationToken(userID uuid.UUID) string {
	// Simple token generation - in production use a proper token service
	return fmt.Sprintf("%s-%d", userID.String(), uuid.New().ID())
}
