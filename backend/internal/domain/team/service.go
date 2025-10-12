// path: backend/internal/domain/team/service.go
// 🆕 NEW - Clean Architecture

package team

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service provides domain-level business logic for teams
type Service struct {
	repo       Repository
	memberRepo MemberRepository
}

// NewService creates a new team domain service
func NewService(repo Repository, memberRepo MemberRepository) *Service {
	return &Service{
		repo:       repo,
		memberRepo: memberRepo,
	}
}

// CreateTeamWithOwner creates a new team and adds the owner as the first member
func (s *Service) CreateTeamWithOwner(ctx context.Context, name, slug, description string, ownerID uuid.UUID) (*Team, *Member, error) {
	// Check if slug already exists
	exists, err := s.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return nil, nil, ErrSlugAlreadyExists
	}

	// Create team entity
	team, err := NewTeam(name, slug, description, ownerID)
	if err != nil {
		return nil, nil, err
	}

	// Create team in repository
	if err := s.repo.Create(ctx, team); err != nil {
		return nil, nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add owner as first member
	member, err := NewMember(team.ID(), ownerID, ownerID, MemberRoleOwner)
	if err != nil {
		// Rollback team creation if member creation fails
		// This should ideally be in a transaction
		_ = s.repo.Delete(ctx, team.ID())
		return nil, nil, fmt.Errorf("failed to create owner member: %w", err)
	}

	// Auto-accept for owner
	member.status = MemberStatusActive
	now := time.Now().UTC()
	member.joinedAt = &now

	if err := s.memberRepo.AddMember(ctx, member); err != nil {
		// Rollback team creation
		_ = s.repo.Delete(ctx, team.ID())
		return nil, nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	return team, member, nil
}

// InviteUserToTeam creates an invitation for a user to join the team
func (s *Service) InviteUserToTeam(ctx context.Context, teamID, inviterID uuid.UUID, userID uuid.UUID, role MemberRole) (*Member, error) {
	// Verify inviter has permission
	inviter, err := s.memberRepo.FindMember(ctx, teamID, inviterID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if !inviter.CanInviteMembers() {
		return nil, ErrInsufficientPermissions
	}

	// Check if user is already a member
	exists, err := s.memberRepo.IsMember(ctx, teamID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if exists {
		return nil, ErrUserAlreadyMember
	}

	// Get team to check limits
	team, err := s.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, ErrTeamNotFound
	}

	// Check member limit
	memberCount, err := s.memberRepo.CountActiveMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}

	if !team.CanAddMember(memberCount) {
		return nil, ErrMemberLimitExceeded
	}

	// Create member invitation
	member, err := NewMember(teamID, userID, inviterID, role)
	if err != nil {
		return nil, err
	}

	// Add member in pending status
	if err := s.memberRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to create member invitation: %w", err)
	}

	return member, nil
}

// AcceptInvitation accepts a team invitation
func (s *Service) AcceptInvitation(ctx context.Context, teamID, userID uuid.UUID) error {
	// Find the member invitation
	member, err := s.memberRepo.FindMember(ctx, teamID, userID)
	if err != nil {
		return ErrInvitationNotFound
	}

	// Accept the invitation
	if err := member.AcceptInvitation(); err != nil {
		return err
	}

	// Update in repository
	if err := s.memberRepo.UpdateMember(ctx, member); err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	return nil
}

// RemoveMemberFromTeam removes a member from the team
func (s *Service) RemoveMemberFromTeam(ctx context.Context, teamID, removerID, userID uuid.UUID) error {
	// Check remover's permission
	remover, err := s.memberRepo.FindMember(ctx, teamID, removerID)
	if err != nil {
		return ErrUnauthorized
	}

	if !remover.CanManageMembers() {
		return ErrInsufficientPermissions
	}

	// Get member to remove
	member, err := s.memberRepo.FindMember(ctx, teamID, userID)
	if err != nil {
		return ErrMemberNotFound
	}

	// Cannot remove owner
	if member.Role() == MemberRoleOwner {
		// Check if there are other owners
		ownerCount, err := s.memberRepo.CountByRole(ctx, teamID, MemberRoleOwner)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if ownerCount <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}

	// Remove member
	if err := member.Remove(removerID); err != nil {
		return err
	}

	// Update in repository
	if err := s.memberRepo.UpdateMember(ctx, member); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// ChangeMemberRole changes a team member's role
func (s *Service) ChangeMemberRole(ctx context.Context, teamID, changerID, userID uuid.UUID, newRole MemberRole) error {
	// Check changer's permission
	changer, err := s.memberRepo.FindMember(ctx, teamID, changerID)
	if err != nil {
		return ErrUnauthorized
	}

	if !changer.CanManageMembers() {
		return ErrInsufficientPermissions
	}

	// Get member to update
	member, err := s.memberRepo.FindMember(ctx, teamID, userID)
	if err != nil {
		return ErrMemberNotFound
	}

	// If demoting an owner, ensure there's another owner
	if member.Role() == MemberRoleOwner && newRole != MemberRoleOwner {
		ownerCount, err := s.memberRepo.CountByRole(ctx, teamID, MemberRoleOwner)
		if err != nil {
			return fmt.Errorf("failed to count owners: %w", err)
		}
		if ownerCount <= 1 {
			return ErrCannotDemoteOnlyOwner
		}
	}

	// Change role
	if err := member.ChangeRole(newRole, changerID); err != nil {
		return err
	}

	// Update in repository
	if err := s.memberRepo.UpdateMember(ctx, member); err != nil {
		return fmt.Errorf("failed to change member role: %w", err)
	}

	return nil
}

// TransferTeamOwnership transfers ownership of a team to another member
func (s *Service) TransferTeamOwnership(ctx context.Context, teamID, currentOwnerID, newOwnerID uuid.UUID) error {
	// Verify current owner
	isOwner, err := s.memberRepo.IsOwner(ctx, teamID, currentOwnerID)
	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !isOwner {
		return ErrUnauthorized
	}

	// Get team
	team, err := s.repo.FindByID(ctx, teamID)
	if err != nil {
		return ErrTeamNotFound
	}

	// Verify new owner is a member
	newOwnerMember, err := s.memberRepo.FindMember(ctx, teamID, newOwnerID)
	if err != nil {
		return ErrMemberNotFound
	}

	if !newOwnerMember.IsActive() {
		return ErrMemberNotInactive
	}

	// Transfer ownership in team entity
	if err := team.TransferOwnership(newOwnerID); err != nil {
		return err
	}

	// Update team
	if err := s.repo.Update(ctx, team); err != nil {
		return fmt.Errorf("failed to transfer ownership: %w", err)
	}

	// Update new owner's role to owner
	if newOwnerMember.Role() != MemberRoleOwner {
		newOwnerMember.role = MemberRoleOwner
		if err := s.memberRepo.UpdateMember(ctx, newOwnerMember); err != nil {
			// Log error but don't fail - team ownership is already transferred
		}
	}

	// Optionally demote old owner to admin
	if currentOwnerID != newOwnerID {
		oldOwnerMember, err := s.memberRepo.FindMember(ctx, teamID, currentOwnerID)
		if err == nil && oldOwnerMember.Role() == MemberRoleOwner {
			oldOwnerMember.role = MemberRoleAdmin
			_ = s.memberRepo.UpdateMember(ctx, oldOwnerMember)
		}
	}

	return nil
}

// UpgradeTeamPlan upgrades a team's subscription plan
func (s *Service) UpgradeTeamPlan(ctx context.Context, teamID uuid.UUID, newPlan Plan, upgradedBy uuid.UUID) error {
	// Verify upgrader has permission
	member, err := s.memberRepo.FindMember(ctx, teamID, upgradedBy)
	if err != nil {
		return ErrUnauthorized
	}

	if !member.CanManageBilling() {
		return ErrInsufficientPermissions
	}

	// Get team
	team, err := s.repo.FindByID(ctx, teamID)
	if err != nil {
		return ErrTeamNotFound
	}

	// Upgrade plan
	if err := team.UpgradePlan(newPlan); err != nil {
		return err
	}

	// Update team
	if err := s.repo.Update(ctx, team); err != nil {
		return fmt.Errorf("failed to upgrade plan: %w", err)
	}

	// TODO: Emit domain event for billing system
	// events.Publish(TeamPlanUpgradedEvent{TeamID: teamID, OldPlan: oldPlan, NewPlan: newPlan})

	return nil
}

// SuspendTeam suspends a team (typically for non-payment)
func (s *Service) SuspendTeam(ctx context.Context, teamID uuid.UUID, reason string) error {
	// Get team
	team, err := s.repo.FindByID(ctx, teamID)
	if err != nil {
		return ErrTeamNotFound
	}

	// Suspend team
	if err := team.Suspend(reason); err != nil {
		return err
	}

	// Update team
	if err := s.repo.Update(ctx, team); err != nil {
		return fmt.Errorf("failed to suspend team: %w", err)
	}

	return nil
}

// CleanupExpiredTrials finds and handles expired trial teams
func (s *Service) CleanupExpiredTrials(ctx context.Context) (int, error) {
	// Find teams with expired trials
	teams, err := s.repo.FindExpiringTrials(ctx, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired trials: %w", err)
	}

	count := 0
	for _, team := range teams {
		// Downgrade to free plan or suspend
		if team.GetTrialDaysRemaining() <= 0 {
			// Downgrade to free plan
			if err := team.DowngradePlan(PlanFree); err == nil {
				if err := s.repo.Update(ctx, team); err == nil {
					count++
				}
			}
		}
	}

	return count, nil
}

// ValidateTeamLimits checks if a team is within its plan limits
func (s *Service) ValidateTeamLimits(ctx context.Context, teamID uuid.UUID) error {
	team, err := s.repo.FindByID(ctx, teamID)
	if err != nil {
		return ErrTeamNotFound
	}

	// Check member count
	memberCount, err := s.memberRepo.CountActiveMembers(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to count members: %w", err)
	}

	if team.limits.MaxMembers != -1 && memberCount > team.limits.MaxMembers {
		return ErrMemberLimitExceeded
	}

	// Additional limit checks would go here
	// - Social accounts
	// - Scheduled posts
	// - Storage usage
	// etc.

	return nil
}

// GenerateInvitationToken generates a secure invitation token
func GenerateInvitationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// TeamSpecification for complex business rules

// ActiveTeamSpecification checks if a team can perform operations
type ActiveTeamSpecification struct{}

func (s ActiveTeamSpecification) IsSatisfiedBy(team *Team) bool {
	return team.IsActive() && !team.IsSuspended()
}

// PaidTeamSpecification checks if a team is on a paid plan
type PaidTeamSpecification struct{}

func (s PaidTeamSpecification) IsSatisfiedBy(team *Team) bool {
	return team.IsPaid()
}
