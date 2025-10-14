// path: backend/internal/infrastructure/persistence/user_repository.go
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

type UserRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewUserRepository(database *sql.DB) user.Repository {
	return &UserRepository{
		db:      database,
		queries: db.New(database),
	}
}

// ============================================================================
// CREATE
// ============================================================================

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	params := db.CreateUserParams{
		Email:         u.Email(),
		EmailVerified: sql.NullBool{Bool: u.IsEmailVerified(), Valid: true},
		PasswordHash:  sql.NullString{String: u.PasswordHash(), Valid: true},
		Username:      u.Username(),
		FirstName:     u.FirstName(),
		LastName:      u.LastName(),
		AvatarUrl:     sql.NullString{String: u.AvatarURL(), Valid: u.AvatarURL() != ""},
		Timezone:      sql.NullString{String: "UTC", Valid: true},
	}

	_, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return user.ErrEmailAlreadyExists
				}
				if pqErr.Constraint == "users_username_key" || pqErr.Constraint == "users_username_unique" {
					return user.ErrUsernameAlreadyExists
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return r.mapToDomain(dbUser), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return r.mapToDomain(dbUser), nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	dbUser, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	return r.mapToDomain(dbUser), nil
}

func (r *UserRepository) FindByEmailOrUsername(ctx context.Context, identifier string) (*user.User, error) {
	u, err := r.FindByEmail(ctx, identifier)
	if err == nil {
		return u, nil
	}
	return r.FindByUsername(ctx, identifier)
}

func (r *UserRepository) FindAll(ctx context.Context, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT * FROM users 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) FindByRole(ctx context.Context, role user.Role, offset, limit int) ([]*user.User, error) {
	// TODO: Add role column to users table, then implement
	return []*user.User{}, nil
}

func (r *UserRepository) FindByStatus(ctx context.Context, status user.Status, offset, limit int) ([]*user.User, error) {
	isActive := status == user.StatusActive
	query := `
		SELECT * FROM users 
		WHERE is_active = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, isActive, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*user.User, error) {
	// TODO: Implement team membership query
	return []*user.User{}, nil
}

func (r *UserRepository) Search(ctx context.Context, query string, offset, limit int) ([]*user.User, error) {
	searchQuery := `
		SELECT * FROM users 
		WHERE (
			email ILIKE $1 OR 
			username ILIKE $1 OR 
			first_name ILIKE $1 OR 
			last_name ILIKE $1
		) AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) FindInactiveSince(ctx context.Context, since time.Time, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT * FROM users 
		WHERE (last_login_at IS NULL OR last_login_at < $1)
		AND deleted_at IS NULL
		ORDER BY last_login_at ASC NULLS FIRST
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) FindRecentlyCreated(ctx context.Context, duration time.Duration, offset, limit int) ([]*user.User, error) {
	since := time.Now().Add(-duration)
	query := `
		SELECT * FROM users 
		WHERE created_at >= $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

func (r *UserRepository) FindUnverifiedOlderThan(ctx context.Context, duration time.Duration) ([]*user.User, error) {
	cutoffTime := time.Now().Add(-duration)
	query := `
		SELECT * FROM users 
		WHERE (email_verified = false OR email_verified IS NULL)
		AND created_at < $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	params := db.UpdateUserProfileParams{
		ID:        u.ID(),
		Username:  sql.NullString{String: u.Username(), Valid: true},
		FirstName: sql.NullString{String: u.FirstName(), Valid: true},
		LastName:  sql.NullString{String: u.LastName(), Valid: true},
		AvatarUrl: sql.NullString{String: u.AvatarURL(), Valid: u.AvatarURL() != ""},
		Timezone:  sql.NullString{String: "UTC", Valid: true},
	}

	_, err := r.queries.UpdateUserProfile(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return user.ErrUserNotFound
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return r.queries.UpdateUserLastLogin(ctx, id)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	params := db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
	}
	return r.queries.UpdateUserPassword(ctx, params)
}

func (r *UserRepository) UpdateEmailVerificationStatus(ctx context.Context, id uuid.UUID, verified bool) error {
	if verified {
		return r.queries.MarkUserEmailVerified(ctx, id)
	}
	// TODO: Add query to mark as unverified if needed
	return nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) error {
	isActive := status == user.StatusActive
	query := `UPDATE users SET is_active = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, isActive)
	return err
}

func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role user.Role) error {
	// TODO: Add role column to users table
	return fmt.Errorf("role column not yet implemented in database")
}

func (r *UserRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status user.Status) error {
	if len(ids) == 0 {
		return nil
	}

	isActive := status == user.StatusActive
	query := `UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, isActive, pq.Array(ids))
	return err
}

// ============================================================================
// DELETE
// ============================================================================

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeleteUser(ctx, id)
}

func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ============================================================================
// EXISTENCE CHECKS
// ============================================================================

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.CheckEmailExists(ctx, email)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.queries.CheckUsernameExists(ctx, username)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ============================================================================
// COUNT METHODS
// ============================================================================

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *UserRepository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	// TODO: Implement when role column is added
	return 0, nil
}

func (r *UserRepository) CountByStatus(ctx context.Context, status user.Status) (int64, error) {
	var count int64
	isActive := status == user.StatusActive
	query := `SELECT COUNT(*) FROM users WHERE is_active = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query, isActive).Scan(&count)
	return count, err
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// ExistsByID checks if a user exists with the given ID
func (r *UserRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) scanUsers(rows *sql.Rows) ([]*user.User, error) {
	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		err := rows.Scan(
			&dbUser.ID,
			&dbUser.Email,
			&dbUser.EmailVerified,
			&dbUser.PasswordHash,
			&dbUser.Username,
			&dbUser.FirstName,
			&dbUser.LastName,
			&dbUser.FullName,
			&dbUser.AvatarUrl,
			&dbUser.Timezone,
			&dbUser.Locale,
			&dbUser.IsActive,
			&dbUser.LastLoginAt,
			&dbUser.CreatedAt,
			&dbUser.UpdatedAt,
			&dbUser.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}
	return users, nil
}

// MarkEmailVerified marks a user's email as verified
func (r *UserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users 
		SET email_verified = TRUE, updated_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}
func (r *UserRepository) mapToDomain(dbUser db.User) *user.User {
	// These are plain strings (NOT NULL in DB)
	username := dbUser.Username
	firstName := dbUser.FirstName
	lastName := dbUser.LastName

	// Handle nullable fields
	passwordHash := ""
	if dbUser.PasswordHash.Valid {
		passwordHash = dbUser.PasswordHash.String
	}

	avatarURL := ""
	if dbUser.AvatarUrl.Valid {
		avatarURL = dbUser.AvatarUrl.String
	}

	// Determine status
	status := user.StatusActive
	if dbUser.IsActive.Valid && !dbUser.IsActive.Bool {
		status = user.StatusInactive
	}

	// Default role
	role := user.RoleUser

	// Email verified
	emailVerified := false
	if dbUser.EmailVerified.Valid {
		emailVerified = dbUser.EmailVerified.Bool
	}

	// Convert timestamps
	var lastLoginAt *time.Time
	if dbUser.LastLoginAt.Valid {
		t := dbUser.LastLoginAt.Time
		lastLoginAt = &t
	}

	var deletedAt *time.Time
	if dbUser.DeletedAt.Valid {
		t := dbUser.DeletedAt.Time
		deletedAt = &t
	}

	return user.Reconstruct(
		dbUser.ID,
		dbUser.Email,
		username,
		passwordHash,
		firstName,
		lastName,
		avatarURL,
		role,
		status,
		emailVerified,
		lastLoginAt,
		dbUser.CreatedAt.Time,
		dbUser.UpdatedAt.Time,
		deletedAt,
	)
}
