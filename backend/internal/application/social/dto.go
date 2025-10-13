// ============================================================================
// FILE: backend/internal/application/social/dto.go
// ============================================================================
package social

import (
	"time"

	"github.com/google/uuid"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
)

type SocialAccountDTO struct {
	ID             uuid.UUID  `json:"id"`
	TeamID         uuid.UUID  `json:"teamId"`
	Platform       string     `json:"platform"`
	PlatformUserID string     `json:"platformUserId"`
	Username       string     `json:"username"`
	DisplayName    string     `json:"displayName"`
	AvatarURL      string     `json:"avatarUrl,omitempty"`
	IsActive       bool       `json:"isActive"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	ConnectedAt    time.Time  `json:"connectedAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type AnalyticsDTO struct {
	Impressions int `json:"impressions"`
	Engagements int `json:"engagements"`
	Likes       int `json:"likes"`
	Shares      int `json:"shares"`
	Comments    int `json:"comments"`
	Clicks      int `json:"clicks"`
}

// MapAccountToDTO converts domain entity to DTO
func MapAccountToDTO(account *socialDomain.Account) *SocialAccountDTO {
	return &SocialAccountDTO{
		ID:             account.ID(),
		TeamID:         account.TeamID(),
		Platform:       string(account.Platform()),
		PlatformUserID: account.PlatformUserID(),
		Username:       account.Username(),
		DisplayName:    account.DisplayName(),
		AvatarURL:      account.AvatarURL(),
		IsActive:       account.IsActive(),
		ExpiresAt:      account.ExpiresAt(),
		ConnectedAt:    account.CreatedAt(),
		UpdatedAt:      account.UpdatedAt(),
	}
}
