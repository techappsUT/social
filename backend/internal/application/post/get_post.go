// ============================================================================
// FILE 5: backend/internal/application/post/get_post.go
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

type GetPostInput struct {
	PostID uuid.UUID `json:"postId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type GetPostOutput struct {
	Post *PostDTO `json:"post"`
}

type GetPostUseCase struct {
	postRepo   postDomain.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewGetPostUseCase(
	postRepo postDomain.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *GetPostUseCase {
	return &GetPostUseCase{
		postRepo:   postRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *GetPostUseCase) Execute(ctx context.Context, input GetPostInput) (*GetPostOutput, error) {
	// 1. Get post
	post, err := uc.postRepo.FindByID(ctx, input.PostID)
	if err != nil {
		return nil, postDomain.ErrPostNotFound
	}

	// 2. Check user is team member
	isMember, err := uc.memberRepo.IsMember(ctx, post.TeamID(), input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	return &GetPostOutput{
		Post: MapPostToDTO(post),
	}, nil
}
