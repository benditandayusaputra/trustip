package service_test

import (
	"context"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
)

type fakeAccounts map[string]domain.Account

func (f fakeAccounts) GetAccount(_ context.Context, userID string) (domain.Account, error) {
	account, ok := f[userID]
	if !ok {
		return domain.Account{}, apperror.NotFound()
	}
	return account, nil
}

func TestGetAccount_MissingUser_IsNotFound(t *testing.T) {
	useCase := service.NewGetAccount(fakeAccounts{})

	_, err := useCase.Execute(t.Context(), actor.Actor{UserID: knownUserID})

	if errorCode(err) != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestGetAccount_FRAUTH06_ReturnsTheSignedInUsersOwnAccount(t *testing.T) {
	sari := domain.Account{ID: knownUserID, Email: "sari@contoh.id", FullName: "Sari Wulandari"}
	useCase := service.NewGetAccount(fakeAccounts{knownUserID: sari, "other": {ID: "other"}})

	account, err := useCase.Execute(t.Context(), actor.Actor{UserID: knownUserID, SessionFamily: "family-1"})

	if err != nil || account != sari {
		t.Fatalf("account = %+v, err = %v", account, err)
	}
}

func TestGetAccount_FRAUTH05_ReportsTheMembershipOfTheSession(t *testing.T) {
	sari := domain.Account{ID: knownUserID, Email: "sari@contoh.id", FullName: "Sari Wulandari"}
	signedIn := actor.Actor{UserID: knownUserID, SessionFamily: "family-1", Membership: storeOwnerMembership()}

	account, err := service.NewGetAccount(fakeAccounts{knownUserID: sari}).Execute(t.Context(), signedIn)

	if err != nil || account.Membership != storeOwnerMembership() {
		t.Fatalf("account = %+v, err = %v", account, err)
	}
}
