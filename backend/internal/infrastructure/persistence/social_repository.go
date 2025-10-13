// ============================================================================
// FILE: backend/internal/infrastructure/persistence/social_repository.go
// ============================================================================
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
	"github.com/techappsUT/social-queue/internal/domain/social"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
)

type SocialRepository struct {
	queries    *db.Queries
	encryption *services.EncryptionService
}

func NewSocialRepository(queries *db.Queries, encryption *services.EncryptionService) social.AccountRepository {
	return &SocialRepository{
		queries:    queries,
		encryption: encryption,
	}
}

// Create persists a new social account
func (r *SocialRepository) Create(ctx context.Context, account *social.Account) error {
	credentials := account.Credentials()

	// Encrypt tokens
	encryptedAccessToken, err := r.encryption.Encrypt(credentials.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken sql.NullString
	if credentials.RefreshToken != "" {
		encrypted, err := r.encryption.Encrypt(credentials.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		encryptedRefreshToken = sql.NullString{String: encrypted, Valid: true}
	}

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(account.Metadata())
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// FIX: Construct pqtype.NullRawMessage properly
	var metadataNullJSON pqtype.NullRawMessage
	if len(metadataJSON) > 0 {
		metadataNullJSON = pqtype.NullRawMessage{
			RawMessage: metadataJSON,
			Valid:      true,
		}
	}

	// Create account record
	dbAccount, err := r.queries.LinkSocialAccount(ctx, db.LinkSocialAccountParams{
		TeamID:         account.TeamID(),
		Platform:       db.SocialPlatform(account.Platform()),
		PlatformUserID: credentials.PlatformUserID,
		Username:       sql.NullString{String: account.Username(), Valid: account.Username() != ""},
		DisplayName:    sql.NullString{String: account.DisplayName(), Valid: account.DisplayName() != ""},
		AvatarUrl:      sql.NullString{String: account.AvatarURL(), Valid: account.AvatarURL() != ""},
		ProfileUrl:     sql.NullString{String: account.ProfileURL(), Valid: account.ProfileURL() != ""},
		AccountType:    sql.NullString{String: string(account.AccountType()), Valid: true},
		Status:         db.NullSocialAccountStatus{SocialAccountStatus: db.SocialAccountStatus(account.Status()), Valid: true},
		Metadata:       metadataNullJSON, // FIX: Use proper type
		ConnectedBy:    uuid.NullUUID{UUID: account.UserID(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	// Create token record
	var expiresAt sql.NullTime
	if credentials.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *credentials.ExpiresAt, Valid: true}
	}

	_, err = r.queries.CreateSocialToken(ctx, db.CreateSocialTokenParams{
		SocialAccountID: dbAccount.ID,
		AccessToken:     encryptedAccessToken,
		RefreshToken:    encryptedRefreshToken,
		TokenType:       sql.NullString{String: "Bearer", Valid: true},
		ExpiresAt:       expiresAt,
		Scope:           sql.NullString{String: "", Valid: false},
	})
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

// FindByID retrieves a social account by ID
func (r *SocialRepository) FindByID(ctx context.Context, id uuid.UUID) (*social.Account, error) {
	row, err := r.queries.GetSocialAccountWithToken(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.mapToAccount(row)
}

// FindByTeamID retrieves all social accounts for a team
func (r *SocialRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*social.Account, error) {
	dbAccounts, err := r.queries.ListSocialAccountsByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	accounts := make([]*social.Account, 0, len(dbAccounts))
	for _, dbAccount := range dbAccounts {
		row, err := r.queries.GetSocialAccountWithToken(ctx, dbAccount.ID)
		if err != nil {
			continue
		}
		account, err := r.mapToAccount(row)
		if err != nil {
			continue
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// FindByTeamAndPlatform retrieves all social accounts for a team and platform (RENAMED)
func (r *SocialRepository) FindByTeamAndPlatform(ctx context.Context, teamID uuid.UUID, platform social.Platform) ([]*social.Account, error) {
	dbAccounts, err := r.queries.ListSocialAccountsByPlatform(ctx, db.ListSocialAccountsByPlatformParams{
		TeamID:   teamID,
		Platform: db.SocialPlatform(platform),
	})
	if err != nil {
		return nil, err
	}

	accounts := make([]*social.Account, 0, len(dbAccounts))
	for _, dbAccount := range dbAccounts {
		row, err := r.queries.GetSocialAccountWithToken(ctx, dbAccount.ID)
		if err != nil {
			continue
		}
		account, err := r.mapToAccount(row)
		if err != nil {
			continue
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Update updates a social account
func (r *SocialRepository) Update(ctx context.Context, account *social.Account) error {
	credentials := account.Credentials()

	// Encrypt tokens
	encryptedAccessToken, err := r.encryption.Encrypt(credentials.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// FIX: Construct sql.NullString properly
	var encryptedRefreshToken sql.NullString
	if credentials.RefreshToken != "" {
		encrypted, err := r.encryption.Encrypt(credentials.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		encryptedRefreshToken = sql.NullString{String: encrypted, Valid: true}
	}

	// Update tokens
	var expiresAt sql.NullTime
	if credentials.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *credentials.ExpiresAt, Valid: true}
	}

	err = r.queries.UpdateSocialToken(ctx, db.UpdateSocialTokenParams{
		SocialAccountID: account.ID(),
		AccessToken:     sql.NullString{String: encryptedAccessToken, Valid: true},
		RefreshToken:    encryptedRefreshToken,
		TokenType:       sql.NullString{String: "Bearer", Valid: true},
		ExpiresAt:       expiresAt,
		Scope:           sql.NullString{}, // Empty scope, not updated
	})
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	// Update account metadata
	metadataJSON, err := json.Marshal(account.Metadata())
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// FIX: Construct pqtype.NullRawMessage properly
	var metadataNullJSON pqtype.NullRawMessage
	if len(metadataJSON) > 0 {
		metadataNullJSON = pqtype.NullRawMessage{
			RawMessage: metadataJSON,
			Valid:      true,
		}
	}

	// FIX: Use UpdateSocialAccountMetadata instead of non-existent UpdateSocialAccount
	return r.queries.UpdateSocialAccountMetadata(ctx, db.UpdateSocialAccountMetadataParams{
		ID:          account.ID(),
		Username:    sql.NullString{String: account.Username(), Valid: account.Username() != ""},
		DisplayName: sql.NullString{String: account.DisplayName(), Valid: account.DisplayName() != ""},
		AvatarUrl:   sql.NullString{String: account.AvatarURL(), Valid: account.AvatarURL() != ""},
		Metadata:    metadataNullJSON,
	})
}

// Delete soft-deletes a social account
func (r *SocialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// FIX: Use SoftDeleteSocialAccount instead of DeleteSocialAccount
	return r.queries.SoftDeleteSocialAccount(ctx, id)
}

// CountByTeamID counts social accounts for a team
func (r *SocialRepository) CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error) {
	return r.queries.CountSocialTokensByTeam(ctx, teamID)
}

// BulkUpdateStatus updates status for multiple accounts
func (r *SocialRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status social.Status) error {
	// Update each account's status in a transaction would be ideal,
	// but for simplicity we'll update one by one
	for _, id := range ids {
		err := r.queries.UpdateSocialAccountStatus(ctx, db.UpdateSocialAccountStatusParams{
			ID:     id,
			Status: db.NullSocialAccountStatus{SocialAccountStatus: db.SocialAccountStatus(status), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to update account %s status: %w", id, err)
		}
	}
	return nil
}

// FindByUserID retrieves all social accounts connected by a user
func (r *SocialRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*social.Account, error) {
	// Note: You may need to add a SQLC query for this
	// For now, returning empty slice
	return []*social.Account{}, nil
}

// FindByPlatform retrieves accounts by platform with pagination
func (r *SocialRepository) FindByPlatform(ctx context.Context, platform social.Platform, offset, limit int) ([]*social.Account, error) {
	// Note: You may need to add a SQLC query for this with pagination
	return []*social.Account{}, nil
}

// FindByPlatformUserID finds an account by platform and platform user ID
func (r *SocialRepository) FindByPlatformUserID(ctx context.Context, platform social.Platform, platformUserID string) (*social.Account, error) {
	// Note: You may need to add a SQLC query for this
	return nil, fmt.Errorf("not implemented")
}

// FindByStatus retrieves accounts by status with pagination
func (r *SocialRepository) FindByStatus(ctx context.Context, status social.Status, offset, limit int) ([]*social.Account, error) {
	// Note: You may need to add a SQLC query for this
	return []*social.Account{}, nil
}

// FindExpiredAccounts retrieves all expired accounts
func (r *SocialRepository) FindExpiredAccounts(ctx context.Context) ([]*social.Account, error) {
	// Note: You may need to add a SQLC query for this
	return []*social.Account{}, nil
}

// FindExpiringAccounts retrieves accounts expiring within specified days
func (r *SocialRepository) FindExpiringAccounts(ctx context.Context, withinDays int) ([]*social.Account, error) {
	// Note: You may need to add a SQLC query for this
	return []*social.Account{}, nil
}

// FindRateLimitedAccounts retrieves all rate-limited accounts
func (r *SocialRepository) FindRateLimitedAccounts(ctx context.Context) ([]*social.Account, error) {
	// Note: You may need to add a SQLC query for this
	return []*social.Account{}, nil
}

// ExistsByTeamAndPlatformUser checks if an account exists
func (r *SocialRepository) ExistsByTeamAndPlatformUser(ctx context.Context, teamID uuid.UUID, platform social.Platform, platformUserID string) (bool, error) {
	// Note: You may need to add a SQLC query for this
	return false, nil
}

// HardDelete permanently deletes an account
func (r *SocialRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	// Note: You may need to add a SQLC query for this
	return fmt.Errorf("not implemented")
}

// DeleteExpiredAccounts deletes accounts expired longer than specified duration
// Returns the number of accounts deleted
func (r *SocialRepository) DeleteExpiredAccounts(ctx context.Context, olderThan time.Duration) (int, error) {
	// Note: You may need to add a SQLC query for this
	return 0, nil
}

// Helper: Map database row to domain entity with decrypted tokens
func (r *SocialRepository) mapToAccount(row db.GetSocialAccountWithTokenRow) (*social.Account, error) {
	// FIX: Handle sql.NullString for AccessToken
	if !row.AccessToken.Valid || row.AccessToken.String == "" {
		return nil, fmt.Errorf("access token is missing")
	}

	// Decrypt tokens
	accessToken, err := r.encryption.Decrypt(row.AccessToken.String) // FIX: Use .String
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	var refreshToken string
	if row.RefreshToken.Valid {
		refreshToken, err = r.encryption.Decrypt(row.RefreshToken.String)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
		}
	}

	// Parse metadata
	var metadata social.AccountMetadata
	// FIX: Check Valid field for pqtype.NullRawMessage
	if row.Metadata.Valid && len(row.Metadata.RawMessage) > 0 {
		if err := json.Unmarshal(row.Metadata.RawMessage, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Build credentials
	var expiresAt *time.Time
	if row.TokenExpiresAt.Valid {
		expiresAt = &row.TokenExpiresAt.Time
	}

	credentials := social.Credentials{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		ExpiresAt:      expiresAt,
		PlatformUserID: row.PlatformUserID,
	}

	// Reconstruct domain entity
	var connectedAt time.Time
	if row.ConnectedAt.Valid {
		connectedAt = row.ConnectedAt.Time
	}

	var updatedAt time.Time
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		deletedAt = &row.DeletedAt.Time
	}

	// FIX: Handle LastSyncedAt pointer conversion
	var lastSyncAt *time.Time
	if row.LastSyncedAt.Valid {
		lastSyncAt = &row.LastSyncedAt.Time
	}

	return social.Reconstruct(
		row.ID,
		row.TeamID,
		row.ConnectedBy.UUID,
		social.Platform(row.Platform),
		social.AccountType(row.AccountType.String),
		row.Username.String,
		row.DisplayName.String,
		row.ProfileUrl.String,
		row.AvatarUrl.String,
		credentials,
		metadata,
		social.Status(row.Status.SocialAccountStatus),
		social.RateLimits{}, // Empty rate limits for now
		lastSyncAt,          // FIX: Pass pointer
		connectedAt,
		nil, // expiresAt for account (not token)
		createdAt,
		updatedAt,
		deletedAt,
	), nil
}
