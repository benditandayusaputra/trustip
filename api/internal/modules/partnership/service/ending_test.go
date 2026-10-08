package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
)

const consignorID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"

type fakeEnder struct {
	state    domain.EndState
	cutoffs  []time.Time
	ended    []domain.Ending
	listings []domain.ListingChange
}

func (f *fakeEnder) LockForEnding(_ context.Context, key domain.EndKey) (domain.EndState, error) {
	f.cutoffs = append(f.cutoffs, key.ConfirmCutoff)
	if key.PartnershipID != f.state.PartnershipID ||
		(key.OrganizationID != f.state.StoreID && key.OrganizationID != f.state.ConsignorID) {
		return domain.EndState{}, apperror.OutsideScope()
	}
	return f.state, nil
}

func (f *fakeEnder) SaveEnding(_ context.Context, ending domain.Ending) error {
	f.ended = append(f.ended, ending)
	f.state.Status = domain.Ended
	return nil
}

func (f *fakeEnder) SetShownOnPublic(_ context.Context, change domain.ListingChange) error {
	f.listings = append(f.listings, change)
	return nil
}

type endingEvents struct {
	settled []string
	seals   []domain.SealRequest
	notices []domain.Notice
	audits  []audit.Entry
}

func (e *endingEvents) SettleOutstanding(_ context.Context, partnershipID string, _ time.Time) error {
	e.settled = append(e.settled, partnershipID)
	return nil
}

func (e *endingEvents) Seal(_ context.Context, request domain.SealRequest) error {
	e.seals = append(e.seals, request)
	return nil
}

func (e *endingEvents) Announce(_ context.Context, notice domain.Notice) error {
	e.notices = append(e.notices, notice)
	return nil
}

func (e *endingEvents) Record(_ context.Context, entry audit.Entry) error {
	e.audits = append(e.audits, entry)
	return nil
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 11, 20, 3, 0, 0, 0, time.UTC)
}

type endingRun struct {
	ender  *fakeEnder
	events *endingEvents
	deps   service.EndDeps
}

func endingScenario() endingRun {
	ender := &fakeEnder{state: domain.EndState{
		PartnershipID: partnershipID, Status: domain.Active, StoreID: storeID, ConsignorID: consignorID,
		StoreName: "Warung Bu Sri", ConsignorName: "Roti Pak Anton",
	}}
	recorded := &endingEvents{}
	return endingRun{ender: ender, events: recorded, deps: service.EndDeps{
		Partnerships: ender, Reader: &fakeReader{}, Settler: recorded, Sealer: recorded, Announcer: recorded,
		Audit: recorded, Tx: inlineTx{}, Clock: fixedClock{},
	}}
}

func owner(organizationID string, organizationType actor.OrganizationType) actor.Actor {
	return actor.Actor{UserID: "0192f0c4-0000-7000-8000-000000000601", Membership: actor.Membership{
		OrganizationID: organizationID, OrganizationType: organizationType, Role: actor.Owner,
	}}
}

func TestEndPartnership_BRPRT03_SettlesFirstThenEndsSealsAndTellsThePartner(t *testing.T) {
	run := endingScenario()
	ender, recorded, deps := run.ender, run.events, run.deps

	ended, err := service.NewEndPartnership(deps).Execute(t.Context(), owner(storeID, actor.Store),
		service.EndCommand{PartnershipID: partnershipID, Reason: "Toko tutup"})

	if err != nil || ended.ID != partnershipID || len(recorded.settled) != 1 || len(ender.ended) != 1 {
		t.Fatalf("ended = %+v, err = %v, settled = %v", ended, err, recorded.settled)
	}
	if recorded.seals[0].Event != "PARTNERSHIP_ENDED" || recorded.notices[0].OrganizationID != consignorID ||
		recorded.audits[0].Action != audit.PartnershipEnded ||
		!ender.cutoffs[0].Equal(fixedClock{}.Now().Add(-48*time.Hour)) {
		t.Fatalf("events = %+v, cutoffs = %v", recorded, ender.cutoffs)
	}
}

func TestEndPartnership_BRPRT02_BlockersStopTheEnding(t *testing.T) {
	run := endingScenario()
	ender, deps := run.ender, run.deps
	ender.state.Blockers = []domain.Blocker{{Kind: domain.UnpaidSettlement, DocumentNumber: "ST-202611-0001"}}

	_, err := service.NewEndPartnership(deps).Execute(t.Context(), owner(consignorID, actor.Consignor),
		service.EndCommand{PartnershipID: partnershipID})

	appErr, _ := errors.AsType[*apperror.Error](err)
	if appErr == nil || appErr.Code() != apperror.CodePartnershipBlocked || len(ender.ended) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestEndPartnership_RSEC36_StaffCannotEndAPartnership(t *testing.T) {
	run := endingScenario()
	recorded, deps := run.events, run.deps

	_, err := service.NewEndPartnership(deps).Execute(t.Context(), staff(storeID),
		service.EndCommand{PartnershipID: partnershipID})

	if appErr, _ := errors.AsType[*apperror.Error](err); appErr == nil || appErr.Code() != apperror.CodeForbidden ||
		len(recorded.settled) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestGetBlockers_BRPRT02_ListsWhatStandsInTheWay(t *testing.T) {
	run := endingScenario()
	ender, deps := run.ender, run.deps
	ender.state.Blockers = []domain.Blocker{{Kind: domain.StockAtStore, Qty: 4}}

	blockers, err := service.NewGetBlockers(deps).Execute(t.Context(), owner(storeID, actor.Store), partnershipID)

	if err != nil || len(blockers) != 1 || blockers[0].Describe() != "stok 4 unit masih di toko" {
		t.Fatalf("blockers = %+v, err = %v", blockers, err)
	}
}

func TestSetShownOnPublic_FRPRT03_OnlyTheStoreChoosesToBeListed(t *testing.T) {
	run := endingScenario()
	ender, deps := run.ender, run.deps
	setter := service.NewSetShownOnPublic(deps)

	_, storeErr := setter.Execute(t.Context(), owner(storeID, actor.Store),
		service.ListingCommand{PartnershipID: partnershipID, ShowOnPublic: true})
	_, consignorErr := setter.Execute(t.Context(), owner(consignorID, actor.Consignor),
		service.ListingCommand{PartnershipID: partnershipID, ShowOnPublic: true})

	if storeErr != nil || len(ender.listings) != 1 || !ender.listings[0].ShowOnPublic {
		t.Fatalf("store = %v, listings = %+v", storeErr, ender.listings)
	}
	if appErr, _ := errors.AsType[*apperror.Error](consignorErr); appErr == nil ||
		appErr.Code() != apperror.CodeForbidden {
		t.Fatalf("consignor = %v", consignorErr)
	}
}
