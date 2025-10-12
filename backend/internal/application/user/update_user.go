// path: backend/internal/application/user/update_user.go
package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// UpdateUserUseCase handles user profile updates
type UpdateUserUseCase struct {
	userRepo user.Repository
	logger   common.Logger
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	UserID    uuid.UUID `json:"userId"`
	FirstName *string   `json:"firstName,omitempty"`
	LastName  *string   `json:"lastName,omitempty"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	Timezone  *string   `json:"timezone,omitempty"`
}

// UpdateUserOutput represents the output after updating a user
type UpdateUserOutput struct {
	User *UserDTO `json:"user"`
}

// NewUpdateUserUseCase creates a new instance
func NewUpdateUserUseCase(
	userRepo user.Repository,
	logger common.Logger,
) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Execute performs the user update
func (uc *UpdateUserUseCase) Execute(ctx context.Context, input UpdateUserInput) (*UpdateUserOutput, error) {
	// Validate input
	if input.UserID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	// Fetch user
	existingUser, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		if err == user.ErrUserNotFound {
			return nil, fmt.Errorf("user not found")
		}
		uc.logger.Error("Failed to fetch user", "userId", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to fetch user")
	}

	// Check if user is active
	if !existingUser.IsActive() {
		return nil, fmt.Errorf("cannot update inactive user")
	}

	// Apply updates using domain methods
	if input.FirstName != nil || input.LastName != nil || input.AvatarURL != nil {
		firstName := existingUser.FirstName()
		if input.FirstName != nil {
			firstName = *input.FirstName
		}

		lastName := existingUser.LastName()
		if input.LastName != nil {
			lastName = *input.LastName
		}

		avatarURL := existingUser.AvatarURL()
		if input.AvatarURL != nil {
			avatarURL = *input.AvatarURL
		}

		if err := existingUser.UpdateProfile(firstName, lastName, avatarURL); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	// Persist changes
	if err := uc.userRepo.Update(ctx, existingUser); err != nil {
		uc.logger.Error("Failed to update user", "userId", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to update user")
	}

	uc.logger.Info("User updated successfully", "userId", input.UserID)

	return &UpdateUserOutput{
		User: uc.mapToDTO(existingUser),
	}, nil
}

func (uc *UpdateUserUseCase) mapToDTO(u *user.User) *UserDTO {
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
