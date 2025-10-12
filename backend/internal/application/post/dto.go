// ============================================================================
// FILE 1: backend/internal/application/post/dto.go
// ============================================================================
package post

import (
	"time"

	"github.com/google/uuid"
	postDomain "github.com/techappsUT/social-queue/internal/domain/post"
)

type PostDTO struct {
	ID          uuid.UUID  `json:"id"`
	TeamID      uuid.UUID  `json:"teamId"`
	CreatedBy   uuid.UUID  `json:"createdBy"`
	Content     string     `json:"content"`
	Platforms   []string   `json:"platforms"`
	MediaURLs   []string   `json:"mediaUrls,omitempty"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func MapPostToDTO(p *postDomain.Post) *PostDTO {
	if p == nil {
		return nil
	}

	platforms := make([]string, 0, len(p.Platforms()))
	for _, platform := range p.Platforms() {
		platforms = append(platforms, string(platform))
	}

	return &PostDTO{
		ID:          p.ID(),
		TeamID:      p.TeamID(),
		CreatedBy:   p.CreatedBy(),
		Content:     p.Content().Text,
		Platforms:   platforms,
		MediaURLs:   p.Content().MediaURLs,
		Status:      string(p.Status()),
		ScheduledAt: p.ScheduleTime(),
		PublishedAt: p.PublishedAt(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}
