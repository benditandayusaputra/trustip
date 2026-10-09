package main

import (
	"strings"
	"sync"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	dropdomain "github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	dropservice "github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/sse"
)

type recordingFeed struct {
	mu        sync.Mutex
	published map[string][]sse.Event
}

func (f *recordingFeed) Publish(userID string, event sse.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.published[userID] = append(f.published[userID], event)
}

func (f *recordingFeed) eventsOf(userID string) []sse.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.published[userID]
}

func TestRecordDrop_FRNTF02_PublishesASummaryToTheConsignorAfterCommit(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	feed := &recordingFeed{published: map[string][]sse.Event{}}
	base := testFoundation(scene.store.pool, scene.store.now)
	base.feed = feed
	deps := buildTradeComponents(base, testSealKeys{}).drops
	command := dropservice.RecordDropCommand{IdempotencyKey: dropKey, Input: dropdomain.RecordInput{
		PartnershipID: partnershipID, AgreementVersion: 1, OccurredAt: claimed.At(testNow()),
		Items: []dropdomain.ItemInput{{ProductID: scene.productID, Qty: 20}},
	}}

	if _, err := dropservice.NewRecordDrop(deps).Execute(t.Context(), scene.store.actor(t), command); err != nil {
		t.Fatalf("record drop: %v", err)
	}

	events := feed.eventsOf(scene.consignor.actor(t).UserID)
	if len(events) != 1 || events[0].Name != "notification" ||
		!strings.Contains(events[0].Data, `"type":"DROP_RECORDED"`) {
		t.Fatalf("events = %+v", events)
	}
	if strings.Contains(events[0].Data, "TM-") || len(feed.eventsOf(scene.store.actor(t).UserID)) != 0 {
		t.Fatalf("BR-NTF-04 live event carries more than a summary or reached the actor: %+v", feed.published)
	}
}
