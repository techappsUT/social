// ============================================================================
// FILE: backend/internal/infrastructure/persistence/post_repository.go
// FIXED VERSION - All compilation errors resolved
// ============================================================================
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	db "github.com/techappsUT/social-queue/internal/db"
	"github.com/techappsUT/social-queue/internal/domain/post"
)

type PostRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewPostRepository(database *sql.DB, queries *db.Queries) *PostRepository {
	return &PostRepository{
		db:      database,
		queries: queries,
	}
}

// ============================================================================
// CREATE
// ============================================================================

func (r *PostRepository) Create(ctx context.Context, p *post.Post) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Note: For now, we'll use the first platform's social account
	// In production, you'd need to handle multiple platforms differently
	var socialAccountID uuid.UUID
	// This is a placeholder - in real implementation, you'd look up the social account
	// based on team + platform combination
	socialAccountID = uuid.New() // TODO: Look up actual social account

	// Prepare JSONB fields - FIXED: Use pqtype.NullRawMessage correctly
	shortenedLinks := pqtype.NullRawMessage{
		RawMessage: []byte("[]"),
		Valid:      true,
	}
	platformOptions := pqtype.NullRawMessage{
		RawMessage: []byte("{}"),
		Valid:      true,
	}

	// Create scheduled post
	scheduledPost, err := qtx.CreateScheduledPost(ctx, db.CreateScheduledPostParams{
		TeamID:                  p.TeamID(),
		CreatedBy:               p.CreatedBy(),
		SocialAccountID:         socialAccountID,
		Content:                 p.Content().Text,
		ContentHtml:             sql.NullString{String: "", Valid: false},
		ShortenedLinks:          shortenedLinks,
		Status:                  db.NullPostStatus{PostStatus: db.PostStatusDraft, Valid: true},
		ScheduledAt:             sql.NullTime{Time: time.Time{}, Valid: false},
		PlatformSpecificOptions: platformOptions,
	})
	if err != nil {
		return fmt.Errorf("failed to create scheduled post: %w", err)
	}

	// Create attachments if any
	for i, mediaURL := range p.Content().MediaURLs {
		var mediaType db.AttachmentType
		if i < len(p.Content().MediaTypes) {
			mediaType = mapMediaTypeToDBType(p.Content().MediaTypes[i])
		} else {
			mediaType = db.AttachmentTypeImage // default
		}

		_, err := qtx.CreatePostAttachment(ctx, db.CreatePostAttachmentParams{
			ScheduledPostID: scheduledPost.ID,
			Type:            mediaType,
			Url:             mediaURL,
			ThumbnailUrl:    sql.NullString{Valid: false},
			FileSize:        sql.NullInt64{Valid: false},
			MimeType:        sql.NullString{Valid: false},
			Width:           sql.NullInt32{Valid: false},
			Height:          sql.NullInt32{Valid: false},
			Duration:        sql.NullInt32{Valid: false},
			AltText:         sql.NullString{Valid: false},
			UploadOrder:     sql.NullInt32{Int32: int32(i), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to create attachment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *PostRepository) FindByID(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	scheduledPost, err := r.queries.GetScheduledPostByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, post.ErrPostNotFound
		}
		return nil, fmt.Errorf("failed to find post: %w", err)
	}

	// Get attachments
	attachments, err := r.queries.ListPostAttachmentsByScheduledPost(ctx, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}

	return r.mapToPost(scheduledPost, attachments), nil
}

func (r *PostRepository) FindByTeamID(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*post.Post, error) {
	scheduledPosts, err := r.queries.ListScheduledPostsByTeam(ctx, db.ListScheduledPostsByTeamParams{
		TeamID: teamID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	posts := make([]*post.Post, 0, len(scheduledPosts))
	for _, sp := range scheduledPosts {
		attachments, _ := r.queries.ListPostAttachmentsByScheduledPost(ctx, sp.ID)
		posts = append(posts, r.mapToPost(sp, attachments))
	}

	return posts, nil
}

func (r *PostRepository) FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*post.Post, error) {
	// Custom query needed - for now return empty
	query := `
		SELECT * FROM scheduled_posts
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find posts by user: %w", err)
	}
	defer rows.Close()

	posts := make([]*post.Post, 0)
	for rows.Next() {
		var sp db.ScheduledPost
		err := rows.Scan(
			&sp.ID, &sp.TeamID, &sp.CreatedBy, &sp.SocialAccountID,
			&sp.Content, &sp.ContentHtml, &sp.ShortenedLinks, &sp.Status,
			&sp.ScheduledAt, &sp.PublishedAt, &sp.PlatformSpecificOptions,
			&sp.ErrorMessage, &sp.RetryCount, &sp.MaxRetries,
			&sp.CreatedAt, &sp.UpdatedAt, &sp.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		attachments, _ := r.queries.ListPostAttachmentsByScheduledPost(ctx, sp.ID)
		posts = append(posts, r.mapToPost(sp, attachments))
	}

	return posts, nil
}

func (r *PostRepository) FindDuePosts(ctx context.Context, before time.Time) ([]*post.Post, error) {
	duePosts, err := r.queries.GetDuePosts(ctx, db.GetDuePostsParams{
		ScheduledAt: sql.NullTime{Time: before, Valid: true},
		Limit:       100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get due posts: %w", err)
	}

	posts := make([]*post.Post, 0, len(duePosts))
	for _, dp := range duePosts {
		sp := db.ScheduledPost{
			ID:                      dp.ID,
			TeamID:                  dp.TeamID,
			CreatedBy:               dp.CreatedBy,
			SocialAccountID:         dp.SocialAccountID,
			Content:                 dp.Content,
			ContentHtml:             dp.ContentHtml,
			ShortenedLinks:          dp.ShortenedLinks,
			Status:                  dp.Status,
			ScheduledAt:             dp.ScheduledAt,
			PublishedAt:             dp.PublishedAt,
			PlatformSpecificOptions: dp.PlatformSpecificOptions,
			ErrorMessage:            dp.ErrorMessage,
			RetryCount:              dp.RetryCount,
			MaxRetries:              dp.MaxRetries,
			CreatedAt:               dp.CreatedAt,
			UpdatedAt:               dp.UpdatedAt,
			DeletedAt:               dp.DeletedAt,
		}
		attachments, _ := r.queries.ListPostAttachmentsByScheduledPost(ctx, dp.ID)
		posts = append(posts, r.mapToPost(sp, attachments))
	}

	return posts, nil
}

func (r *PostRepository) FindScheduled(ctx context.Context, offset, limit int) ([]*post.Post, error) {
	query := `
		SELECT * FROM scheduled_posts
		WHERE status = 'scheduled' AND deleted_at IS NULL
		ORDER BY scheduled_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find scheduled posts: %w", err)
	}
	defer rows.Close()

	return r.scanPostRows(ctx, rows)
}

func (r *PostRepository) FindPublished(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*post.Post, error) {
	query := `
		SELECT * FROM scheduled_posts
		WHERE team_id = $1 AND status = 'published' AND deleted_at IS NULL
		ORDER BY published_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find published posts: %w", err)
	}
	defer rows.Close()

	return r.scanPostRows(ctx, rows)
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *PostRepository) Update(ctx context.Context, p *post.Post) error {
	scheduleTime := sql.NullTime{Valid: false}
	if p.ScheduleTime() != nil {
		scheduleTime = sql.NullTime{Time: *p.ScheduleTime(), Valid: true}
	}

	_, err := r.queries.UpdateScheduledPost(ctx, db.UpdateScheduledPostParams{
		ID:          p.ID(),
		Content:     sql.NullString{String: p.Content().Text, Valid: true},
		ScheduledAt: scheduleTime,
	})
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	// Update status if changed
	var dbStatus db.PostStatus
	switch p.Status() {
	case post.StatusDraft:
		dbStatus = db.PostStatusDraft
	case post.StatusScheduled:
		dbStatus = db.PostStatusScheduled
	case post.StatusQueued:
		dbStatus = db.PostStatusQueued
	case post.StatusPublishing:
		dbStatus = db.PostStatusProcessing
	case post.StatusPublished:
		dbStatus = db.PostStatusPublished
	case post.StatusFailed:
		dbStatus = db.PostStatusFailed
	case post.StatusCanceled:
		dbStatus = db.PostStatusCancelled
	default:
		dbStatus = db.PostStatusDraft
	}

	err = r.queries.UpdateScheduledPostStatus(ctx, db.UpdateScheduledPostStatusParams{
		ID:     p.ID(),
		Status: db.NullPostStatus{PostStatus: dbStatus, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update post status: %w", err)
	}

	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *PostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteScheduledPost(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

// ============================================================================
// COUNT & STATS
// ============================================================================

func (r *PostRepository) CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error) {
	count, err := r.queries.CountScheduledPostsByTeam(ctx, teamID)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}
	return count, nil
}

func (r *PostRepository) CountScheduledByTeam(ctx context.Context, teamID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*) FROM scheduled_posts
		WHERE team_id = $1 AND status IN ('scheduled', 'queued') AND deleted_at IS NULL
	`
	var count int64
	err := r.db.QueryRowContext(ctx, query, teamID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count scheduled posts: %w", err)
	}
	return count, nil
}

func (r *PostRepository) CountPublishedToday(ctx context.Context, teamID uuid.UUID) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	count, err := r.queries.CountPostsByTeamAndDateRange(ctx, db.CountPostsByTeamAndDateRangeParams{
		TeamID:        teamID,
		PublishedAt:   sql.NullTime{Time: startOfDay, Valid: true},
		PublishedAt_2: sql.NullTime{Time: endOfDay, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count published today: %w", err)
	}
	return count, nil
}

// FIXED: Removed TeamID field from PostStatistics
func (r *PostRepository) GetTeamPostStats(ctx context.Context, teamID uuid.UUID) (*post.PostStatistics, error) {
	stats := &post.PostStatistics{}

	// Total posts
	total, err := r.CountByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	stats.TotalPosts = total

	// Scheduled posts
	scheduled, err := r.CountScheduledByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	stats.ScheduledPosts = scheduled

	// Published posts
	publishedCount, err := r.queries.CountPostsByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	stats.PublishedPosts = publishedCount

	return stats, nil
}

// ============================================================================
// ADDITIONAL METHODS (Placeholders for interface compliance)
// ============================================================================

func (r *PostRepository) FindByStatus(ctx context.Context, status post.Status, offset, limit int) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindQueued(ctx context.Context, limit int) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindFailed(ctx context.Context, limit int) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindByPlatform(ctx context.Context, platform post.Platform, offset, limit int) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindByTeamAndPlatform(ctx context.Context, teamID uuid.UUID, platform post.Platform, offset, limit int) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindScheduledBetween(ctx context.Context, teamID uuid.UUID, start, end time.Time) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindPublishedBetween(ctx context.Context, teamID uuid.UUID, start, end time.Time) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) FindCreatedToday(ctx context.Context, teamID uuid.UUID) ([]*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status post.Status) error {
	return nil // Implement if needed
}

func (r *PostRepository) MarkOverduePosts(ctx context.Context, before time.Time) (int, error) {
	return 0, nil // Implement if needed
}

func (r *PostRepository) GetNextInQueue(ctx context.Context) (*post.Post, error) {
	return nil, nil // Implement if needed
}

func (r *PostRepository) LockForProcessing(ctx context.Context, id uuid.UUID) error {
	return nil // Implement if needed
}

func (r *PostRepository) UnlockPost(ctx context.Context, id uuid.UUID) error {
	return nil // Implement if needed
}

func (r *PostRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM scheduled_posts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostRepository) DeleteOldDrafts(ctx context.Context, olderThan time.Time) (int, error) {
	return 0, nil // Implement if needed
}

func (r *PostRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE scheduled_posts SET deleted_at = NULL, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func (r *PostRepository) mapToPost(sp db.ScheduledPost, attachments []db.PostAttachment) *post.Post {
	// Extract media URLs and types
	mediaURLs := make([]string, 0, len(attachments))
	mediaTypes := make([]post.MediaType, 0, len(attachments))
	for _, att := range attachments {
		mediaURLs = append(mediaURLs, att.Url)
		mediaTypes = append(mediaTypes, mapDBTypeToMediaType(att.Type))
	}

	// Build content
	content := post.Content{
		Text:       sp.Content,
		MediaURLs:  mediaURLs,
		MediaTypes: mediaTypes,
		Hashtags:   []string{},
		Mentions:   []string{},
	}

	// Parse platforms from DB (for now, default to Twitter)
	platforms := []post.Platform{post.PlatformTwitter}

	// Build post entity
	var scheduleTime *time.Time
	if sp.ScheduledAt.Valid {
		scheduleTime = &sp.ScheduledAt.Time
	}

	var publishedAt *time.Time
	if sp.PublishedAt.Valid {
		publishedAt = &sp.PublishedAt.Time
	}

	var deletedAt *time.Time
	if sp.DeletedAt.Valid {
		deletedAt = &sp.DeletedAt.Time
	}

	// Map status
	var status post.Status
	if sp.Status.Valid {
		switch sp.Status.PostStatus {
		case db.PostStatusDraft:
			status = post.StatusDraft
		case db.PostStatusScheduled:
			status = post.StatusScheduled
		case db.PostStatusQueued:
			status = post.StatusQueued
		case db.PostStatusProcessing:
			status = post.StatusPublishing
		case db.PostStatusPublished:
			status = post.StatusPublished
		case db.PostStatusFailed:
			status = post.StatusFailed
		case db.PostStatusCancelled:
			status = post.StatusCanceled
		default:
			status = post.StatusDraft
		}
	} else {
		status = post.StatusDraft
	}

	var createdAt, updatedAt time.Time
	if sp.CreatedAt.Valid {
		createdAt = sp.CreatedAt.Time
	}
	if sp.UpdatedAt.Valid {
		updatedAt = sp.UpdatedAt.Time
	}

	// FIXED: Use Reconstruct (not ReconstructPost)
	return post.Reconstruct(
		sp.ID,
		sp.TeamID,
		sp.CreatedBy,
		content,
		platforms,
		scheduleTime,
		publishedAt,
		status,
		post.PriorityNormal,
		post.Metadata{},
		nil,
		createdAt,
		updatedAt,
		deletedAt,
	)
}

func (r *PostRepository) scanPostRows(ctx context.Context, rows *sql.Rows) ([]*post.Post, error) {
	posts := make([]*post.Post, 0)
	for rows.Next() {
		var sp db.ScheduledPost
		err := rows.Scan(
			&sp.ID, &sp.TeamID, &sp.CreatedBy, &sp.SocialAccountID,
			&sp.Content, &sp.ContentHtml, &sp.ShortenedLinks, &sp.Status,
			&sp.ScheduledAt, &sp.PublishedAt, &sp.PlatformSpecificOptions,
			&sp.ErrorMessage, &sp.RetryCount, &sp.MaxRetries,
			&sp.CreatedAt, &sp.UpdatedAt, &sp.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		attachments, _ := r.queries.ListPostAttachmentsByScheduledPost(ctx, sp.ID)
		posts = append(posts, r.mapToPost(sp, attachments))
	}
	return posts, nil
}

func mapMediaTypeToDBType(mt post.MediaType) db.AttachmentType {
	switch mt {
	case post.MediaTypeImage:
		return db.AttachmentTypeImage
	case post.MediaTypeVideo:
		return db.AttachmentTypeVideo
	case post.MediaTypeGIF:
		return db.AttachmentTypeGif
	default:
		return db.AttachmentTypeImage
	}
}

func mapDBTypeToMediaType(dt db.AttachmentType) post.MediaType {
	switch dt {
	case db.AttachmentTypeImage:
		return post.MediaTypeImage
	case db.AttachmentTypeVideo:
		return post.MediaTypeVideo
	case db.AttachmentTypeGif:
		return post.MediaTypeGIF
	default:
		return post.MediaTypeImage
	}
}
