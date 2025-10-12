// ============================================================================
// FILE 3: backend/internal/application/post/update_post.go
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

type UpdatePostInput struct {
	PostID      uuid.UUID             `json:"postId" validate:"required"`
	UserID      uuid.UUID             `json:"userId" validate:"required"`
	Content     *string               `json:"content,omitempty"`
	Platforms   []postDomain.Platform `json:"platforms,omitempty"`
	Attachments []string              `json:"attachments,omitempty"`
}

type UpdatePostOutput struct {
	Post *PostDTO `json:"post"`
}

type UpdatePostUseCase struct {
	postRepo   postDomain.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewUpdatePostUseCase(
	postRepo postDomain.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *UpdatePostUseCase {
	return &UpdatePostUseCase{
		postRepo:   postRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *UpdatePostUseCase) Execute(ctx context.Context, input UpdatePostInput) (*UpdatePostOutput, error) {
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

	canEdit := post.CreatedBy() == input.UserID ||
		member.Role() == team.MemberRoleOwner ||
		member.Role() == team.MemberRoleAdmin

	if !canEdit {
		return nil, fmt.Errorf("access denied: cannot edit this post")
	}

	// 3. Check if post can be edited
	if post.Status() == postDomain.StatusPublished {
		return nil, postDomain.ErrCannotEditPublished
	}

	// 4. Update content if provided
	if input.Content != nil {
		newContent := post.Content()
		newContent.Text = *input.Content
		if len(input.Attachments) > 0 {
			newContent.MediaURLs = input.Attachments
		}
		if err := post.UpdateContent(newContent); err != nil {
			return nil, err
		}
	}

	// 5. Update platforms if provided
	if len(input.Platforms) > 0 {
		if err := post.UpdatePlatforms(input.Platforms); err != nil {
			return nil, err
		}
	}

	// 6. Save changes
	if err := uc.postRepo.Update(ctx, post); err != nil {
		uc.logger.Error("Failed to update post", "postId", input.PostID, "error", err)
		return nil, fmt.Errorf("failed to update post")
	}

	uc.logger.Info("Post updated", "postId", input.PostID)

	return &UpdatePostOutput{
		Post: MapPostToDTO(post),
	}, nil
}
