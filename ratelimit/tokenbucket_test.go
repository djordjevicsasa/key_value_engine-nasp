package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketRateLimit(t *testing.T) {
	tb := NewTokenBucket(2, 50)

	if !tb.Allow() {
		t.Error("First request should be allowed")
	}
	if !tb.Allow() {
		t.Error("Second request should be allowed")
	}
	if tb.Allow() {
		t.Error("Third request should be denied because bucket is empty")
	}

	time.Sleep(60 * time.Millisecond)

	if !tb.Allow() {
		t.Error("Request should be allowed after waiting for refill")
	}
	if tb.Allow() {
		t.Error("Bucket should be empty again")
	}
}
