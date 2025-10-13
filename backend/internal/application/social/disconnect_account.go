// ============================================================================
// FILE: backend/internal/application/social/disconnect_account.go
// ============================================================================
package social

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/common"
	socialDomain "github.com/techappsUT/social-queue/internal/domain/social"
	"github.com/techappsUT/social-queue/internal/domain/team"
)

type DisconnectAccountInput struct {
	AccountID uuid.UUID `json:"accountId" validate:"required"`
	UserID    uuid.UUID `json:"userId" validate:"required"`
}

type DisconnectAccountUseCase struct {
	socialRepo socialDomain.AccountRepository // FIXED
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewDisconnectAccountUseCase(
	socialRepo socialDomain.AccountRepository, // FIXED
	memberRepo team.MemberRepository,
	logger common.Logger,
) *DisconnectAccountUseCase {
	return &DisconnectAccountUseCase{
		socialRepo: socialRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *DisconnectAccountUseCase) Execute(ctx context.Context, input DisconnectAccountInput) error {
	// 1. Get account
	account, err := uc.socialRepo.FindByID(ctx, input.AccountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	// 2. Verify user is team member
	_, err = uc.memberRepo.FindMember(ctx, account.TeamID(), input.UserID)
	if err != nil {
		return fmt.Errorf("access denied: not a team member")
	}

	// 3. Soft delete account
	if err := uc.socialRepo.Delete(ctx, input.AccountID); err != nil {
		uc.logger.Error("Failed to disconnect account", "accountId", input.AccountID, "error", err)
		return fmt.Errorf("failed to disconnect account")
	}

	uc.logger.Info("Social account disconnected", "accountId", input.AccountID)
	return nil
}
