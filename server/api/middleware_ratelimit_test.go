package api

import (
	"testing"
	"time"

	"github.com/workfit/tester/assert"
)

func TestRateLimiterAllowsUpToCapacity(t *testing.T) {
	limiter := newRateLimiter(5, 1.0, time.Minute)
	// 5 capacity → first 5 calls succeed, 6th fails.
	for i := 0; i < 5; i++ {
		assert.For(t, i).ThatActual(limiter.Allow("1.2.3.4")).IsTrue()
	}
	assert.For(t).ThatActual(limiter.Allow("1.2.3.4")).IsFalse()
}

func TestRateLimiterRefills(t *testing.T) {
	// 1 token capacity, refill at 100 tokens/sec → after 20ms ≈ 2 tokens worth.
	limiter := newRateLimiter(1, 100.0, time.Minute)
	assert.For(t).ThatActual(limiter.Allow("ip")).IsTrue()
	assert.For(t).ThatActual(limiter.Allow("ip")).IsFalse()
	time.Sleep(20 * time.Millisecond)
	assert.For(t).ThatActual(limiter.Allow("ip")).IsTrue()
}

func TestRateLimiterPerIP(t *testing.T) {
	limiter := newRateLimiter(1, 0.001, time.Minute)
	assert.For(t).ThatActual(limiter.Allow("a")).IsTrue()
	// "a" is now empty; "b" is independent.
	assert.For(t).ThatActual(limiter.Allow("a")).IsFalse()
	assert.For(t).ThatActual(limiter.Allow("b")).IsTrue()
	assert.For(t).ThatActual(limiter.Allow("b")).IsFalse()
}

func TestRateLimiterEvictsIdle(t *testing.T) {
	limiter := newRateLimiter(1, 1.0, 10*time.Millisecond)
	limiter.Allow("ip")
	limiter.mu.Lock()
	_, hadBucket := limiter.buckets["ip"]
	limiter.mu.Unlock()
	assert.For(t).ThatActual(hadBucket).IsTrue()

	time.Sleep(15 * time.Millisecond)
	limiter.evictIdle()

	limiter.mu.Lock()
	_, stillHere := limiter.buckets["ip"]
	limiter.mu.Unlock()
	assert.For(t).ThatActual(stillHere).IsFalse()
}
