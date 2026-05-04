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
