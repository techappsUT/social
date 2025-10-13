// ============================================================================
// FILE: backend/internal/application/social/get_analytics.go
// COMPLETE FIX - Correct imports and type mapping
// ============================================================================
package social

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	socialAdapter "github.com/techappsUT/social-queue/internal/adapters/social" // FIX: Import adapter package
	"github.com/techappsUT/social-queue/internal/application/common"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
)

type GetAnalyticsInput struct {
	AccountID uuid.UUID `json:"accountId" validate:"required"`
	PostID    string    `json:"postId" validate:"required"`
	UserID    uuid.UUID `json:"userId" validate:"required"` // For authorization
}

type GetAnalyticsOutput struct {
	Analytics *AnalyticsDTO `json:"analytics"`
}

type GetAnalyticsUseCase struct {
	accountRepo socialDomain.AccountRepository
	adapters    map[socialDomain.Platform]socialAdapter.Adapter // FIX: Use adapter.Adapter type
	cache       common.CacheService
	logger      common.Logger
}

func NewGetAnalyticsUseCase(
	accountRepo socialDomain.AccountRepository,
	adapters map[socialDomain.Platform]socialAdapter.Adapter, // FIX: Use adapter.Adapter type
	cache common.CacheService,
	logger common.Logger,
) *GetAnalyticsUseCase {
	return &GetAnalyticsUseCase{
		accountRepo: accountRepo,
		adapters:    adapters,
		cache:       cache,
		logger:      logger,
	}
}

func (uc *GetAnalyticsUseCase) Execute(ctx context.Context, input GetAnalyticsInput) (*GetAnalyticsOutput, error) {
	// 1. Build cache key
	cacheKey := fmt.Sprintf("analytics:%s:%s", input.AccountID.String(), input.PostID)

	// 2. Check cache (cache stores JSON string)
	cached, err := uc.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		// Unmarshal the JSON string to AnalyticsDTO
		var analyticsDTO AnalyticsDTO
		if err := json.Unmarshal([]byte(cached), &analyticsDTO); err == nil {
			uc.logger.Debug("Analytics cache hit", "accountId", input.AccountID)
			return &GetAnalyticsOutput{Analytics: &analyticsDTO}, nil
		}
		// If unmarshal fails, continue to fetch fresh data
		uc.logger.Warn("Failed to unmarshal cached analytics", "error", err)
	}

	// 3. Get account
	account, err := uc.accountRepo.FindByID(ctx, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	// 4. Authorization check
	if account.UserID() != input.UserID {
		return nil, fmt.Errorf("unauthorized access to account")
	}

	// 5. Get adapter
	adapter, ok := uc.adapters[account.Platform()]
	if !ok {
		return nil, fmt.Errorf("unsupported platform")
	}

	// 6. Fetch from platform
	credentials := account.Credentials()

	// FIX: Use socialAdapter.Token (from adapters/social package)
	token := &socialAdapter.Token{
		AccessToken:    credentials.AccessToken,
		RefreshToken:   credentials.RefreshToken,
		ExpiresAt:      credentials.ExpiresAt,
		Scopes:         credentials.Scope,
		PlatformUserID: credentials.PlatformUserID,
	}

	// FIX: GetPostAnalytics returns *socialAdapter.Analytics
	analytics, err := adapter.GetPostAnalytics(ctx, token, input.PostID)
	if err != nil {
		uc.logger.Error("Failed to fetch analytics", "accountId", input.AccountID, "error", err)
		return nil, fmt.Errorf("failed to fetch analytics")
	}

	// 7. Convert adapter Analytics to DTO
	// FIX: Adapter returns Analytics with Engagements (int)
	analyticsDTO := &AnalyticsDTO{
		Impressions: analytics.Impressions,
		Engagements: analytics.Engagements, // FIX: Use Engagements from adapter
		Likes:       analytics.Likes,
		Shares:      analytics.Shares,
		Comments:    analytics.Comments,
		Clicks:      analytics.Clicks,
	}

	// 8. Marshal to JSON string before caching
	analyticsJSON, err := json.Marshal(analyticsDTO)
	if err != nil {
		uc.logger.Warn("Failed to marshal analytics for cache", "error", err)
	} else {
		// Cache result (15 minutes)
		if err := uc.cache.Set(ctx, cacheKey, string(analyticsJSON), 15*time.Minute); err != nil {
			uc.logger.Warn("Failed to cache analytics", "error", err)
		}
	}

	return &GetAnalyticsOutput{Analytics: analyticsDTO}, nil
}
