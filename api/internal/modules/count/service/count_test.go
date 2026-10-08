package service_test

import (
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
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
)

const (
	partnershipID  = "0192f0c4-0000-7000-8000-000000000201"
	breadID        = "0192f0c4-0000-7000-8000-000000000101"
	storeID        = "0192f0c4-0000-7000-8000-000000000401"
	consignorID    = "0192f0c4-0000-7000-8000-000000000402"
	cashierID      = "0192f0c4-0000-7000-8000-000000000501"
	idempotencyKey = "0192f0c4-0000-7000-8000-000000000999"
)

type fakeCounts struct {
	sheet     domain.Sheet
	saved     map[string]domain.Count
	locks     int
	sequences []domain.SequenceKey
	listed    []domain.ListQuery
}

func newFakeCounts() *fakeCounts {
	terms := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	return &fakeCounts{saved: map[string]domain.Count{}, sheet: domain.Sheet{
		PartnershipID: partnershipID, IsActive: true, AgreementVersion: 1, StoreID: storeID, StoreName: "Warung Bu Sri",
		ConsignorID: consignorID,
		Products: []domain.StockedProduct{{ProductID: breadID, Batches: []domain.PricedBatch{
			{Stock: stock.Batch{ID: "bread-1", Balance: 20}, Terms: terms},
		}}},
	}}
}

func (f *fakeCounts) LockCount(ctx context.Context, key domain.CountKey) error {
	_, err := f.GetCount(ctx, key)
	return err
}

func (f *fakeCounts) SaveConfirmation(_ context.Context, decision domain.Decision) error {
	count := f.saved[decision.CountID]
	if count.Status != domain.Recorded {
		return domain.ErrNotPending
	}
	count.Status = domain.Confirmed
	f.saved[decision.CountID] = count
	return nil
}

func (f *fakeCounts) ListCounts(_ context.Context, query domain.ListQuery) ([]domain.Summary, error) {
	f.listed = append(f.listed, query)
	return []domain.Summary{}, nil
}

func (f *fakeCounts) LockSheet(ctx context.Context, key domain.SheetKey) (domain.Sheet, error) {
	f.locks++
	return f.GetSheet(ctx, key)
}

func (f *fakeCounts) GetSheet(_ context.Context, key domain.SheetKey) (domain.Sheet, error) {
	if key.StoreID != f.sheet.StoreID || key.PartnershipID != f.sheet.PartnershipID {
		return domain.Sheet{}, apperror.OutsideScope()
	}
	return f.sheet, nil
}

func (f *fakeCounts) NextNumber(_ context.Context, key domain.SequenceKey) (int, error) {
	f.sequences = append(f.sequences, key)
	return len(f.sequences), nil
}

func (f *fakeCounts) CreateCount(_ context.Context, count domain.Count) error {
	f.saved[count.ID] = count
	return nil
}

func (f *fakeCounts) FindByKey(_ context.Context, key domain.IdempotencyKey) (domain.Stored, bool, error) {
	for _, count := range f.saved {
		if count.StoreID == key.StoreID && count.IdempotencyKey == key.Key {
			return domain.Stored{CountID: count.ID, Fingerprint: count.Fingerprint}, true, nil
		}
	}
	return domain.Stored{}, false, nil
}

func (f *fakeCounts) GetCount(_ context.Context, key domain.CountKey) (domain.View, error) {
	count, ok := f.saved[key.CountID]
	if !ok || (key.OrganizationID != count.StoreID && key.OrganizationID != count.ConsignorID) {
		return domain.View{}, apperror.OutsideScope()
	}
	return domain.View{Count: count}, nil
}

type events struct {
	seals   []domain.SealRequest
	notices []domain.Notice
	audits  []audit.Entry
}

func (e *events) Seal(_ context.Context, request domain.SealRequest) error {
	e.seals = append(e.seals, request)
	return nil
}

func (e *events) Check(context.Context, domain.SealCheck) (domain.SealVerdict, error) {
	return domain.SealVerdict{IsValid: true}, nil
}

func (e *events) Announce(_ context.Context, notice domain.Notice) error {
	e.notices = append(e.notices, notice)
	return nil
}

func (e *events) Record(_ context.Context, entry audit.Entry) error {
	e.audits = append(e.audits, entry)
	return nil
}

type issuer struct {
	calls []time.Time
}

func (i *issuer) IssueOnCount(
	_ context.Context, _ string, countedAt time.Time,
) (domain.SettlementRef, bool, error) {
	i.calls = append(i.calls, countedAt)
	return domain.SettlementRef{ID: "settlement", Number: "ST-202611-0001"}, true, nil
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 11, 10, 5, 0, 0, 0, time.UTC)
}

type sequentialIDs struct{ next int }

func (s *sequentialIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

type scenario struct {
	counts *fakeCounts
	events *events
	issuer *issuer
	deps   service.Deps
}

func newScenario() scenario {
	counts, recorded, settlements := newFakeCounts(), &events{}, &issuer{}
	return scenario{counts: counts, events: recorded, issuer: settlements, deps: service.Deps{
		Counts: counts, Sealer: recorded, Seals: recorded, Announcer: recorded, Audit: recorded,
		Issuer: settlements, Tx: inlineTx{}, Clock: fixedClock{}, IDs: &sequentialIDs{},
	}}
}

func cashier() actor.Actor {
	return actor.Actor{UserID: cashierID, Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func consignorStaff() actor.Actor {
	return actor.Actor{UserID: "0192f0c4-0000-7000-8000-000000000502", Membership: actor.Membership{
		OrganizationID: consignorID, OrganizationType: actor.Consignor, Role: actor.Staff,
	}}
}

func command(recorded int) service.RecordCountCommand {
	return service.RecordCountCommand{IdempotencyKey: idempotencyKey, Input: domain.RecordInput{
		PartnershipID: partnershipID, AgreementVersion: 1, CountedAt: fixedClock{}.Now(), Occasion: "STORE_CLOSING",
		Lines: []domain.RequestLine{{
			ProductID: breadID, RecordedQty: recorded, CountedQty: 2, ReturnedQty: 2, ReturnReason: "UNSOLD",
		}},
	}}
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.Error](err); ok {
		return appErr.Code()
	}
	return ""
}

func TestRecordCount_RSEC36_ConsignorMayNotCount(t *testing.T) {
	run := newScenario()

	_, err := service.NewRecordCount(run.deps).Execute(t.Context(), consignorStaff(), command(20))

	if codeOf(err) != apperror.CodeForbidden || run.counts.locks != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestRecordCount_BRAUTHZ02_OtherStoresPartnershipIsNotFound(t *testing.T) {
	run := newScenario()
	run.counts.sheet.StoreID = "0192f0c4-0000-7000-8000-000000000409"

	_, err := service.NewRecordCount(run.deps).Execute(t.Context(), cashier(), command(20))

	if codeOf(err) != apperror.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}

func TestRecordCount_BRSTK01_StockSeenEarlierIsRefusedAfterLocking(t *testing.T) {
	run := newScenario()

	_, err := service.NewRecordCount(run.deps).Execute(t.Context(), cashier(), command(19))

	if codeOf(err) != apperror.CodeStockChanged || run.counts.locks != 1 || len(run.counts.saved) != 0 {
		t.Fatalf("err = %v, locks = %d", err, run.counts.locks)
	}
}

func TestRecordCount_BRCNT04_SavesSealsAnnouncesAndAudits(t *testing.T) {
	run := newScenario()

	receipt, err := service.NewRecordCount(run.deps).Execute(t.Context(), cashier(), command(20))
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	saved := run.counts.saved[receipt.View.ID]
	if receipt.IsReplay || saved.Number != "HS-202611-0001" || saved.SoldQty != 18 ||
		saved.Split.ConsignorShare != 31500 || len(saved.Movements()) != 2 {
		t.Fatalf("saved = %+v", saved)
	}
	assertCountEvents(t, run.events)
	if len(run.issuer.calls) != 1 || !run.issuer.calls[0].Equal(saved.CountedAt) {
		t.Fatalf("BR-CNT-04 settlement calls = %v", run.issuer.calls)
	}
}

func assertCountEvents(t *testing.T, recorded *events) {
	t.Helper()
	if len(recorded.seals) != 1 || recorded.seals[0].Event != "COUNT_RECORDED" ||
		len(recorded.notices) != 1 || recorded.audits[0].Action != audit.CountRecorded {
		t.Fatalf("events = %+v", recorded)
	}
}

func TestRecordCount_BRCNT07_SameKeyReplaysAndAnotherBodyIsRefused(t *testing.T) {
	run := newScenario()
	record := service.NewRecordCount(run.deps)
	first, firstErr := record.Execute(t.Context(), cashier(), command(20))
	second, secondErr := record.Execute(t.Context(), cashier(), command(20))
	changed := command(20)
	changed.Input.Lines[0].CountedQty = 3

	_, changedErr := record.Execute(t.Context(), cashier(), changed)

	if errors.Join(firstErr, secondErr) != nil || !second.IsReplay || second.View.ID != first.View.ID {
		t.Fatalf("first = %+v, second = %+v", first, second)
	}
	if codeOf(changedErr) != apperror.CodeIdempotencyKeyReused || len(run.counts.saved) != 1 {
		t.Fatalf("changed = %v", changedErr)
	}
}

func TestGetCount_BothPartiesReadTheSameCount(t *testing.T) {
	run := newScenario()
	receipt, err := service.NewRecordCount(run.deps).Execute(t.Context(), cashier(), command(20))
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	get := service.NewGetCount(run.deps)

	byStore, storeErr := get.Execute(t.Context(), cashier(), receipt.View.ID)
	byConsignor, consignorErr := get.Execute(t.Context(), consignorStaff(), receipt.View.ID)

	if errors.Join(storeErr, consignorErr) != nil || byStore.Number != byConsignor.Number {
		t.Fatalf("store = %+v, consignor = %+v", byStore, byConsignor)
	}
}

func TestGetSheet_RSEC36_OnlyTheStoreOpensTheSheet(t *testing.T) {
	run := newScenario()

	sheet, storeErr := service.NewGetSheet(run.deps).Execute(t.Context(), cashier(), partnershipID)
	_, consignorErr := service.NewGetSheet(run.deps).Execute(t.Context(), consignorStaff(), partnershipID)

	if storeErr != nil || len(sheet.Products) != 1 || codeOf(consignorErr) != apperror.CodeForbidden {
		t.Fatalf("sheet = %+v, store = %v, consignor = %v", sheet, storeErr, consignorErr)
	}
}
