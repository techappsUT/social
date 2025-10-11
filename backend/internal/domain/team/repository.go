// path: backend/internal/domain/team/repository.go
// 🆕 NEW - Clean Architecture

package team

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for team persistence
type Repository interface {
	// Team operations
	Create(ctx context.Context, team *Team) error
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Team, error)
	FindBySlug(ctx context.Context, slug string) (*Team, error)
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Team, error)
	FindAll(ctx context.Context, offset, limit int) ([]*Team, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	Count(ctx context.Context) (int64, error)
	CountByPlan(ctx context.Context, plan Plan) (int64, error)
	CountByStatus(ctx context.Context, status Status) (int64, error)

	// Search and filtering
	FindByStatus(ctx context.Context, status Status, offset, limit int) ([]*Team, error)
	FindByPlan(ctx context.Context, plan Plan, offset, limit int) ([]*Team, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*Team, error)
	FindExpiringTrials(ctx context.Context, daysUntilExpiry int) ([]*Team, error)

	// Member-related queries
	FindByMemberID(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	GetMemberCount(ctx context.Context, teamID uuid.UUID) (int, error)

	// Cleanup operations
	FindInactiveSince(ctx context.Context, since time.Time) ([]*Team, error)
	HardDelete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error
}

// MemberRepository defines the interface for team member persistence
type MemberRepository interface {
	// CRUD operations
	AddMember(ctx context.Context, member *Member) error
	UpdateMember(ctx context.Context, member *Member) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error

	// Queries
	FindMember(ctx context.Context, teamID, userID uuid.UUID) (*Member, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Member, error)
	FindTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*Member, error)
	FindUserMemberships(ctx context.Context, userID uuid.UUID) ([]*Member, error)
	FindByRole(ctx context.Context, teamID uuid.UUID, role MemberRole) ([]*Member, error)
	FindByStatus(ctx context.Context, teamID uuid.UUID, status MemberStatus) ([]*Member, error)

	// Counts
	CountMembers(ctx context.Context, teamID uuid.UUID) (int, error)
	CountActiveMembers(ctx context.Context, teamID uuid.UUID) (int, error)
	CountByRole(ctx context.Context, teamID uuid.UUID, role MemberRole) (int, error)

	// Existence checks
	IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	IsOwner(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	HasPermission(ctx context.Context, teamID, userID uuid.UUID, permission string) (bool, error)

	// Bulk operations
	RemoveAllMembers(ctx context.Context, teamID uuid.UUID) error
	FindPendingInvitations(ctx context.Context, teamID uuid.UUID) ([]*Member, error)

	// Cleanup
	RemoveExpiredInvitations(ctx context.Context, expiryDuration time.Duration) (int, error)
}

// InvitationRepository handles team invitations (optional, can be part of MemberRepository)
type InvitationRepository interface {
	CreateInvitation(ctx context.Context, invitation *Invitation) error
	FindInvitation(ctx context.Context, id uuid.UUID) (*Invitation, error)
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	FindTeamInvitations(ctx context.Context, teamID uuid.UUID) ([]*Invitation, error)
	FindUserInvitations(ctx context.Context, email string) ([]*Invitation, error)
	UpdateInvitation(ctx context.Context, invitation *Invitation) error
	DeleteInvitation(ctx context.Context, id uuid.UUID) error
	DeleteExpiredInvitations(ctx context.Context) (int, error)
}

// Invitation represents an invitation to join a team
type Invitation struct {
	ID         uuid.UUID
	TeamID     uuid.UUID
	Email      string
	Role       MemberRole
	Token      string
	InvitedBy  uuid.UUID
	Status     InvitationStatus
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AcceptedAt *time.Time
}

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusCanceled InvitationStatus = "canceled"
)

// TeamStatistics holds aggregated team statistics
type TeamStatistics struct {
	TotalTeams            int64
	ActiveTeams           int64
	TrialTeams            int64
	PaidTeams             int64
	SuspendedTeams        int64
	TotalMembers          int64
	AverageTeamSize       float64
	TeamsByPlan           map[Plan]int64
	TeamsCreatedToday     int64
	TeamsCreatedThisWeek  int64
	TeamsCreatedThisMonth int64
}

// AdvancedRepository extends Repository with complex operations
type AdvancedRepository interface {
	Repository

	// Analytics and reporting
	GetStatistics(ctx context.Context) (*TeamStatistics, error)
	FindTeamsNearLimits(ctx context.Context) ([]*Team, error)
	FindOverageTeams(ctx context.Context) ([]*Team, error)

	// Batch operations
	BulkUpdatePlan(ctx context.Context, teamIDs []uuid.UUID, plan Plan) error
	BulkSuspend(ctx context.Context, teamIDs []uuid.UUID, reason string) error

	// Advanced queries
	FindTeamsWithExpiredTrials(ctx context.Context) ([]*Team, error)
	FindTeamsForBilling(ctx context.Context, billingDate time.Time) ([]*Team, error)
}

// CacheRepository defines caching operations for teams
type CacheRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetBySlug(ctx context.Context, slug string) (*Team, error)
	Set(ctx context.Context, team *Team, ttl time.Duration) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteBySlug(ctx context.Context, slug string) error
	InvalidateUserTeams(ctx context.Context, userID uuid.UUID) error
}

// TransactionalRepository provides transactional operations
type TransactionalRepository interface {
	Repository

	// Transaction management
	BeginTransaction(ctx context.Context) (Transaction, error)

	// Transactional operations
	CreateTeamWithOwner(ctx context.Context, team *Team, ownerID uuid.UUID) error
	TransferOwnership(ctx context.Context, teamID, fromUserID, toUserID uuid.UUID) error
	DeleteTeamAndMembers(ctx context.Context, teamID uuid.UUID) error
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
}

// QueryOptions for advanced filtering
type QueryOptions struct {
	Offset         int
	Limit          int
	SortBy         string
	SortOrder      string // "asc" or "desc"
	IncludeDeleted bool
}

// TeamFilter for complex team queries
type TeamFilter struct {
	Query         string
	OwnerID       *uuid.UUID
	Status        *Status
	Plan          *Plan
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	HasMembers    *bool
	MinMembers    *int
	MaxMembers    *int
}
