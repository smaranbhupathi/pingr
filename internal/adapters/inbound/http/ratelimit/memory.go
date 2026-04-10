package ratelimit

// MemoryStore implements Store using an in-process sliding window algorithm.
//
// Algorithm — Sliding Window:
//   Each key has a list of timestamps (one per request).
//   On every request we:
//     1. Drop timestamps older than `window` (they're outside the window)
//     2. Count what's left
//     3. If count >= limit → reject
//     4. Otherwise append current timestamp → allow
//
//   This fixes the "boundary problem" of fixed windows:
//   a fixed window resets at :00, so 5 requests at :59 + 5 requests at :01
//   = 10 requests in 2 seconds. Sliding window always looks back exactly
//   `window` duration from NOW, so that burst is caught.
//
// Tradeoff vs Redis:
//   This store is per-process. If you run 2 API instances, each has its own
//   map and doesn't know about the other's requests. For multi-instance
//   deployments, implement Store with Redis INCR + EXPIRE (atomic operations).

import (
	"context"
	"sync"
	"time"
)

// bucket holds the request timestamps for a single key.
type bucket struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// MemoryStore is safe for concurrent use.
type MemoryStore struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	done    chan struct{}
}

// NewMemoryStore creates a MemoryStore and starts a background cleanup goroutine
// that evicts idle buckets to prevent unbounded memory growth.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		buckets: make(map[string]*bucket),
		done:    make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Allow implements Store using a sliding window counter.
func (s *MemoryStore) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	// Get or create the bucket for this key.
	s.mu.Lock()
	b, ok := s.buckets[key]
	if !ok {
		b = &bucket{}
		s.buckets[key] = b
	}
	s.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// Prune timestamps that have fallen outside the window.
	// Reuse the same slice to avoid allocation.
	valid := b.timestamps[:0]
	for _, t := range b.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	b.timestamps = valid

	if len(b.timestamps) >= limit {
		return false, nil // rate limit exceeded
	}

	b.timestamps = append(b.timestamps, now)
	return true, nil
}

// Close stops the background cleanup goroutine.
// Call this when the server shuts down.
func (s *MemoryStore) Close() {
	close(s.done)
}

// cleanup runs every 5 minutes and removes buckets that have no recent
// timestamps. Without this, every unique IP that ever hit the server
// would stay in memory forever.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			for key, b := range s.buckets {
				b.mu.Lock()
				if len(b.timestamps) == 0 {
					delete(s.buckets, key)
				}
				b.mu.Unlock()
			}
			s.mu.Unlock()

		case <-s.done:
			return
		}
	}
}
