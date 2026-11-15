package twitch

import (
	"slices"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

const (
	urgentRewardWindow = 2 * time.Hour
	rewardSwitchMargin = 5 * time.Minute
)

type urgentReward struct {
	index       int
	latestStart time.Time
	badge       bool
}

func applyUrgentReward(streamers []*model.Streamer, online []int, watching map[int]struct{}, add func(int) bool, previous map[string]bool, now time.Time) {
	candidates := make([]urgentReward, 0, len(online))
	var earliest time.Time
	for _, idx := range online {
		if _, selected := watching[idx]; selected {
			continue
		}
		start, badge := streamerRewardStart(streamers[idx], now)
		if start.IsZero() {
			continue
		}
		candidates = append(candidates, urgentReward{idx, start, badge})
		if earliest.IsZero() || start.Before(earliest) {
			earliest = start
		}
	}

	// Within the switching margin prefer continuity, then badges. Compare
	// against the true earliest start so near-ties cannot accumulate drift.
	var best *urgentReward
	for i := range candidates {
		c := &candidates[i]
		if c.latestStart.Sub(earliest) > rewardSwitchMargin {
			continue
		}
		if best == nil || preferUrgentReward(*c, *best, streamers, previous) {
			best = c
		}
	}
	if best != nil {
		add(best.index)
	}
}

func preferUrgentReward(a, b urgentReward, streamers []*model.Streamer, previous map[string]bool) bool {
	wasA, wasB := previous[streamers[a.index].Username], previous[streamers[b.index].Username]
	if wasA != wasB {
		return wasA
	}
	if a.badge != b.badge {
		return a.badge
	}
	return a.latestStart.Before(b.latestStart)
}

func streamerRewardStart(s *model.Streamer, now time.Time) (time.Time, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	var best time.Time
	if s.IsBadgeWatched {
		best = urgentStart(s.BadgeReward, now)
	}
	badge := !best.IsZero()
	if s.Settings == nil || !s.Settings.ClaimDrops {
		return best, badge
	}
	for i := range s.Stream.Campaigns {
		campaign := &s.Stream.Campaigns[i]
		if !campaignMatchesStreamer(campaign, s) || now.Before(campaign.StartAt) ||
			(len(campaign.Channels) > 0 && !slices.Contains(campaign.Channels, s.ChannelID)) {
			continue
		}
		if start := campaignRewardStart(campaign, now); !start.IsZero() && (best.IsZero() || start.Before(best)) {
			best, badge = start, false
		}
	}
	return best, badge
}

func campaignRewardStart(c *model.Campaign, now time.Time) time.Time {
	var earliest time.Time
	for _, drop := range c.Drops {
		start := urgentStart(availableDropDeadline(c.EndAt, drop, now), now)
		if !start.IsZero() && (earliest.IsZero() || start.Before(earliest)) {
			earliest = start
		}
	}
	return earliest
}

func availableDropDeadline(campaignEnd time.Time, d *model.Drop, now time.Time) model.RewardDeadline {
	if d.IsClaimed || d.IsClaimable || d.IsSynthetic || d.RequiresSubscription || d.HasPreconditionsMet == nil ||
		!*d.HasPreconditionsMet || now.Before(d.StartAt) || d.EndAt.IsZero() || campaignEnd.IsZero() ||
		d.CurrentMinutesWatched < 0 || d.MinutesRequired <= d.CurrentMinutesWatched {
		return model.RewardDeadline{}
	}
	end := d.EndAt
	if campaignEnd.Before(end) {
		end = campaignEnd
	}
	remaining := d.MinutesRequired - d.CurrentMinutesWatched
	// Avoid duration overflow from malformed external values.
	if remaining > int(end.Sub(now)/time.Minute) {
		return model.RewardDeadline{}
	}
	return model.RewardDeadline{EndAt: end, Remaining: time.Duration(remaining) * time.Minute}
}

func urgentStart(reward model.RewardDeadline, now time.Time) time.Time {
	if reward.EndAt.IsZero() || reward.Remaining <= 0 {
		return time.Time{}
	}
	start := reward.LatestStart()
	slack := start.Sub(now)
	if slack < model.RewardDeadlineSafetyMargin || slack > urgentRewardWindow {
		return time.Time{}
	}
	return start
}
