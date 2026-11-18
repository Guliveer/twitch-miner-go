package twitch

import (
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// detectStalledDropCredit cools down streamers whose watched drops stopped
// advancing.
//
// Twitch acknowledges minute-watched events it does not count, so spade
// acceptance alone cannot prove the account is being credited. Drop progress
// comes from inventory and only moves when watch time is actually counted,
// which makes it the reliable signal.
//
// Progress is account-wide, not per channel, so two streamers sharing a
// campaign would make an advancing drop look like credit for both. Only
// campaigns matched to a single streamer are used, where a missed advance can
// be attributed to that streamer.
func (c *Client) detectStalledDropCredit(streamers []*model.Streamer, now time.Time) {
	matched := make(map[string]int, len(streamers))
	for _, streamer := range streamers {
		streamer.Mu.RLock()
		if streamer.Settings != nil && streamer.Settings.ClaimDrops {
			for i := range streamer.Stream.Campaigns {
				matched[streamer.Stream.Campaigns[i].ID]++
			}
		}
		streamer.Mu.RUnlock()
	}

	for _, streamer := range streamers {
		streamer.Mu.Lock()
		if streamer.Settings == nil || !streamer.Settings.ClaimDrops {
			streamer.Mu.Unlock()
			continue
		}
		drop, idle := stalledDropCredit(streamer, matched, now)
		if drop != nil {
			streamer.Stream.StalledCooldownUntil = now.Add(constants.StalledCooldownDuration)
		}
		username := streamer.Username
		streamer.Mu.Unlock()

		if drop != nil {
			c.Log.Warn("Streamer drops stopped advancing, cooling down",
				"streamer", username,
				"drop", drop.Name,
				"progress", drop.ProgressBar(),
				"stalled_for", idle.Round(time.Minute))
		}
	}
}

// stalledDropCredit returns the drop whose progress has not moved for longer
// than DropCreditStallThreshold while the streamer was still being watched.
func stalledDropCredit(streamer *model.Streamer, matched map[string]int, now time.Time) (*model.Drop, time.Duration) {
	lastAttempt := streamer.Stream.LastMinuteAttemptedAt
	lastAccepted := streamer.Stream.LastMinuteCreditedAt
	if lastAttempt.IsZero() || lastAccepted.IsZero() {
		return nil, 0
	}
	// Without a recent attempt the channel was not being watched, so a frozen
	// drop says nothing about its credit.
	if now.Sub(lastAttempt) > constants.DropCreditStallThreshold ||
		now.Sub(lastAccepted) > constants.DropCreditStallThreshold {
		return nil, 0
	}
	if now.Before(streamer.Stream.StalledCooldownUntil) {
		return nil, 0
	}

	for i := range streamer.Stream.Campaigns {
		campaign := &streamer.Stream.Campaigns[i]
		if matched[campaign.ID] != 1 {
			continue
		}
		for _, drop := range campaign.Drops {
			if !drop.AwaitingWatchCredit(now) {
				continue
			}
			if idle := streamer.Stream.ObserveDropCredit(drop.ID, drop.CurrentMinutesWatched, now); idle > constants.DropCreditStallThreshold {
				return drop, idle
			}
		}
	}
	return nil, 0
}
