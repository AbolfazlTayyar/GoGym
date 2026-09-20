package coach

import (
	"sync"
	"time"
)

const (
	// authRateLimitMax is the number of requests a single key (an IP or a
	// phone number) may make within authRateLimitWindow before /auth/*
	// starts rejecting it with 429.
	authRateLimitMax    = 10
	authRateLimitWindow = time.Minute
)

// rateLimiter is a simple in-process, per-key sliding-window limiter. It's
// enough for v1's single-instance deployment; a multi-instance deployment
// would need a shared store (e.g. Redis) instead, since each process would
// otherwise track its own independent counts.
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

// Allow reports whether key is still under the limit, and records this call
// as an attempt against it.
func (r *rateLimiter) Allow(key string) bool {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := now.Add(-r.window)
	kept := r.attempts[key][:0]
	for _, t := range r.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= r.max {
		r.attempts[key] = kept
		return false
	}

	r.attempts[key] = append(kept, now)
	return true
}
