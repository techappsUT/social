// backend/internal/infrastructure/persistence/user_repository.go
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

func NewUserRepository(database *sql.DB, queries *db.Queries) user.Repository {
	return &UserRepository{
		db:      database,
		queries: queries,
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
			if pqErr.Code == "23505" {
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
		SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name,
		       avatar_url, timezone, locale, is_active,
		       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
		       last_login_at, created_at, updated_at, deleted_at
		FROM users 
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
			&dbUser.VerificationToken,
			&dbUser.VerificationTokenExpiresAt,
			&dbUser.ResetToken,
			&dbUser.ResetTokenExpiresAt,
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

	// Update password if changed
	if u.PasswordHash() != "" {
		err = r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
			ID:           u.ID(),
			PasswordHash: sql.NullString{String: u.PasswordHash(), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}
	}

	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	err := r.queries.UpdateUserLastLogin(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	err := r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: sql.NullString{String: passwordHash, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateEmailVerificationStatus(ctx context.Context, id uuid.UUID, verified bool) error {
	// Use the ClearVerificationToken query which also sets email_verified = TRUE
	err := r.queries.ClearVerificationToken(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}
	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return user.ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ============================================================================
// EXISTS CHECKS
// ============================================================================

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.CheckEmailExists(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.queries.CheckUsernameExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

// ============================================================================
// COUNTS & STATS (Placeholder implementations)
// ============================================================================

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *UserRepository) FindByRole(ctx context.Context, role user.Role, offset, limit int) ([]*user.User, error) {
	// Placeholder - implement when role column is added
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) FindByStatus(ctx context.Context, status user.Status, offset, limit int) ([]*user.User, error) {
	// Placeholder - implement when needed
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*user.User, error) {
	// Placeholder - implement when needed
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) CountByRole(ctx context.Context, role user.Role) (int64, error) {
	// Placeholder
	return 0, fmt.Errorf("not implemented")
}

func (r *UserRepository) CountByStatus(ctx context.Context, status user.Status) (int64, error) {
	// Placeholder
	return 0, fmt.Errorf("not implemented")
}

func (r *UserRepository) Search(ctx context.Context, query string, offset, limit int) ([]*user.User, error) {
	// Placeholder
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) FindInactiveSince(ctx context.Context, since time.Time, offset, limit int) ([]*user.User, error) {
	// Placeholder
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) FindRecentlyCreated(ctx context.Context, duration time.Duration, offset, limit int) ([]*user.User, error) {
	// Placeholder
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.Status) error {
	// Placeholder
	return fmt.Errorf("not implemented")
}

// ============================================================================
// MAPPING
// ============================================================================

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
