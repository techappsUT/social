// path: backend/internal/infrastructure/persistence/team_repository.go
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"
	"github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type TeamRepository struct {
	database *sql.DB
	queries  *db.Queries
}

func NewTeamRepository(database *sql.DB) team.Repository {
	return &TeamRepository{
		database: database,
		queries:  db.New(database),
	}
}

func (r *TeamRepository) Create(ctx context.Context, t *team.Team) error {
	settingsJSON, err := json.Marshal(t.Settings())
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	params := db.CreateTeamParams{
		Name:      t.Name(),
		Slug:      t.Slug(),
		AvatarUrl: sql.NullString{Valid: false},
		Settings:  pqtype.NullRawMessage{RawMessage: settingsJSON, Valid: true},
		CreatedBy: uuid.NullUUID{UUID: t.OwnerID(), Valid: true},
	}

	_, err = r.queries.CreateTeam(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return team.ErrTeamAlreadyExists
			}
		}
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

func (r *TeamRepository) Update(ctx context.Context, t *team.Team) error {
	settingsJSON, err := json.Marshal(t.Settings())
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	params := db.UpdateTeamParams{
		ID:        t.ID(),
		Name:      sql.NullString{String: t.Name(), Valid: true},
		Slug:      sql.NullString{String: t.Slug(), Valid: true},
		AvatarUrl: sql.NullString{Valid: false},
		Settings:  pqtype.NullRawMessage{RawMessage: settingsJSON, Valid: true},
	}

	_, err = r.queries.UpdateTeam(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return team.ErrTeamNotFound
		}
		return fmt.Errorf("failed to update team: %w", err)
	}

	return nil
}

func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteTeam(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}

func (r *TeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*team.Team, error) {
	dbTeam, err := r.queries.GetTeamByID(ctx, id) // ✅ FIXED: Use GetTeamByID
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, team.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	return r.mapToTeamEntity(dbTeam)
}

func (r *TeamRepository) FindBySlug(ctx context.Context, slug string) (*team.Team, error) {
	dbTeam, err := r.queries.GetTeamBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, team.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to find team by slug: %w", err)
	}

	return r.mapToTeamEntity(dbTeam)
}

func (r *TeamRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*team.Team, error) {
	params := db.ListTeamsByUserParams{
		UserID: ownerID,
	}

	dbTeams, err := r.queries.ListTeamsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by owner: %w", err)
	}

	teams := make([]*team.Team, 0, len(dbTeams))
	for _, dbTeam := range dbTeams {
		t, err := r.mapToTeamEntity(dbTeam)
		if err != nil {
			continue
		}
		teams = append(teams, t)
	}

	return teams, nil
}

func (r *TeamRepository) FindAll(ctx context.Context, offset, limit int) ([]*team.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	_, err := r.queries.GetTeamBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *TeamRepository) Count(ctx context.Context) (int64, error) {
	// Not implemented in current queries
	return 0, fmt.Errorf("not implemented")
}

func (r *TeamRepository) CountByPlan(ctx context.Context, plan team.Plan) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *TeamRepository) CountByStatus(ctx context.Context, status team.Status) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *TeamRepository) FindByStatus(ctx context.Context, status team.Status, offset, limit int) ([]*team.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamRepository) FindByPlan(ctx context.Context, plan team.Plan, offset, limit int) ([]*team.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamRepository) Search(ctx context.Context, query string, offset, limit int) ([]*team.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamRepository) FindExpiringTrials(ctx context.Context, daysUntilExpiry int) ([]*team.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamRepository) FindByMemberID(ctx context.Context, userID uuid.UUID) ([]*team.Team, error) {
	params := db.ListTeamsByUserParams{
		UserID: userID,
	}

	dbTeams, err := r.queries.ListTeamsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by member: %w", err)
	}

	teams := make([]*team.Team, 0, len(dbTeams))
	for _, dbTeam := range dbTeams {
		t, err := r.mapToTeamEntity(dbTeam)
		if err != nil {
			continue
		}
		teams = append(teams, t)
	}

	return teams, nil
}

func (r *TeamRepository) GetMemberCount(ctx context.Context, teamID uuid.UUID) (int, error) {
	count, err := r.queries.CountTeamMembers(ctx, teamID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *TeamRepository) FindInactiveSince(ctx context.Context, since time.Time) ([]*team.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *TeamRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *TeamRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *TeamRepository) mapToTeamEntity(dbTeam db.Team) (*team.Team, error) {
	var settings team.TeamSettings
	if dbTeam.Settings.Valid && len(dbTeam.Settings.RawMessage) > 0 {
		if err := json.Unmarshal(dbTeam.Settings.RawMessage, &settings); err != nil {
			// Use default settings if unmarshal fails
			settings = team.TeamSettings{
				Timezone:            "UTC",
				DefaultPostTime:     "10:00",
				EnableNotifications: true,
				EnableAnalytics:     true,
				RequireApproval:     false,
				AutoSchedule:        false,
				Language:            "en",
				DateFormat:          "MM/DD/YYYY",
			}
		}
	} else {
		// Default settings
		settings = team.TeamSettings{
			Timezone:            "UTC",
			DefaultPostTime:     "10:00",
			EnableNotifications: true,
			EnableAnalytics:     true,
			RequireApproval:     false,
			AutoSchedule:        false,
			Language:            "en",
			DateFormat:          "MM/DD/YYYY",
		}
	}

	ownerID := uuid.Nil
	if dbTeam.CreatedBy.Valid {
		ownerID = dbTeam.CreatedBy.UUID
	}

	// Use NewTeam with correct signature
	t, err := team.NewTeam(dbTeam.Name, dbTeam.Slug, "", ownerID)
	if err != nil {
		return nil, err
	}

	return t, nil
}
