// ============================================================================
// FILE: backend/internal/application/team/update_member_role.go
// ============================================================================
package team

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type UpdateMemberRoleInput struct {
	TeamID    uuid.UUID             `json:"teamId" validate:"required"`
	UserID    uuid.UUID             `json:"userId" validate:"required"` // Member whose role to update
	NewRole   teamDomain.MemberRole `json:"newRole" validate:"required"`
	UpdaterID uuid.UUID             `json:"updaterId" validate:"required"` // Who is updating
}

type UpdateMemberRoleOutput struct {
	Member *MemberDTO `json:"member"`
}

type UpdateMemberRoleUseCase struct {
	teamRepo   teamDomain.Repository
	memberRepo teamDomain.MemberRepository
	userRepo   user.Repository
	logger     common.Logger
}

func NewUpdateMemberRoleUseCase(
	teamRepo teamDomain.Repository,
	memberRepo teamDomain.MemberRepository,
	userRepo user.Repository,
	logger common.Logger,
) *UpdateMemberRoleUseCase {
	return &UpdateMemberRoleUseCase{
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (uc *UpdateMemberRoleUseCase) Execute(ctx context.Context, input UpdateMemberRoleInput) (*UpdateMemberRoleOutput, error) {
	// 1. Validate new role
	if !uc.isValidRole(input.NewRole) {
		return nil, fmt.Errorf("invalid role: must be owner, admin, editor, or viewer")
	}

	// 2. Check updater authorization (only owner can change roles)
	isOwner, err := uc.memberRepo.IsOwner(ctx, input.TeamID, input.UpdaterID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	if !isOwner {
		return nil, fmt.Errorf("access denied: only team owner can change member roles")
	}

	// 3. Cannot update own role
	if input.UserID == input.UpdaterID {
		return nil, fmt.Errorf("cannot change your own role")
	}

	// 4. Get member to update
	member, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("member not found")
	}

	// 5. Check if role is actually changing
	if member.Role() == input.NewRole {
		return nil, fmt.Errorf("member already has this role")
	}

	// 6. If demoting from owner, ensure there's another owner
	if member.Role() == teamDomain.MemberRoleOwner && input.NewRole != teamDomain.MemberRoleOwner {
		ownerCount, err := uc.memberRepo.CountByRole(ctx, input.TeamID, teamDomain.MemberRoleOwner)
		if err != nil {
			uc.logger.Error("Failed to count owners", "error", err)
			return nil, fmt.Errorf("failed to verify owner count")
		}

		if ownerCount <= 1 {
			return nil, fmt.Errorf("cannot demote the last owner")
		}
	}

	// 7. If demoting from admin, ensure there's at least one admin or owner left
	if member.Role() == teamDomain.MemberRoleAdmin && input.NewRole != teamDomain.MemberRoleAdmin {
		adminCount, err := uc.memberRepo.CountByRole(ctx, input.TeamID, teamDomain.MemberRoleAdmin)
		if err != nil {
			uc.logger.Error("Failed to count admins", "error", err)
			return nil, fmt.Errorf("failed to verify admin count")
		}

		ownerCount, err := uc.memberRepo.CountByRole(ctx, input.TeamID, teamDomain.MemberRoleOwner)
		if err != nil {
			uc.logger.Error("Failed to count owners", "error", err)
			return nil, fmt.Errorf("failed to verify owner count")
		}

		// Prevent demoting last admin if there's no other admin or owner
		if adminCount <= 1 && ownerCount == 0 {
			return nil, fmt.Errorf("cannot demote the last admin")
		}
	}

	// 8. Update member role using domain method
	if err := member.ChangeRole(input.NewRole, input.UpdaterID); err != nil {
		return nil, fmt.Errorf("failed to change role: %w", err)
	}

	// 9. Persist the change
	if err := uc.memberRepo.UpdateMember(ctx, member); err != nil {
		uc.logger.Error("Failed to update member role",
			"teamId", input.TeamID,
			"userId", input.UserID,
			"error", err)
		return nil, fmt.Errorf("failed to update member role")
	}

	// 10. Get user details for response
	u, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		uc.logger.Warn("Failed to get user details", "userId", input.UserID)
	}

	// 11. Map to DTO
	memberDTO := MapMemberToDTO(member, u)

	uc.logger.Info("Member role updated",
		"teamId", input.TeamID,
		"userId", input.UserID,
		"oldRole", member.Role(),
		"newRole", input.NewRole,
		"updaterId", input.UpdaterID)

	return &UpdateMemberRoleOutput{
		Member: memberDTO,
	}, nil
}

// isValidRole checks if the role is valid
func (uc *UpdateMemberRoleUseCase) isValidRole(role teamDomain.MemberRole) bool {
	validRoles := []teamDomain.MemberRole{
		teamDomain.MemberRoleOwner,
		teamDomain.MemberRoleAdmin,
		teamDomain.MemberRoleEditor,
		teamDomain.MemberRoleViewer,
	}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}

	return false
}
