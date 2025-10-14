// path: backend/internal/application/dto/mappers.go
package dto

// import (
// 	"time"

// 	"github.com/google/uuid"
// 	postDomain "github.com/techappsUT/social-queue/internal/domain/post"
// 	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
// 	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
// 	userDomain "github.com/techappsUT/social-queue/internal/domain/user"
// )

// // ============================================================================
// // USER DTOs
// // ============================================================================

// type UserDTO struct {
// 	ID        uuid.UUID `json:"id"`
// 	Email     string    `json:"email"`
// 	FirstName *string   `json:"firstName,omitempty"`
// 	LastName  *string   `json:"lastName,omitempty"`
// 	AvatarURL *string   `json:"avatarUrl,omitempty"`
// 	CreatedAt time.Time `json:"createdAt"`
// 	UpdatedAt time.Time `json:"updatedAt"`
// }

// func MapUserToDTO(user *userDomain.User) *UserDTO {
// 	if user == nil {
// 		return nil
// 	}
// 	return &UserDTO{
// 		ID:        user.ID(),
// 		Email:     user.Email(),
// 		FirstName: user.FirstName(),
// 		LastName:  user.LastName(),
// 		AvatarURL: user.AvatarURL(),
// 		CreatedAt: user.CreatedAt(),
// 		UpdatedAt: user.UpdatedAt(),
// 	}
// }

// func MapUsersToDTO(users []*userDomain.User) []*UserDTO {
// 	dtos := make([]*UserDTO, len(users))
// 	for i, user := range users {
// 		dtos[i] = MapUserToDTO(user)
// 	}
// 	return dtos
// }

// // ============================================================================
// // TEAM DTOs
// // ============================================================================

// type TeamDTO struct {
// 	ID          uuid.UUID  `json:"id"`
// 	Name        string     `json:"name"`
// 	Slug        string     `json:"slug"`
// 	AvatarURL   *string    `json:"avatarUrl,omitempty"`
// 	IsActive    bool       `json:"isActive"`
// 	MemberCount int        `json:"memberCount,omitempty"`
// 	CreatedBy   *uuid.UUID `json:"createdBy,omitempty"`
// 	CreatedAt   time.Time  `json:"createdAt"`
// 	UpdatedAt   time.Time  `json:"updatedAt"`
// }

// func MapTeamToDTO(team *teamDomain.Team) *TeamDTO {
// 	if team == nil {
// 		return nil
// 	}
// 	dto := &TeamDTO{
// 		ID:        team.ID(),
// 		Name:      team.Name(),
// 		Slug:      team.Slug(),
// 		IsActive:  team.IsActive(),
// 		CreatedAt: team.CreatedAt(),
// 		UpdatedAt: team.UpdatedAt(),
// 	}
// 	if avatarURL := team.AvatarURL(); avatarURL != nil {
// 		dto.AvatarURL = avatarURL
// 	}
// 	if createdBy := team.CreatedBy(); createdBy != nil {
// 		dto.CreatedBy = createdBy
// 	}
// 	return dto
// }

// func MapTeamsToDTO(teams []*teamDomain.Team) []*TeamDTO {
// 	dtos := make([]*TeamDTO, len(teams))
// 	for i, team := range teams {
// 		dtos[i] = MapTeamToDTO(team)
// 	}
// 	return dtos
// }

// // ============================================================================
// // TEAM MEMBER DTOs
// // ============================================================================

// type MemberDTO struct {
// 	ID        uuid.UUID  `json:"id"`
// 	TeamID    uuid.UUID  `json:"teamId"`
// 	UserID    uuid.UUID  `json:"userId"`
// 	Role      string     `json:"role"`
// 	Status    string     `json:"status"`
// 	User      *UserDTO   `json:"user,omitempty"`
// 	JoinedAt  *time.Time `json:"joinedAt,omitempty"`
// 	CreatedAt time.Time  `json:"createdAt"`
// 	UpdatedAt time.Time  `json:"updatedAt"`
// }

// func MapMemberToDTO(member *teamDomain.Member) *MemberDTO {
// 	if member == nil {
// 		return nil
// 	}
// 	return &MemberDTO{
// 		ID:        member.ID(),
// 		TeamID:    member.TeamID(),
// 		UserID:    member.UserID(),
// 		Role:      string(member.Role()),
// 		Status:    string(member.Status()),
// 		JoinedAt:  member.JoinedAt(),
// 		CreatedAt: member.CreatedAt(),
// 		UpdatedAt: member.UpdatedAt(),
// 	}
// }

// // MapMemberWithUserToDTO includes user details
// func MapMemberWithUserToDTO(member *teamDomain.Member, user *userDomain.User) *MemberDTO {
// 	dto := MapMemberToDTO(member)
// 	if dto != nil {
// 		dto.User = MapUserToDTO(user)
// 	}
// 	return dto
// }

// func MapMembersToDTO(members []*teamDomain.Member) []*MemberDTO {
// 	dtos := make([]*MemberDTO, len(members))
// 	for i, member := range members {
// 		dtos[i] = MapMemberToDTO(member)
// 	}
// 	return dtos
// }

// // ============================================================================
// // POST DTOs
// // ============================================================================

// type PostDTO struct {
// 	ID          uuid.UUID  `json:"id"`
// 	TeamID      uuid.UUID  `json:"teamId"`
// 	AuthorID    uuid.UUID  `json:"authorId"`
// 	Content     string     `json:"content"`
// 	Platforms   []string   `json:"platforms"`
// 	Attachments []string   `json:"attachments,omitempty"`
// 	Status      string     `json:"status"`
// 	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
// 	PublishedAt *time.Time `json:"publishedAt,omitempty"`
// 	Author      *UserDTO   `json:"author,omitempty"`
// 	CreatedAt   time.Time  `json:"createdAt"`
// 	UpdatedAt   time.Time  `json:"updatedAt"`
// }

// func MapPostToDTO(post *postDomain.Post) *PostDTO {
// 	if post == nil {
// 		return nil
// 	}

// 	// Convert platforms
// 	platforms := make([]string, len(post.Platforms()))
// 	for i, p := range post.Platforms() {
// 		platforms[i] = string(p)
// 	}

// 	return &PostDTO{
// 		ID:          post.ID(),
// 		TeamID:      post.TeamID(),
// 		AuthorID:    post.AuthorID(),
// 		Content:     post.Content(),
// 		Platforms:   platforms,
// 		Attachments: post.Attachments(),
// 		Status:      string(post.Status()),
// 		ScheduledAt: post.ScheduledAt(),
// 		PublishedAt: post.PublishedAt(),
// 		CreatedAt:   post.CreatedAt(),
// 		UpdatedAt:   post.UpdatedAt(),
// 	}
// }

// func MapPostsToDTO(posts []*postDomain.Post) []*PostDTO {
// 	dtos := make([]*PostDTO, len(posts))
// 	for i, post := range posts {
// 		dtos[i] = MapPostToDTO(post)
// 	}
// 	return dtos
// }

// // ============================================================================
// // SOCIAL ACCOUNT DTOs
// // ============================================================================

// type SocialAccountDTO struct {
// 	ID             uuid.UUID  `json:"id"`
// 	TeamID         uuid.UUID  `json:"teamId"`
// 	Platform       string     `json:"platform"`
// 	Username       string     `json:"username"`
// 	DisplayName    string     `json:"displayName"`
// 	AvatarURL      *string    `json:"avatarUrl,omitempty"`
// 	Status         string     `json:"status"`
// 	ConnectedAt    time.Time  `json:"connectedAt"`
// 	LastSyncedAt   *time.Time `json:"lastSyncedAt,omitempty"`
// 	TokenExpiresAt *time.Time `json:"tokenExpiresAt,omitempty"`
// }

// func MapSocialAccountToDTO(account *socialDomain.Account) *SocialAccountDTO {
// 	if account == nil {
// 		return nil
// 	}
// 	return &SocialAccountDTO{
// 		ID:           account.ID(),
// 		TeamID:       account.TeamID(),
// 		Platform:     string(account.Platform()),
// 		Username:     account.Username(),
// 		DisplayName:  account.DisplayName(),
// 		AvatarURL:    account.AvatarURL(),
// 		Status:       string(account.Status()),
// 		ConnectedAt:  account.ConnectedAt(),
// 		LastSyncedAt: account.LastSyncedAt(),
// 	}
// }

// func MapSocialAccountsToDTO(accounts []*socialDomain.Account) []*SocialAccountDTO {
// 	dtos := make([]*SocialAccountDTO, len(accounts))
// 	for i, account := range accounts {
// 		dtos[i] = MapSocialAccountToDTO(account)
// 	}
// 	return dtos
// }

// // ============================================================================
// // PAGINATION DTO
// // ============================================================================

// type PaginatedResponse struct {
// 	Data       interface{} `json:"data"`
// 	Page       int         `json:"page"`
// 	PerPage    int         `json:"perPage"`
// 	Total      int         `json:"total"`
// 	TotalPages int         `json:"totalPages"`
// }

// func NewPaginatedResponse(data interface{}, page, perPage, total int) *PaginatedResponse {
// 	totalPages := (total + perPage - 1) / perPage
// 	return &PaginatedResponse{
// 		Data:       data,
// 		Page:       page,
// 		PerPage:    perPage,
// 		Total:      total,
// 		TotalPages: totalPages,
// 	}
// }
