// ============================================================================
// FILE: backend/internal/application/team/remove_member.go
// ============================================================================
package team

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
)

type RemoveMemberInput struct {
	TeamID    uuid.UUID `json:"teamId" validate:"required"`
	UserID    uuid.UUID `json:"userId" validate:"required"`    // Member to remove
	RemoverID uuid.UUID `json:"removerId" validate:"required"` // Who is removing
}

type RemoveMemberUseCase struct {
	teamRepo   teamDomain.Repository
	memberRepo teamDomain.MemberRepository
	logger     common.Logger
}

func NewRemoveMemberUseCase(
	teamRepo teamDomain.Repository,
	memberRepo teamDomain.MemberRepository,
	logger common.Logger,
) *RemoveMemberUseCase {
	return &RemoveMemberUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *RemoveMemberUseCase) Execute(ctx context.Context, input RemoveMemberInput) error {
	// 1. Check remover authorization (must be admin or owner)
	remover, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.RemoverID)
	if err != nil {
		return fmt.Errorf("access denied: not a team member")
	}

	if remover.Role() != teamDomain.MemberRoleOwner && remover.Role() != teamDomain.MemberRoleAdmin {
		return fmt.Errorf("access denied: only admins and owners can remove members")
	}

	// 2. Check if user trying to remove themselves
	if input.UserID == input.RemoverID {
		return fmt.Errorf("cannot remove yourself from the team, use leave team instead")
	}

	// 3. Get member to remove
	member, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return fmt.Errorf("member not found")
	}

	// 4. Cannot remove the team owner
	isOwner, err := uc.memberRepo.IsOwner(ctx, input.TeamID, input.UserID)
	if err != nil {
		uc.logger.Error("Failed to check owner status", "error", err)
		return fmt.Errorf("failed to verify member status")
	}

	if isOwner {
		return fmt.Errorf("cannot remove team owner, transfer ownership first")
	}

	// 5. If removing an admin, check if they're the last admin
	if member.Role() == teamDomain.MemberRoleAdmin {
		adminCount, err := uc.memberRepo.CountByRole(ctx, input.TeamID, teamDomain.MemberRoleAdmin)
		if err != nil {
			uc.logger.Error("Failed to count admins", "error", err)
			return fmt.Errorf("failed to verify admin count")
		}

		// Check if there's also an owner (owner can manage the team)
		ownerCount, err := uc.memberRepo.CountByRole(ctx, input.TeamID, teamDomain.MemberRoleOwner)
		if err != nil {
			uc.logger.Error("Failed to count owners", "error", err)
			return fmt.Errorf("failed to verify owner count")
		}

		// Prevent removing last admin if there's no owner
		if adminCount <= 1 && ownerCount == 0 {
			return fmt.Errorf("cannot remove the last admin")
		}
	}

	// 6. Non-owners can only remove members with lower or equal roles
	if remover.Role() != teamDomain.MemberRoleOwner {
		// Admin can only remove editors and viewers
		if member.Role() == teamDomain.MemberRoleAdmin || member.Role() == teamDomain.MemberRoleOwner {
			return fmt.Errorf("insufficient permissions: admins cannot remove other admins or owners")
		}
	}

	// 7. Remove member (soft delete)
	if err := uc.memberRepo.RemoveMember(ctx, input.TeamID, input.UserID); err != nil {
		uc.logger.Error("Failed to remove member",
			"teamId", input.TeamID,
			"userId", input.UserID,
			"error", err)
		return fmt.Errorf("failed to remove member")
	}

	uc.logger.Info("Member removed from team",
		"teamId", input.TeamID,
		"userId", input.UserID,
		"removerId", input.RemoverID,
		"memberRole", member.Role())

	return nil
}
