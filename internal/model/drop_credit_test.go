package model

import (
	"testing"
	"time"
)

func TestDropAwaitingWatchCredit(t *testing.T) {
	t.Parallel()

	now := time.Now()
	met := true

	tests := []struct {
		name   string
		mutate func(*Drop)
		want   bool
	}{
		{name: "in progress", mutate: func(*Drop) {}, want: true},
		{name: "completed", mutate: func(d *Drop) { d.CurrentMinutesWatched = d.MinutesRequired }},
		{name: "claimed", mutate: func(d *Drop) { d.IsClaimed = true }},
		{name: "claimable", mutate: func(d *Drop) { d.IsClaimable = true }},
		{name: "synthetic", mutate: func(d *Drop) { d.IsSynthetic = true }},
		{name: "subscription required", mutate: func(d *Drop) { d.RequiresSubscription = true }},
		{name: "preconditions not met", mutate: func(d *Drop) {
			blocked := false
			d.HasPreconditionsMet = &blocked
		}},
		{name: "preconditions unknown", mutate: func(d *Drop) { d.HasPreconditionsMet = nil }},
		{name: "not started", mutate: func(d *Drop) {
			d.StartAt = now.Add(time.Hour)
			d.EndAt = now.Add(2 * time.Hour)
		}},
		{name: "already ended", mutate: func(d *Drop) {
			d.StartAt = now.Add(-2 * time.Hour)
			d.EndAt = now.Add(-time.Hour)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			drop := NewDrop("d1", "Test Drop", nil, 60, now.Add(-time.Hour), now.Add(time.Hour))
			drop.HasPreconditionsMet = &met
			tt.mutate(drop)

			if got := drop.AwaitingWatchCredit(now); got != tt.want {
				t.Errorf("AwaitingWatchCredit() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestObserveDropCreditBaselinesFirstSight(t *testing.T) {
	t.Parallel()

	s := NewStream()
	now := time.Now()

	if idle := s.ObserveDropCredit("d1", 10, now); idle != 0 {
		t.Fatalf("first observation should only baseline, got %s", idle)
	}
	if idle := s.ObserveDropCredit("d1", 10, now.Add(20*time.Minute)); idle != 20*time.Minute {
		t.Fatalf("unchanged progress should report idle time, got %s", idle)
	}
}

func TestObserveDropCreditResetsWhenProgressMoves(t *testing.T) {
	t.Parallel()

	s := NewStream()
	now := time.Now()

	s.ObserveDropCredit("d1", 10, now)
	s.ObserveDropCredit("d1", 10, now.Add(20*time.Minute))

	if idle := s.ObserveDropCredit("d1", 11, now.Add(30*time.Minute)); idle != 0 {
		t.Fatalf("progress that moved starts a new stretch, got %s", idle)
	}
}

// A long pause means the channel was not being watched, so the progress it
// stood still for must not be charged to the new watch segment.
func TestMarkMinuteWatchAttemptResetsCreditTrackerOnNewSegment(t *testing.T) {
	t.Parallel()

	s := NewStream()
	now := time.Now()
	s.ObserveDropCredit("d1", 10, now)
	s.LastMinuteAttemptedAt = now
	s.MarkMinuteWatchAttempt()

	if idle := s.ObserveDropCredit("d1", 10, now.Add(10*time.Minute)); idle != 10*time.Minute {
		t.Fatalf("continuous watching keeps the baseline, got %s", idle)
	}

	s.LastMinuteAttemptedAt = time.Now().Add(-10 * time.Minute)
	s.MarkMinuteWatchAttempt()

	if idle := s.ObserveDropCredit("d1", 10, time.Now()); idle != 0 {
		t.Fatalf("a new watch segment should baseline again, got %s", idle)
	}
}
