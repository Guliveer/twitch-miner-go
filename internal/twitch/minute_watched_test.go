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

func TestSelectStreamersToWatch_DropsOnly_SkippedWhenNoCampaignIDs(t *testing.T) {
	t.Parallel()

	done := makeOnlineStreamer("done")
	done.Settings.DropsOnly = true
	// Campaigns alone (stale detail from a previous sync) must not pass the gate.
	done.Stream.Campaigns = []model.Campaign{{ID: "c1"}}

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

func TestSelectStreamersToWatch_DropsOnly_IncludedWhenHasCampaignIDs(t *testing.T) {
	t.Parallel()

	active := makeOnlineStreamer("active")
	active.Settings.DropsOnly = true
	active.Stream.CampaignIDs = []string{"c1"}

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

func TestSelectStreamersToWatch_Streak_ShortGapResolvedNotPrioritized(t *testing.T) {
	t.Parallel()

	s := makeOnlineStreamer("carry")
	s.Stream.IsWatchStreakMissing = true
	s.Stream.MinuteWatched = 2

	// Simulate: streak was resolved in the previous segment.
	s.Stream.IsWatchStreakMissing = false
	s.Stream.MinuteWatched = 8

	// Offline briefly, then back online — carryover should mark streak as resolved.
	s.Mu.Lock()
	s.IsOnline = false
	s.OfflineAt = time.Now().Add(-10 * time.Minute)
	s.Mu.Unlock()

	s.Mu.Lock()
	s.OnlineAt = time.Now().Add(-2 * time.Minute)
	s.IsOnline = true
	shortGap := !s.OfflineAt.IsZero() && time.Since(s.OfflineAt) <= 30*time.Minute
	streakResolved := !s.Stream.IsWatchStreakMissing
	s.Stream.InitWatchStreak()
	if shortGap && streakResolved {
		s.Stream.IsWatchStreakMissing = false
	}
	s.Mu.Unlock()

	other := makeOnlineStreamer("other")
	other.Stream.IsWatchStreakMissing = true
	other.Stream.MinuteWatched = 2

	selected := SelectStreamersToWatch(
		[]*model.Streamer{s, other},
		[]model.Priority{model.PriorityStreak, model.PriorityOrder},
		1,
	)

	if len(selected) != 1 || selected[0].Username != "other" {
		t.Fatalf("resolved streak + short gap should not be prioritized, got %v", selected)
	}
}

func TestSelectStreamersToWatch_Streak_ShortGapUnresolvedStillPrioritized(t *testing.T) {
	t.Parallel()

	s := makeOnlineStreamer("unresolved")
	s.Stream.IsWatchStreakMissing = true
	s.Stream.MinuteWatched = 2

	// Offline briefly, then back online — streak was never resolved.
	s.Mu.Lock()
	s.IsOnline = false
	s.OfflineAt = time.Now().Add(-10 * time.Minute)
	s.Mu.Unlock()

	s.Mu.Lock()
	s.OnlineAt = time.Now().Add(-2 * time.Minute)
	s.IsOnline = true
	shortGap := !s.OfflineAt.IsZero() && time.Since(s.OfflineAt) <= 30*time.Minute
	streakResolved := !s.Stream.IsWatchStreakMissing
	s.Stream.InitWatchStreak()
	if shortGap && streakResolved {
		s.Stream.IsWatchStreakMissing = false
	}
	s.Mu.Unlock()

	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("streak should be missing after short gap with no prior resolution")
	}

	other := makeOnlineStreamer("other")
	other.Stream.IsWatchStreakMissing = false

	selected := SelectStreamersToWatch(
		[]*model.Streamer{other, s},
		[]model.Priority{model.PriorityStreak, model.PriorityOrder},
		1,
	)

	if len(selected) != 1 || selected[0].Username != "unresolved" {
		t.Fatalf("unresolved streak + short gap should be prioritized, got %v", selected)
	}
}

func TestSelectStreamersToWatch_Streak_LongGapResolvedResets(t *testing.T) {
	t.Parallel()

	s := makeOnlineStreamer("longgap")
	s.Stream.IsWatchStreakMissing = false
	s.Stream.MinuteWatched = 8

	// Offline long (>30 min), then back online — streak should reset.
	s.Mu.Lock()
	s.IsOnline = false
	s.OfflineAt = time.Now().Add(-60 * time.Minute)
	s.Mu.Unlock()

	s.Mu.Lock()
	s.OnlineAt = time.Now().Add(-2 * time.Minute)
	s.IsOnline = true
	shortGap := !s.OfflineAt.IsZero() && time.Since(s.OfflineAt) <= 30*time.Minute
	streakResolved := !s.Stream.IsWatchStreakMissing
	s.Stream.InitWatchStreak()
	if shortGap && streakResolved {
		s.Stream.IsWatchStreakMissing = false
	}
	s.Mu.Unlock()

	if !s.Stream.IsWatchStreakMissing {
		t.Fatal("after long gap, streak should reset to missing")
	}
}

func TestSelectStreamersToWatch_Freeze_FrozenStreamerSkipped(t *testing.T) {
	t.Parallel()

	frozen := makeOnlineStreamer("frozen")
	frozen.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)

	healthy := makeOnlineStreamer("healthy")
	healthy.Stream.LastMinuteCreditedAt = time.Now().Add(-1 * time.Minute)

	selected := SelectStreamersToWatch(
		[]*model.Streamer{frozen, healthy},
		[]model.Priority{model.PriorityOrder},
		2,
	)

	if len(selected) != 1 || selected[0].Username != "healthy" {
		t.Fatalf("expected only 'healthy', got %v", selected)
	}
}

func TestSelectStreamersToWatch_Freeze_NeverCreditedNotFrozen(t *testing.T) {
	t.Parallel()

	newbie := makeOnlineStreamer("newbie")
	newbie.Stream.LastMinuteCreditedAt = time.Time{}

	selected := SelectStreamersToWatch(
		[]*model.Streamer{newbie},
		[]model.Priority{model.PriorityOrder},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("streamer with zero LastMinuteCreditedAt should not be frozen, got %d selected", len(selected))
	}
}

func TestSelectStreamersToWatch_Freeze_AllFrozenReturnsNone(t *testing.T) {
	t.Parallel()

	a := makeOnlineStreamer("a")
	a.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)

	b := makeOnlineStreamer("b")
	b.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)

	selected := SelectStreamersToWatch(
		[]*model.Streamer{a, b},
		[]model.Priority{model.PriorityOrder},
		2,
	)

	if len(selected) != 0 {
		t.Fatalf("expected 0 selected when all frozen, got %d", len(selected))
	}
}

func TestSelectStreamersToWatch_Freeze_BelowThresholdNotFrozen(t *testing.T) {
	t.Parallel()

	recently := makeOnlineStreamer("recently")
	recently.Stream.LastMinuteCreditedAt = time.Now().Add(-3 * time.Minute)

	selected := SelectStreamersToWatch(
		[]*model.Streamer{recently},
		[]model.Priority{model.PriorityOrder},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("streamer credited 3min ago should not be frozen, got %d selected", len(selected))
	}
}

func TestSelectStreamersToWatch_EndingSoonest(t *testing.T) {
	t.Parallel()

	earlyEnd := makeOnlineStreamer("early")
	earlyEnd.Stream.Campaigns = []model.Campaign{
		{ID: "c1", EndAt: time.Now().Add(1 * time.Hour), IsWithinTimeWindow: true},
	}

	lateEnd := makeOnlineStreamer("late")
	lateEnd.Stream.Campaigns = []model.Campaign{
		{ID: "c2", EndAt: time.Now().Add(10 * time.Hour), IsWithinTimeWindow: true},
	}

	selected := SelectStreamersToWatch(
		[]*model.Streamer{lateEnd, earlyEnd},
		[]model.Priority{model.PriorityEndingSoonest},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("expected 1 selected streamer, got %d", len(selected))
	}
	if selected[0].Username != "early" {
		t.Fatalf("expected 'early' (soonest ending) to be selected, got %s", selected[0].Username)
	}
}

func TestSelectStreamersToWatch_EndingSoonest_SkipsNoCampaigns(t *testing.T) {
	t.Parallel()

	hasCampaigns := makeOnlineStreamer("has")
	hasCampaigns.Stream.Campaigns = []model.Campaign{
		{ID: "c1", EndAt: time.Now().Add(5 * time.Hour), IsWithinTimeWindow: true},
	}

	noCampaigns := makeOnlineStreamer("none")

	selected := SelectStreamersToWatch(
		[]*model.Streamer{noCampaigns, hasCampaigns},
		[]model.Priority{model.PriorityEndingSoonest},
		1,
	)

	if len(selected) != 1 || selected[0].Username != "has" {
		t.Fatalf("expected 'has' to be selected, got %v", selected)
	}
}

func TestSelectStreamersToWatch_LowAvailabilityFirst(t *testing.T) {
	t.Parallel()

	moreDrops := makeOnlineStreamer("many")
	moreDrops.Stream.Campaigns = []model.Campaign{
		{ID: "c1", Drops: make([]*model.Drop, 5), IsWithinTimeWindow: true},
	}

	fewerDrops := makeOnlineStreamer("few")
	fewerDrops.Stream.Campaigns = []model.Campaign{
		{ID: "c2", Drops: make([]*model.Drop, 1), IsWithinTimeWindow: true},
	}

	selected := SelectStreamersToWatch(
		[]*model.Streamer{moreDrops, fewerDrops},
		[]model.Priority{model.PriorityLowAvailabilityFirst},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("expected 1 selected streamer, got %d", len(selected))
	}
	if selected[0].Username != "few" {
		t.Fatalf("expected 'few' (lowest availability) to be selected, got %s", selected[0].Username)
	}
}

func TestSelectStreamersToWatch_LowAvailabilityFirst_SkipsNoCampaigns(t *testing.T) {
	t.Parallel()

	hasDrops := makeOnlineStreamer("has")
	hasDrops.Stream.Campaigns = []model.Campaign{
		{ID: "c1", Drops: make([]*model.Drop, 3), IsWithinTimeWindow: true},
	}

	noDrops := makeOnlineStreamer("none")

	selected := SelectStreamersToWatch(
		[]*model.Streamer{noDrops, hasDrops},
		[]model.Priority{model.PriorityLowAvailabilityFirst},
		1,
	)

	if len(selected) != 1 || selected[0].Username != "has" {
		t.Fatalf("expected 'has' to be selected, got %v", selected)
	}
}

func TestSelectStreamersToWatch_EndingSoonest_PicksFirstEnd(t *testing.T) {
	t.Parallel()

	multi := makeOnlineStreamer("multi")
	multi.Stream.Campaigns = []model.Campaign{
		{ID: "c1", EndAt: time.Now().Add(20 * time.Hour), IsWithinTimeWindow: true},
		{ID: "c2", EndAt: time.Now().Add(2 * time.Hour), IsWithinTimeWindow: true},
	}

	single := makeOnlineStreamer("single")
	single.Stream.Campaigns = []model.Campaign{
		{ID: "c3", EndAt: time.Now().Add(10 * time.Hour), IsWithinTimeWindow: true},
	}

	selected := SelectStreamersToWatch(
		[]*model.Streamer{single, multi},
		[]model.Priority{model.PriorityEndingSoonest},
		1,
	)

	if len(selected) != 1 || selected[0].Username != "multi" {
		t.Fatalf("expected 'multi' (has soonest ending campaign at 2h) to be selected, got %v", selected)
	}
}

func TestStalledCooldown_FrozenStreamerGetsCooldown(t *testing.T) {
	t.Parallel()

	s := makeOnlineStreamer("frozen")
	s.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)

	SelectStreamersToWatch(
		[]*model.Streamer{s},
		[]model.Priority{model.PriorityOrder},
		1,
	)

	s.Mu.RLock()
	cooldown := s.Stream.StalledCooldownUntil
	s.Mu.RUnlock()

	if cooldown.IsZero() {
		t.Fatal("expected StalledCooldownUntil to be set after freeze detection")
	}
}

func TestStalledCooldown_InCooldownSkipped(t *testing.T) {
	t.Parallel()

	s := makeOnlineStreamer("cooldown")
	s.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)
	s.Stream.StalledCooldownUntil = time.Now().Add(15 * time.Minute)

	other := makeOnlineStreamer("other")
	other.Stream.LastMinuteCreditedAt = time.Now().Add(-1 * time.Minute)

	selected := SelectStreamersToWatch(
		[]*model.Streamer{s, other},
		[]model.Priority{model.PriorityOrder},
		1,
	)

	if len(selected) != 1 || selected[0].Username != "other" {
		t.Fatalf("expected 'other', got %v", selected)
	}
}

func TestStalledCooldown_ExpiredCooldownAllowed(t *testing.T) {
	t.Parallel()

	s := makeOnlineStreamer("expired")
	s.Stream.LastMinuteCreditedAt = time.Now().Add(-1 * time.Minute)
	s.Stream.StalledCooldownUntil = time.Now().Add(-1 * time.Minute)

	selected := SelectStreamersToWatch(
		[]*model.Streamer{s},
		[]model.Priority{model.PriorityOrder},
		1,
	)

	if len(selected) != 1 {
		t.Fatalf("expected streamer with expired cooldown to be selected, got %d", len(selected))
	}
}

func TestStalledCooldown_AllInCooldownReturnsNone(t *testing.T) {
	t.Parallel()

	a := makeOnlineStreamer("a")
	a.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)
	a.Stream.StalledCooldownUntil = time.Now().Add(20 * time.Minute)

	b := makeOnlineStreamer("b")
	b.Stream.LastMinuteCreditedAt = time.Now().Add(-10 * time.Minute)
	b.Stream.StalledCooldownUntil = time.Now().Add(20 * time.Minute)

	selected := SelectStreamersToWatch(
		[]*model.Streamer{a, b},
		[]model.Priority{model.PriorityOrder},
		2,
	)

	if len(selected) != 0 {
		t.Fatalf("expected 0 when all in cooldown, got %d", len(selected))
	}
}
