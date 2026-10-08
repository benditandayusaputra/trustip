package main

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	dropdomain "github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	droprepository "github.com/benditandayusaputra/trustip/api/internal/modules/drop/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

type jumpingWaits struct {
	clock  *clock.Manual
	wakes  int
	parked chan struct{}
}

func (w *jumpingWaits) After(wait time.Duration) <-chan time.Time {
	fired := make(chan time.Time, 1)
	if w.wakes == 0 {
		close(w.parked)
		return fired
	}
	w.wakes--
	w.clock.Advance(wait)
	fired <- w.clock.Now()
	return fired
}

func (scene dropScene) runJobsUntil(t *testing.T, until time.Time, wakes int) {
	t.Helper()
	now := scene.store.now
	now.Advance(until.Sub(now.Now()))
	waits := &jumpingWaits{clock: now, wakes: wakes, parked: make(chan struct{})}
	trade := buildTradeComponents(scene.store.pool, now, testSealKeys{})
	jobs := newScheduler(schedulerParts{
		runners: buildJobRunners(scene.store.pool, trade), clock: now, waits: waits,
		log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	ctx, cancel := context.WithCancel(t.Context())
	finished := make(chan struct{})
	go func() {
		jobs.Run(ctx)
		close(finished)
	}()
	<-waits.parked
	cancel()
	<-finished
	scene.store.post("/api/v1/sessions/refresh")
	scene.consignor.post("/api/v1/sessions/refresh")
}

func jakartaMidnight(month time.Month, day int) time.Time {
	return time.Date(testNow().Year(), month, day, 0, 0, 0, 0, period.Jakarta())
}

func TestScheduler_FRSTL02_WeeklySettlementIsIssuedOnMondayAfterTheWeek(t *testing.T) {
	scene := recordedDrop(t)
	scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 2, returned: 2, reason: "UNSOLD"}))

	scene.runJobsUntil(t, jakartaMidnight(10, 26), 2)

	listed := scene.consignor.get(settlementsPath + "?partnership_id=" + scene.partnershipID)
	requireCode(t, listed, "OK")
	loadContract(t).requireConforms(t, listed)
	if !strings.Contains(string(listed.body), `"number":"ST-202610-0001"`) ||
		!strings.Contains(string(listed.body), `"schedule":"WEEKLY"`) ||
		!strings.Contains(string(listed.body), `"start":"2026-10-18T17:00:00Z","end":"2026-10-25T17:00:00Z"`) ||
		!strings.Contains(string(listed.body), `"consignor_share":31500`) {
		t.Fatalf("listed = %s", listed.body)
	}
	running := scene.store.get("/api/v1/partnerships/" + scene.partnershipID + "/running-sales")
	if !strings.Contains(string(running.body), `"has_sales":false`) {
		t.Fatalf("running = %s", running.body)
	}
}

func TestScheduler_FRSTL02_MonthlySettlementWaitsForTheFirstOfTheMonth(t *testing.T) {
	scene := recordedDropWith(t, "MONTHLY")
	scene.store.recordCount(countBody(scene.partnershipID, scene.productID, countLine{recorded: 20, counted: 12}))

	scene.runJobsUntil(t, jakartaMidnight(10, 26), 2)
	beforeMonthEnd := scene.consignor.get(settlementsPath + "?partnership_id=" + scene.partnershipID)
	scene.runJobsUntil(t, jakartaMidnight(11, 1), 2)
	afterMonthEnd := scene.consignor.get(settlementsPath + "?partnership_id=" + scene.partnershipID)

	if !strings.Contains(string(beforeMonthEnd.body), `"data":[]`) {
		t.Fatalf("before = %s", beforeMonthEnd.body)
	}
	if !strings.Contains(string(afterMonthEnd.body), `"schedule":"MONTHLY"`) ||
		!strings.Contains(string(afterMonthEnd.body), `"start":"2026-09-30T17:00:00Z","end":"2026-10-31T17:00:00Z"`) ||
		!strings.Contains(string(afterMonthEnd.body), `"consignor_share":14000`) {
		t.Fatalf("after = %s", afterMonthEnd.body)
	}
}

func TestScheduler_SpecTwentyOneFive_UnansweredDropIsStoredAsConfirmedAfterFortyEightHours(t *testing.T) {
	scene := recordedDrop(t)
	dropID := strings.TrimPrefix(scene.path, dropsPath+"/")

	scene.runJobsUntil(t, jakartaMidnight(10, 22), 1)

	view, err := droprepository.NewDrops(scene.store.pool).GetDrop(t.Context(), dropdomain.DropKey{
		OrganizationID: scene.store.actor(t).Membership.OrganizationID, DropID: dropID,
	})
	if err != nil || view.Status != dropdomain.Confirmed || !view.AutoConfirmed ||
		!view.ConfirmedAt.Equal(view.ReceivedAt.Add(48*time.Hour)) {
		t.Fatalf("view = %+v, err = %v", view, err)
	}
}
