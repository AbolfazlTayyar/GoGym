package coach

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_AllowsUpToMax(t *testing.T) {
	limiter := newRateLimiter(3, time.Minute)

	for i := 1; i <= 3; i++ {
		assert.True(t, limiter.Allow("caller"), "attempt %d should be allowed", i)
	}
	assert.False(t, limiter.Allow("caller"), "attempt past the max should be rejected")
	assert.True(t, limiter.Allow("other-caller"), "a different key has its own budget")
}

func TestRateLimiter_EvictsExpiredKeys(t *testing.T) {
	window := 20 * time.Millisecond
	limiter := newRateLimiter(authRateLimitMax, window)

	require.True(t, limiter.Allow("first-caller"))
	require.Len(t, limiter.attempts, 1)

	time.Sleep(window + 10*time.Millisecond)
	require.True(t, limiter.Allow("second-caller"))

	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	assert.NotContains(t, limiter.attempts, "first-caller", "expired key should have been swept")
	assert.Len(t, limiter.attempts, 1)
}

func TestRateLimiter_SweepKeepsLiveKeys(t *testing.T) {
	window := 50 * time.Millisecond
	limiter := newRateLimiter(2, window)

	require.True(t, limiter.Allow("caller"))

	// Force a sweep while "caller" is still within its window.
	limiter.mu.Lock()
	limiter.lastSweep = time.Now().Add(-window)
	limiter.mu.Unlock()

	require.True(t, limiter.Allow("caller"))
	assert.False(t, limiter.Allow("caller"), "earlier attempts must still count toward the max")
}
