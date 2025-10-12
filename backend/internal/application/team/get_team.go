// ============================================================================
// FILE: backend/internal/application/team/get_team.go
// ============================================================================
package team

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type GetTeamInput struct {
	TeamID uuid.UUID `json:"teamId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type GetTeamOutput struct {
	Team *TeamDTO `json:"team"`
}

type GetTeamUseCase struct {
	teamRepo   team.Repository
	memberRepo team.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewGetTeamUseCase(
	teamRepo team.Repository,
	memberRepo team.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *GetTeamUseCase {
	return &GetTeamUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *GetTeamUseCase) Execute(ctx context.Context, input GetTeamInput) (*GetTeamOutput, error) {
	// 1. Check user is team member
	isMember, err := uc.memberRepo.IsMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: user is not a team member")
	}

	// 2. Get team
	t, err := uc.teamRepo.FindByID(ctx, input.TeamID)
	if err != nil {
		return nil, err
	}

	// 3. Get members
	members, err := uc.memberRepo.FindTeamMembers(ctx, input.TeamID)
	if err != nil {
		uc.logger.Warn("Failed to get members", "teamId", input.TeamID)
		members = []*team.Member{}
	}

	return &GetTeamOutput{
		Team: MapTeamToDTO(t, members, uc.userRepo, ctx),
	}, nil
}
