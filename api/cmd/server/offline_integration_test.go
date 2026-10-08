package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type delivery struct {
	recordedAt time.Time
	isOffline  bool
}

func offlineDropBody(partnershipID, productID string, sent delivery) string {
	return fmt.Sprintf(`{"partnership_id":"%s","agreement_version":1,"client_recorded_at":"%s",`+
		`"recorded_offline":%t,"items":[{"product_id":"%s","qty":20}]}`,
		partnershipID, instant(sent.recordedAt), sent.isOffline, productID)
}

func withDelivery(body string, recordedAt time.Time) string {
	delivery := fmt.Sprintf(`{"client_recorded_at":"%s","recorded_offline":true,`, instant(recordedAt))
	return strings.Replace(body, "{", delivery, 1)
}

func requireFields(t *testing.T, body string, fields ...string) {
	t.Helper()
	for _, field := range fields {
		if !strings.Contains(body, field) {
			t.Fatalf("missing %s in %s", field, body)
		}
	}
}

func instant(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}

func TestRecordDrop_FRPWA10_OfflineDropKeepsBothTimesAndCountsTheWindowFromReceipt(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	receivedAt := scene.store.now.Now()
	recordedAt := receivedAt.Add(-2 * time.Hour)
	body := offlineDropBody(partnershipID, scene.productID, delivery{recordedAt: recordedAt, isOffline: true})

	got := scene.store.recordDrop(body, dropKey)

	requireCreated(t, got)
	loadContract(t).requireConforms(t, got)
	requireFields(t, string(got.body), `"recorded_offline":true`, `"is_valid":true`,
		`"client_recorded_at":"`+instant(recordedAt)+`"`, `"occurred_at":"`+instant(recordedAt)+`"`,
		`"received_at":"`+instant(receivedAt)+`"`,
		`"cancel_deadline":"`+instant(receivedAt.Add(15*time.Minute))+`"`)
}

func TestRecordDrop_BRPWA07_DeviceClockFarAheadIsReplacedByTheReceipt(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	receivedAt := scene.store.now.Now()

	got := scene.store.recordDrop(
		offlineDropBody(partnershipID, scene.productID, delivery{recordedAt: receivedAt.Add(time.Hour)}), dropKey)

	requireCreated(t, got)
	requireFields(t, string(got.body), `"client_recorded_at":"`+instant(receivedAt)+`"`,
		`"occurred_at":"`+instant(receivedAt)+`"`)
}

func TestRecordDrop_BRPWA03_ResendMarkedOfflineReplaysTheFirstSend(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	recordedAt := scene.store.now.Now().Add(-time.Minute)
	first := scene.store.recordDrop(offlineDropBody(partnershipID, scene.productID, delivery{recordedAt: recordedAt}),
		dropKey)
	scene.store.now.Advance(time.Minute)
	resent := offlineDropBody(partnershipID, scene.productID, delivery{recordedAt: recordedAt, isOffline: true})

	second := scene.store.recordDrop(resent, dropKey)

	requireCreated(t, first)
	requireCreated(t, second)
	if second.header.Get("Idempotent-Replay") != "true" ||
		quotedAfter(string(first.body), `"id":"`) != quotedAfter(string(second.body), `"id":"`) {
		t.Fatalf("first = %s\nsecond = %s", first.body, second.body)
	}
	if got := productStock(t, scene, partnershipID); got != 20 {
		t.Fatalf("stock = %d, want 20", got)
	}
}

func TestRecordCount_FRPWA10_OfflineCountKeepsBothTimes(t *testing.T) {
	scene := recordedDrop(t)
	recordedAt := scene.store.now.Now().Add(time.Hour)
	scene.store.now.Advance(3 * time.Hour)
	scene.store.post("/api/v1/sessions/refresh")
	receivedAt := scene.store.now.Now()

	got := scene.store.recordCount(withDelivery(
		countBody(scene.partnershipID, scene.productID, countLine{recorded: 20, counted: 12}), recordedAt))

	requireCreated(t, got)
	loadContract(t).requireConforms(t, got)
	requireFields(t, string(got.body), `"recorded_offline":true`, `"is_valid":true`,
		`"client_recorded_at":"`+instant(recordedAt)+`"`, `"counted_at":"`+instant(recordedAt)+`"`,
		`"received_at":"`+instant(receivedAt)+`"`)
}
