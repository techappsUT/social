// path: backend/internal/application/user/delete_user.go
package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// DeleteUserUseCase handles user account deletion (soft delete)
type DeleteUserUseCase struct {
	userRepo     user.Repository
	tokenService common.TokenService
	logger       common.Logger
}

// DeleteUserInput represents the input for deleting a user
type DeleteUserInput struct {
	UserID uuid.UUID `json:"userId"`
	Reason string    `json:"reason,omitempty"` // Optional: reason for deletion
}

// DeleteUserOutput represents the output after deleting a user
type DeleteUserOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewDeleteUserUseCase creates a new instance
func NewDeleteUserUseCase(
	userRepo user.Repository,
	tokenService common.TokenService,
	logger common.Logger,
) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo:     userRepo,
		tokenService: tokenService,
		logger:       logger,
	}
}

// Execute performs the user deletion (soft delete)
func (uc *DeleteUserUseCase) Execute(ctx context.Context, input DeleteUserInput) (*DeleteUserOutput, error) {
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

	// Check if user is already deleted (deletedAt will be non-nil)
	if existingUser.DeletedAt() != nil {
		return nil, fmt.Errorf("user already deleted")
	}

	// Perform soft delete using domain method
	if err := existingUser.SoftDelete(); err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	// Persist deletion
	if err := uc.userRepo.Update(ctx, existingUser); err != nil {
		uc.logger.Error("Failed to soft delete user", "userId", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to delete user")
	}

	// Log deletion
	uc.logger.Info("User account deleted",
		"userId", input.UserID,
		"email", existingUser.Email(),
		"reason", input.Reason)

	// TODO: Additional cleanup tasks (run async)
	// - Revoke all refresh tokens
	// - Remove from teams
	// - Archive posts
	// - Send confirmation email

	return &DeleteUserOutput{
		Success: true,
		Message: "Account successfully deleted",
	}, nil
}
