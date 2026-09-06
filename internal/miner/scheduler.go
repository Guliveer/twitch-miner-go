package miner

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/twitch"
)

// watchOptions builds the watch-set configuration from the account config.
// Pointer fields are filled in by applyDefaults, but a config assembled in
// code (tests, the DB path) may leave them nil, so each falls back.
func (m *Miner) watchOptions() twitch.WatchOptions {
	opts := twitch.WatchOptions{
		Priorities:    m.priorities,
		MaxWatch:      constants.MaxWatchStreams,
		StreakWatch:   constants.StreakWatchStreams,
		StreakMinutes: constants.WatchStreakMinutes,
		Preferred:     m.cfg.PreferredStreamers,
	}

	if m.cfg.MaxWatchStreams != nil {
		opts.MaxWatch = *m.cfg.MaxWatchStreams
	}
	if m.cfg.StreakWatchStreams != nil {
		opts.StreakWatch = *m.cfg.StreakWatchStreams
	}
	if m.cfg.WatchStreakMinutes != nil {
		opts.StreakMinutes = *m.cfg.WatchStreakMinutes
	}

	return opts
}

func (m *Miner) runMinuteWatcher(ctx context.Context) error {
	ticker := time.NewTicker(constants.DefaultMinuteWatchedInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			streamers := m.getStreamers()
			set := twitch.SelectWatchSet(streamers, m.watchOptions())
			toWatch := set.Streamers

			m.logWatchModeChange(set)
			m.logWatchingChanges(toWatch)

			if len(toWatch) > 0 {
				if err := m.twitch.SendMinuteWatchedEvents(ctx, toWatch); err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					m.log.Debug("Minute watched error", "error", err)
				}
			}
		}
	}
}

// logWatchModeChange reports when the watch set switches between chasing
// pending streaks on a narrow set and running at full width, so the shrunken
// channel list is never a mystery in the log.
func (m *Miner) logWatchModeChange(set twitch.WatchSet) {
	// Nobody online yet says nothing about the mode. Reporting it here would
	// announce a spurious "widened to 0" on every gap in coverage, and would
	// consume the transition that the first real selection should announce.
	if len(set.Streamers) == 0 {
		return
	}

	m.lastWatchingMu.Lock()
	changed := m.lastStreakHarvest == nil || *m.lastStreakHarvest != set.StreakHarvest
	harvest := set.StreakHarvest
	m.lastStreakHarvest = &harvest
	m.lastWatchingMu.Unlock()

	if !changed {
		return
	}

	if set.StreakHarvest {
		m.log.Info("🔥 Chasing watch streaks — narrowing watch set",
			"streams", set.Width,
			"minutes_per_channel", m.watchOptions().StreakMinutes)
		return
	}
	m.log.Info("✅ All watch streaks settled — widening watch set", "streams", set.Width)
}

// logWatchingChanges compares the current set of watched streamers with the
// previous set and logs which streamers started or stopped being watched.
func (m *Miner) logWatchingChanges(toWatch []*model.Streamer) {
	currentSet := make(map[string]bool, len(toWatch))
	for _, s := range toWatch {
		s.Mu.RLock()
		currentSet[s.Username] = true
		s.Mu.RUnlock()
	}

	m.lastWatchingMu.Lock()
	defer m.lastWatchingMu.Unlock()

	for username := range currentSet {
		if !m.lastWatching[username] {
			m.log.Info("👀 Watching", "streamer", username)
		}
	}

	for username := range m.lastWatching {
		if !currentSet[username] {
			m.log.Info("💤 Stopped watching", "streamer", username)
		}
	}

	m.lastWatching = currentSet
}

func (m *Miner) runCampaignSync(ctx context.Context) error {
	ticker := time.NewTicker(constants.DefaultCampaignSyncInterval)
	defer ticker.Stop()

	// Keep the loop alive even when no streamer can claim drops yet: streamers
	// discovered later (e.g. by the category watcher) need campaign sync on the
	// next tick, otherwise drops-only watching never starts.
	sync := func() {
		streamers := m.getStreamers()
		hasDrops := false
		for _, s := range streamers {
			s.Mu.RLock()
			hasDrops = s.Settings != nil && s.Settings.ClaimDrops
			s.Mu.RUnlock()
			if hasDrops {
				break
			}
		}
		if !hasDrops {
			return
		}
		if err := m.twitch.SyncCampaigns(ctx, streamers); err != nil {
			if ctx.Err() != nil {
				return
			}
			m.log.Warn("Campaign sync failed", "error", err)
		}
		// Hint GC to reclaim transient campaign sync allocations
		runtime.GC()
	}

	sync()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sync()
		}
	}
}

func (m *Miner) runContextRefresh(ctx context.Context) error {
	ticker := time.NewTicker(constants.DefaultCampaignSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			streamers := m.getStreamers()
			for _, s := range streamers {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				s.Mu.RLock()
				isOnline := s.IsOnline
				s.Mu.RUnlock()

				if isOnline {
					if err := m.twitch.LoadChannelPointsContext(ctx, s); err != nil {
						m.log.Debug("Context refresh failed",
							"streamer", s.Username, "error", err)
					}
				}
			}
		}
	}
}

// runPredictionSweeper discards tracked predictions whose outcome never
// arrived. handlePredictionResult is the normal exit, but a "prediction-result"
// that Twitch never sends — a dropped PubSub message, a bet placed just as the
// event was cancelled — would otherwise pin the entry for the process lifetime.
func (m *Miner) runPredictionSweeper(ctx context.Context) error {
	ticker := time.NewTicker(constants.PredictionSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			m.sweepStalePredictions(time.Now())
		}
	}
}

// sweepStalePredictions removes predictions older than the retention window.
// It is separate from runPredictionSweeper so tests can drive it directly.
func (m *Miner) sweepStalePredictions(now time.Time) int {
	var stale []string

	m.eventsPredictionsMu.RLock()
	for id, event := range m.eventsPredictions {
		event.Mu.Lock()
		age := now.Sub(event.CreatedAt)
		inFlight := event.PlacementInFlight
		event.Mu.Unlock()

		if !inFlight && age > constants.PredictionRetention {
			stale = append(stale, id)
		}
	}
	m.eventsPredictionsMu.RUnlock()

	for _, id := range stale {
		m.log.Debug("Sweeping prediction whose result never arrived", "event_id", id)
		m.forgetPrediction(id)
	}

	return len(stale)
}

func (m *Miner) runMonitorLoop(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(20+rand.IntN(40)) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Reset(time.Duration(20+rand.IntN(40)) * time.Second)

			streamers := m.getStreamers()
			for _, s := range streamers {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if err := m.twitch.CheckStreamerOnline(ctx, s); err != nil {
					m.log.Debug("Online check failed",
						"streamer", s.Username, "error", err)
				}
			}
		}
	}
}
