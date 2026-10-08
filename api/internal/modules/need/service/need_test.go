package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/service"
)

const (
	storeID  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	snacksID = "01a1161e-0a49-71f7-a49a-3cf81042e3b0"
	needID   = "0192f0c4-0000-7000-8000-000000000001"
)

func postedAt() time.Time { return time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC) }

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

type fixedIDs struct{}

func (fixedIDs) NewID() (string, error) { return needID, nil }

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

type fakeAudit struct{ entries []audit.Entry }

func (f *fakeAudit) Record(_ context.Context, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type fakeNeeds struct {
	stored    map[string]domain.Need
	createErr error
	saveErr   error
	changes   []domain.Change
	expiredAt time.Time
}

func newFakeNeeds() *fakeNeeds { return &fakeNeeds{stored: map[string]domain.Need{}} }

func (f *fakeNeeds) CreateNeed(_ context.Context, posting domain.Posting) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.stored[posting.ID] = domain.Need{
		ID: posting.ID, StoreID: posting.StoreID, Draft: posting.Draft, Status: domain.Open,
		ExpiresAt: posting.ExpiresAt, CreatedAt: posting.At,
	}
	return nil
}

func (f *fakeNeeds) LockNeed(ctx context.Context, key domain.NeedKey) (domain.Need, error) {
	return f.GetNeed(ctx, key)
}

func (f *fakeNeeds) SaveChange(_ context.Context, change domain.Change) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.changes = append(f.changes, change)
	need := f.stored[change.NeedID]
	need.Status, need.ExpiresAt = change.Status, change.ExpiresAt
	f.stored[change.NeedID] = need
	return nil
}

func (f *fakeNeeds) ExpireOpen(_ context.Context, now time.Time) (int, error) {
	f.expiredAt = now
	return 1, nil
}

func (f *fakeNeeds) GetNeed(_ context.Context, key domain.NeedKey) (domain.Need, error) {
	need, ok := f.stored[key.NeedID]
	if !ok || need.StoreID != key.StoreID {
		return domain.Need{}, apperror.NotFound()
	}
	return need, nil
}

func (f *fakeNeeds) ListNeeds(_ context.Context, _ domain.ListQuery) ([]domain.Need, error) {
	needs := []domain.Need{}
	for _, need := range f.stored {
		needs = append(needs, need)
	}
	return needs, nil
}

func (f *fakeNeeds) ListResponses(_ context.Context, _ domain.NeedKey) ([]domain.Response, error) {
	return []domain.Response{}, nil
}

type fakeInterest struct {
	interested []domain.Interested
	expiring   []domain.Expiring
	window     domain.ReminderWindow
}

func (f *fakeInterest) ListInterested(context.Context, domain.Need) ([]domain.Interested, error) {
	return f.interested, nil
}

func (f *fakeInterest) ClaimExpiring(_ context.Context, window domain.ReminderWindow) ([]domain.Expiring, error) {
	f.window = window
	return f.expiring, nil
}

type fakeAnnouncer struct{ notices []domain.Notice }

func (f *fakeAnnouncer) Announce(_ context.Context, notice domain.Notice) error {
	f.notices = append(f.notices, notice)
	return nil
}

type scene struct {
	needs     *fakeNeeds
	audit     *fakeAudit
	interest  *fakeInterest
	announcer *fakeAnnouncer
	deps      service.Deps
}

func newScene(at time.Time) scene {
	needs, trail, interest, announcer := newFakeNeeds(), &fakeAudit{}, &fakeInterest{}, &fakeAnnouncer{}
	return scene{needs: needs, audit: trail, interest: interest, announcer: announcer, deps: service.Deps{
		Writer: needs, Reader: needs, Interest: interest, Announcer: announcer, Audit: trail, Tx: inlineTx{},
		Clock: fixedClock{at: at}, IDs: fixedIDs{},
	}}
}

func storeOwner() actor.Actor {
	return actor.Actor{UserID: "u1", Membership: actor.Membership{
		OrganizationID: storeID, OrganizationType: actor.Store, Role: actor.Owner,
	}}
}

func form() domain.Form {
	return domain.Form{Title: "Butuh camilan kantin", CategoryID: snacksID}
}

func post(world scene) (domain.Need, error) {
	need, err := service.NewPostNeed(world.deps).Execute(context.Background(), storeOwner(), form())
	if err != nil {
		return domain.Need{}, fmt.Errorf("post: %w", err)
	}
	return need, nil
}

func TestPostNeed_FRNED01_StoreOwnerPostsANeed(t *testing.T) {
	world := newScene(postedAt())

	need, err := post(world)

	if err != nil || need.Status != domain.Open || len(world.audit.entries) != 1 ||
		world.audit.entries[0].Action != audit.NeedPosted {
		t.Fatalf("need = %+v, err = %v", need, err)
	}
}

func TestPostNeed_BRAUTHZ01_OnlyTheStoreOwner(t *testing.T) {
	world := newScene(postedAt())
	staff := storeOwner()
	staff.Membership.Role = actor.Staff

	_, err := service.NewPostNeed(world.deps).Execute(context.Background(), staff, form())

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestPostNeed_BRNED01_SixthOpenNeedIsRefused(t *testing.T) {
	world := newScene(postedAt())
	world.needs.createErr = domain.ErrOpenLimit

	_, err := post(world)

	if apperror.From(err).Code() != apperror.CodeNeedLimitReached {
		t.Fatalf("err = %v", err)
	}
}

func TestPostNeed_UnknownCategoryIsAFieldError(t *testing.T) {
	world := newScene(postedAt())
	world.needs.createErr = domain.ErrUnknownCategory

	_, err := post(world)

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

func TestCloseNeed_FRNED06_ClosedNeedCannotBeClosedAgain(t *testing.T) {
	world := newScene(postedAt())
	if _, err := post(world); err != nil {
		t.Fatal(err)
	}
	closer := service.NewCloseNeed(world.deps)

	closed, err := closer.Execute(context.Background(), storeOwner(), needID)
	_, againErr := closer.Execute(context.Background(), storeOwner(), needID)

	if err != nil || closed.Status != domain.Closed || apperror.From(againErr).Code() != apperror.CodeNeedNotOpen {
		t.Fatalf("closed = %+v, err = %v, again = %v", closed, err, againErr)
	}
}

func TestExtendNeed_BRNED01_ReopeningAboveTheLimitIsRefused(t *testing.T) {
	world := newScene(postedAt())
	if _, err := post(world); err != nil {
		t.Fatal(err)
	}
	world.needs.saveErr = domain.ErrOpenLimit

	_, err := service.NewExtendNeed(world.deps).Execute(context.Background(), storeOwner(), needID)

	if apperror.From(err).Code() != apperror.CodeNeedLimitReached {
		t.Fatalf("err = %v", err)
	}
}

func TestGetNeed_BRNED03_ShowsAnExpiredNeedAsExpired(t *testing.T) {
	world := newScene(postedAt())
	if _, err := post(world); err != nil {
		t.Fatal(err)
	}

	detail, err := service.NewGetNeed(world.needs, fixedClock{at: postedAt().Add(domain.ShowingPeriod)}).Execute(
		context.Background(), storeOwner(), needID)

	if err != nil || detail.Need.Status != domain.Expired {
		t.Fatalf("detail = %+v, err = %v", detail, err)
	}
}

func TestListNeeds_RejectsAnUnknownStatus(t *testing.T) {
	world := newScene(postedAt())

	_, err := service.NewListNeeds(world.needs, world.deps.Clock).Execute(context.Background(), storeOwner(),
		service.ListRequest{Status: "LOST"})

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

func TestExpireNeeds_BRNED03_MarksExpiredNeedsAtTheGivenTime(t *testing.T) {
	world := newScene(postedAt())

	expired, err := service.NewExpireNeeds(world.needs).Execute(context.Background(), postedAt())

	if err != nil || expired != 1 || !world.needs.expiredAt.Equal(postedAt()) {
		t.Fatalf("expired = %d, err = %v", expired, err)
	}
}

func TestPostNeed_FRNED07_NearbyConsignorsOfTheCategoryAreTold(t *testing.T) {
	world := newScene(postedAt())
	world.interest.interested = []domain.Interested{
		{OrganizationID: "c1", DistanceMeters: 1200}, {OrganizationID: "c2", DistanceMeters: 9000},
	}

	_, err := post(world)

	if err != nil || len(world.announcer.notices) != 2 || world.announcer.notices[1].OrganizationID != "c2" ||
		world.announcer.notices[0].Type != "NEED_POSTED" {
		t.Fatalf("notices = %+v, err = %v", world.announcer.notices, err)
	}
}

func TestRemindExpiringNeeds_BRNED03_RemindsEachClaimedNeedOnce(t *testing.T) {
	world := newScene(postedAt())
	world.interest.expiring = []domain.Expiring{{NeedID: "n1", StoreID: storeID, Title: "Butuh roti"}}

	reminded, err := service.NewRemindExpiringNeeds(world.deps).Execute(context.Background(), postedAt())

	if err != nil || reminded != 1 || len(world.announcer.notices) != 1 ||
		!world.interest.window.Until.Equal(postedAt().Add(domain.ExpiryReminderBefore)) {
		t.Fatalf("reminded = %d, notices = %+v, err = %v", reminded, world.announcer.notices, err)
	}
}
