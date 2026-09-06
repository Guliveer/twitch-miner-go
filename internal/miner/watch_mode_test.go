package miner

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/twitch"
)

func harvestState(m *Miner) *bool {
	m.lastWatchingMu.Lock()
	defer m.lastWatchingMu.Unlock()
	return m.lastStreakHarvest
}

// An empty set means nobody is online, which says nothing about the mode. It
// must not be recorded, or the first real selection would find the state
// already set and stay silent.
func TestWatchModeNotRecordedWhileNobodyOnline(t *testing.T) {
	m, _ := newTestMiner(t)

	m.logWatchModeChange(twitch.WatchSet{})

	if harvestState(m) != nil {
		t.Fatal("an empty watch set must leave the mode unrecorded")
	}
}

func TestWatchModeRecordedOnFirstRealSelection(t *testing.T) {
	m, _ := newTestMiner(t)

	m.logWatchModeChange(twitch.WatchSet{
		Streamers:     []*model.Streamer{model.NewStreamer("a")},
		StreakHarvest: true,
		Width:         2,
	})

	state := harvestState(m)
	if state == nil || !*state {
		t.Fatalf("expected harvest mode to be recorded, got %v", state)
	}

	m.logWatchModeChange(twitch.WatchSet{
		Streamers:     []*model.Streamer{model.NewStreamer("a")},
		StreakHarvest: false,
		Width:         20,
	})

	state = harvestState(m)
	if state == nil || *state {
		t.Fatalf("expected the switch back to full width to be recorded, got %v", state)
	}
}
