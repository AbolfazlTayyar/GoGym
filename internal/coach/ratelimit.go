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
	mu        sync.Mutex
	attempts  map[string][]time.Time
	max       int
	window    time.Duration
	lastSweep time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts:  make(map[string][]time.Time),
		max:       max,
		window:    window,
		lastSweep: time.Now(),
	}
}

// Allow reports whether key is still under the limit, and records this call
// as an attempt against it.
func (r *rateLimiter) Allow(key string) bool {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sweep(now)

	kept := prune(r.attempts[key], now.Add(-r.window))

	if len(kept) >= r.max {
		r.attempts[key] = kept
		return false
	}

	r.attempts[key] = append(kept, now)
	return true
}

// sweep drops keys whose attempts have all aged out. Pruning on its own only
// happens when a key is hit again, so without this a key seen exactly once —
// every distinct client IP that ever reaches /auth/* — would sit in the map
// for the life of the process. Runs at most once per window; the caller holds
// the mutex.
func (r *rateLimiter) sweep(now time.Time) {
	if now.Sub(r.lastSweep) < r.window {
		return
	}
	r.lastSweep = now

	cutoff := now.Add(-r.window)
	for key, attempts := range r.attempts {
		kept := prune(attempts, cutoff)
		if len(kept) == 0 {
			delete(r.attempts, key)
			continue
		}
		r.attempts[key] = kept
	}
}

// prune returns attempts with everything at or before cutoff dropped,
// filtering in place so the backing array is reused.
func prune(attempts []time.Time, cutoff time.Time) []time.Time {
	kept := attempts[:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}
