package revocation

import (
	"sync"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

func start() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func TestIsRevoked_RSEC19_HoldsUntilTheAccessTokenWouldExpire(t *testing.T) {
	now := clock.NewManual(start())
	list := New(now)
	list.Revoke("family-1", start().Add(15*time.Minute))

	now.Advance(15*time.Minute - time.Second)
	stillRevoked := list.IsRevoked("family-1")
	now.Advance(time.Second)
	released := list.IsRevoked("family-1")

	if !stillRevoked || released || list.IsRevoked("family-2") {
		t.Fatalf("still revoked = %v, released = %v", stillRevoked, released)
	}
}

func TestRevoke_EarlierDeadlineNeverShortensTheBan(t *testing.T) {
	now := clock.NewManual(start())
	list := New(now)
	list.Revoke("family-1", start().Add(15*time.Minute))
	list.Revoke("family-1", start().Add(time.Minute))

	now.Advance(10 * time.Minute)

	if !list.IsRevoked("family-1") {
		t.Fatal("a later deadline must win")
	}
}

func TestIsRevoked_ExpiredEntriesAreSweptAway(t *testing.T) {
	now := clock.NewManual(start())
	list := New(now)
	list.Revoke("family-1", start().Add(time.Minute))

	now.Advance(2 * time.Minute)
	list.IsRevoked("family-2")

	if list.tracked() != 0 {
		t.Fatalf("tracked = %d", list.tracked())
	}
}

func TestList_IsSafeForConcurrentUse(t *testing.T) {
	list := New(clock.NewManual(start()))
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			list.Revoke("family-1", start().Add(time.Minute))
			list.IsRevoked("family-1")
		})
	}
	workers.Wait()

	if !list.IsRevoked("family-1") {
		t.Fatal("family must stay revoked")
	}
}
