package request

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DefaultCapacity constant determines how much capacity the requests slice should allocate in advance
const DefaultCapacity = 1000

type snapshot struct {
	Requests []int64 `json:"requests"`
}

// Counter is the request request counter interface
type Counter interface {
	Count() int
	Snapshot() ([]byte, error)
	Clear()
}

// slidingWindowCounter counts requests within the given window with a millisecond accuracy
type slidingWindowCounter struct {
	lock sync.RWMutex

	requests []int64
	window   time.Duration
}

func NewCounter(window time.Duration) Counter {
	return &slidingWindowCounter{
		lock: sync.RWMutex{},

		requests: make([]int64, 0, DefaultCapacity),
		window:   window,
	}
}

func NewCounterFromSnapshot(snapshotJSON []byte, window time.Duration) (Counter, error) {
	var snap snapshot
	if err := json.Unmarshal(snapshotJSON, &snap); err != nil {
		return nil, fmt.Errorf("unable to unmarshal the snapshot: %w", err)
	}

	return &slidingWindowCounter{
		lock:     sync.RWMutex{},
		requests: snap.Requests,
		window:   window,
	}, nil
}

// Count increases the request count by 1 and returns the number of requests within the given window
func (s *slidingWindowCounter) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now().UTC()
	windowStart := now.Add(-s.window)

	discardCount := 0
	for _, el := range s.requests {
		if time.UnixMilli(el).Before(windowStart) {
			discardCount++
		}
	}

	s.requests = append(s.requests[discardCount:], now.UnixMilli())

	return len(s.requests)
}

// Snapshot creates a snapshot of the current state
func (s *slidingWindowCounter) Snapshot() ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	snap := snapshot{
		Requests: s.requests,
	}

	return json.Marshal(snap)
}

func (s *slidingWindowCounter) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	clear(s.requests)
} 