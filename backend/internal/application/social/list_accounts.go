// ============================================================================
// FILE: backend/internal/application/social/list_accounts.go
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

type ListAccountsInput struct {
	TeamID uuid.UUID `json:"teamId" validate:"required"`
	UserID uuid.UUID `json:"userId" validate:"required"`
}

type ListAccountsOutput struct {
	Accounts []*SocialAccountDTO `json:"accounts"`
}

type ListAccountsUseCase struct {
	socialRepo socialDomain.AccountRepository // FIXED
	memberRepo team.MemberRepository
	logger     common.Logger
}

func NewListAccountsUseCase(
	socialRepo socialDomain.AccountRepository, // FIXED
	memberRepo team.MemberRepository,
	logger common.Logger,
) *ListAccountsUseCase {
	return &ListAccountsUseCase{
		socialRepo: socialRepo,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

func (uc *ListAccountsUseCase) Execute(ctx context.Context, input ListAccountsInput) (*ListAccountsOutput, error) {
	// 1. Verify user is team member
	_, err := uc.memberRepo.FindMember(ctx, input.TeamID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("access denied: not a team member")
	}

	// 2. Get all accounts for team
	accounts, err := uc.socialRepo.FindByTeamID(ctx, input.TeamID)
	if err != nil {
		uc.logger.Error("Failed to list accounts", "teamId", input.TeamID, "error", err)
		return nil, fmt.Errorf("failed to list accounts")
	}

	// 3. Map to DTOs
	accountDTOs := make([]*SocialAccountDTO, 0, len(accounts))
	for _, account := range accounts {
		accountDTOs = append(accountDTOs, MapAccountToDTO(account))
	}

	return &ListAccountsOutput{Accounts: accountDTOs}, nil
}
