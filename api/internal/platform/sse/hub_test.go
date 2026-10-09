package sse

import "testing"

func newTestHub() *Hub {
	return NewHub(make(chan struct{}), Limits{StreamsPerUser: 3, Backlog: 2})
}

func isClosed(stream *Stream) bool {
	select {
	case _, open := <-stream.Events():
		return !open
	default:
		return false
	}
}

func TestPublish_OpenStream_ReceivesEventOfItsUserOnly(t *testing.T) {
	hub := newTestHub()
	mine := hub.Open("user-a")
	theirs := hub.Open("user-b")

	hub.Publish("user-a", Event{Name: "notification", Data: `{"id":"1"}`})

	if got := <-mine.Events(); got.Data != `{"id":"1"}` {
		t.Fatalf("event = %+v", got)
	}
	if len(theirs.Events()) != 0 {
		t.Fatal("another user's stream received the event")
	}
}

func TestOpen_FourthStreamOfOneUser_ClosesTheOldest(t *testing.T) {
	hub := newTestHub()
	oldest := hub.Open("user-a")
	middle := hub.Open("user-a")
	hub.Open("user-a")

	newest := hub.Open("user-a")
	hub.Publish("user-a", Event{Name: "notification", Data: "{}"})

	if !isClosed(oldest) || len(middle.Events()) != 1 || len(newest.Events()) != 1 {
		t.Fatal("the oldest stream should make room for the newest")
	}
}

func TestClose_FreesASlotForTheUser(t *testing.T) {
	hub := newTestHub()
	streams := make([]*Stream, 0, 3)
	for range 3 {
		stream := hub.Open("user-a")
		streams = append(streams, stream)
	}

	hub.Close(streams[0])
	hub.Close(streams[0])
	hub.Open("user-a")

	if !isClosed(streams[0]) || isClosed(streams[1]) {
		t.Fatal("closing one stream should leave the others open")
	}
}

func TestPublish_RPERF07_SlowStreamIsDroppedInsteadOfQueued(t *testing.T) {
	hub := newTestHub()
	slow := hub.Open("user-a")

	for range 3 {
		hub.Publish("user-a", Event{Name: "notification", Data: "{}"})
	}

	received := 0
	for range slow.Events() {
		received++
	}
	if received != 2 {
		t.Fatalf("received %d events before the drop, want the backlog of 2", received)
	}
}

func TestPublish_NoStream_IsIgnored(t *testing.T) {
	hub := newTestHub()

	hub.Publish("nobody", Event{Name: "notification", Data: "{}"})

	if stream := hub.Open("nobody"); len(stream.Events()) != 0 {
		t.Fatal("a stream opened later received an old event")
	}
}

func TestDone_FollowsTheShutdownSignal(t *testing.T) {
	shutdown := make(chan struct{})
	hub := NewHub(shutdown, Limits{StreamsPerUser: 1, Backlog: 1})

	close(shutdown)

	if _, open := <-hub.Done(); open {
		t.Fatal("done channel still open after shutdown")
	}
}
