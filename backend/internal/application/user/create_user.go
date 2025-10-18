// backend/internal/application/user/create_user.go
// ✅ FINAL FIX - Generate token BEFORE creating user

package user

import (
	"context"
	"fmt"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type CreateUserUseCase struct {
	userRepo     user.Repository
	userService  *user.Service
	tokenService common.TokenService
	emailService common.EmailService
	logger       common.Logger
}

type CreateUserInput struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=30"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type CreateUserOutput struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
}

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

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	// 1. Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// ✅ 2. Generate verification token BEFORE creating user
	verificationToken, err := user.GenerateToken()
	if err != nil {
		uc.logger.Error("Failed to generate verification token", "error", err)
		return nil, fmt.Errorf("failed to generate verification token")
	}
	tokenExpiry := time.Now().Add(24 * time.Hour)

	uc.logger.Info("✅ Generated verification token",
		"tokenPrefix", verificationToken[:10]+"...",
		"expiresAt", tokenExpiry)

	// ✅ 3. Create user via domain service WITH token
	newUser, err := uc.userService.CreateUserWithToken(
		ctx,
		input.Email,
		input.Username,
		input.Password,
		input.FirstName,
		input.LastName,
		verificationToken,
		tokenExpiry,
	)
	if err != nil {
		uc.logger.Error("Failed to create user", "error", err)
		return nil, uc.mapDomainError(err)
	}

	uc.logger.Info("✅ User created successfully with verification token",
		"userId", newUser.ID(),
		"email", newUser.Email())

	// 4. Generate JWT tokens
	accessToken, err := uc.tokenService.GenerateAccessToken(
		newUser.ID().String(),
		newUser.Email(),
		string(newUser.Role()),
	)
	if err != nil {
		uc.logger.Error("Failed to generate access token", "error", err)
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(newUser.ID().String())
	if err != nil {
		uc.logger.Error("Failed to generate refresh token", "error", err)
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// 5. Send verification email asynchronously
	go func() {
		if err := uc.emailService.SendVerificationEmail(
			context.Background(),
			newUser.Email(),
			verificationToken,
		); err != nil {
			uc.logger.Error("Failed to send verification email",
				"email", newUser.Email(),
				"error", err)
		}
	}()

	// 6. Send welcome email asynchronously
	go func() {
		if err := uc.emailService.SendWelcomeEmail(
			context.Background(),
			newUser.Email(),
			newUser.FirstName(),
		); err != nil {
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
