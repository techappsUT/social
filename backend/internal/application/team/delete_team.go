// ============================================================================
// path: backend/internal/application/team/delete_team.go
// ============================================================================
package team

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type DeleteTeamInput struct {
	TeamID uuid.UUID `json:"teamId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type DeleteTeamUseCase struct {
	teamRepo   team.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewDeleteTeamUseCase(
	teamRepo team.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *DeleteTeamUseCase {
	return &DeleteTeamUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *DeleteTeamUseCase) Execute(ctx context.Context, input DeleteTeamInput) error {
	// Check authorization (must be owner)
	isOwner, err := uc.memberRepo.IsOwner(ctx, input.TeamID, input.UserID)
	if err != nil || !isOwner {
		return fmt.Errorf("access denied: only team owner can delete team")
	}

	// Soft delete team
	if err := uc.teamRepo.Delete(ctx, input.TeamID); err != nil {
		uc.logger.Error("Failed to delete team", "teamId", input.TeamID, "error", err)
		return fmt.Errorf("failed to delete team")
	}

	uc.logger.Info("Team deleted", "teamId", input.TeamID, "userId", input.UserID)

	return nil
}
