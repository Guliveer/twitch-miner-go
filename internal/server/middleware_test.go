package server

import (
	"testing"
	"time"
)

func newTestRateLimiter(t *testing.T, rate int, period time.Duration, start time.Time) (*rateLimiter, *time.Time) {
	t.Helper()

	rl := newRateLimiter(rate, period)
	now := start
	rl.now = func() time.Time { return now }
	return rl, &now
}

func TestRateLimiterRefillsUnderSteadyTraffic(t *testing.T) {
	rl, now := newTestRateLimiter(t, 2, time.Second, time.Unix(0, 0))

	const key = "10.0.0.1"

	// The first request creates the bucket with rate-1 tokens, the second drains it.
	if !rl.allow(key) {
		t.Fatal("expected first request to be allowed")
	}
	if !rl.allow(key) {
		t.Fatal("expected second request to be allowed")
	}
	if rl.allow(key) {
		t.Fatal("expected request beyond the burst to be limited")
	}

	// One token per 500ms, so polling every 250ms must still let half through.
	allowed := 0
	for range 8 {
		*now = now.Add(250 * time.Millisecond)
		if rl.allow(key) {
			allowed++
		}
	}

	if allowed != 4 {
		t.Fatalf("expected 4 of 8 requests to refill and pass, got %d", allowed)
	}
}

func TestRateLimiterDoesNotExceedLimitAfterIdle(t *testing.T) {
	rl, now := newTestRateLimiter(t, 2, time.Second, time.Unix(0, 0))

	const key = "10.0.0.2"

	*now = now.Add(10 * time.Second)
	if !rl.allow(key) {
		t.Fatal("expected first request after idle to be allowed")
	}
	if !rl.allow(key) {
		t.Fatal("expected second request after idle to be allowed")
	}
	if rl.allow(key) {
		t.Fatal("expected the bucket to stay capped at the configured limit")
	}
}
