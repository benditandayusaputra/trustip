package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

type AccountReader interface {
	GetAccount(ctx context.Context, userID string) (domain.Account, error)
}

type GetAccount struct {
	accounts AccountReader
}

func NewGetAccount(accounts AccountReader) *GetAccount {
	return &GetAccount{accounts: accounts}
}

func (uc *GetAccount) Execute(ctx context.Context, signedIn actor.Actor) (domain.Account, error) {
	account, err := uc.accounts.GetAccount(ctx, signedIn.UserID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("get account of %s: %w", signedIn.UserID, err)
	}
	return account, nil
}
