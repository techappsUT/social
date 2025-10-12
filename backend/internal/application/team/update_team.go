// ============================================================================
// path: backend/internal/application/team/update_team.go
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

type UpdateTeamInput struct {
	TeamID uuid.UUID `json:"teamId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
	Name   *string   `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
}

type UpdateTeamOutput struct {
	Team *TeamDTO `json:"team"`
}

type UpdateTeamUseCase struct {
	teamRepo   team.Repository
	memberRepo team.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewUpdateTeamUseCase(
	teamRepo team.Repository,
	memberRepo team.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *UpdateTeamUseCase {
	return &UpdateTeamUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *UpdateTeamUseCase) Execute(ctx context.Context, input UpdateTeamInput) (*UpdateTeamOutput, error) {
	// 1. Check authorization
	member, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	if member.Role() != team.MemberRoleOwner && member.Role() != team.MemberRoleAdmin {
		return nil, fmt.Errorf("access denied: admin role required")
	}

	// 2. Get team
	t, err := uc.teamRepo.FindByID(ctx, input.TeamID)
	if err != nil {
		return nil, err
	}

	// 3. Update fields using domain method UpdateProfile
	if input.Name != nil {
		// Use UpdateProfile which exists in your domain
		if err := t.UpdateProfile(*input.Name, t.Description(), t.AvatarURL()); err != nil {
			return nil, err
		}
	}

	// 4. Persist
	if err := uc.teamRepo.Update(ctx, t); err != nil {
		uc.logger.Error("Failed to update team", "teamId", input.TeamID, "error", err)
		return nil, fmt.Errorf("failed to update team")
	}

	// 5. Get members
	members, _ := uc.memberRepo.FindTeamMembers(ctx, input.TeamID)

	uc.logger.Info("Team updated", "teamId", input.TeamID, "userId", input.UserID)

	return &UpdateTeamOutput{
		Team: MapTeamToDTO(t, members, uc.userRepo, ctx),
	}, nil
}
