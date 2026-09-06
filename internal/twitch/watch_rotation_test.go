package twitch

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// streakPending builds an online streamer that still owes a watch streak.
func streakPending(username string, minuteWatched float64) *model.Streamer {
	s := makeOnlineStreamer(username)
	s.Settings.WatchStreak = true
	s.Stream.IsWatchStreakMissing = true
	s.Stream.MinuteWatched = minuteWatched
	return s
}

// streakSettled builds an online streamer whose streak has already landed.
func streakSettled(username string) *model.Streamer {
	s := makeOnlineStreamer(username)
	s.Settings.WatchStreak = true
	s.Stream.IsWatchStreakMissing = false
	return s
}

func usernames(set WatchSet) []string {
	out := make([]string, 0, len(set.Streamers))
	for _, s := range set.Streamers {
		out = append(out, s.Username)
	}
	return out
}

func rotationOptions(preferred ...string) WatchOptions {
	return WatchOptions{
		Priorities: []model.Priority{
			model.PriorityStreak,
			model.PriorityPreferred,
			model.PriorityOrder,
		},
		MaxWatch:      20,
		StreakWatch:   2,
		StreakMinutes: 10,
		Preferred:     preferred,
	}
}

func TestWatchSetNarrowsWhileStreaksPending(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakSettled("filler1"),
		streakPending("needs-a", 3),
		streakSettled("filler2"),
		streakPending("needs-b", 1),
		streakSettled("filler3"),
	}

	set := SelectWatchSet(streamers, rotationOptions())

	if !set.StreakHarvest {
		t.Error("expected streak harvest mode while streaks are pending")
	}
	if set.Width != 2 {
		t.Errorf("expected the set to narrow to 2, got %d", set.Width)
	}
	if len(set.Streamers) != 2 {
		t.Fatalf("expected 2 watched streamers, got %v", usernames(set))
	}
	for _, s := range set.Streamers {
		if s.Username != "needs-a" && s.Username != "needs-b" {
			t.Errorf("a settled channel took a streak slot: %s", s.Username)
		}
	}
}

func TestWatchSetWidensOnceStreaksSettled(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakSettled("a"), streakSettled("b"), streakSettled("c"),
		streakSettled("d"), streakSettled("e"),
	}

	set := SelectWatchSet(streamers, rotationOptions())

	if set.StreakHarvest {
		t.Error("no streak is pending, harvest mode should be off")
	}
	if set.Width != 20 {
		t.Errorf("expected full width 20, got %d", set.Width)
	}
	if len(set.Streamers) != 5 {
		t.Fatalf("expected all 5 streamers watched, got %v", usernames(set))
	}
}

// A channel that used up its slot without the streak arriving must release it,
// otherwise the narrowed set would never widen again.
func TestWatchSetReleasesChannelPastStreakWindow(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakPending("exhausted", 12),
		streakSettled("settled"),
	}

	set := SelectWatchSet(streamers, rotationOptions())

	if set.StreakHarvest {
		t.Error("a channel past its streak window should not hold the set narrow")
	}
	if len(set.Streamers) != 2 {
		t.Fatalf("expected the widened set to include both, got %v", usernames(set))
	}
}

// The slot goes to whoever is furthest along, so a channel mid-window finishes
// instead of the set churning every tick.
func TestStreakSlotsGoToChannelsClosestToTheirStreak(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakPending("fresh", 0),
		streakPending("almost", 8),
		streakPending("halfway", 4),
	}

	set := SelectWatchSet(streamers, rotationOptions())

	got := usernames(set)
	if len(got) != 2 || got[0] != "almost" || got[1] != "halfway" {
		t.Fatalf("expected [almost halfway], got %v", got)
	}
}

func TestPreferredChannelsTakeSlotsInListedOrder(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakSettled("random1"),
		streakSettled("insym"),
		streakSettled("random2"),
		streakSettled("jimpanse247"),
		streakSettled("gronkh"),
	}

	opts := rotationOptions("jimpanse", "jimpanse247", "insym", "gronkh")
	opts.MaxWatch = 3

	got := usernames(SelectWatchSet(streamers, opts))

	// jimpanse is not live, so the list falls through to the next entries.
	want := []string{"jimpanse247", "insym", "gronkh"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestPreferredMatchingIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakSettled("other"),
		streakSettled("insym"),
	}

	opts := rotationOptions("  InSym  ")
	opts.MaxWatch = 1

	got := usernames(SelectWatchSet(streamers, opts))
	if len(got) != 1 || got[0] != "insym" {
		t.Fatalf("expected [insym], got %v", got)
	}
}

// Streak harvesting outranks the preferred list: that is the ordering the
// operator chose, favourites resume once nothing is pending.
func TestStreaksOutrankPreferredChannels(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakSettled("insym"),
		streakPending("needs-a", 2),
		streakPending("needs-b", 5),
	}

	got := usernames(SelectWatchSet(streamers, rotationOptions("insym")))

	if len(got) != 2 {
		t.Fatalf("expected 2 watched streamers, got %v", got)
	}
	for _, name := range got {
		if name == "insym" {
			t.Fatalf("preferred channel took a slot while streaks were pending: %v", got)
		}
	}
}

func TestStreakWatchZeroDisablesNarrowing(t *testing.T) {
	t.Parallel()

	streamers := []*model.Streamer{
		streakPending("needs-a", 1),
		streakSettled("b"),
		streakSettled("c"),
	}

	opts := rotationOptions()
	opts.StreakWatch = 0

	set := SelectWatchSet(streamers, opts)
	if set.StreakHarvest {
		t.Error("streak_watch_streams=0 must disable the narrowing")
	}
	if len(set.Streamers) != 3 {
		t.Fatalf("expected all 3 watched, got %v", usernames(set))
	}
}

func TestWatchSetEmptyWhenNobodyOnline(t *testing.T) {
	t.Parallel()

	set := SelectWatchSet(nil, rotationOptions())
	if len(set.Streamers) != 0 || set.StreakHarvest {
		t.Fatalf("expected an empty set, got %+v", set)
	}
}
