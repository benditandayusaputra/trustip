package service_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/service"
)

type storedAgreement struct{ view domain.View }

func (s storedAgreement) ListVersions(context.Context, domain.PartnershipKey) ([]domain.Agreement, error) {
	return []domain.Agreement{s.view.Agreement}, nil
}

func (s storedAgreement) GetAgreement(_ context.Context, key domain.VersionKey) (domain.View, error) {
	parties := []string{s.view.Consignor.OrganizationID, s.view.Store.OrganizationID}
	isParty := slices.Contains(parties, key.OrganizationID)
	if !isParty || key.PartnershipID != s.view.Agreement.PartnershipID {
		return domain.View{}, apperror.NotFound()
	}
	return s.view, nil
}

type hashChecker struct{ sealedHash []byte }

func (h hashChecker) Check(_ context.Context, check domain.SealCheck) (domain.SealVerdict, error) {
	current, err := canonical.Digest(check.Content)
	if err != nil {
		return domain.SealVerdict{}, fmt.Errorf("digest: %w", err)
	}
	return domain.SealVerdict{IsValid: string(current) == string(h.sealedHash)}, nil
}

func agreementView() domain.View {
	founding := domain.Found(input(), domain.FoundingIDs{PartnershipID: "k1", AgreementID: "a1"})
	return domain.View{
		Agreement: founding.Agreement, Consignor: domain.Party{OrganizationID: "c1", Name: "Roti Pak Anton"},
		Store: domain.Party{OrganizationID: "s1", Name: "Warung Bu Sri"},
	}
}

func member(organizationID string) actor.Actor {
	return actor.Actor{UserID: "u", Membership: actor.Membership{
		OrganizationID: organizationID, OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func TestGetAgreement_FRAGR01_BothPartiesSeeTheSameAgreementAndSeal(t *testing.T) {
	view := agreementView()
	sealedHash, _ := canonical.Digest(view.Agreement.Content())
	reader := service.NewGetAgreement(storedAgreement{view: view}, hashChecker{sealedHash: sealedHash})

	fromStore, storeErr := reader.Execute(t.Context(), member("s1"), service.VersionRequest{PartnershipID: "k1"})
	active := service.VersionRequest{PartnershipID: "k1"}
	fromConsignor, consignorErr := reader.Execute(t.Context(), member("c1"), active)

	if storeErr != nil || consignorErr != nil || !fromStore.Seal.IsValid ||
		fromStore.View.Agreement.ID != fromConsignor.View.Agreement.ID {
		t.Fatalf("store = %+v, consignor = %+v", fromStore, fromConsignor)
	}
}

func TestGetAgreement_RSEC33_ChangedPriceMakesTheSealInvalid(t *testing.T) {
	view := agreementView()
	sealedHash, _ := canonical.Digest(view.Agreement.Content())
	view.Agreement.Terms.PaymentTermsDays = 30

	got, err := service.NewGetAgreement(storedAgreement{view: view}, hashChecker{sealedHash: sealedHash}).
		Execute(t.Context(), member("s1"), service.VersionRequest{PartnershipID: "k1"})

	if err != nil || got.Seal.IsValid {
		t.Fatalf("got = %+v, err = %v", got, err)
	}
}

func TestGetAgreement_BRAUTHZ02_OtherOrganizationGetsNotFound(t *testing.T) {
	_, err := service.NewGetAgreement(storedAgreement{view: agreementView()}, hashChecker{}).
		Execute(t.Context(), member("elsewhere"), service.VersionRequest{PartnershipID: "k1"})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}
