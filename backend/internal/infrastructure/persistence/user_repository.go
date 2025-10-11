// path: backend/internal/infrastructure/persistence/user_repository.go
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
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

// Create persists a new user
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	params := db.CreateUserParams{
		Email:         u.Email(),
		EmailVerified: &[]bool{u.IsEmailVerified()}[0],
		PasswordHash:  &[]string{u.PasswordHash()}[0],
		FullName:      &[]string{fmt.Sprintf("%s %s", u.FirstName(), u.LastName())}[0],
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
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
	_, err := r.db.ExecContext(
		ctx,
		query,
		u.ID(),
		u.Email(),
		fullName,
		time.Now(),
		u.LastLoginAt(),
	)

	return err
}

// Delete performs a soft delete on a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeleteUser(ctx, id)
}

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

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		if err := rows.Scan(&dbUser); err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
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

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		if err := rows.Scan(&dbUser); err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
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

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		if err := rows.Scan(&dbUser); err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
}

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
// ADVANCED REPOSITORY METHODS (from domain/user/repository.go)
// ============================================================================

// FindByVerificationToken finds a user by email verification token
func (r *UserRepository) FindByVerificationToken(ctx context.Context, token string) (*user.User, error) {
	// TODO: Implement when tokens table is added
	return nil, user.ErrUserNotFound
}

// FindByPasswordResetToken finds a user by password reset token
func (r *UserRepository) FindByPasswordResetToken(ctx context.Context, token string) (*user.User, error) {
	// TODO: Implement when tokens table is added
	return nil, user.ErrUserNotFound
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, hashedPassword, time.Now())
	return err
}

// UpdateEmailVerified marks email as verified
func (r *UserRepository) UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	query := `UPDATE users SET email_verified = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, verified, time.Now())
	return err
}

// UpdateLastLogin updates last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, timestamp time.Time) error {
	query := `UPDATE users SET last_login_at = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, timestamp, time.Now())
	return err
}

// BulkUpdateStatus updates status for multiple users
func (r *UserRepository) BulkUpdateStatus(ctx context.Context, userIDs []uuid.UUID, status user.Status) error {
	if len(userIDs) == 0 {
		return nil
	}

	isActive := status == user.StatusActive

	// Build placeholders for IN clause
	query := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = ANY($3::uuid[]) AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, isActive, time.Now(), userIDs)
	return err
}

// Search performs text search on users
func (r *UserRepository) Search(ctx context.Context, query string, offset, limit int) ([]*user.User, error) {
	searchQuery := `
		SELECT * FROM users 
		WHERE (
			email ILIKE $1 OR
			full_name ILIKE $1
		) AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		if err := rows.Scan(&dbUser); err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
}

// GetTeamMembers gets all users in a specific team
func (r *UserRepository) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*user.User, error) {
	return r.FindByTeamID(ctx, teamID, 0, 1000) // Get up to 1000 members
}

// GetActiveUsersCount returns count of active users
func (r *UserRepository) GetActiveUsersCount(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE last_login_at > $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query, since).Scan(&count)
	return count, err
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

// mapToDomain maps database user to domain user
func (r *UserRepository) mapToDomain(dbUser db.User) *user.User {
	// Extract first and last name from full_name
	firstName := ""
	lastName := ""
	if dbUser.FullName != nil {
		// Simple split - in production use proper parsing
		firstName = *dbUser.FullName
	}

	// Default password hash if nil
	passwordHash := ""
	if dbUser.PasswordHash != nil {
		passwordHash = *dbUser.PasswordHash
	}

	// Map status
	status := user.StatusActive
	if dbUser.IsActive != nil && !*dbUser.IsActive {
		status = user.StatusInactive
	}

	// Default role
	role := user.RoleUser

	// Handle email verified
	emailVerified := false
	if dbUser.EmailVerified != nil {
		emailVerified = *dbUser.EmailVerified
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
		"", // username - TODO: Add to DB
		passwordHash,
		firstName,
		lastName,
		"", // avatarURL
		role,
		status,
		emailVerified,
		lastLoginAt,
		dbUser.CreatedAt.Time,
		dbUser.UpdatedAt.Time,
		deletedAt,
	)
}

// path: backend/internal/infrastructure/persistence/user_repository.go
// ADD THESE MISSING METHODS TO YOUR EXISTING UserRepository

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

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		// Properly scan all columns
		err := scanUser(rows, &dbUser)
		if err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
}

// FindRecentlyCreated finds recently created users
// func (r *UserRepository) FindRecentlyCreated(ctx context.Context, since time.Time, offset, limit int) ([]*user.User, error) {
// 	query := `
// 		SELECT * FROM users
// 		WHERE created_at > $1 AND deleted_at IS NULL
// 		ORDER BY created_at DESC
// 		LIMIT $2 OFFSET $3
// 	`

// 	rows, err := r.db.QueryContext(ctx, query, since, limit, offset)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var users []*user.User
// 	for rows.Next() {
// 		var dbUser db.User
// 		err := scanUser(rows, &dbUser)
// 		if err != nil {
// 			return nil, err
// 		}
// 		users = append(users, r.mapToDomain(dbUser))
// 	}

// 	return users, nil
// }

// FindPendingVerification finds users with unverified emails
func (r *UserRepository) FindPendingVerification(ctx context.Context, offset, limit int) ([]*user.User, error) {
	query := `
		SELECT * FROM users 
		WHERE (email_verified = false OR email_verified IS NULL)
		AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		err := scanUser(rows, &dbUser)
		if err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
}

// BulkCreate creates multiple users at once
func (r *UserRepository) BulkCreate(ctx context.Context, users []*user.User) error {
	if len(users) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, u := range users {
		fullName := fmt.Sprintf("%s %s", u.FirstName(), u.LastName())
		_, err = stmt.ExecContext(
			ctx,
			u.Email(),
			u.PasswordHash(),
			fullName,
			u.IsEmailVerified(),
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateAvatar updates user avatar URL
func (r *UserRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, avatarURL, time.Now())
	return err
}

// UpdateProfile updates user profile information
func (r *UserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error {
	fullName := fmt.Sprintf("%s %s", firstName, lastName)
	query := `UPDATE users SET full_name = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, fullName, time.Now())
	return err
}

// GetStatistics returns user statistics
func (r *UserRepository) GetStatistics(ctx context.Context) (*user.Statistics, error) {
	stats := &user.Statistics{}

	// Total users
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Active users (logged in last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE last_login_at > $1 AND deleted_at IS NULL`,
		thirtyDaysAgo,
	).Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	// Verified users
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE email_verified = true AND deleted_at IS NULL`,
	).Scan(&stats.VerifiedUsers)
	if err != nil {
		return nil, err
	}

	// New users (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE created_at > $1 AND deleted_at IS NULL`,
		sevenDaysAgo,
	).Scan(&stats.NewUsersThisWeek)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Helper function to scan user from rows
func scanUser(rows *sql.Rows, u *db.User) error {
	return rows.Scan(
		&u.ID,
		&u.Email,
		&u.EmailVerified,
		&u.PasswordHash,
		&u.FullName,
		&u.AvatarUrl,
		&u.Timezone,
		&u.Locale,
		&u.IsActive,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
}

// path: backend/internal/infrastructure/persistence/user_repository.go

// FindRecentlyCreated finds recently created users
func (r *UserRepository) FindRecentlyCreated(ctx context.Context, since time.Duration, offset, limit int) ([]*user.User, error) {
	// Convert duration to absolute time
	sinceTime := time.Now().Add(-since)

	query := `
		SELECT * FROM users 
		WHERE created_at > $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, sinceTime, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dbUser db.User
		err := scanUser(rows, &dbUser)
		if err != nil {
			return nil, err
		}
		users = append(users, r.mapToDomain(dbUser))
	}

	return users, nil
}
