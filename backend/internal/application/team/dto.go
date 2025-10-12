// path: backend/internal/application/team/dto.go
package team

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/domain/team"
	"github.com/techappsUT/social-queue/internal/domain/user"
)

// TeamDTO represents a team data transfer object
type TeamDTO struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Settings    TeamSettingsDTO `json:"settings"`
	Members     []MemberDTO     `json:"members"`
	MemberCount int             `json:"memberCount"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// TeamSettingsDTO represents team settings
type TeamSettingsDTO struct {
	Timezone            string `json:"timezone"`
	DefaultPostTime     string `json:"defaultPostTime"`
	EnableNotifications bool   `json:"enableNotifications"`
	EnableAnalytics     bool   `json:"enableAnalytics"`
	RequireApproval     bool   `json:"requireApproval"`
	AutoSchedule        bool   `json:"autoSchedule"`
	Language            string `json:"language"`
	DateFormat          string `json:"dateFormat"`
}

// MemberDTO represents a team member data transfer object
type MemberDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joinedAt"`
}

// MapTeamToDTO maps a team entity to a DTO
func MapTeamToDTO(t *team.Team, members []*team.Member, userRepo user.Repository, ctx context.Context) *TeamDTO {
	memberDTOs := make([]MemberDTO, 0, len(members))

	for _, m := range members {
		// Get user details for this member
		u, err := userRepo.FindByID(ctx, m.UserID())

		memberDTO := MemberDTO{
			ID:       m.ID(),
			UserID:   m.UserID(),
			Role:     string(m.Role()),
			JoinedAt: time.Now(), // Default if no joinedAt
		}

		// Add user details if found
		if err == nil && u != nil {
			memberDTO.Email = u.Email()
			memberDTO.Username = u.Username()
			memberDTO.FirstName = u.FirstName()
			memberDTO.LastName = u.LastName()
			memberDTO.AvatarURL = u.AvatarURL()
		}

		if m.JoinedAt() != nil {
			memberDTO.JoinedAt = *m.JoinedAt()
		}

		memberDTOs = append(memberDTOs, memberDTO)
	}

	// Map TeamSettings to TeamSettingsDTO
	settings := t.Settings()
	settingsDTO := TeamSettingsDTO{
		Timezone:            settings.Timezone,
		DefaultPostTime:     settings.DefaultPostTime,
		EnableNotifications: settings.EnableNotifications,
		EnableAnalytics:     settings.EnableAnalytics,
		RequireApproval:     settings.RequireApproval,
		AutoSchedule:        settings.AutoSchedule,
		Language:            settings.Language,
		DateFormat:          settings.DateFormat,
	}

	return &TeamDTO{
		ID:          t.ID(),
		Name:        t.Name(),
		Slug:        t.Slug(),
		Settings:    settingsDTO,
		Members:     memberDTOs,
		MemberCount: len(members),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

// MapMemberToDTO maps a member entity to a DTO
func MapMemberToDTO(m *team.Member, u *user.User) *MemberDTO {
	dto := &MemberDTO{
		ID:       m.ID(),
		UserID:   m.UserID(),
		Role:     string(m.Role()),
		JoinedAt: time.Now(),
	}

	if m.JoinedAt() != nil {
		dto.JoinedAt = *m.JoinedAt()
	}

	if u != nil {
		dto.Email = u.Email()
		dto.Username = u.Username()
		dto.FirstName = u.FirstName()
		dto.LastName = u.LastName()
		dto.AvatarURL = u.AvatarURL()
	}

	return dto
}
