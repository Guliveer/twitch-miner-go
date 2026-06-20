package twitch

import (
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func makeOnlineStreamer(username string) *model.Streamer {
	s := model.NewStreamer(username)
	s.Settings = model.DefaultStreamerSettings()
	s.IsOnline = true
	s.OnlineAt = time.Now().Add(-2 * time.Minute)
	return s
}

func withCampaigns(s *model.Streamer) *model.Streamer {
	s.Stream.Campaigns = []model.Campaign{{ID: "c1"}}
	return s
}

func TestSelectStreamersToWatchPreservesPriorityOrder(t *testing.T) {
	t.Parallel()

	first := makeOnlineStreamer("first")
	second := makeOnlineStreamer("second")
	third := makeOnlineStreamer("third")

	selected := SelectStreamersToWatch(
		[]*model.Streamer{first, second, third},
		[]model.Priority{model.PriorityOrder},
		2,
	)

	if len(selected) != 2 {
		t.Fatalf("expected 2 selected streamers, got %d", len(selected))
	}
	if selected[0].Username != "first" || selected[1].Username != "second" {
		t.Fatalf("expected stable order [first second], got [%s %s]", selected[0].Username, selected[1].Username)
	}
}

func TestSelectStreamersToWatch_DropsOnly_SkippedWhenNoCampaigns(t *testing.T) {
	t.Parallel()

	done := makeOnlineStreamer("done")
	done.Settings.DropsOnly = true

	other := makeOnlineStreamer("other")

	selected := SelectStreamersToWatch(
		[]*model.Streamer{done, other},
		[]model.Priority{model.PriorityOrder},
		2,
	)

	if len(selected) != 1 || selected[0].Username != "other" {
		t.Fatalf("expected only 'other', got %v", selected)
	}
}

func TestSelectStreamersToWatch_DropsOnly_IncludedWhenHasCampaigns(t *testing.T) {
	t.Parallel()

	active := withCampaigns(makeOnlineStreamer("active"))
	active.Settings.DropsOnly = true

	selected := SelectStreamersToWatch(
		[]*model.Streamer{active},
		[]model.Priority{model.PriorityOrder},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("expected 1 selected streamer, got %d", len(selected))
	}
}

func TestSelectStreamersToWatchPrioritizesMissingStreak(t *testing.T) {
	t.Parallel()

	streak := makeOnlineStreamer("streak")
	streak.Stream.IsWatchStreakMissing = true
	streak.Stream.MinuteWatched = 2
	streak.Settings.WatchStreak = true

	other := makeOnlineStreamer("other")
	other.Stream.IsWatchStreakMissing = false

	selected := SelectStreamersToWatch(
		[]*model.Streamer{other, streak},
		[]model.Priority{model.PriorityStreak, model.PriorityOrder},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("expected 1 selected streamer, got %d", len(selected))
	}
	if selected[0].Username != "streak" {
		t.Fatalf("expected streak-priority streamer to be selected, got %s", selected[0].Username)
	}
}
