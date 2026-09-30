package coach

import (
	"sync"
	"time"
)

const (
	authRateLimitMax    = 10
	authRateLimitWindow = time.Minute
)

// rateLimiter is in-process only; a multi-instance deployment would need a shared store such as Redis.
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

// sweep evicts keys never hit again (e.g. one-off IPs), which Allow's pruning can't reach; caller holds mu.
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

// prune filters in place, reusing attempts' backing array.
func prune(attempts []time.Time, cutoff time.Time) []time.Time {
	kept := attempts[:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}
