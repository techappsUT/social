// ===========================================================================
// FILE: backend/internal/application/team/get_pending_invitations.go
// NEW - Get user's pending team invitations
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

type GetPendingInvitationsInput struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type GetPendingInvitationsOutput struct {
	Invitations []*InvitationDTO `json:"invitations"`
}

type InvitationDTO struct {
	TeamID      string `json:"teamId"`
	TeamName    string `json:"teamName"`
	TeamSlug    string `json:"teamSlug"`
	Role        string `json:"role"`
	InvitedBy   string `json:"invitedBy"`
	InvitedAt   string `json:"invitedAt"`
	InviterName string `json:"inviterName,omitempty"`
}

type GetPendingInvitationsUseCase struct {
	teamRepo   teamDomain.Repository
	memberRepo teamDomain.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewGetPendingInvitationsUseCase(
	teamRepo teamDomain.Repository,
	memberRepo teamDomain.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *GetPendingInvitationsUseCase {
	return &GetPendingInvitationsUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *GetPendingInvitationsUseCase) Execute(ctx context.Context, input GetPendingInvitationsInput) (*GetPendingInvitationsOutput, error) {
	// 1. Get all user memberships
	memberships, err := uc.memberRepo.FindUserMemberships(ctx, input.UserID)
	if err != nil {
		uc.logger.Error("Failed to get user memberships", "userId", input.UserID, "error", err)
		return nil, fmt.Errorf("failed to get invitations")
	}

	// 2. Filter pending invitations
	var invitations []*InvitationDTO
	for _, member := range memberships {
		if member.Status() != teamDomain.MemberStatusPending {
			continue
		}

		// Get team details
		team, err := uc.teamRepo.FindByID(ctx, member.TeamID())
		if err != nil {
			uc.logger.Warn("Failed to get team", "teamId", member.TeamID())
			continue
		}

		// Get inviter details (optional)
		var inviterName string
		if inviter, err := uc.userRepo.FindByID(ctx, member.InvitedBy()); err == nil {
			inviterName = fmt.Sprintf("%s %s", inviter.FirstName(), inviter.LastName())
		}

		invitations = append(invitations, &InvitationDTO{
			TeamID:      member.TeamID().String(),
			TeamName:    team.Name(),
			TeamSlug:    team.Slug(),
			Role:        string(member.Role()),
			InvitedBy:   member.InvitedBy().String(),
			InvitedAt:   member.InvitedAt().Format("2006-01-02T15:04:05Z"),
			InviterName: inviterName,
		})
	}

	uc.logger.Info("Retrieved pending invitations", "userId", input.UserID, "count", len(invitations))

	return &GetPendingInvitationsOutput{
		Invitations: invitations,
	}, nil
}
