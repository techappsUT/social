// ===========================================================================
// FILE: backend/internal/infrastructure/persistence/team_repository.go
// CRITICAL FIX: Use domain team ID, don't let database generate it
// ===========================================================================

package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// ✅ FIXED: Now manually inserts with the domain team ID
func (r *TeamRepository) Create(ctx context.Context, t *team.Team) error {
	settingsJSON, err := json.Marshal(t.Settings())
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// ✅ FIX: Use raw SQL with the team's domain ID
	query := `
		INSERT INTO teams (
			id,          -- ✅ ADDED: Use domain ID
			name,
			slug,
			avatar_url,
			settings,
			created_by,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`

	avatarURL := sql.NullString{Valid: false}
	if t.AvatarURL() != "" {
		avatarURL = sql.NullString{String: t.AvatarURL(), Valid: true}
	}

	createdBy := uuid.NullUUID{Valid: false}
	if t.OwnerID() != uuid.Nil {
		createdBy = uuid.NullUUID{UUID: t.OwnerID(), Valid: true}
	}

	// ✅ FIX: Execute with the domain team ID as the first parameter
	_, err = r.database.ExecContext(ctx, query,
		t.ID(), // ✅ Use domain ID
		t.Name(),
		t.Slug(),
		avatarURL,
		settingsJSON,
		createdBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

// Rest of the repository methods stay the same...

func (r *TeamRepository) Update(ctx context.Context, t *team.Team) error {
	settingsJSON, err := json.Marshal(t.Settings())
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	params := db.UpdateTeamParams{
		ID:       t.ID(),
		Name:     sql.NullString{String: t.Name(), Valid: true},
		Slug:     sql.NullString{String: t.Slug(), Valid: true},
		Settings: pqtype.NullRawMessage{Valid: true, RawMessage: settingsJSON},
	}

	if t.AvatarURL() != "" {
		params.AvatarUrl = sql.NullString{String: t.AvatarURL(), Valid: true}
	}

	_, err = r.queries.UpdateTeam(ctx, params)
	return err
}

func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeleteTeam(ctx, id)
}

func (r *TeamRepository) FindByID(ctx context.Context, id uuid.UUID) (*team.Team, error) {
	dbTeam, err := r.queries.GetTeamByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, err
	}

	return r.mapToTeamEntity(dbTeam)
}

func (r *TeamRepository) FindBySlug(ctx context.Context, slug string) (*team.Team, error) {
	dbTeam, err := r.queries.GetTeamBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, err
	}

	return r.mapToTeamEntity(dbTeam)
}

func (r *TeamRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	_, err := r.queries.GetTeamBySlug(ctx, slug)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *TeamRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*team.Team, error) {
	dbTeams, err := r.queries.ListTeamsByUser(ctx, userID)
	if err != nil {
		return nil, err
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

// ============================================================================
// MISSING REQUIRED INTERFACE METHODS - Add these
// ============================================================================

func (r *TeamRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM teams WHERE deleted_at IS NULL`
	var count int64
	err := r.database.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *TeamRepository) CountByPlan(ctx context.Context, plan team.Plan) (int64, error) {
	// Plan is not stored in DB yet, return 0
	return 0, nil
}

func (r *TeamRepository) CountByStatus(ctx context.Context, status team.Status) (int64, error) {
	// Status is not stored in DB yet, return count of active teams if status is active
	if status == team.StatusActive {
		return r.Count(ctx)
	}
	return 0, nil
}

func (r *TeamRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*team.Team, error) {
	query := `
		SELECT id, name, slug, avatar_url, settings, is_active, created_by, created_at, updated_at, deleted_at
		FROM teams
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.database.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTeamsFromRows(rows)
}

func (r *TeamRepository) FindAll(ctx context.Context, offset, limit int) ([]*team.Team, error) {
	query := `
		SELECT id, name, slug, avatar_url, settings, is_active, created_by, created_at, updated_at, deleted_at
		FROM teams
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.database.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTeamsFromRows(rows)
}

func (r *TeamRepository) FindByStatus(ctx context.Context, status team.Status, offset, limit int) ([]*team.Team, error) {
	// Status not stored yet, return all active teams
	if status == team.StatusActive {
		return r.FindAll(ctx, offset, limit)
	}
	return []*team.Team{}, nil
}

func (r *TeamRepository) FindByPlan(ctx context.Context, plan team.Plan, offset, limit int) ([]*team.Team, error) {
	// Plan not stored yet
	return []*team.Team{}, nil
}

func (r *TeamRepository) Search(ctx context.Context, query string, offset, limit int) ([]*team.Team, error) {
	sqlQuery := `
		SELECT id, name, slug, avatar_url, settings, is_active, created_by, created_at, updated_at, deleted_at
		FROM teams
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR slug ILIKE $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := r.database.QueryContext(ctx, sqlQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTeamsFromRows(rows)
}

func (r *TeamRepository) FindExpiringTrials(ctx context.Context, daysUntilExpiry int) ([]*team.Team, error) {
	// Trial expiry not implemented yet
	return []*team.Team{}, nil
}

func (r *TeamRepository) FindByMemberID(ctx context.Context, userID uuid.UUID) ([]*team.Team, error) {
	// This is the same as ListByUserID
	return r.ListByUserID(ctx, userID)
}

// ============================================================================
// HELPER METHODS
// ============================================================================

func (r *TeamRepository) scanTeamsFromRows(rows *sql.Rows) ([]*team.Team, error) {
	teams := []*team.Team{}

	for rows.Next() {
		var dbTeam db.Team
		err := rows.Scan(
			&dbTeam.ID,
			&dbTeam.Name,
			&dbTeam.Slug,
			&dbTeam.AvatarUrl,
			&dbTeam.Settings,
			&dbTeam.IsActive,
			&dbTeam.CreatedBy,
			&dbTeam.CreatedAt,
			&dbTeam.UpdatedAt,
			&dbTeam.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		t, err := r.mapToTeamEntity(dbTeam)
		if err != nil {
			continue // Skip invalid teams
		}
		teams = append(teams, t)
	}

	return teams, rows.Err()
}

func (r *TeamRepository) mapToTeamEntity(dbTeam db.Team) (*team.Team, error) {
	var settings team.TeamSettings
	if dbTeam.Settings.Valid && len(dbTeam.Settings.RawMessage) > 0 {
		if err := json.Unmarshal(dbTeam.Settings.RawMessage, &settings); err != nil {
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

	return team.ReconstructTeam(
		dbTeam.ID, // Use database ID
		dbTeam.Name,
		dbTeam.Slug,
		"", // description not stored yet
		ownerID,
		settings,
	)
}
