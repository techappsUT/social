// path: backend/internal/domain/team/member.go
// 🆕 NEW - Clean Architecture

package team

import (
	"time"

	"github.com/google/uuid"
)

// Member represents a user's membership in a team
// This is a value object that connects User and Team entities
type Member struct {
	id        uuid.UUID
	teamID    uuid.UUID
	userID    uuid.UUID
	role      MemberRole
	status    MemberStatus
	invitedBy uuid.UUID
	invitedAt time.Time
	joinedAt  *time.Time
	leftAt    *time.Time
}

// MemberRole represents the role of a member in a team
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleEditor MemberRole = "editor"
	MemberRoleViewer MemberRole = "viewer"
)

// MemberStatus represents the status of a membership
type MemberStatus string

const (
	MemberStatusPending  MemberStatus = "pending"  // Invited but not accepted
	MemberStatusActive   MemberStatus = "active"   // Active member
	MemberStatusInactive MemberStatus = "inactive" // Deactivated
	MemberStatusLeft     MemberStatus = "left"     // Left the team
)

// NewMember creates a new team member
func NewMember(teamID, userID, invitedBy uuid.UUID, role MemberRole) (*Member, error) {
	if teamID == uuid.Nil {
		return nil, ErrInvalidTeamID
	}

	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if !isValidMemberRole(role) {
		return nil, ErrInvalidMemberRole
	}

	now := time.Now().UTC()

	return &Member{
		id:        uuid.New(),
		teamID:    teamID,
		userID:    userID,
		role:      role,
		status:    MemberStatusPending,
		invitedBy: invitedBy,
		invitedAt: now,
	}, nil
}

// ReconstructMember recreates a member from persistence
func ReconstructMember(
	id uuid.UUID,
	teamID uuid.UUID,
	userID uuid.UUID,
	role MemberRole,
	status MemberStatus,
	invitedBy uuid.UUID,
	invitedAt time.Time,
	joinedAt *time.Time,
	leftAt *time.Time,
) *Member {
	return &Member{
		id:        id,
		teamID:    teamID,
		userID:    userID,
		role:      role,
		status:    status,
		invitedBy: invitedBy,
		invitedAt: invitedAt,
		joinedAt:  joinedAt,
		leftAt:    leftAt,
	}
}

// Getters

func (m *Member) ID() uuid.UUID        { return m.id }
func (m *Member) TeamID() uuid.UUID    { return m.teamID }
func (m *Member) UserID() uuid.UUID    { return m.userID }
func (m *Member) Role() MemberRole     { return m.role }
func (m *Member) Status() MemberStatus { return m.status }
func (m *Member) InvitedBy() uuid.UUID { return m.invitedBy }
func (m *Member) InvitedAt() time.Time { return m.invitedAt }
func (m *Member) JoinedAt() *time.Time { return m.joinedAt }
func (m *Member) LeftAt() *time.Time   { return m.leftAt }

// Business Logic

// AcceptInvitation marks the member as having accepted the invitation
func (m *Member) AcceptInvitation() error {
	if m.status != MemberStatusPending {
		return ErrInvitationAlreadyAccepted
	}

	now := time.Now().UTC()
	m.status = MemberStatusActive
	m.joinedAt = &now
	return nil
}

// DeclineInvitation declines the invitation
func (m *Member) DeclineInvitation() error {
	if m.status != MemberStatusPending {
		return ErrInvitationAlreadyProcessed
	}

	now := time.Now().UTC()
	m.status = MemberStatusInactive
	m.leftAt = &now
	return nil
}

// ChangeRole updates the member's role
func (m *Member) ChangeRole(newRole MemberRole, changedBy uuid.UUID) error {
	if !isValidMemberRole(newRole) {
		return ErrInvalidMemberRole
	}

	if m.role == newRole {
		return ErrSameRole
	}

	// Cannot demote an owner unless there's another owner
	if m.role == MemberRoleOwner && newRole != MemberRoleOwner {
		// This check should be done at the service level with repository
		// to ensure there's at least one owner
		return ErrCannotDemoteOnlyOwner
	}

	m.role = newRole
	return nil
}

// Leave marks the member as having left the team
func (m *Member) Leave() error {
	if m.status == MemberStatusLeft {
		return ErrMemberAlreadyLeft
	}

	// Cannot leave if owner (must transfer ownership first)
	if m.role == MemberRoleOwner {
		return ErrOwnerCannotLeave
	}

	now := time.Now().UTC()
	m.status = MemberStatusLeft
	m.leftAt = &now
	return nil
}

// Remove forcefully removes a member from the team
func (m *Member) Remove(removedBy uuid.UUID) error {
	if m.status == MemberStatusLeft {
		return ErrMemberAlreadyLeft
	}

	// Cannot remove owner
	if m.role == MemberRoleOwner {
		return ErrCannotRemoveOwner
	}

	now := time.Now().UTC()
	m.status = MemberStatusLeft
	m.leftAt = &now
	return nil
}

// Deactivate temporarily deactivates a member
func (m *Member) Deactivate() error {
	if m.status == MemberStatusInactive {
		return ErrMemberAlreadyInactive
	}

	if m.status == MemberStatusLeft {
		return ErrMemberAlreadyLeft
	}

	m.status = MemberStatusInactive
	return nil
}

// Reactivate reactivates a deactivated member
func (m *Member) Reactivate() error {
	if m.status != MemberStatusInactive {
		return ErrMemberNotInactive
	}

	m.status = MemberStatusActive
	return nil
}

// Permission Checks

// CanManageTeam checks if the member can manage team settings
func (m *Member) CanManageTeam() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner || m.role == MemberRoleAdmin)
}

// CanManageMembers checks if the member can manage other members
func (m *Member) CanManageMembers() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner || m.role == MemberRoleAdmin)
}

// CanCreatePosts checks if the member can create posts
func (m *Member) CanCreatePosts() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner ||
			m.role == MemberRoleAdmin ||
			m.role == MemberRoleEditor)
}

// CanEditPosts checks if the member can edit posts
func (m *Member) CanEditPosts() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner ||
			m.role == MemberRoleAdmin ||
			m.role == MemberRoleEditor)
}

// CanDeletePosts checks if the member can delete posts
func (m *Member) CanDeletePosts() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner || m.role == MemberRoleAdmin)
}

// CanViewAnalytics checks if the member can view analytics
func (m *Member) CanViewAnalytics() bool {
	return m.status == MemberStatusActive // All active members can view
}

// CanManageBilling checks if the member can manage billing
func (m *Member) CanManageBilling() bool {
	return m.status == MemberStatusActive && m.role == MemberRoleOwner
}

// CanManageSocialAccounts checks if the member can manage social accounts
func (m *Member) CanManageSocialAccounts() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner || m.role == MemberRoleAdmin)
}

// CanInviteMembers checks if the member can invite new members
func (m *Member) CanInviteMembers() bool {
	return m.status == MemberStatusActive &&
		(m.role == MemberRoleOwner || m.role == MemberRoleAdmin)
}

// IsActive checks if the member is active
func (m *Member) IsActive() bool {
	return m.status == MemberStatusActive
}

// IsPending checks if the member invitation is pending
func (m *Member) IsPending() bool {
	return m.status == MemberStatusPending
}

// HasJoined checks if the member has joined the team
func (m *Member) HasJoined() bool {
	return m.joinedAt != nil && m.status == MemberStatusActive
}

// GetMembershipDuration returns how long the member has been part of the team
func (m *Member) GetMembershipDuration() time.Duration {
	if m.joinedAt == nil {
		return 0
	}

	if m.leftAt != nil {
		return m.leftAt.Sub(*m.joinedAt)
	}

	return time.Since(*m.joinedAt)
}

// Helper Functions

func isValidMemberRole(role MemberRole) bool {
	switch role {
	case MemberRoleOwner, MemberRoleAdmin, MemberRoleEditor, MemberRoleViewer:
		return true
	default:
		return false
	}
}

// Permission Matrix for reference
//
// | Permission              | Owner | Admin | Editor | Viewer |
// |------------------------|-------|-------|--------|--------|
// | Manage Team Settings   | ✅    | ✅    | ❌     | ❌     |
// | Manage Members         | ✅    | ✅    | ❌     | ❌     |
// | Create Posts           | ✅    | ✅    | ✅     | ❌     |
// | Edit Posts             | ✅    | ✅    | ✅     | ❌     |
// | Delete Posts           | ✅    | ✅    | ❌     | ❌     |
// | View Analytics         | ✅    | ✅    | ✅     | ✅     |
// | Manage Billing         | ✅    | ❌    | ❌     | ❌     |
// | Manage Social Accounts | ✅    | ✅    | ❌     | ❌     |
// | Invite Members         | ✅    | ✅    | ❌     | ❌     |
