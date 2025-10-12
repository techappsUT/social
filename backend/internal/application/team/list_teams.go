// ============================================================================
// FILE: backend/internal/application/team/list_teams.go
// ============================================================================
package team

import (
	"context"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type ListTeamsInput struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type ListTeamsOutput struct {
	Teams []TeamDTO `json:"teams"`
	Total int       `json:"total"`
}

type ListTeamsUseCase struct {
	teamRepo   team.Repository
	memberRepo team.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewListTeamsUseCase(
	teamRepo team.Repository,
	memberRepo team.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *ListTeamsUseCase {
	return &ListTeamsUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *ListTeamsUseCase) Execute(ctx context.Context, input ListTeamsInput) (*ListTeamsOutput, error) {
	// Get teams by user
	teams, err := uc.teamRepo.FindByMemberID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	teamDTOs := make([]TeamDTO, 0, len(teams))
	for _, t := range teams {
		members, _ := uc.memberRepo.FindTeamMembers(ctx, t.ID())
		teamDTO := MapTeamToDTO(t, members, uc.userRepo, ctx)
		teamDTOs = append(teamDTOs, *teamDTO)
	}

	return &ListTeamsOutput{
		Teams: teamDTOs,
		Total: len(teamDTOs),
	}, nil
}
