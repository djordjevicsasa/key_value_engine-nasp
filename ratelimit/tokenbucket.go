package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	maxTokens      int
	tokens         int
	refillInterval time.Duration
	lastRefill     time.Time
	mu             sync.Mutex
}

func NewTokenBucket(maxTokens int, refillIntervalMs int) *TokenBucket {
	return &TokenBucket{
		maxTokens:      maxTokens,
		tokens:         maxTokens,
		refillInterval: time.Duration(refillIntervalMs) * time.Millisecond,
		lastRefill:     time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Koliko intervala je proslo
	intervals := int(elapsed / tb.refillInterval)
	if intervals > 0 {
		tb.tokens += intervals
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefill = tb.lastRefill.Add(time.Duration(intervals) * tb.refillInterval)
	}
}
