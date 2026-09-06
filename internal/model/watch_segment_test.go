package model

import (
	"testing"
	"time"
)

// The watch set rotates, so a channel routinely goes unwatched for long
// stretches. Charging that gap to the channel would push it past its streak
// window and retire it from the rotation without it ever having been watched.
func TestMinuteWatchedIgnoresUnwatchedGaps(t *testing.T) {
	s := NewStream()

	s.UpdateMinuteWatched() // baseline
	s.minuteWatchedTimestamp = time.Now().Add(-25 * time.Minute)
	s.UpdateMinuteWatched()

	if s.MinuteWatched != 0 {
		t.Fatalf("a 25 minute gap is not watch time, got %.2f minutes", s.MinuteWatched)
	}
}

func TestMinuteWatchedCountsContinuousWatching(t *testing.T) {
	s := NewStream()

	s.UpdateMinuteWatched()
	s.minuteWatchedTimestamp = time.Now().Add(-20 * time.Second)
	s.UpdateMinuteWatched()

	if s.MinuteWatched < 0.3 || s.MinuteWatched > 0.4 {
		t.Fatalf("expected roughly 20 seconds of watch time, got %.2f minutes", s.MinuteWatched)
	}
}

func TestMinuteWatchedAccumulatesAcrossSegments(t *testing.T) {
	s := NewStream()

	// One watched minute, a long rotation away, then another watched minute.
	s.UpdateMinuteWatched()
	s.minuteWatchedTimestamp = time.Now().Add(-1 * time.Minute)
	s.UpdateMinuteWatched()

	s.minuteWatchedTimestamp = time.Now().Add(-30 * time.Minute)
	s.UpdateMinuteWatched()

	s.minuteWatchedTimestamp = time.Now().Add(-1 * time.Minute)
	s.UpdateMinuteWatched()

	if s.MinuteWatched < 1.9 || s.MinuteWatched > 2.1 {
		t.Fatalf("expected the two watched minutes only, got %.2f", s.MinuteWatched)
	}
}

// A frozen channel is one we are still sending events for while Twitch has
// stopped crediting them.
func TestStalledWhenCreditsStopButWeKeepSending(t *testing.T) {
	s := NewStream()
	s.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)
	s.LastMinuteAttemptedAt = time.Now()

	if !s.IsMinuteWatchStalled(5 * time.Minute) {
		t.Fatal("a channel we keep sending for without credit is stalled")
	}
}

// This is the case that put rotated-out channels into a 30 minute cooldown.
func TestRotatedOutChannelIsNotStalled(t *testing.T) {
	s := NewStream()
	s.LastMinuteCreditedAt = time.Now().Add(-25 * time.Minute)
	s.LastMinuteAttemptedAt = time.Now().Add(-25 * time.Minute)

	if s.IsMinuteWatchStalled(5 * time.Minute) {
		t.Fatal("a channel we stopped watching must not be reported as frozen")
	}
}

func TestNeverCreditedIsNotStalled(t *testing.T) {
	s := NewStream()
	s.LastMinuteAttemptedAt = time.Now()

	if s.IsMinuteWatchStalled(5 * time.Minute) {
		t.Fatal("a channel that was never credited is new, not frozen")
	}
}

func TestNeverAttemptedIsNotStalled(t *testing.T) {
	s := NewStream()
	s.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)

	if s.IsMinuteWatchStalled(5 * time.Minute) {
		t.Fatal("without an attempt there is nothing to judge")
	}
}

func TestHealthyChannelIsNotStalled(t *testing.T) {
	s := NewStream()
	s.LastMinuteCreditedAt = time.Now().Add(-30 * time.Second)
	s.LastMinuteAttemptedAt = time.Now()

	if s.IsMinuteWatchStalled(5 * time.Minute) {
		t.Fatal("a recently credited channel is healthy")
	}
}

func TestMarkMinuteWatchAttemptStamps(t *testing.T) {
	s := NewStream()
	if !s.LastMinuteAttemptedAt.IsZero() {
		t.Fatal("a fresh stream has no attempt recorded")
	}

	s.MarkMinuteWatchAttempt()
	if time.Since(s.LastMinuteAttemptedAt) > time.Second {
		t.Fatal("attempt timestamp was not stamped")
	}
}
