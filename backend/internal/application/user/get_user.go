// path: backend/internal/application/user/get_user.go
package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// GetUserUseCase handles retrieving a single user
type GetUserUseCase struct {
	userRepo user.Repository
	logger   common.Logger
}

// GetUserInput represents the input for getting a user
type GetUserInput struct {
	UserID uuid.UUID `json:"userId"`
}

// GetUserOutput represents the output after getting a user
type GetUserOutput struct {
	User *UserDTO `json:"user"`
}

// NewGetUserUseCase creates a new instance
func NewGetUserUseCase(
	userRepo user.Repository,
	logger common.Logger,
) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Execute retrieves a user by ID
func (uc *GetUserUseCase) Execute(ctx context.Context, input GetUserInput) (*GetUserOutput, error) {
	// Validate input
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	// Fetch user
	foundUser, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, fmt.Errorf("user not found")
		}
		uc.logger.Error("Failed to fetch user", "userId", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to fetch user")
	}

	uc.logger.Debug("User retrieved successfully", "userId", input.UserID)

	return &GetUserOutput{
		User: uc.mapToDTO(foundUser),
	}, nil
}

func (uc *GetUserUseCase) mapToDTO(u *user.User) *UserDTO {
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
