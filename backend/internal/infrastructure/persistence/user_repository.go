// path: backend/internal/infrastructure/persistence/user_repository.go
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq" // PostgreSQL driver for database/sql
	"github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// UserRepository implements user.Repository using PostgreSQL
type UserRepository struct {
	queries *db.Queries
	db      *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(database *sql.DB) user.Repository {
	return &UserRepository{
		queries: db.New(database),
		db:      database,
	}
}

// ============================================================================
// CORE CRUD OPERATIONS
// ============================================================================

// Create persists a new user
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	params := db.CreateUserParams{
		Email: u.Email(),
		EmailVerified: sql.NullBool{
			Bool:  u.IsEmailVerified(),
			Valid: true,
		},
		PasswordHash: sql.NullString{
			String: u.PasswordHash(),
			Valid:  true,
		},
		FullName: sql.NullString{
			String: fmt.Sprintf("%s %s", u.FirstName(), u.LastName()),
			Valid:  true,
		},
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return user.ErrEmailAlreadyExists
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Update the user ID from database
	u.SetID(dbUser.ID)
	return nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users 
		SET 
			email = $2,
			full_name = $3,
			updated_at = $4,
			last_login_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	fullName := fmt.Sprintf("%s %s", u.FirstName(), u.LastName())
	result, err := r.db.ExecContext(
		ctx,
		query,
		u.ID(),
		u.Email(),
		fullName,
		time.Now(),
		u.LastLoginAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
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

// Delete performs a soft delete on a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeleteUser(ctx, id)
}

// ============================================================================
// FINDER METHODS
// ============================================================================

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return r.mapToDomain(dbUser), nil
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return r.mapToDomain(dbUser), nil
}

// FindByUsername retrieves a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	// TODO: Implement when username column is added to database
	return nil, user.ErrUserNotFound
}

// FindByEmailOrUsername retrieves a user by email or username
func (r *UserRepository) FindByEmailOrUsername(ctx context.Context, identifier string) (*user.User, error) {
	// Try email first
	u, err := r.FindByEmail(ctx, identifier)
	if err == nil {
		return u, nil
	}

	// Try username
	return r.FindByUsername(ctx, identifier)
}

// FindAll retrieves all users with pagination
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

// FindByRole retrieves users by role
func (r *UserRepository) FindByRole(ctx context.Context, role user.Role, offset, limit int) ([]*user.User, error) {
	// TODO: Add role column to database, then implement
	return []*user.User{}, nil
}

// FindByStatus retrieves users by status
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

// FindByTeamID retrieves users by team ID
func (r *UserRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT u.* FROM users u
		JOIN team_memberships tm ON u.id = tm.user_id
		WHERE tm.team_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

// ============================================================================
// ADVANCED FINDER METHODS
// ============================================================================

// Search searches users by query string
func (r *UserRepository) Search(ctx context.Context, searchQuery string, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT * FROM users 
		WHERE (
			email ILIKE $1 OR 
			full_name ILIKE $1
		) AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + searchQuery + "%"
	rows, err := r.db.QueryContext(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

// FindInactiveSince finds users inactive since a given time
func (r *UserRepository) FindInactiveSince(ctx context.Context, since time.Time, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT * FROM users 
		WHERE (last_login_at < $1 OR last_login_at IS NULL) 
		AND deleted_at IS NULL
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

// FindRecentlyCreated finds recently created users
func (r *UserRepository) FindRecentlyCreated(ctx context.Context, duration time.Duration, offset, limit int) ([]*user.User, error) {
	since := time.Now().Add(-duration)

	query := `
		SELECT * FROM users
		WHERE created_at > $1 AND deleted_at IS NULL
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

// FindUnverifiedOlderThan finds unverified users older than specified duration
func (r *UserRepository) FindUnverifiedOlderThan(ctx context.Context, duration time.Duration) ([]*user.User, error) {
	cutoffTime := time.Now().Add(-duration)

	query := `
		SELECT * FROM users 
		WHERE (email_verified = false OR email_verified IS NULL)
		AND created_at < $1
		AND deleted_at IS NULL
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
// EXISTENCE CHECKS
// ============================================================================

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsByUsername checks if a user exists with the given username
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	// TODO: Implement when username column is added
	return false, nil
}

// ============================================================================
// COUNT METHODS
// ============================================================================

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// CountByRole returns count by role
func (r *UserRepository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	// TODO: Add role column to database
	return 0, nil
}

// CountByStatus returns count by status
func (r *UserRepository) CountByStatus(ctx context.Context, status user.Status) (int64, error) {
	var count int64
	isActive := status == user.StatusActive
	query := `SELECT COUNT(*) FROM users WHERE is_active = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query, isActive).Scan(&count)
	return count, err
}

// ============================================================================
// UPDATE METHODS
// ============================================================================

// UpdateLastLogin updates only the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	query := `UPDATE users SET last_login_at = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, lastLoginAt, time.Now())
	return err
}

// UpdatePassword updates only the password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, passwordHash, time.Now())
	return err
}

// UpdateEmailVerificationStatus updates the email verification status
func (r *UserRepository) UpdateEmailVerificationStatus(ctx context.Context, id uuid.UUID, verified bool) error {
	query := `UPDATE users SET email_verified = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, verified, time.Now())
	return err
}

// UpdateStatus updates only the user status
func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) error {
	isActive := status == user.StatusActive
	query := `UPDATE users SET is_active = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, isActive, time.Now())
	return err
}

// UpdateRole updates only the user role
func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role user.Role) error {
	// TODO: Add role column to database
	return nil
}

// BulkUpdateStatus updates status for multiple users
func (r *UserRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status user.Status) error {
	if len(ids) == 0 {
		return nil
	}

	isActive := status == user.StatusActive

	// Build placeholders for IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+2)
	args[0] = isActive
	args[1] = time.Now()

	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args[i+2] = id
	}

	query := fmt.Sprintf(
		"UPDATE users SET is_active = $1, updated_at = $2 WHERE id IN (%s) AND deleted_at IS NULL",
		strings.Join(placeholders, ","),
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// ============================================================================
// RESTORE AND DELETE METHODS
// ============================================================================

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NULL, updated_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

// HardDelete permanently deletes a user (use with extreme caution)
func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// scanUsers scans multiple user rows
func (r *UserRepository) scanUsers(rows *sql.Rows) ([]*user.User, error) {
	var users []*user.User
	for rows.Next() {
		dbUser := db.User{}
		err := rows.Scan(
			&dbUser.ID,
			&dbUser.Email,
			&dbUser.PasswordHash,
			&dbUser.FullName,
			&dbUser.AvatarUrl,
			&dbUser.IsActive,
			&dbUser.EmailVerified,
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

// mapToDomain maps database user to domain user
func (r *UserRepository) mapToDomain(dbUser db.User) *user.User {
	// Extract first and last name from full_name
	firstName := ""
	lastName := ""
	if dbUser.FullName.Valid && dbUser.FullName.String != "" {
		parts := strings.SplitN(dbUser.FullName.String, " ", 2)
		if len(parts) > 0 {
			firstName = parts[0]
		}
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	// Handle password hash
	passwordHash := ""
	if dbUser.PasswordHash.Valid {
		passwordHash = dbUser.PasswordHash.String
	}

	// Determine status
	status := user.StatusActive
	if dbUser.IsActive.Valid && !dbUser.IsActive.Bool {
		status = user.StatusInactive
	}

	// Default role
	role := user.RoleUser

	// Handle email verified
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

	// Avatar URL
	avatarURL := ""
	if dbUser.AvatarUrl.Valid {
		avatarURL = dbUser.AvatarUrl.String
	}

	return user.Reconstruct(
		dbUser.ID,
		dbUser.Email,
		"", // username - TODO: Add to DB
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
