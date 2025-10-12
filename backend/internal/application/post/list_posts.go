// ============================================================================
// FILE 6: backend/internal/application/post/list_posts.go
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

type ListPostsInput struct {
	TeamID uuid.UUID          `json:"teamId" validate:"required"`
	UserID uuid.UUID          `json:"userId" validate:"required"`
	Status *postDomain.Status `json:"status,omitempty"`
	Offset int                `json:"offset"`
	Limit  int                `json:"limit"`
}

type ListPostsOutput struct {
	Posts []PostDTO `json:"posts"`
	Total int       `json:"total"`
}

type ListPostsUseCase struct {
	postRepo   postDomain.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewListPostsUseCase(
	postRepo postDomain.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *ListPostsUseCase {
	return &ListPostsUseCase{
		postRepo:   postRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *ListPostsUseCase) Execute(ctx context.Context, input ListPostsInput) (*ListPostsOutput, error) {
	// 1. Check user is team member
	isMember, err := uc.memberRepo.IsMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	// 2. Set default pagination
	if input.Limit == 0 {
		input.Limit = 20
	}

	// 3. Get posts
	var posts []*postDomain.Post
	if input.Status != nil {
		posts, err = uc.postRepo.FindByStatus(ctx, *input.Status, input.Offset, input.Limit)
	} else {
		posts, err = uc.postRepo.FindByTeamID(ctx, input.TeamID, input.Offset, input.Limit)
	}

	if err != nil {
		uc.logger.Error("Failed to list posts", "teamId", input.TeamID, "error", err)
		return nil, fmt.Errorf("failed to list posts")
	}

	// 4. Map to DTOs
	postDTOs := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		postDTOs = append(postDTOs, *MapPostToDTO(p))
	}

	// 5. Get total count
	total, err := uc.postRepo.CountByTeamID(ctx, input.TeamID)
	if err != nil {
		total = int64(len(postDTOs))
	}

	return &ListPostsOutput{
		Posts: postDTOs,
		Total: int(total),
	}, nil
}
