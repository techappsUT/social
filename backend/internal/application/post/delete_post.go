// ============================================================================
// FILE 4: backend/internal/application/post/delete_post.go
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

type DeletePostInput struct {
	PostID uuid.UUID `json:"postId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type DeletePostUseCase struct {
	postRepo   postDomain.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewDeletePostUseCase(
	postRepo postDomain.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *DeletePostUseCase {
	return &DeletePostUseCase{
		postRepo:   postRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *DeletePostUseCase) Execute(ctx context.Context, input DeletePostInput) error {
	// 1. Get post
	post, err := uc.postRepo.FindByID(ctx, input.PostID)
	if err != nil {
		return postDomain.ErrPostNotFound
	}

	// 2. Check authorization
	member, err := uc.memberRepo.FindMember(ctx, post.TeamID(), input.UserID)
	if err != nil {
		return fmt.Errorf("access denied: not a team member")
	}

	canDelete := post.CreatedBy() == input.UserID ||
		member.Role() == team.MemberRoleOwner ||
		member.Role() == team.MemberRoleAdmin

	if !canDelete {
		return fmt.Errorf("access denied: cannot delete this post")
	}

	// 3. Cancel if scheduled
	if post.IsScheduled() {
		if err := post.Cancel(); err != nil {
			return err
		}
		if err := uc.postRepo.Update(ctx, post); err != nil {
			return fmt.Errorf("failed to cancel post")
		}
	}

	// 4. Soft delete
	if err := uc.postRepo.Delete(ctx, input.PostID); err != nil {
		uc.logger.Error("Failed to delete post", "postId", input.PostID, "error", err)
		return fmt.Errorf("failed to delete post")
	}

	uc.logger.Info("Post deleted", "postId", input.PostID)

	return nil
}
