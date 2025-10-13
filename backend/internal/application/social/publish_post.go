// ============================================================================
// FILE: backend/internal/application/social/publish_post.go
// ============================================================================
package social

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/adapters/social"
	"github.com/techappsUT/social-queue/internal/application/common"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type PublishPostInput struct {
	AccountID uuid.UUID `json:"accountId" validate:"required"`
	UserID    uuid.UUID `json:"userId" validate:"required"`
	Content   string    `json:"content" validate:"required"`
	MediaURLs []string  `json:"mediaUrls,omitempty"`
}

type PublishPostOutput struct {
	PlatformPostID string    `json:"platformPostId"`
	URL            string    `json:"url"`
	PublishedAt    time.Time `json:"publishedAt"`
}

type PublishPostUseCase struct {
	socialRepo socialDomain.AccountRepository // FIXED
	memberRepo team.MemberRepository
	adapters   map[socialDomain.Platform]social.Adapter
	logger     common.Logger
}

func NewPublishPostUseCase(
	socialRepo socialDomain.AccountRepository, // FIXED
	memberRepo team.MemberRepository,
	adapters map[socialDomain.Platform]social.Adapter,
	logger common.Logger,
) *PublishPostUseCase {
	return &PublishPostUseCase{
		socialRepo: socialRepo,
		memberRepo: memberRepo,
		adapters:   adapters,
		logger:     logger,
	}
}

func (uc *PublishPostUseCase) Execute(ctx context.Context, input PublishPostInput) (*PublishPostOutput, error) {
	// 1. Get account
	account, err := uc.socialRepo.FindByID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	// 2. Verify user is team member
	_, err = uc.memberRepo.FindMember(ctx, account.TeamID(), input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied")
	}

	// 3. Check account is active
	if account.Status() != socialDomain.StatusActive {
		return nil, fmt.Errorf("account is not active")
	}

	// 4. Get adapter
	adapter, ok := uc.adapters[account.Platform()]
	if !ok {
		return nil, fmt.Errorf("unsupported platform")
	}

	// 5. Prepare token
	credentials := account.Credentials()
	token := &social.Token{
		AccessToken:    credentials.AccessToken,
		RefreshToken:   credentials.RefreshToken,
		ExpiresAt:      credentials.ExpiresAt,
		Scopes:         credentials.Scope,
		PlatformUserID: credentials.PlatformUserID,
	}

	// 6. Prepare content
	content := &social.PostContent{
		Text:      input.Content,
		MediaURLs: input.MediaURLs,
	}

	// 7. Publish to platform
	result, err := adapter.PublishPost(ctx, token, content)
	if err != nil {
		uc.logger.Error("Failed to publish post",
			"accountId", input.AccountID,
			"platform", account.Platform(),
			"error", err)
		return nil, fmt.Errorf("failed to publish: %w", err)
	}

	uc.logger.Info("Post published",
		"accountId", input.AccountID,
		"platform", account.Platform(),
		"postId", result.PlatformPostID)

	return &PublishPostOutput{
		PlatformPostID: result.PlatformPostID,
		URL:            result.URL,
		PublishedAt:    result.PublishedAt,
	}, nil
}
