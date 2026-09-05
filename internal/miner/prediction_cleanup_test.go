package miner

import (
	"context"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// trackedPrediction registers an event under test and returns it.
func trackedPrediction(m *Miner, eventID string, createdAt time.Time) *model.EventPrediction {
	event := &model.EventPrediction{
		EventID:   eventID,
		Title:     "Who will win?",
		CreatedAt: createdAt,
		Status:    "ACTIVE",
	}
	m.eventsPredictions[eventID] = event
	return event
}

// predictionUpdate builds the eventDict half of a prediction-updated message.
func predictionUpdate(eventID, status string) map[string]any {
	return map[string]any{
		"id":     eventID,
		"status": status,
	}
}

func trackedCount(m *Miner) int {
	m.eventsPredictionsMu.RLock()
	defer m.eventsPredictionsMu.RUnlock()
	return len(m.eventsPredictions)
}

func TestCanceledPredictionWithoutStakeIsDropped(t *testing.T) {
	m, _ := newTestMiner(t)
	trackedPrediction(m, "evt-1", time.Now())

	m.handlePredictionUpdated(context.Background(), nil,
		predictionUpdate("evt-1", "CANCELED"), "evt-1", "CANCELED",
		&model.Message{Timestamp: time.Now()})

	if got := trackedCount(m); got != 0 {
		t.Fatalf("expected the unstaked CANCELED prediction to be dropped, %d still tracked", got)
	}
}

func TestResolvedPredictionWithStakeIsKept(t *testing.T) {
	m, _ := newTestMiner(t)
	event := trackedPrediction(m, "evt-2", time.Now())
	event.BetPlaced = true

	m.handlePredictionUpdated(context.Background(), nil,
		predictionUpdate("evt-2", "RESOLVED"), "evt-2", "RESOLVED",
		&model.Message{Timestamp: time.Now()})

	// The stake means a "prediction-result" is still coming; dropping the
	// event here would lose the points accounting.
	if got := trackedCount(m); got != 1 {
		t.Fatalf("expected the staked prediction to be kept, %d tracked", got)
	}
}

func TestActivePredictionIsKept(t *testing.T) {
	m, _ := newTestMiner(t)
	trackedPrediction(m, "evt-3", time.Now())

	m.handlePredictionUpdated(context.Background(), nil,
		predictionUpdate("evt-3", "ACTIVE"), "evt-3", "ACTIVE",
		&model.Message{Timestamp: time.Now()})

	if got := trackedCount(m); got != 1 {
		t.Fatalf("expected the active prediction to be kept, %d tracked", got)
	}
}

func TestForgetPredictionStopsPendingTimer(t *testing.T) {
	m, _ := newTestMiner(t)
	trackedPrediction(m, "evt-4", time.Now())

	fired := make(chan struct{}, 1)
	m.pendingTimers["evt-4"] = time.AfterFunc(50*time.Millisecond, func() {
		fired <- struct{}{}
	})

	m.forgetPrediction("evt-4")

	select {
	case <-fired:
		t.Fatal("pending placement timer fired after the prediction was dropped")
	case <-time.After(150 * time.Millisecond):
	}

	m.pendingTimersMu.Lock()
	remaining := len(m.pendingTimers)
	m.pendingTimersMu.Unlock()
	if remaining != 0 {
		t.Fatalf("expected the timer entry to be removed, %d remain", remaining)
	}
}

func TestSweepDiscardsPredictionsPastRetention(t *testing.T) {
	m, _ := newTestMiner(t)
	now := time.Now()

	// Staked but never resolved — the case handlePredictionUpdated cannot fix.
	old := trackedPrediction(m, "evt-old", now.Add(-constants.PredictionRetention-time.Hour))
	old.BetPlaced = true

	trackedPrediction(m, "evt-fresh", now.Add(-time.Minute))

	inFlight := trackedPrediction(m, "evt-inflight", now.Add(-constants.PredictionRetention-time.Hour))
	inFlight.PlacementInFlight = true

	if swept := m.sweepStalePredictions(now); swept != 1 {
		t.Fatalf("expected exactly one sweep, got %d", swept)
	}

	if _, ok := m.eventsPredictions["evt-old"]; ok {
		t.Error("stale prediction survived the sweep")
	}
	if _, ok := m.eventsPredictions["evt-fresh"]; !ok {
		t.Error("recent prediction was swept")
	}
	if _, ok := m.eventsPredictions["evt-inflight"]; !ok {
		t.Error("prediction with a placement in flight was swept")
	}
}
