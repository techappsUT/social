// ============================================================================
// FILE 7: backend/internal/application/post/publish_now.go
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

type PublishNowInput struct {
	PostID uuid.UUID `json:"postId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type PublishNowOutput struct {
	Post *PostDTO `json:"post"`
}

type PublishNowUseCase struct {
	postRepo   postDomain.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewPublishNowUseCase(
	postRepo postDomain.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *PublishNowUseCase {
	return &PublishNowUseCase{
		postRepo:   postRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *PublishNowUseCase) Execute(ctx context.Context, input PublishNowInput) (*PublishNowOutput, error) {
	// 1. Get post
	post, err := uc.postRepo.FindByID(ctx, input.PostID)
	if err != nil {
		return nil, postDomain.ErrPostNotFound
	}

	// 2. Check authorization
	member, err := uc.memberRepo.FindMember(ctx, post.TeamID(), input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	canPublish := post.CreatedBy() == input.UserID ||
		member.Role() == team.MemberRoleOwner ||
		member.Role() == team.MemberRoleAdmin

	if !canPublish {
		return nil, fmt.Errorf("access denied: cannot publish this post")
	}

	// 3. Validate post can be published
	if post.Status() == postDomain.StatusPublished {
		return nil, fmt.Errorf("post is already published")
	}

	if post.Content().Text == "" {
		return nil, postDomain.ErrEmptyContent
	}

	// 4. Queue for immediate publishing
	if err := post.Queue(); err != nil {
		return nil, err
	}

	// 5. Save changes
	if err := uc.postRepo.Update(ctx, post); err != nil {
		uc.logger.Error("Failed to queue post for publishing", "postId", input.PostID, "error", err)
		return nil, fmt.Errorf("failed to queue post")
	}

	uc.logger.Info("Post queued for immediate publishing", "postId", input.PostID)

	return &PublishNowOutput{
		Post: MapPostToDTO(post),
	}, nil
}
