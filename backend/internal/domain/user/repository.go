// path: backend/internal/domain/user/repository.go
// 🆕 NEW - Clean Architecture

package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for user persistence
// This interface is part of the domain layer and will be implemented
// by the infrastructure layer (e.g., PostgreSQL, MongoDB, etc.)
type Repository interface {
	// Create persists a new user
	Create(ctx context.Context, user *User) error

	// Update persists changes to an existing user
	Update(ctx context.Context, user *User) error

	// Delete performs a soft delete on a user
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)

	// FindByEmail retrieves a user by their email address
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByUsername retrieves a user by their username
	FindByUsername(ctx context.Context, username string) (*User, error)

	// FindByEmailOrUsername retrieves a user by email or username
	// Used for login scenarios
	FindByEmailOrUsername(ctx context.Context, identifier string) (*User, error)

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user exists with the given username
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// FindAll retrieves all users with pagination
	FindAll(ctx context.Context, offset, limit int) ([]*User, error)

	// FindByRole retrieves all users with a specific role
	FindByRole(ctx context.Context, role Role, offset, limit int) ([]*User, error)

	// FindByStatus retrieves all users with a specific status
	FindByStatus(ctx context.Context, status Status, offset, limit int) ([]*User, error)

	// FindByTeamID retrieves all users belonging to a specific team
	FindByTeamID(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// CountByRole returns the count of users by role
	CountByRole(ctx context.Context, role Role) (int64, error)

	// CountByStatus returns the count of users by status
	CountByStatus(ctx context.Context, status Status) (int64, error)

	// Search searches users by query string (searches in email, username, first name, last name)
	Search(ctx context.Context, query string, offset, limit int) ([]*User, error)

	// FindInactiveSince finds users who haven't logged in since the given time
	FindInactiveSince(ctx context.Context, since time.Time, offset, limit int) ([]*User, error)

	// FindRecentlyCreated finds users created within the specified duration
	FindRecentlyCreated(ctx context.Context, duration time.Duration, offset, limit int) ([]*User, error)

	// UpdateLastLogin updates only the last login timestamp
	UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error

	// UpdatePassword updates only the password hash
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// UpdateEmailVerificationStatus updates the email verification status
	UpdateEmailVerificationStatus(ctx context.Context, id uuid.UUID, verified bool) error

	// UpdateStatus updates only the user status
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error

	// UpdateRole updates only the user role
	UpdateRole(ctx context.Context, id uuid.UUID, role Role) error

	// BulkUpdateStatus updates status for multiple users
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status Status) error

	// FindUnverifiedOlderThan finds unverified users older than specified duration
	FindUnverifiedOlderThan(ctx context.Context, duration time.Duration) ([]*User, error)

	// Restore restores a soft-deleted user
	Restore(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a user (use with extreme caution)
	HardDelete(ctx context.Context, id uuid.UUID) error
}

// CacheRepository defines caching operations for users
// This is optional and can be implemented by Redis or similar
type CacheRepository interface {
	// GetByID retrieves a user from cache by ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail retrieves a user from cache by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Set stores a user in cache
	Set(ctx context.Context, user *User, ttl time.Duration) error

	// Delete removes a user from cache
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByEmail removes a user from cache by email
	DeleteByEmail(ctx context.Context, email string) error

	// Invalidate invalidates all user-related cache entries
	Invalidate(ctx context.Context, id uuid.UUID) error
}

// QueryOptions represents options for querying users
type QueryOptions struct {
	Offset         int
	Limit          int
	SortBy         string
	SortOrder      string // "asc" or "desc"
	IncludeDeleted bool
}

// SearchCriteria represents search criteria for advanced user searches
type SearchCriteria struct {
	Query           string
	Role            *Role
	Status          *Status
	EmailVerified   *bool
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
	LastLoginAfter  *time.Time
	LastLoginBefore *time.Time
	TeamID          *uuid.UUID
}

// AdvancedRepository extends Repository with more complex operations
// This is optional and can be implemented as needed
type AdvancedRepository interface {
	Repository

	// FindByCriteria performs an advanced search with multiple criteria
	FindByCriteria(ctx context.Context, criteria SearchCriteria, opts QueryOptions) ([]*User, int64, error)

	// FindUsersWithoutTeams finds users who are not members of any team
	FindUsersWithoutTeams(ctx context.Context, offset, limit int) ([]*User, error)

	// FindDuplicateEmails finds users with potentially duplicate emails (for data cleanup)
	FindDuplicateEmails(ctx context.Context) (map[string][]uuid.UUID, error)

	// GetStatistics returns user statistics (total, active, suspended, etc.)
	GetStatistics(ctx context.Context) (*UserStatistics, error)
}

// UserStatistics holds aggregated user statistics
type UserStatistics struct {
	Total              int64
	Active             int64
	Inactive           int64
	Suspended          int64
	Pending            int64
	EmailVerified      int64
	EmailUnverified    int64
	DeletedCount       int64
	CreatedToday       int64
	CreatedThisWeek    int64
	CreatedThisMonth   int64
	LastLoginToday     int64
	LastLoginThisWeek  int64
	LastLoginThisMonth int64
	ByRole             map[Role]int64
}

// Transaction represents a database transaction
// Used for operations that need to be atomic
type Transaction interface {
	Commit() error
	Rollback() error
}

// TransactionalRepository provides transactional operations
type TransactionalRepository interface {
	Repository

	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (Transaction, error)

	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(repo Repository) error) error
}
