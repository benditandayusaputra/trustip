package service_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
)

const (
	partnershipID  = "0192f0c4-0000-7000-8000-000000000201"
	agreementID    = "0192f0c4-0000-7000-8000-000000000301"
	breadID        = "0192f0c4-0000-7000-8000-000000000101"
	storeID        = "0192f0c4-0000-7000-8000-000000000401"
	consignorID    = "0192f0c4-0000-7000-8000-000000000402"
	cashierID      = "0192f0c4-0000-7000-8000-000000000501"
	idempotencyKey = "0192f0c4-0000-7000-8000-000000000999"
)

var errStoreDown = errors.New("store is down")

type fakeDrops struct {
	setting    domain.Setting
	saved      map[string]domain.Drop
	sequences  []domain.SequenceKey
	duplicate  *domain.Drop
	failCreate error
	stockUsed  bool
	reversals  []domain.Movement
}

func (f *fakeDrops) LockDrop(ctx context.Context, key domain.DropKey) error {
	_, err := f.GetDrop(ctx, key)
	return err
}

func (f *fakeDrops) IsStockUsed(context.Context, string) (bool, error) {
	return f.stockUsed, nil
}

func (f *fakeDrops) SaveDecision(_ context.Context, decision domain.Decision, reversal []domain.Movement) error {
	drop := f.saved[decision.DropID]
	if drop.Status != domain.Recorded {
		return domain.ErrNotPending
	}
	drop.Status = decision.Status
	f.saved[decision.DropID] = drop
	f.reversals = append(f.reversals, reversal...)
	return nil
}

func newFakeDrops() *fakeDrops {
	terms := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	return &fakeDrops{
		saved: map[string]domain.Drop{},
		setting: domain.Setting{
			PartnershipID: partnershipID, IsActive: true, AgreementID: agreementID, AgreementVersion: 1,
			StoreID: storeID, StoreName: "Warung Bu Sri", ConsignorID: consignorID,
			Products: map[string]domain.AgreedProduct{
				breadID: {ProductID: breadID, ShelfLife: stock.ShelfLifeOf(3), Terms: terms},
			},
		},
	}
}

func (f *fakeDrops) NextNumber(_ context.Context, key domain.SequenceKey) (int, error) {
	f.sequences = append(f.sequences, key)
	return len(f.sequences), nil
}

func (f *fakeDrops) CreateDrop(_ context.Context, drop domain.Drop) error {
	if f.failCreate != nil {
		return f.failCreate
	}
	if f.duplicate != nil {
		f.saved[f.duplicate.ID] = *f.duplicate
		return domain.ErrDuplicateKey
	}
	f.saved[drop.ID] = drop
	return nil
}

func (f *fakeDrops) GetSetting(_ context.Context, key domain.SettingKey) (domain.Setting, error) {
	if key.StoreID != f.setting.StoreID || key.PartnershipID != f.setting.PartnershipID {
		return domain.Setting{}, apperror.OutsideScope()
	}
	return f.setting, nil
}

func (f *fakeDrops) FindByKey(_ context.Context, key domain.IdempotencyKey) (domain.Stored, bool, error) {
	for _, drop := range f.saved {
		if drop.StoreID == key.StoreID && drop.IdempotencyKey == key.Key {
			return domain.Stored{DropID: drop.ID, Fingerprint: drop.Fingerprint}, true, nil
		}
	}
	return domain.Stored{}, false, nil
}

func (f *fakeDrops) GetDrop(_ context.Context, key domain.DropKey) (domain.View, error) {
	drop, ok := f.saved[key.DropID]
	if !ok || (key.OrganizationID != drop.StoreID && key.OrganizationID != drop.ConsignorID) {
		return domain.View{}, apperror.OutsideScope()
	}
	return domain.View{Drop: drop, StoreName: "Warung Bu Sri", ConsignorName: "Roti Pak Anton"}, nil
}

type recorder struct {
	seals   []domain.SealRequest
	notices []domain.Notice
	audits  []audit.Entry
	checks  []domain.SealCheck
}

func (r *recorder) Seal(_ context.Context, request domain.SealRequest) error {
	r.seals = append(r.seals, request)
	return nil
}

func (r *recorder) Check(_ context.Context, check domain.SealCheck) (domain.SealVerdict, error) {
	r.checks = append(r.checks, check)
	return domain.SealVerdict{IsValid: true}, nil
}

func (r *recorder) Announce(_ context.Context, notice domain.Notice) error {
	r.notices = append(r.notices, notice)
	return nil
}

func (r *recorder) Record(_ context.Context, entry audit.Entry) error {
	r.audits = append(r.audits, entry)
	return nil
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

type movingClock struct{ at *time.Time }

func (c movingClock) Now() time.Time { return *c.at }

type sequentialIDs struct{ next int }

func (s *sequentialIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

type world struct {
	drops    *fakeDrops
	recorder *recorder
	deps     service.Deps
	now      time.Time
	clock    *time.Time
}

func newWorld() world {
	drops, events := newFakeDrops(), &recorder{}
	now := time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)
	current := now
	return world{drops: drops, recorder: events, now: now, clock: &current, deps: service.Deps{
		Drops: drops, Sealer: events, Seals: events, Announcer: events, Audit: events, Tx: inlineTx{},
		Clock: movingClock{at: &current}, IDs: &sequentialIDs{},
	}}
}

func (w world) recorded(t *testing.T) string {
	t.Helper()
	receipt, err := service.NewRecordDrop(w.deps).Execute(t.Context(), cashier(), w.command())
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	return receipt.View.ID
}

func cashier() actor.Actor {
	return actor.Actor{UserID: cashierID, Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func consignorOwner() actor.Actor {
	return actor.Actor{UserID: "0192f0c4-0000-7000-8000-000000000502", Membership: actor.Membership{
		OrganizationID: consignorID, OrganizationType: actor.Consignor, Role: actor.Owner,
	}}
}

func (w world) command() service.RecordDropCommand {
	return service.RecordDropCommand{IdempotencyKey: idempotencyKey, Input: domain.RecordInput{
		PartnershipID: partnershipID, AgreementVersion: 1, OccurredAt: w.now,
		Items: []domain.ItemInput{{ProductID: breadID, Qty: 20}},
	}}
}

func codeOf(err error) apperror.Code {
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		return ""
	}
	return appErr.Code()
}

func TestRecordDrop_RSEC36_ConsignorMayNotRecord(t *testing.T) {
	scenario := newWorld()

	_, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), consignorOwner(), scenario.command())

	if codeOf(err) != apperror.CodeForbidden || len(scenario.drops.saved) != 0 {
		t.Fatalf("err = %v, saved = %d", err, len(scenario.drops.saved))
	}
}

func TestRecordDrop_BRAUTHZ02_OtherStoresPartnershipIsNotFound(t *testing.T) {
	scenario := newWorld()
	scenario.drops.setting.StoreID = "0192f0c4-0000-7000-8000-000000000409"

	_, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), cashier(), scenario.command())

	if codeOf(err) != apperror.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

func TestRecordDrop_BRDRP01_EndedPartnershipIsRefused(t *testing.T) {
	scenario := newWorld()
	scenario.drops.setting.IsActive = false

	_, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), cashier(), scenario.command())

	if codeOf(err) != apperror.CodePartnershipNotActive || len(scenario.recorder.seals) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestRecordDrop_StoreFailureIsWrapped(t *testing.T) {
	scenario := newWorld()
	scenario.drops.failCreate = errStoreDown

	_, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), cashier(), scenario.command())

	if !errors.Is(err, errStoreDown) || len(scenario.recorder.notices) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestRecordDrop_BRDRP03_SavesSealsAnnouncesAndAudits(t *testing.T) {
	scenario := newWorld()

	receipt, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), cashier(), scenario.command())
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	saved := scenario.drops.saved[receipt.View.ID]
	if receipt.IsReplay || saved.Number != "TM-202611-0001" || saved.TotalQty() != 20 || saved.RecordedBy != cashierID {
		t.Fatalf("receipt = %+v, saved = %+v", receipt, saved)
	}
	if scenario.drops.sequences[0] != (domain.SequenceKey{OrganizationID: storeID, Prefix: "TM", Period: "202611"}) {
		t.Fatalf("sequence = %+v", scenario.drops.sequences)
	}
	assertRecordedEvents(t, scenario.recorder)
	if !receipt.View.Seal.IsValid {
		t.Fatalf("seal = %+v", receipt.View.Seal)
	}
}

func assertRecordedEvents(t *testing.T, events *recorder) {
	t.Helper()
	if len(events.seals) != 1 || len(events.notices) != 1 || len(events.audits) != 1 || len(events.checks) != 1 {
		t.Fatalf("events = %+v", events)
	}
	if events.audits[0].Action != audit.DropRecorded || events.notices[0].OrganizationID != consignorID {
		t.Fatalf("audit = %+v, notice = %+v", events.audits[0], events.notices[0])
	}
}

func TestRecordDrop_BRDRP07_SameKeyAndBodyReplaysTheFirstDrop(t *testing.T) {
	scenario := newWorld()
	record := service.NewRecordDrop(scenario.deps)
	first, firstErr := record.Execute(t.Context(), cashier(), scenario.command())

	second, secondErr := record.Execute(t.Context(), cashier(), scenario.command())

	if errors.Join(firstErr, secondErr) != nil || !second.IsReplay || second.View.ID != first.View.ID ||
		len(scenario.drops.saved) != 1 || len(scenario.recorder.seals) != 1 {
		t.Fatalf("first = %+v, second = %+v, err = %v", first, second, errors.Join(firstErr, secondErr))
	}
}

func TestRecordDrop_RRES06_SameKeyWithAnotherBodyIsRefused(t *testing.T) {
	scenario := newWorld()
	record := service.NewRecordDrop(scenario.deps)
	if _, err := record.Execute(t.Context(), cashier(), scenario.command()); err != nil {
		t.Fatalf("first: %v", err)
	}
	changed := scenario.command()
	changed.Input.Items[0].Qty = 21

	_, err := record.Execute(t.Context(), cashier(), changed)

	if codeOf(err) != apperror.CodeIdempotencyKeyReused || len(scenario.drops.saved) != 1 {
		t.Fatalf("err = %v", err)
	}
}

func TestRecordDrop_BRDRP07_ConcurrentDuplicateReadsTheWinner(t *testing.T) {
	scenario := newWorld()
	fingerprint, err := scenario.command().Input.Fingerprint()
	if err != nil {
		t.Fatalf("fingerprint: %v", err)
	}
	winner := domain.Drop{
		ID: "0192f0c4-0000-7000-8000-000000000777", StoreID: storeID, ConsignorID: consignorID,
		IdempotencyKey: idempotencyKey, Fingerprint: fingerprint,
	}
	scenario.drops.duplicate = &winner

	receipt, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), cashier(), scenario.command())

	if err != nil || !receipt.IsReplay || receipt.View.ID != winner.ID ||
		!bytes.Equal(receipt.View.Fingerprint, fingerprint) {
		t.Fatalf("receipt = %+v, err = %v", receipt, err)
	}
}

func TestGetDrop_BRAUTHZ02_BothPartiesReadTheSameDrop(t *testing.T) {
	scenario := newWorld()
	receipt, err := service.NewRecordDrop(scenario.deps).Execute(t.Context(), cashier(), scenario.command())
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	get := service.NewGetDrop(scenario.deps)

	byStore, storeErr := get.Execute(t.Context(), cashier(), receipt.View.ID)
	byConsignor, consignorErr := get.Execute(t.Context(), consignorOwner(), receipt.View.ID)

	if errors.Join(storeErr, consignorErr) != nil || byStore.Number != byConsignor.Number {
		t.Fatalf("store = %+v, consignor = %+v", byStore, byConsignor)
	}
}

func TestGetDrop_RSEC36_RequiresAMembership(t *testing.T) {
	scenario := newWorld()

	_, err := service.NewGetDrop(scenario.deps).Execute(t.Context(), actor.Actor{UserID: cashierID}, "x")

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v, want FORBIDDEN", err)
	}
}
