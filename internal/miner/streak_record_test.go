package miner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/streakstore"
)

// pointsEarnedMessage builds the PubSub message Twitch sends when points land.
func pointsEarnedMessage(reasonCode string, amount int) *model.Message {
	return &model.Message{
		Type: model.MsgTypePointsEarned,
		Data: map[string]any{
			"balance": map[string]any{"balance": 1000},
			"point_gain": map[string]any{
				"total_points": amount,
				"reason_code":  reasonCode,
			},
		},
	}
}

func onlineStreamerWithBroadcast(username, broadcastID string) *model.Streamer {
	s := model.NewStreamer(username)
	s.Settings = model.DefaultStreamerSettings()
	s.IsOnline = true
	s.Stream.BroadcastID = broadcastID
	s.Stream.IsWatchStreakMissing = true
	return s
}

// The reason code arrives as "WATCH_STREAK" but is stored under the mapped
// event name, which is how the flag stopped being cleared in the first place.
func TestWatchStreakPayoutClearsTheFlag(t *testing.T) {
	m, _ := newTestMiner(t)
	s := onlineStreamerWithBroadcast("krissi", "b-1")

	m.handleCommunityPoints(context.Background(), pointsEarnedMessage("WATCH_STREAK", 350), s)

	if s.Stream.IsWatchStreakMissing {
		t.Fatal("a WATCH_STREAK payout must release the channel from the rotation")
	}
	if got := s.History[string(model.EventGainForWatchStreak)]; got == nil || got.Amount != 350 {
		t.Errorf("expected the payout recorded under the mapped event name, got %+v", got)
	}
}

func TestOrdinaryWatchPayoutLeavesTheFlag(t *testing.T) {
	m, _ := newTestMiner(t)
	s := onlineStreamerWithBroadcast("krissi", "b-1")

	m.handleCommunityPoints(context.Background(), pointsEarnedMessage("WATCH", 10), s)

	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("plain watch points are not a streak and must not release the channel")
	}
}

func TestRememberWatchStreakIsSafeWithoutAStore(t *testing.T) {
	m, _ := newTestMiner(t)
	m.streaks = nil

	// Must not panic; persistence is optional.
	m.rememberWatchStreak("krissi", "b-1")

	if m.WatchStreakEarned("krissi", "b-1") {
		t.Error("without a store nothing can be reported as earned")
	}
}

// End-to-end over the production path: the PubSub payload drives the real
// handler, which writes through the real store to disk, and a fresh store
// opened from that directory — as a restart would — still knows about it.
func TestWatchStreakPayoutIsPersistedAndSurvivesReopen(t *testing.T) {
	dir := t.TempDir()

	m, _ := newTestMiner(t)
	store, err := streakstore.Open(dir, "3jakec")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	m.streaks = store

	s := onlineStreamerWithBroadcast("krissi", "broadcast-42")
	m.handleCommunityPoints(context.Background(), pointsEarnedMessage("WATCH_STREAK", 350), s)

	if _, err := os.Stat(filepath.Join(dir, "3jakec.json")); err != nil {
		t.Fatalf("payout was not written to disk: %v", err)
	}

	restarted, err := streakstore.Open(dir, "3jakec")
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	if !restarted.Earned("krissi", "broadcast-42") {
		t.Fatal("the payout did not survive a restart")
	}
	// A new broadcast must still be chased.
	if restarted.Earned("krissi", "broadcast-43") {
		t.Error("a later broadcast must not inherit the record")
	}
}

// Without a broadcast ID there is nothing to key the record on, and guessing
// would suppress a real chase.
func TestWatchStreakPayoutWithoutBroadcastIsNotPersisted(t *testing.T) {
	dir := t.TempDir()

	m, _ := newTestMiner(t)
	store, err := streakstore.Open(dir, "3jakec")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	m.streaks = store

	s := onlineStreamerWithBroadcast("krissi", "")
	m.handleCommunityPoints(context.Background(), pointsEarnedMessage("WATCH_STREAK", 350), s)

	if store.Len() != 0 {
		t.Error("a payout without a broadcast ID must not be recorded")
	}
	// The in-memory flag is still cleared — the payout did happen.
	if s.Stream.IsWatchStreakMissing {
		t.Error("the payout should still release the channel for this run")
	}
}
