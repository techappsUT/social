// ===========================================================================
// FILE: backend/internal/application/team/accept_invitation.go
// NEW - Complete invitation acceptance use case
// ===========================================================================
package team

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type AcceptInvitationInput struct {
	TeamID uuid.UUID `json:"teamId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"` // From auth context
}

type AcceptInvitationOutput struct {
	Team   *TeamDTO   `json:"team"`
	Member *MemberDTO `json:"member"`
}

type AcceptInvitationUseCase struct {
	teamRepo   teamDomain.Repository
	memberRepo teamDomain.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewAcceptInvitationUseCase(
	teamRepo teamDomain.Repository,
	memberRepo teamDomain.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *AcceptInvitationUseCase {
	return &AcceptInvitationUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *AcceptInvitationUseCase) Execute(ctx context.Context, input AcceptInvitationInput) (*AcceptInvitationOutput, error) {
	// 1. Find the pending invitation
	member, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		uc.logger.Error("Invitation not found",
			"teamId", input.TeamID,
			"userId", input.UserID,
			"error", err)
		return nil, fmt.Errorf("invitation not found")
	}

	// 2. Check if already accepted
	if member.Status() != teamDomain.MemberStatusPending {
		uc.logger.Warn("Invitation already processed",
			"teamId", input.TeamID,
			"userId", input.UserID,
			"status", member.Status())
		return nil, fmt.Errorf("invitation already processed")
	}

	// 3. Accept the invitation (domain logic)
	if err := member.AcceptInvitation(); err != nil {
		uc.logger.Error("Failed to accept invitation", "error", err)
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	// 4. Persist the change
	if err := uc.memberRepo.UpdateMember(ctx, member); err != nil {
		uc.logger.Error("Failed to update member",
			"teamId", input.TeamID,
			"userId", input.UserID,
			"error", err)
		return nil, fmt.Errorf("failed to update membership")
	}

	// 5. Get team details
	team, err := uc.teamRepo.FindByID(ctx, input.TeamID)
	if err != nil {
		uc.logger.Error("Failed to get team", "teamId", input.TeamID, "error", err)
		return nil, fmt.Errorf("failed to get team details")
	}

	// 6. Get all team members
	members, err := uc.memberRepo.FindTeamMembers(ctx, input.TeamID)
	if err != nil {
		uc.logger.Warn("Failed to get team members", "teamId", input.TeamID)
		members = []*teamDomain.Member{}
	}

	// 7. Get user details
	u, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		uc.logger.Warn("Failed to get user details", "userId", input.UserID)
	}

	uc.logger.Info("Invitation accepted successfully",
		"teamId", input.TeamID,
		"userId", input.UserID,
		"role", member.Role())

	// 8. Map to DTOs
	teamDTO := MapTeamToDTO(team, members, uc.userRepo, ctx)
	memberDTO := MapMemberToDTO(member, u)

	return &AcceptInvitationOutput{
		Team:   teamDTO,
		Member: memberDTO,
	}, nil
}
