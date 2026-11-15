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

func TestRateLimiterEvictsIdleClients(t *testing.T) {
	rl, now := newTestRateLimiter(t, 2, time.Second, time.Unix(0, 0))

	for _, key := range []string{"10.0.1.1", "10.0.1.2", "10.0.1.3"} {
		rl.allow(key)
	}
	if got := trackedClients(rl); got != 3 {
		t.Fatalf("expected 3 tracked clients, got %d", got)
	}

	*now = now.Add(2 * time.Second)
	rl.allow("10.0.1.1")

	if got := trackedClients(rl); got != 1 {
		t.Fatalf("expected the 2 idle clients to be evicted, %d still tracked", got)
	}
}

func TestRateLimiterKeepsActiveClients(t *testing.T) {
	rl, now := newTestRateLimiter(t, 2, time.Second, time.Unix(0, 0))

	const active = "10.0.2.1"
	rl.allow(active)

	*now = now.Add(500 * time.Millisecond)
	rl.allow(active)
	rl.allow("10.0.2.2")

	if got := trackedClients(rl); got != 2 {
		t.Fatalf("expected both clients to stay tracked, got %d", got)
	}
}

func trackedClients(rl *rateLimiter) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.clients)
}

func TestNewRateLimiterRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		rate   int
		period time.Duration
	}{
		{name: "zero rate", rate: 0, period: time.Minute},
		{name: "negative rate", rate: -1, period: time.Minute},
		{name: "zero period", rate: 1, period: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected the constructor to panic")
				}
			}()
			newRateLimiter(tc.rate, tc.period)
		})
	}
}
