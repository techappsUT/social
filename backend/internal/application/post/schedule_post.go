// ============================================================================
// FILE 2: backend/internal/application/post/schedule_post.go
// ============================================================================
package post

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	postDomain "github.com/techappsUT/social-queue/internal/domain/post"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type SchedulePostInput struct {
	PostID      uuid.UUID `json:"postId" validate:"required"`
	UserID      uuid.UUID `json:"userId" validate:"required"`
	ScheduledAt time.Time `json:"scheduledAt" validate:"required"`
	Timezone    string    `json:"timezone,omitempty"`
}

type SchedulePostOutput struct {
	Post *PostDTO `json:"post"`
}

type SchedulePostUseCase struct {
	postRepo   postDomain.Repository
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewSchedulePostUseCase(
	postRepo postDomain.Repository,
	memberRepo team.MemberRepository,
	logger common.Logger,
) *SchedulePostUseCase {
	return &SchedulePostUseCase{
		postRepo:   postRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *SchedulePostUseCase) Execute(ctx context.Context, input SchedulePostInput) (*SchedulePostOutput, error) {
	// 1. Get post
	post, err := uc.postRepo.FindByID(ctx, input.PostID)
	if err != nil {
		return nil, postDomain.ErrPostNotFound
	}

	// 2. Check authorization (author or admin)
	member, err := uc.memberRepo.FindMember(ctx, post.TeamID(), input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	canSchedule := post.CreatedBy() == input.UserID ||
		member.Role() == team.MemberRoleOwner ||
		member.Role() == team.MemberRoleAdmin

	if !canSchedule {
		return nil, fmt.Errorf("access denied: cannot schedule this post")
	}

	// 3. Validate schedule time
	if input.ScheduledAt.Before(time.Now()) {
		return nil, postDomain.ErrScheduleTimeInPast
	}

	// 4. Schedule the post
	if err := post.Schedule(input.ScheduledAt); err != nil {
		return nil, err
	}

	// 5. Save changes
	if err := uc.postRepo.Update(ctx, post); err != nil {
		uc.logger.Error("Failed to schedule post", "postId", input.PostID, "error", err)
		return nil, fmt.Errorf("failed to update post")
	}

	uc.logger.Info("Post scheduled", "postId", input.PostID, "scheduledAt", input.ScheduledAt)

	return &SchedulePostOutput{
		Post: MapPostToDTO(post),
	}, nil
}
