package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

func TestConsumeQuota_BRPRP06_RSEC44_DatabaseStopsTheEleventhProposalOfTheDay(t *testing.T) {
	scene := newWorld(t)
	key := domain.QuotaKeyOf(consignorID, base())

	for range domain.DailyNewProposalLimit {
		scene.must(scene.proposals.ConsumeQuota(t.Context(), key))
	}
	err := scene.proposals.ConsumeQuota(t.Context(), key)
	nextDay := scene.proposals.ConsumeQuota(t.Context(), domain.QuotaKeyOf(consignorID, base().Add(24*time.Hour)))

	if !errors.Is(err, domain.ErrQuotaExhausted) || nextDay != nil {
		t.Fatalf("err = %v, next day = %v", err, nextDay)
	}
}

func TestGetDirectParties_FRPRP01_OnlyListedActiveTargets(t *testing.T) {
	scene := newWorld(t)
	pair := domain.DirectPair{SenderID: consignorID, TargetID: storeID}

	_, unlisted := scene.proposals.GetDirectParties(t.Context(), pair)
	scene.must(scene.queries.SetOrganizationListed(t.Context(), sqlc.SetOrganizationListedParams{
		ID: storeID, IsListed: true, UpdatedAt: base(),
	}))
	parties, err := scene.proposals.GetDirectParties(t.Context(), pair)

	if apperror.From(unlisted).Code() != apperror.CodeNotFound || err != nil ||
		parties.Target.Type != "STORE" || parties.SenderName == "" || parties.ArePartners ||
		!parties.LastDeclinedAt.IsZero() {
		t.Fatalf("unlisted = %v, parties = %+v, err = %v", unlisted, parties, err)
	}
}

func TestGetDirectParties_BRPRP06_RemembersTheLastDecline(t *testing.T) {
	scene := newWorld(t)
	scene.must(scene.queries.SetOrganizationListed(t.Context(), sqlc.SetOrganizationListedParams{
		ID: storeID, IsListed: true, UpdatedAt: base(),
	}))
	created := invitation(1, consignorID, breadID)
	scene.invite(created)
	scene.must(scene.proposals.ClaimInvitation(t.Context(), domain.Claim{
		ProposalID: created.IDs.ProposalID, ConsignorID: consignorID, StoreID: storeID, ClaimerID: storeID, At: base(),
	}))
	declinedAt := base().Add(time.Hour)
	scene.must(scene.proposals.SaveDecision(t.Context(), domain.Decision{
		ProposalID: created.IDs.ProposalID, Status: domain.Declined, DecidedBy: userID, Reason: "Rak penuh.",
		RevisionNo: 1, At: declinedAt,
	}))

	parties, err := scene.proposals.GetDirectParties(t.Context(), domain.DirectPair{
		SenderID: consignorID, TargetID: storeID,
	})

	if err != nil || !parties.LastDeclinedAt.Equal(declinedAt) {
		t.Fatalf("parties = %+v, err = %v", parties, err)
	}
}

func TestConsumeQuota_BRAUTHZ06_SuspendedOrganizationCannotSend(t *testing.T) {
	scene := newWorld(t)
	_, err := scene.queries.SetOrganizationStatus(t.Context(), sqlc.SetOrganizationStatusParams{
		ID: consignorID, Status: "SUSPENDED", UpdatedAt: base(),
	})
	scene.must(err)

	suspended := scene.proposals.ConsumeQuota(t.Context(), domain.QuotaKeyOf(consignorID, base()))
	active := scene.proposals.ConsumeQuota(t.Context(), domain.QuotaKeyOf(storeID, base()))

	if !errors.Is(suspended, domain.ErrSenderSuspended) || active != nil {
		t.Fatalf("suspended = %v, active = %v", suspended, active)
	}
}
