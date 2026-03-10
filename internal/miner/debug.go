package miner

import (
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/pubsub"
	"github.com/Guliveer/twitch-miner-go/internal/twitch"
)

// DebugSnapshot is a serializable runtime snapshot for one miner instance.
type DebugSnapshot struct {
	Account              string                      `json:"account"`
	IsRunning            bool                        `json:"is_running"`
	Watching             []DebugWatchingEntry        `json:"watching"`
	ActivePredictions    []DebugPredictionEntry      `json:"active_predictions"`
	PubSubConnections    []pubsub.ConnectionSnapshot `json:"pubsub_connections,omitempty"`
	PendingTimerCount    int                         `json:"pending_timer_count"`
	TrackedStreamerCount int                         `json:"tracked_streamer_count"`
}

// DebugWatchingEntry describes one streamer selected for minute-watched events.
type DebugWatchingEntry struct {
	Username           string  `json:"username"`
	DisplayName        string  `json:"display_name,omitempty"`
	ChannelPoints      int     `json:"channel_points"`
	WatchStreakMissing bool    `json:"watch_streak_missing"`
	MinuteWatched      float64 `json:"minute_watched"`
	DropsEnabled       bool    `json:"drops_enabled"`
}

// DebugPredictionEntry describes one active prediction and its scheduling state.
type DebugPredictionEntry struct {
	EventID           string    `json:"event_id"`
	Streamer          string    `json:"streamer"`
	Title             string    `json:"title"`
	Status            string    `json:"status"`
	ScheduledFor      time.Time `json:"scheduled_for,omitempty"`
	BetPlaced         bool      `json:"bet_placed"`
	BetConfirmed      bool      `json:"bet_confirmed"`
	PlacementInFlight bool      `json:"placement_in_flight"`
	PlacementAttempts int       `json:"placement_attempts"`
	LastAttemptAt     time.Time `json:"last_attempt_at,omitempty"`
	LastFailedReason  string    `json:"last_failed_reason,omitempty"`
	DecisionChoice    int       `json:"decision_choice"`
	DecisionAmount    int       `json:"decision_amount"`
	DecisionOutcomeID string    `json:"decision_outcome_id,omitempty"`
}

// DebugSnapshot returns a read-only runtime view of the miner internals.
func (m *Miner) DebugSnapshot() DebugSnapshot {
	streamers := m.getStreamers()
	watching := twitch.SelectStreamersToWatch(streamers, m.priorities, m.cfg.MaxWatchStreams)

	watchingEntries := make([]DebugWatchingEntry, 0, len(watching))
	for _, streamer := range watching {
		streamer.Mu.RLock()
		entry := DebugWatchingEntry{
			Username:           streamer.Username,
			DisplayName:        streamer.DisplayName,
			ChannelPoints:      streamer.ChannelPoints,
			WatchStreakMissing: streamer.Stream != nil && streamer.Stream.WatchStreakMissing,
			DropsEnabled:       streamer.Settings != nil && streamer.Settings.ClaimDrops,
		}
		if streamer.Stream != nil {
			entry.MinuteWatched = streamer.Stream.MinuteWatched
		}
		streamer.Mu.RUnlock()
		watchingEntries = append(watchingEntries, entry)
	}

	m.eventsPredictionsMu.RLock()
	predictions := make([]DebugPredictionEntry, 0, len(m.eventsPredictions))
	for _, prediction := range m.eventsPredictions {
		prediction.Mu.Lock()
		streamerName := ""
		if prediction.Streamer != nil {
			streamerName = prediction.Streamer.Username
		}
		predictions = append(predictions, DebugPredictionEntry{
			EventID:           prediction.EventID,
			Streamer:          streamerName,
			Title:             prediction.Title,
			Status:            prediction.Status,
			ScheduledFor:      prediction.ScheduledFor,
			BetPlaced:         prediction.BetPlaced,
			BetConfirmed:      prediction.BetConfirmed,
			PlacementInFlight: prediction.PlacementInFlight,
			PlacementAttempts: prediction.PlacementAttempts,
			LastAttemptAt:     prediction.LastAttemptAt,
			LastFailedReason:  prediction.LastFailedReason,
			DecisionChoice:    prediction.Bet.Decision.Choice,
			DecisionAmount:    prediction.Bet.Decision.Amount,
			DecisionOutcomeID: prediction.Bet.Decision.OutcomeID,
		})
		prediction.Mu.Unlock()
	}
	m.eventsPredictionsMu.RUnlock()

	m.pendingTimersMu.Lock()
	pendingTimerCount := len(m.pendingTimers)
	m.pendingTimersMu.Unlock()

	var pubsubConnections []pubsub.ConnectionSnapshot
	if m.pubsub != nil {
		pubsubConnections = m.pubsub.Snapshot()
	}

	return DebugSnapshot{
		Account:              m.cfg.Username,
		IsRunning:            m.IsRunning(),
		Watching:             watchingEntries,
		ActivePredictions:    predictions,
		PubSubConnections:    pubsubConnections,
		PendingTimerCount:    pendingTimerCount,
		TrackedStreamerCount: len(streamers),
	}
}
