package ratelimit

import (
	"context"
	"time"
)

// Store is the single interface that rate limit state storage must implement.
//
// To swap backends, implement this interface and pass your implementation
// to Middleware(). Nothing else in the codebase needs to change.
//
// Current implementations:
//   - NewMemoryStore() — in-process sliding window, single instance only
//
// Future implementations (when you need multi-instance):
//   - NewRedisStore(client) — atomic INCR/EXPIRE, shared across all instances
type Store interface {
	// Allow reports whether the request identified by key should proceed.
	//
	// key    — unique identifier, typically "clientIP:routePath"
	// limit  — max number of requests allowed in the window
	// window — the rolling time window (e.g. 1 minute)
	//
	// Returns true  → request is allowed, proceed to handler
	// Returns false → request exceeds limit, respond with 429
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}
