package sse

import (
	"slices"
	"sync"
)

type Event struct {
	Name string
	Data string
}

type Limits struct {
	StreamsPerUser int
	Backlog        int
}

type Stream struct {
	userID string
	events chan Event
}

func (s *Stream) Events() <-chan Event {
	return s.events
}

type Hub struct {
	mu      sync.Mutex
	limits  Limits
	done    <-chan struct{}
	streams map[string][]*Stream
}

func NewHub(done <-chan struct{}, limits Limits) *Hub {
	return &Hub{limits: limits, done: done, streams: map[string][]*Stream{}}
}

func (h *Hub) Done() <-chan struct{} {
	return h.done
}

func (h *Hub) Open(userID string) *Stream {
	h.mu.Lock()
	defer h.mu.Unlock()
	for len(h.streams[userID]) >= h.limits.StreamsPerUser {
		h.detach(h.streams[userID][0])
	}
	stream := &Stream{userID: userID, events: make(chan Event, h.limits.Backlog)}
	h.streams[userID] = append(h.streams[userID], stream)
	return stream
}

func (h *Hub) Close(stream *Stream) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.detach(stream)
}

func (h *Hub) Publish(userID string, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, stream := range slices.Clone(h.streams[userID]) {
		select {
		case stream.events <- event:
		default:
			h.detach(stream)
		}
	}
}

func (h *Hub) detach(stream *Stream) {
	userStreams := h.streams[stream.userID]
	index := slices.Index(userStreams, stream)
	if index < 0 {
		return
	}
	close(stream.events)
	remaining := slices.Delete(userStreams, index, index+1)
	if len(remaining) == 0 {
		delete(h.streams, stream.userID)
		return
	}
	h.streams[stream.userID] = remaining
}
