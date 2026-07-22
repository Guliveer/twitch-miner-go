package model

import (
	"testing"
	"time"
)

func TestSetOnline_Carryover_StreakResolved_ShortGap(t *testing.T) {
	t.Parallel()

	s := NewStreamer("carry")
	s.Settings = DefaultStreamerSettings()
	s.IsOnline = true
	s.OnlineAt = time.Now().Add(-50 * time.Minute)
	s.Stream.InitWatchStreak()

	// Simulate: streak was resolved in the previous segment.
	s.Stream.IsWatchStreakMissing = false
	s.Stream.MinuteWatched = 8

	// Streamer goes offline briefly (<30 min).
	s.SetOffline()
	if s.OfflineAt.IsZero() {
		t.Fatal("OfflineAt should be set after SetOffline")
	}

	// Come back online within 30 minutes.
	time.Sleep(50 * time.Millisecond)
	s.SetOnline()

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// The streak was already resolved — should NOT be marked as missing again.
	if s.Stream.IsWatchStreakMissing {
		t.Error("streak was resolved before offline; should carry over as resolved (not missing)")
	}
}

func TestSetOnline_Carryover_StreakUnresolved_ShortGap(t *testing.T) {
	t.Parallel()

	s := NewStreamer("unresolved")
	s.Settings = DefaultStreamerSettings()
	s.IsOnline = true
	s.OnlineAt = time.Now().Add(-50 * time.Minute)
	s.Stream.InitWatchStreak()

	// Streak was NOT resolved — still missing.
	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("streak should be missing after InitWatchStreak")
	}

	// Streamer goes offline briefly.
	s.SetOffline()

	// Come back online within 30 minutes.
	time.Sleep(50 * time.Millisecond)
	s.SetOnline()

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// Streak was never resolved — should remain missing.
	if !s.Stream.IsWatchStreakMissing {
		t.Error("streak was never resolved; should remain missing after short offline gap")
	}
}

func TestSetOnline_Carryover_StreakResolved_LongGap(t *testing.T) {
	t.Parallel()

	s := NewStreamer("longgap")
	s.Settings = DefaultStreamerSettings()
	s.IsOnline = true
	s.OnlineAt = time.Now().Add(-5 * time.Hour)
	s.Stream.InitWatchStreak()

	// Streak resolved in previous segment.
	s.Stream.IsWatchStreakMissing = false

	// Simulate long offline gap (>30 min) by setting OfflineAt in the past.
	s.IsOnline = false
	s.OfflineAt = time.Now().Add(-60 * time.Minute)

	// Come back online after >30 min.
	s.SetOnline()

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// Long gap — streak state should be fully reset (fresh segment).
	if !s.Stream.IsWatchStreakMissing {
		t.Error("after long offline gap, streak should be reset to missing (new segment)")
	}
	if s.Stream.MinuteWatched != 0 {
		t.Errorf("after long offline gap, MinuteWatched should reset to 0, got %f", s.Stream.MinuteWatched)
	}
}

func TestSetOnline_Carryover_FirstOnline(t *testing.T) {
	t.Parallel()

	s := NewStreamer("first")
	s.Settings = DefaultStreamerSettings()

	// First time online — OfflineAt is zero.
	s.SetOnline()

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// Should start with streak missing.
	if !s.Stream.IsWatchStreakMissing {
		t.Error("first-time online should start with streak missing")
	}
}

func TestSetOnline_AlreadyOnline_NoOp(t *testing.T) {
	t.Parallel()

	s := NewStreamer("noop")
	s.Settings = DefaultStreamerSettings()
	s.IsOnline = true
	s.Stream.IsWatchStreakMissing = false

	s.SetOnline()

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// Should not reset — already online.
	if s.Stream.IsWatchStreakMissing {
		t.Error("SetOnline on already-online streamer should be a no-op")
	}
}
