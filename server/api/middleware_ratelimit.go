package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter is a token-bucket per-IP rate limiter. The framework does not
// ship a rate limiter today; this is the first one and is intended for the
// /api/join and /api/join/seat endpoints (spec §6.1, §6.2). Token-bucket
// rather than fixed-window because it absorbs short bursts (e.g. a phone
// hitting both /api/join and /api/join/seat in quick succession after the
// user enters a code).
type rateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*ipBucket
	capacity    int     // max tokens (and starting tokens for a fresh IP)
	refill      float64 // tokens per second
	idleTTL     time.Duration
	lastCleanup time.Time
}

type ipBucket struct {
	tokens     float64
	lastRefill time.Time
}

func newRateLimiter(capacity int, refillPerSecond float64, idleTTL time.Duration) *rateLimiter {
	r := &rateLimiter{
		buckets:     make(map[string]*ipBucket),
		capacity:    capacity,
		refill:      refillPerSecond,
		idleTTL:     idleTTL,
		lastCleanup: time.Now(),
	}
	return r
}

// Allow takes one token for the given IP. Returns true iff the bucket had a
// token available; on false the caller should respond with 429 Too Many
// Requests.
func (r *rateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastCleanup) >= r.idleTTL/2 {
		r.evictIdleLocked(now)
		r.lastCleanup = now
	}
	b, ok := r.buckets[ip]
	if !ok {
		b = &ipBucket{tokens: float64(r.capacity), lastRefill: now}
		r.buckets[ip] = b
	}

	// Refill since lastRefill (capped at capacity).
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * r.refill
	if b.tokens > float64(r.capacity) {
		b.tokens = float64(r.capacity)
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}
	return false
}

func (r *rateLimiter) evictIdle() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.evictIdleLocked(time.Now())
}

func (r *rateLimiter) evictIdleLocked(now time.Time) {
	cutoff := now.Add(-r.idleTTL)
	for ip, b := range r.buckets {
		if b.lastRefill.Before(cutoff) {
			delete(r.buckets, ip)
		}
	}
}

// rateLimitMiddleware returns a gin handler that consumes one token per
// request from the given limiter, keyed by ClientIP. On rate-limit failure
// returns 429 and aborts the chain.
func rateLimitMiddleware(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded; slow down and try again shortly",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
