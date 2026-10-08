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
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
)

const (
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	consignorID   = "0192f0c4-0000-7000-8000-000000000402"
)

func countedAt() time.Time {
	return time.Date(2026, 11, 10, 20, 0, 0, 0, period.Jakarta())
}

type fakeSettlements struct {
	ledger  domain.Ledger
	created []domain.Settlement
	listed  []domain.ListQuery
	dueKeys []domain.DueKey
	locks   int
}

func (f *fakeSettlements) GetLedger(_ context.Context, partnershipID string) (domain.Ledger, error) {
	if partnershipID != f.ledger.PartnershipID {
		return domain.Ledger{}, apperror.OutsideScope()
	}
	return f.ledger, nil
}

func (f *fakeSettlements) NextNumber(context.Context, domain.SequenceKey) (int, error) {
	return len(f.created) + 1, nil
}

func (f *fakeSettlements) CreateSettlement(_ context.Context, settlement domain.Settlement) error {
	f.created = append(f.created, settlement)
	return nil
}

func (f *fakeSettlements) GetSettlement(_ context.Context, key domain.SettlementKey) (domain.View, error) {
	for _, settlement := range f.created {
		isParty := key.OrganizationID == settlement.ConsignorID || key.OrganizationID == settlement.StoreID
		if settlement.ID == key.SettlementID && isParty {
			return domain.View{Settlement: settlement}, nil
		}
	}
	return domain.View{}, apperror.OutsideScope()
}

func (f *fakeSettlements) ListSettlements(_ context.Context, query domain.ListQuery) ([]domain.ListItem, error) {
	f.listed = append(f.listed, query)
	items := []domain.ListItem{}
	for _, settlement := range f.created {
		items = append(items, domain.ListItem{
			ID: settlement.ID, Status: settlement.Status, DueDate: settlement.DueDate,
		})
	}
	return items, nil
}

func (f *fakeSettlements) GetRunning(_ context.Context, key domain.RunningKey) (domain.Running, error) {
	if key.OrganizationID != storeID && key.OrganizationID != consignorID {
		return domain.Running{}, apperror.OutsideScope()
	}
	return domain.Running{PartnershipID: key.PartnershipID}, nil
}

func (f *fakeSettlements) ListDue(_ context.Context, key domain.DueKey) ([]string, error) {
	f.dueKeys = append(f.dueKeys, key)
	return []string{f.ledger.PartnershipID}, nil
}

func (f *fakeSettlements) LockLedger(ctx context.Context, partnershipID string) (domain.Ledger, error) {
	f.locks++
	return f.GetLedger(ctx, partnershipID)
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
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

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time {
	return c.now
}

type sequentialIDs struct{ next int }

func (s *sequentialIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

type scenario struct {
	settlements *fakeSettlements
	events      *events
	deps        service.Deps
}

func newScenario(schedule terms.Schedule, quantities ...int) scenario {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	allocations := []domain.Allocation{}
	for index, qty := range quantities {
		allocations = append(allocations, domain.Allocation{
			ID: fmt.Sprintf("sale-%d", index), ProductID: "roti", OccurredAt: countedAt(), Qty: qty, Terms: bread,
			Split: bread.SplitOf(qty),
		})
	}
	settlements := &fakeSettlements{ledger: domain.Ledger{
		PartnershipID: partnershipID, ConsignorID: consignorID, StoreID: storeID, StoreName: "Warung Bu Sri",
		Schedule: schedule, PaymentTermsDays: 7, Unsettled: allocations,
	}}
	recorded := &events{}
	return scenario{settlements: settlements, events: recorded, deps: service.Deps{
		Settlements: settlements, Sealer: recorded, Seals: recorded, Announcer: recorded, Audit: recorded,
		Tx: inlineTx{}, Clock: fixedClock{now: countedAt().Add(time.Minute)}, IDs: &sequentialIDs{},
	}}
}

func TestIssueOnCount_FRSTL01_IssuesTheConsignorShareOfTheCount(t *testing.T) {
	run := newScenario(terms.ScheduleOnCount, 18)

	settlement, isIssued, err := service.NewIssuer(run.deps).IssueOnCount(t.Context(), partnershipID, countedAt())

	if err != nil || !isIssued || settlement.Number != "ST-202611-0001" ||
		settlement.Summary.Split.ConsignorShare != 31500 || len(run.settlements.created) != 1 {
		t.Fatalf("settlement = %+v, err = %v", settlement, err)
	}
	assertIssueEvents(t, run.events)
}

func assertIssueEvents(t *testing.T, recorded *events) {
	t.Helper()
	if len(recorded.seals) != 1 || recorded.seals[0].Event != "SETTLEMENT_ISSUED" ||
		len(recorded.notices) != 2 || len(recorded.audits) != 2 ||
		recorded.audits[0].Action != audit.SettlementIssued || recorded.audits[0].ActorID != "" {
		t.Fatalf("events = %+v", recorded)
	}
}

func TestIssueOnCount_FRSTL02_WeeklyPartnershipWaitsForTheScheduler(t *testing.T) {
	run := newScenario(terms.ScheduleWeekly, 18)

	_, isIssued, err := service.NewIssuer(run.deps).IssueOnCount(t.Context(), partnershipID, countedAt())

	if err != nil || isIssued || len(run.settlements.created) != 0 || len(run.events.seals) != 0 {
		t.Fatalf("issued = %v, err = %v", isIssued, err)
	}
}

func TestIssueOnCount_BRSTL03_CountWithoutSalesIssuesNothing(t *testing.T) {
	run := newScenario(terms.ScheduleOnCount)

	_, isIssued, err := service.NewIssuer(run.deps).IssueOnCount(t.Context(), partnershipID, countedAt())

	if err != nil || isIssued || len(run.settlements.created) != 0 {
		t.Fatalf("issued = %v, err = %v", isIssued, err)
	}
}

func storeOwner() actor.Actor {
	return actor.Actor{UserID: "0192f0c4-0000-7000-8000-000000000501", Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Owner,
	}}
}

func stranger() actor.Actor {
	return actor.Actor{UserID: "0192f0c4-0000-7000-8000-000000000509", Membership: actor.Membership{
		OrganizationID: "0192f0c4-0000-7000-8000-000000000409", OrganizationType: actor.Store, Role: actor.Owner,
	}}
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.Error](err); ok {
		return appErr.Code()
	}
	return ""
}

func TestGetSettlement_FRSTL04_PartiesReadItWithStandingAndSeal(t *testing.T) {
	run := newScenario(terms.ScheduleOnCount, 18)
	settlement, _, err := service.NewIssuer(run.deps).IssueOnCount(t.Context(), partnershipID, countedAt())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	get := service.NewGetSettlement(run.deps)

	view, readErr := get.Execute(t.Context(), storeOwner(), settlement.ID)
	_, strangerErr := get.Execute(t.Context(), stranger(), settlement.ID)

	if readErr != nil || !view.Seal.IsValid || view.Standing.Age != domain.AgeNotDue {
		t.Fatalf("view = %+v, err = %v", view, readErr)
	}
	if codeOf(strangerErr) != apperror.CodeNotFound {
		t.Fatalf("stranger = %v", strangerErr)
	}
}

func TestListSettlements_FRSTL09_ScopesToTheActorAndDerivesTheAge(t *testing.T) {
	run := newScenario(terms.ScheduleOnCount, 18)
	if _, _, err := service.NewIssuer(run.deps).IssueOnCount(t.Context(), partnershipID, countedAt()); err != nil {
		t.Fatalf("issue: %v", err)
	}
	run.deps.Clock = fixedClock{now: countedAt().AddDate(0, 0, 10)}

	page, err := service.NewListSettlements(run.deps).Execute(t.Context(), storeOwner(), service.ListRequest{
		Filter: domain.ListFilter{Age: "LATE_1_7"},
	})

	query := run.settlements.listed[0]
	if err != nil || query.OrganizationID != storeID || query.Status != "ISSUED" || len(page.Items) != 1 ||
		page.Items[0].Standing.DaysLate != 3 {
		t.Fatalf("page = %+v, query = %+v, err = %v", page, query, err)
	}
}

func TestListSettlements_RefusesAnUnknownAge(t *testing.T) {
	run := newScenario(terms.ScheduleOnCount)

	_, err := service.NewListSettlements(run.deps).Execute(t.Context(), storeOwner(), service.ListRequest{
		Filter: domain.ListFilter{Age: "ANCIENT"},
	})

	if codeOf(err) != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

func TestGetRunning_FRSTL10_RequiresMembership(t *testing.T) {
	run := newScenario(terms.ScheduleWeekly, 18)

	running, err := service.NewGetRunning(run.deps).Execute(t.Context(), storeOwner(), partnershipID)
	_, outsiderErr := service.NewGetRunning(run.deps).Execute(t.Context(), actor.Actor{}, partnershipID)

	if err != nil || running.PartnershipID != partnershipID || codeOf(outsiderErr) != apperror.CodeForbidden {
		t.Fatalf("running = %+v, err = %v, outsider = %v", running, err, outsiderErr)
	}
}
