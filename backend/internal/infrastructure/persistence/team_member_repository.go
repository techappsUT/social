// ============================================================================
// FILE: backend/internal/infrastructure/persistence/team_member_repository.go
// FIXED VERSION - Corrected sql.NullTime to time.Time conversion
// ============================================================================
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type TeamMemberRepository struct {
	database *sql.DB
	queries  *db.Queries
}

func NewTeamMemberRepository(database *sql.DB) team.MemberRepository {
	return &TeamMemberRepository{
		database: database,
		queries:  db.New(database),
	}
}

func (r *TeamMemberRepository) AddMember(ctx context.Context, member *team.Member) error {
	// Get role ID by name
	dbRole, err := r.queries.GetRoleByName(ctx, string(member.Role()))
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	params := db.AddTeamMemberParams{
		TeamID: member.TeamID(),
		UserID: member.UserID(),
		RoleID: dbRole.ID,
	}

	_, err = r.queries.AddTeamMember(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return team.ErrUserAlreadyMember
			}
		}
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

func (r *TeamMemberRepository) UpdateMember(ctx context.Context, member *team.Member) error {
	// Get role ID by name
	dbRole, err := r.queries.GetRoleByName(ctx, string(member.Role()))
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	params := db.UpdateTeamMemberRoleParams{
		TeamID: member.TeamID(),
		RoleID: dbRole.ID,
		UserID: member.UserID(),
	}

	_, err = r.queries.UpdateTeamMemberRole(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	return nil
}

func (r *TeamMemberRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	params := db.RemoveTeamMemberParams{
		TeamID: teamID,
		UserID: userID,
	}

	err := r.queries.RemoveTeamMember(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

func (r *TeamMemberRepository) FindMember(ctx context.Context, teamID, userID uuid.UUID) (*team.Member, error) {
	params := db.GetTeamMemberByUserIDParams{
		TeamID: teamID,
		UserID: userID,
	}

	dbMember, err := r.queries.GetTeamMemberByUserID(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, team.ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to find member: %w", err)
	}

	// FIXED: Convert sql.NullTime to time.Time and *time.Time
	var createdAt time.Time
	if dbMember.CreatedAt.Valid {
		createdAt = dbMember.CreatedAt.Time
	} else {
		createdAt = time.Now().UTC()
	}

	var joinedAt *time.Time
	if dbMember.CreatedAt.Valid {
		t := dbMember.CreatedAt.Time
		joinedAt = &t
	}

	// Reconstruct domain member
	role := team.MemberRole(dbMember.RoleName)
	member := team.ReconstructMember(
		dbMember.ID,
		dbMember.TeamID,
		dbMember.UserID,
		role,
		team.MemberStatusActive, // Assuming active if found
		uuid.Nil,                // invitedBy - not in current schema
		createdAt,               // FIXED: Now using time.Time
		joinedAt,                // FIXED: Now using *time.Time
		nil,                     // leftAt
	)

	return member, nil
}

func (r *TeamMemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*team.Member, error) {
	// Not directly supported by current queries
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) FindTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*team.Member, error) {
	dbMembers, err := r.queries.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team members: %w", err)
	}

	members := make([]*team.Member, 0, len(dbMembers))
	for _, dbMember := range dbMembers {
		// FIXED: Convert sql.NullTime to time.Time and *time.Time
		var createdAt time.Time
		if dbMember.CreatedAt.Valid {
			createdAt = dbMember.CreatedAt.Time
		} else {
			createdAt = time.Now().UTC()
		}

		var joinedAt *time.Time
		if dbMember.CreatedAt.Valid {
			t := dbMember.CreatedAt.Time
			joinedAt = &t
		}

		role := team.MemberRole(dbMember.RoleName)
		member := team.ReconstructMember(
			dbMember.ID,
			dbMember.TeamID,
			dbMember.UserID,
			role,
			team.MemberStatusActive,
			uuid.Nil,
			createdAt, // FIXED: Now using time.Time
			joinedAt,  // FIXED: Now using *time.Time
			nil,
		)
		members = append(members, member)
	}

	return members, nil
}

func (r *TeamMemberRepository) FindUserMemberships(ctx context.Context, userID uuid.UUID) ([]*team.Member, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) FindByRole(ctx context.Context, teamID uuid.UUID, role team.MemberRole) ([]*team.Member, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) FindByStatus(ctx context.Context, teamID uuid.UUID, status team.MemberStatus) ([]*team.Member, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) CountMembers(ctx context.Context, teamID uuid.UUID) (int, error) {
	count, err := r.queries.CountTeamMembers(ctx, teamID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *TeamMemberRepository) CountActiveMembers(ctx context.Context, teamID uuid.UUID) (int, error) {
	return r.CountMembers(ctx, teamID)
}

func (r *TeamMemberRepository) CountByRole(ctx context.Context, teamID uuid.UUID, role team.MemberRole) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) IsMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	params := db.ExistsTeamMemberParams{
		TeamID: teamID,
		UserID: userID,
	}

	exists, err := r.queries.ExistsTeamMember(ctx, params)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TeamMemberRepository) IsOwner(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	member, err := r.FindMember(ctx, teamID, userID)
	if err != nil {
		return false, nil // Not a member = not an owner
	}

	return member.Role() == team.MemberRoleOwner, nil
}

func (r *TeamMemberRepository) HasPermission(ctx context.Context, teamID, userID uuid.UUID, permission string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) RemoveAllMembers(ctx context.Context, teamID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) FindPendingInvitations(ctx context.Context, teamID uuid.UUID) ([]*team.Member, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamMemberRepository) RemoveExpiredInvitations(ctx context.Context, expiryDuration time.Duration) (int, error) {
	return 0, fmt.Errorf("not implemented")
}
