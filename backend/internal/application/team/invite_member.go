// ============================================================================
// FILE: backend/internal/application/team/invite_member.go
// FINAL FIXED VERSION - Copy this EXACTLY
// ============================================================================
package team

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/domain/user"
	"github.com/techappsUT/social-queue/internal/middleware"
)

type InviteMemberInput struct {
	TeamID    uuid.UUID             `json:"teamId" validate:"required"`
	Email     string                `json:"email" validate:"required,email"`
	Role      teamDomain.MemberRole `json:"role" validate:"required"`
	InviterID uuid.UUID             `json:"inviterId" validate:"required"`
}

type InviteMemberOutput struct {
	Member *MemberDTO `json:"member"`
}

type InviteMemberUseCase struct {
	validator    *validator.Validate // ADD THIS
	teamRepo     teamDomain.Repository
	memberRepo   teamDomain.MemberRepository
	userRepo     user.Repository
	emailService common.EmailService
	logger       common.Logger
}

func NewInviteMemberUseCase(
	teamRepo teamDomain.Repository,
	memberRepo teamDomain.MemberRepository,
	userRepo user.Repository,
	emailService common.EmailService,
	logger common.Logger,
) *InviteMemberUseCase {
	return &InviteMemberUseCase{
		teamRepo:     teamRepo,
		memberRepo:   memberRepo,
		userRepo:     userRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *InviteMemberUseCase) Execute(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error) {
	// 1. Validate email format
	// if err := uc.validateEmail(input.Email); err != nil {
	// 	return nil, err
	// }
	// ✅ Single validation call
	if err := uc.validator.Struct(input); err != nil {
		// return nil, formatValidationError(err)
		fields := middleware.FormatValidationErrors(err)
		return nil, fmt.Errorf("validation failed: %v", fields)
	}

	// 2. Validate role
	if !uc.isValidRole(input.Role) {
		return nil, fmt.Errorf("invalid role: must be owner, admin, editor, or viewer")
	}

	// 3. Check inviter authorization (must be admin or owner)
	inviter, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.InviterID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	if inviter.Role() != teamDomain.MemberRoleOwner && inviter.Role() != teamDomain.MemberRoleAdmin {
		return nil, fmt.Errorf("access denied: only admins and owners can invite members")
	}

	// 4. Get team to check limits
	t, err := uc.teamRepo.FindByID(ctx, input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("team not found")
	}

	// 5. Check member count limit
	memberCount, err := uc.memberRepo.CountActiveMembers(ctx, input.TeamID)
	if err != nil {
		uc.logger.Error("Failed to count members", "teamId", input.TeamID, "error", err)
		return nil, fmt.Errorf("failed to check team capacity")
	}

	if !t.CanAddMember(memberCount) {
		return nil, fmt.Errorf("team member limit reached")
	}

	// 6. Check if user with this email exists
	invitedUser, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// User doesn't exist yet - we'll create pending invitation
		uc.logger.Info("User not found, will create pending invitation", "email", input.Email)
	}

	var invitedUserID uuid.UUID
	if invitedUser != nil {
		invitedUserID = invitedUser.ID()

		// 7. Check if already a member
		isMember, err := uc.memberRepo.IsMember(ctx, input.TeamID, invitedUserID)
		if err != nil {
			uc.logger.Error("Failed to check membership", "error", err)
			return nil, fmt.Errorf("failed to check membership")
		}

		if isMember {
			return nil, fmt.Errorf("user is already a team member")
		}
	} else {
		// Generate a temporary UUID for pending user
		invitedUserID = uuid.New()
	}

	// 8. Create member invitation
	member, err := teamDomain.NewMember(input.TeamID, invitedUserID, input.InviterID, input.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create member invitation: %w", err)
	}

	// 9. Add member to repository (in pending status)
	if err := uc.memberRepo.AddMember(ctx, member); err != nil {
		uc.logger.Error("Failed to add member", "teamId", input.TeamID, "error", err)
		return nil, fmt.Errorf("failed to create invitation")
	}

	// 10. Send invitation email
	go func() {
		// Generate invitation token (in production, use a proper token service)
		inviteToken := fmt.Sprintf("%s-%s", input.TeamID.String(), invitedUserID.String())

		if err := uc.emailService.SendInvitationEmail(context.Background(), input.Email, t.Name(), inviteToken); err != nil {
			uc.logger.Error("Failed to send invitation email", "email", input.Email, "error", err)
		}
	}()

	// 11. Map to DTO
	memberDTO := MapMemberToDTO(member, invitedUser)
	memberDTO.Email = input.Email // Ensure email is set even if user doesn't exist yet

	uc.logger.Info("Member invited",
		"teamId", input.TeamID,
		"email", input.Email,
		"role", input.Role,
		"inviterId", input.InviterID)

	return &InviteMemberOutput{
		Member: memberDTO,
	}, nil
}

// validateEmail validates email format
// func (uc *InviteMemberUseCase) validateEmail(email string) error {
// 	if email == "" {
// 		return fmt.Errorf("email is required")
// 	}

// 	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
// 	if !emailRegex.MatchString(email) {
// 		return fmt.Errorf("invalid email format")
// 	}

// 	return nil
// }

// isValidRole checks if the role is valid
func (uc *InviteMemberUseCase) isValidRole(role teamDomain.MemberRole) bool {
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

// Helper to format validator errors
// func formatValidationError(err error) error {
// 	if validationErrors, ok := err.(validator.ValidationErrors); ok {
// 		var errors []string
// 		for _, e := range validationErrors {
// 			errors = append(errors, fmt.Sprintf("%s: %s", e.Field(), e.Tag()))
// 		}
// 		return fmt.Errorf("validation failed: %s", strings.Join(errors, ", "))
// 	}
// 	return err
// }
