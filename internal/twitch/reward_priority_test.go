package twitch

import (
	"slices"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func deadlineDropStreamer(name string, now time.Time, deadline time.Duration, required, watched int) *model.Streamer {
	s := makeOnlineStreamer(name)
	s.Settings.ClaimDrops = true
	s.ChannelID = name
	s.Stream.CampaignIDs = []string{name}
	d := model.NewDrop(name, name, nil, required, now.Add(-time.Hour), now.Add(deadline))
	d.Update(true, watched, "", false)
	s.Stream.Campaigns = []model.Campaign{{ID: name, EndAt: now.Add(deadline), Drops: []*model.Drop{d}}}
	return s
}

func deadlineBadgeStreamer(now time.Time, deadline, remaining time.Duration) *model.Streamer {
	s := makeOnlineStreamer("badge")
	s.IsBadgeWatched = true
	s.BadgeReward = model.RewardDeadline{EndAt: now.Add(deadline), Remaining: remaining}
	return s
}

func TestUrgentRewardPriority(t *testing.T) {
	t.Parallel()

	now := time.Now()
	drop := deadlineDropStreamer("drop", now, 5*time.Hour, 250, 10)
	badge := deadlineBadgeStreamer(now, 2*time.Hour, 30*time.Minute)
	month := deadlineDropStreamer("month", now, 30*24*time.Hour, 60, 0)

	opts := WatchOptions{MaxWatch: 2, Priorities: []model.Priority{
		model.PriorityUrgentRewards, model.PriorityUrgentRewards, model.PriorityOrder,
	}}
	streamers := []*model.Streamer{month, badge, drop}
	if got := usernames(selectWatchSetAt(streamers, opts, now)); !slices.Equal(got, []string{"drop", "month"}) {
		t.Fatalf("selection = %v; want tightest slack and one ordinary slot", got)
	}
	opts.Priorities = []model.Priority{model.PriorityBadges, model.PriorityOrder}
	if got := usernames(selectWatchSetAt(streamers, opts, now)); !slices.Equal(got, []string{"badge", "month"}) {
		t.Fatalf("selection without opt-in = %v", got)
	}
}

func TestUrgentRewardTieAndContinuity(t *testing.T) {
	t.Parallel()

	now := time.Now()
	drop := deadlineDropStreamer("drop", now, time.Hour, 32, 0)
	badge := deadlineBadgeStreamer(now, time.Hour, 30*time.Minute)
	opts := WatchOptions{MaxWatch: 1, Priorities: []model.Priority{model.PriorityUrgentRewards}}
	streamers := []*model.Streamer{drop, badge}
	if got := usernames(selectWatchSetAt(streamers, opts, now)); !slices.Equal(got, []string{"badge"}) {
		t.Fatalf("near-tie = %v; want badge", got)
	}
	opts.PreviouslyWatched = map[string]bool{"drop": true}
	if got := usernames(selectWatchSetAt(streamers, opts, now)); !slices.Equal(got, []string{"drop"}) {
		t.Fatalf("near-tie with watch in progress = %v; want continuity", got)
	}
	badge.BadgeReward.Remaining = 40 * time.Minute
	if got := usernames(selectWatchSetAt(streamers, opts, now)); !slices.Equal(got, []string{"badge"}) {
		t.Fatalf("decisively more urgent badge = %v", got)
	}
}

func TestUrgentRewardRejectsUnusableRewards(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*model.Streamer, *model.Drop, time.Time){
		"too late":             func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.MinutesRequired = 61 },
		"inside safety margin": func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.MinutesRequired = 56 },
		"expired":              func(_ *model.Streamer, d *model.Drop, now time.Time) { d.EndAt = now },
		"not started":          func(_ *model.Streamer, d *model.Drop, now time.Time) { d.StartAt = now.Add(time.Minute) },
		"claimed":              func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.IsClaimed = true },
		"claimable":            func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.IsClaimable = true },
		"paid":                 func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.RequiresSubscription = true },
		"locked":               func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.Update(false, 0, "", false) },
		"unknown progress":     func(_ *model.Streamer, d *model.Drop, _ time.Time) { d.HasPreconditionsMet = nil },
		"unconfirmed channel":  func(s *model.Streamer, _ *model.Drop, _ time.Time) { s.Stream.CampaignIDs = []string{"other"} },
		"restricted channel": func(s *model.Streamer, _ *model.Drop, _ time.Time) {
			s.Stream.Campaigns[0].Channels = []string{"other"}
		},
		"drops disabled": func(s *model.Streamer, _ *model.Drop, _ time.Time) { s.Settings.ClaimDrops = false },
		"offline":        func(s *model.Streamer, _ *model.Drop, _ time.Time) { s.IsOnline = false },
		"cooldown": func(s *model.Streamer, _ *model.Drop, now time.Time) {
			s.Stream.StalledCooldownUntil = now.Add(time.Minute)
		},
	}
	for name, modify := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			now := time.Now()
			drop := deadlineDropStreamer("drop", now, time.Hour, 30, 0)
			modify(drop, drop.Stream.Campaigns[0].Drops[0], now)
			plain := makeOnlineStreamer("plain")
			opts := WatchOptions{MaxWatch: 1, Priorities: []model.Priority{model.PriorityUrgentRewards, model.PriorityOrder}}
			if got := usernames(selectWatchSetAt([]*model.Streamer{plain, drop}, opts, now)); !slices.Equal(got, []string{"plain"}) {
				t.Fatalf("selection = %v; want normal fallback", got)
			}
		})
	}
}

func TestUrgentRewardKeepsStreakWidth(t *testing.T) {
	t.Parallel()

	now := time.Now()
	badge := deadlineBadgeStreamer(now, time.Hour, 30*time.Minute)
	opts := rotationOptions()
	opts.StreakWatch = 1
	opts.Priorities = []model.Priority{model.PriorityUrgentRewards, model.PriorityStreak, model.PriorityOrder}
	got := selectWatchSetAt([]*model.Streamer{streakPending("streak", 1), badge}, opts, now)
	if !got.StreakHarvest || got.Width != 1 || !slices.Equal(usernames(got), []string{"badge"}) {
		t.Fatalf("selection = %+v (%v); want one badge slot", got, usernames(got))
	}
}

func TestUrgentRewardUsesNearestTierAndEarlierDeadline(t *testing.T) {
	t.Parallel()

	now := time.Now()
	s := deadlineDropStreamer("drop", now, 4*time.Hour, 60, 0)
	c := &s.Stream.Campaigns[0]
	later := model.NewDrop("later", "Later tier", nil, 220, now.Add(-time.Hour), c.EndAt)
	later.Update(true, 0, "", false)
	c.Drops = append(c.Drops, later)
	// The higher tier is urgent, but the nearer reward has three hours of slack.
	if got := campaignRewardStart(c, now); !got.IsZero() {
		t.Fatalf("unexpected urgent higher tier: %s", got)
	}
	c.Drops[0].EndAt = now.Add(90 * time.Minute)
	if got, want := campaignRewardStart(c, now), now.Add(30*time.Minute); !got.Equal(want) {
		t.Fatalf("latest start = %s; want %s", got, want)
	}
	c.Drops[0].EndAt = now.Add(8 * time.Hour)
	c.EndAt = now.Add(75 * time.Minute)
	if got, want := campaignRewardStart(c, now), now.Add(15*time.Minute); !got.Equal(want) {
		t.Fatalf("campaign must cap the reward deadline: %s; want %s", got, want)
	}
}
