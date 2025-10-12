// ============================================================================
// FILE 1: backend/internal/application/post/create_draft.go
// ============================================================================
package post

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	postDomain "github.com/techappsUT/social-queue/internal/domain/post"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type CreateDraftInput struct {
	TeamID      uuid.UUID             `json:"teamId" validate:"required"`
	AuthorID    uuid.UUID             `json:"authorId" validate:"required"`
	Content     string                `json:"content" validate:"required"`
	Platforms   []postDomain.Platform `json:"platforms" validate:"required,min=1"`
	Attachments []string              `json:"attachments,omitempty"`
}

type CreateDraftOutput struct {
	Post *PostDTO `json:"post"`
}

type CreateDraftUseCase struct {
	postRepo   postDomain.Repository
	teamRepo   team.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewCreateDraftUseCase(
	postRepo postDomain.Repository,
	teamRepo team.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *CreateDraftUseCase {
	return &CreateDraftUseCase{
		postRepo:   postRepo,
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *CreateDraftUseCase) Execute(ctx context.Context, input CreateDraftInput) (*CreateDraftOutput, error) {
	// 1. Validate author is team member
	isMember, err := uc.memberRepo.IsMember(ctx, input.TeamID, input.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("author is not a team member")
	}

	// 2. Validate content
	if input.Content == "" {
		return nil, postDomain.ErrEmptyContent
	}

	// 3. Validate platforms
	if len(input.Platforms) == 0 {
		return nil, postDomain.ErrNoPlatformsSelected
	}

	for _, platform := range input.Platforms {
		if !isValidPlatform(platform) {
			return nil, postDomain.ErrInvalidPlatform
		}
	}

	// 4. Build content
	content := postDomain.Content{
		Text:      input.Content,
		MediaURLs: input.Attachments,
	}

	// 5. Create post entity
	post, err := postDomain.NewPost(input.TeamID, input.AuthorID, content, input.Platforms)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// 6. Save to repository
	if err := uc.postRepo.Create(ctx, post); err != nil {
		uc.logger.Error("Failed to create draft", "error", err)
		return nil, fmt.Errorf("failed to save post")
	}

	uc.logger.Info("Draft created", "postId", post.ID(), "teamId", input.TeamID)

	return &CreateDraftOutput{
		Post: MapPostToDTO(post),
	}, nil
}

func isValidPlatform(p postDomain.Platform) bool {
	validPlatforms := []postDomain.Platform{
		postDomain.PlatformTwitter,
		postDomain.PlatformFacebook,
		postDomain.PlatformLinkedIn,
		postDomain.PlatformInstagram,
	}
	for _, vp := range validPlatforms {
		if p == vp {
			return true
		}
	}
	return false
}
