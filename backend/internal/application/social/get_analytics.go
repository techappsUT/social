// ============================================================================
// FILE: backend/internal/application/social/get_analytics.go
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

type GetAnalyticsInput struct {
	AccountID uuid.UUID `json:"accountId" validate:"required"`
	PostID    string    `json:"postId" validate:"required"`
	UserID    uuid.UUID `json:"userId" validate:"required"`
}

type GetAnalyticsOutput struct {
	Analytics *AnalyticsDTO `json:"analytics"`
}

type GetAnalyticsUseCase struct {
	socialRepo socialDomain.AccountRepository // FIXED
	memberRepo team.MemberRepository
	adapters   map[socialDomain.Platform]social.Adapter
	cache      common.CacheService
	logger     common.Logger
}

func NewGetAnalyticsUseCase(
	socialRepo socialDomain.AccountRepository, // FIXED
	memberRepo team.MemberRepository,
	adapters map[socialDomain.Platform]social.Adapter,
	cache common.CacheService,
	logger common.Logger,
) *GetAnalyticsUseCase {
	return &GetAnalyticsUseCase{
		socialRepo: socialRepo,
		memberRepo: memberRepo,
		adapters:   adapters,
		cache:      cache,
		logger:     logger,
	}
}

func (uc *GetAnalyticsUseCase) Execute(ctx context.Context, input GetAnalyticsInput) (*GetAnalyticsOutput, error) {
	// 1. Get account
	account, err := uc.socialRepo.FindByID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	// 2. Verify access
	_, err = uc.memberRepo.FindMember(ctx, account.TeamID(), input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied")
	}

	// 3. Check cache first
	cacheKey := fmt.Sprintf("analytics:%s:%s", input.AccountID, input.PostID)
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if dto, ok := cached.(*AnalyticsDTO); ok {
			return &GetAnalyticsOutput{Analytics: dto}, nil
		}
	}

	// 4. Get adapter
	adapter, ok := uc.adapters[account.Platform()]
	if !ok {
		return nil, fmt.Errorf("unsupported platform")
	}

	// 5. Fetch from platform
	credentials := account.Credentials()
	token := &social.Token{
		AccessToken:    credentials.AccessToken,
		ExpiresAt:      credentials.ExpiresAt,
		PlatformUserID: credentials.PlatformUserID,
	}

	analytics, err := adapter.GetPostAnalytics(ctx, token, input.PostID)
	if err != nil {
		uc.logger.Error("Failed to fetch analytics", "accountId", input.AccountID, "error", err)
		return nil, fmt.Errorf("failed to fetch analytics")
	}

	// 6. Convert to DTO
	analyticsDTO := &AnalyticsDTO{
		Impressions: analytics.Impressions,
		Engagements: analytics.Engagements,
		Likes:       analytics.Likes,
		Shares:      analytics.Shares,
		Comments:    analytics.Comments,
		Clicks:      analytics.Clicks,
	}

	// 7. Cache result (15 minutes)
	uc.cache.Set(ctx, cacheKey, analyticsDTO, 15*time.Minute)

	return &GetAnalyticsOutput{Analytics: analyticsDTO}, nil
}
