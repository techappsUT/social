// ============================================================================
// path: backend/internal/application/team/create_team.go
// REMOVE generateSlug from here since it's now in helpers.go
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

type CreateTeamInput struct {
	Name    string    `json:"name" validate:"required,min=3,max=100"`
	OwnerID uuid.UUID `json:"ownerId" validate:"required"`
}

type CreateTeamOutput struct {
	Team *TeamDTO `json:"team"`
}

type CreateTeamUseCase struct {
	teamRepo   team.Repository
	memberRepo team.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewCreateTeamUseCase(
	teamRepo team.Repository,
	memberRepo team.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *CreateTeamUseCase {
	return &CreateTeamUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *CreateTeamUseCase) Execute(ctx context.Context, input CreateTeamInput) (*CreateTeamOutput, error) {
	// 1. Validate owner exists
	owner, err := uc.userRepo.FindByID(ctx, input.OwnerID)
	if err != nil {
		uc.logger.Error("Owner not found", "ownerId", input.OwnerID, "error", err)
		return nil, fmt.Errorf("owner not found")
	}

	if !owner.IsActive() {
		return nil, fmt.Errorf("owner account is not active")
	}

	// 2. Create team entity (using correct signature)
	slug := generateSlug(input.Name)
	newTeam, err := team.NewTeam(input.Name, slug, "", input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid team data: %w", err)
	}

	// 3. Persist team
	if err := uc.teamRepo.Create(ctx, newTeam); err != nil {
		uc.logger.Error("Failed to create team", "error", err)
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// 4. Add owner as member
	member, err := team.NewMember(newTeam.ID(), input.OwnerID, input.OwnerID, team.MemberRoleOwner)
	if err != nil {
		// Rollback
		_ = uc.teamRepo.Delete(ctx, newTeam.ID())
		return nil, fmt.Errorf("failed to create owner member: %w", err)
	}

	// Auto-accept for owner
	_ = member.AcceptInvitation()

	if err := uc.memberRepo.AddMember(ctx, member); err != nil {
		uc.logger.Error("Failed to add owner to team", "teamId", newTeam.ID(), "error", err)
		_ = uc.teamRepo.Delete(ctx, newTeam.ID())
		return nil, fmt.Errorf("failed to add owner to team")
	}

	// 5. Get members with user info
	members, err := uc.memberRepo.FindTeamMembers(ctx, newTeam.ID())
	if err != nil {
		uc.logger.Warn("Failed to get team members", "teamId", newTeam.ID())
		members = []*team.Member{}
	}

	// 6. Map to DTO
	teamDTO := MapTeamToDTO(newTeam, members, uc.userRepo, ctx)

	uc.logger.Info("Team created successfully", "teamId", newTeam.ID(), "ownerId", input.OwnerID)

	return &CreateTeamOutput{
		Team: teamDTO,
	}, nil
}
