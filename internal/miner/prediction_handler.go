package miner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/jsonutil"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func (m *Miner) handlePredictionsChannel(ctx context.Context, msg *model.Message, streamer *model.Streamer) {
	if streamer == nil || msg.Data == nil {
		return
	}

	eventDict, _ := msg.Data["event"].(map[string]any)
	if eventDict == nil {
		return
	}

	eventID, _ := eventDict["id"].(string)
	eventStatus, _ := eventDict["status"].(string)

	switch msg.Type {
	case model.MsgTypePredictionEvent:
		m.handlePredictionCreated(ctx, streamer, eventDict, eventID, eventStatus, msg)
	case model.MsgTypePredictionUpdate:
		m.handlePredictionUpdated(ctx, streamer, eventDict, eventID, eventStatus, msg)
	case model.MsgTypePredictionLocked:
		m.handlePredictionLocked(eventID, eventStatus)
	}
}

func (m *Miner) handlePredictionCreated(
	ctx context.Context,
	streamer *model.Streamer,
	eventDict map[string]any,
	eventID, eventStatus string,
	msg *model.Message,
) {
	m.eventsPredictionsMu.RLock()
	_, exists := m.eventsPredictions[eventID]
	m.eventsPredictionsMu.RUnlock()
	if exists {
		return
	}

	if eventStatus != "ACTIVE" {
		return
	}

	streamer.Mu.RLock()
	isOnline := streamer.IsOnline
	makePredictions := streamer.Settings != nil && streamer.Settings.MakePredictions
	balance := streamer.ChannelPoints
	var betSettings *model.BetSettings
	if streamer.Settings != nil {
		betSettings = streamer.Settings.Bet
	}
	username := streamer.Username
	category := streamer.ResolveCategory()
	streamer.Mu.RUnlock()

	if !makePredictions || !isOnline || betSettings == nil {
		return
	}

	predictionWindowSeconds := jsonutil.FloatFromAny(eventDict["prediction_window_seconds"])

	actualWindow := model.GetPredictionWindow(betSettings, predictionWindowSeconds)

	outcomes := parseOutcomes(eventDict["outcomes"])
	if len(outcomes) == 0 {
		m.log.Warn("Skipping prediction without outcomes",
			"streamer", username,
			"event_id", eventID)
		return
	}

	createdAtStr, _ := eventDict["created_at"].(string)
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		createdAt = time.Now()
	}

	event := model.NewEventPrediction(
		streamer,
		eventID,
		jsonutil.StringFromAny(eventDict["title"]),
		createdAt,
		actualWindow,
		eventStatus,
		outcomes,
	)

	secondsUntilClose := event.ClosingBetAfter(msg.Timestamp)
	if secondsUntilClose <= 0 {
		m.log.Debug("Prediction window already closed",
			"streamer", username, "event_id", eventID)
		return
	}

	// Store event before balance check so dedup works even when bet is rejected.
	// Without this, repeated prediction updates re-trigger handlePredictionCreated
	// and spam "Insufficient points" every second.
	m.eventsPredictionsMu.Lock()
	m.eventsPredictions[eventID] = event
	m.eventsPredictionsMu.Unlock()

	if betSettings.MinimumPoints > 0 && balance < betSettings.MinimumPoints {
		m.log.Event(ctx, model.EventBetFilters,
			"Insufficient points for bet",
			"streamer", username,
			"category", category,
			"balance", balance,
			"minimum", betSettings.MinimumPoints)
		return
	}

	m.log.Event(ctx, model.EventBetStart,
		fmt.Sprintf("Placing bet in %.0fs", secondsUntilClose),
		"streamer", username,
		"category", category,
		"title", event.Title)

	m.schedulePredictionAttempt(ctx, streamer, event, secondsUntilClose)
}

func (m *Miner) handlePredictionUpdated(
	ctx context.Context,
	streamer *model.Streamer,
	eventDict map[string]any,
	eventID, eventStatus string,
	msg *model.Message,
) {
	m.eventsPredictionsMu.RLock()
	event, ok := m.eventsPredictions[eventID]
	m.eventsPredictionsMu.RUnlock()
	if !ok {
		if eventStatus == "ACTIVE" {
			m.handlePredictionCreated(ctx, streamer, eventDict, eventID, eventStatus, msg)
		}
		return
	}

	event.Mu.Lock()
	defer event.Mu.Unlock()

	event.Status = eventStatus

	if !event.BetPlaced && !event.BetConfirmed {
		outcomes := parseOutcomes(eventDict["outcomes"])
		if len(outcomes) > 0 {
			event.Bet.UpdateOutcomes(outcomes)
		}
	}

	m.pendingTimersMu.Lock()
	_, hasTimer := m.pendingTimers[eventID]
	m.pendingTimersMu.Unlock()
	if !hasTimer && eventStatus == "ACTIVE" && !event.BetPlaced && !event.BetConfirmed {
		delay := event.ClosingBetAfter(time.Now())
		if delay < 0 {
			delay = 0
		}
		go m.schedulePredictionAttempt(ctx, streamer, event, delay)
	}
}

func (m *Miner) handlePredictionLocked(eventID, eventStatus string) {
	m.eventsPredictionsMu.RLock()
	event, ok := m.eventsPredictions[eventID]
	m.eventsPredictionsMu.RUnlock()
	if !ok {
		return
	}

	event.Mu.Lock()
	event.Status = eventStatus
	event.Mu.Unlock()

	m.log.Debug("Prediction locked", "event_id", eventID)
}

func (m *Miner) handlePredictionsUser(ctx context.Context, msg *model.Message, streamer *model.Streamer) {
	if msg.Data == nil {
		return
	}

	prediction, _ := msg.Data["prediction"].(map[string]any)
	if prediction == nil {
		return
	}

	eventID, _ := prediction["event_id"].(string)

	m.eventsPredictionsMu.RLock()
	event, ok := m.eventsPredictions[eventID]
	m.eventsPredictionsMu.RUnlock()
	if !ok {
		return
	}

	switch msg.Type {
	case "prediction-result":
		m.handlePredictionResult(ctx, event, prediction, streamer)
	case "prediction-made":
		m.handlePredictionMade(event)
	}
}

func (m *Miner) handlePredictionResult(ctx context.Context, event *model.EventPrediction, prediction map[string]any, streamer *model.Streamer) {
	event.Mu.Lock()
	betConfirmed := event.BetConfirmed
	betPlaced := event.BetPlaced
	event.Mu.Unlock()

	if !betConfirmed && !betPlaced {
		return
	}

	result, _ := prediction["result"].(map[string]any)
	if result == nil {
		return
	}

	resultType, _ := result["type"].(string)
	pointsWon := jsonutil.IntFromAny(result["points_won"])

	event.Mu.Lock()
	points := event.ParseResult(resultType, pointsWon)

	var notifyEvent model.Event
	switch resultType {
	case "WIN":
		notifyEvent = model.EventBetWin
	case "LOSE":
		notifyEvent = model.EventBetLose
	case "REFUND":
		notifyEvent = model.EventBetRefund
	default:
		notifyEvent = model.EventBetGeneral
	}

	choiceStr := "unknown"
	if event.Bet.Decision.Choice >= 0 && event.Bet.Decision.Choice < len(event.Bet.Outcomes) {
		chosen := event.Bet.Outcomes[event.Bet.Decision.Choice]
		choiceStr = fmt.Sprintf("%s (%s)", chosen.Title, chosen.Color)
	}
	eventTitle := event.Title
	resultString := event.Result.ResultString
	event.Mu.Unlock()

	m.pendingTimersMu.Lock()
	if t, ok := m.pendingTimers[event.EventID]; ok {
		t.Stop()
		delete(m.pendingTimers, event.EventID)
	}
	m.pendingTimersMu.Unlock()

	streamerName := ""
	streamerCategory := ""
	if streamer != nil {
		streamer.Mu.RLock()
		streamerName = streamer.Username
		streamerCategory = streamer.ResolveCategory()
		streamer.Mu.RUnlock()
	}

	m.log.Event(ctx, notifyEvent,
		"Prediction result",
		"streamer", streamerName,
		"category", streamerCategory,
		"title", eventTitle,
		"choice", choiceStr,
		"result", resultString)

	if streamer != nil {
		streamer.Mu.Lock()
		streamer.UpdateHistory("PREDICTION", points["gained"], 1)

		switch resultType {
		case "REFUND":
			streamer.UpdateHistory("REFUND", -points["placed"], -1)
		case "WIN":
			streamer.UpdateHistory("PREDICTION", -points["won"], -1)
		}
		streamer.Mu.Unlock()
	}

	// Clean up resolved prediction to prevent unbounded map growth (OOM fix).
	m.eventsPredictionsMu.Lock()
	delete(m.eventsPredictions, event.EventID)
	m.eventsPredictionsMu.Unlock()
}

func (m *Miner) handlePredictionMade(event *model.EventPrediction) {
	event.Mu.Lock()
	event.BetConfirmed = true
	event.BetPlaced = true
	event.PlacementInFlight = false
	event.LastFailedReason = ""
	event.Mu.Unlock()
	m.log.Debug("Prediction confirmed", "event_id", event.EventID)
}

func (m *Miner) schedulePredictionAttempt(ctx context.Context, streamer *model.Streamer, event *model.EventPrediction, delaySeconds float64) {
	if event == nil {
		return
	}

	if delaySeconds < 0 {
		delaySeconds = 0
	}

	event.Mu.Lock()
	if event.Status != "ACTIVE" || event.BetConfirmed {
		event.Mu.Unlock()
		return
	}
	event.ScheduledFor = time.Now().Add(time.Duration(delaySeconds * float64(time.Second)))
	eventID := event.EventID
	event.Mu.Unlock()

	timer := time.AfterFunc(time.Duration(delaySeconds*float64(time.Second)), func() {
		m.executePredictionAttempt(ctx, streamer, eventID)
	})

	m.pendingTimersMu.Lock()
	if old, ok := m.pendingTimers[eventID]; ok {
		old.Stop()
	}
	m.pendingTimers[eventID] = timer
	m.pendingTimersMu.Unlock()
}

func (m *Miner) executePredictionAttempt(ctx context.Context, streamer *model.Streamer, eventID string) {
	m.pendingTimersMu.Lock()
	delete(m.pendingTimers, eventID)
	m.pendingTimersMu.Unlock()

	m.eventsPredictionsMu.RLock()
	event, ok := m.eventsPredictions[eventID]
	m.eventsPredictionsMu.RUnlock()
	if !ok {
		return
	}

	event.Mu.Lock()
	if event.PlacementInFlight || event.BetConfirmed {
		event.Mu.Unlock()
		return
	}
	event.PlacementInFlight = true
	event.PlacementAttempts++
	event.LastAttemptAt = time.Now()
	event.Mu.Unlock()

	err := m.twitch.MakePrediction(ctx, streamer, event)

	event.Mu.Lock()
	event.PlacementInFlight = false
	if err != nil {
		event.LastFailedReason = err.Error()
	}
	remaining := event.ClosingBetAfter(time.Now())
	attempts := event.PlacementAttempts
	event.Mu.Unlock()

	if err == nil {
		return
	}

	streamerName := ""
	if streamer != nil {
		streamer.Mu.RLock()
		streamerName = streamer.Username
		streamer.Mu.RUnlock()
	}

	m.log.Warn("Failed to place prediction",
		"streamer", streamerName,
		"event_id", eventID,
		"attempts", attempts,
		"remaining_seconds", fmt.Sprintf("%.2f", remaining),
		"error", err)

	if attempts < 3 && remaining > 2 && isTransientPredictionError(err) && ctx.Err() == nil {
		retryDelay := 2.0
		if remaining < retryDelay {
			retryDelay = remaining
		}
		m.log.Info("Retrying prediction placement",
			"streamer", streamerName,
			"event_id", eventID,
			"retry_in_seconds", fmt.Sprintf("%.2f", retryDelay))
		m.schedulePredictionAttempt(ctx, streamer, event, retryDelay)
	}
}

func isTransientPredictionError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, gql.ErrCircuitOpen) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "eof") ||
		strings.Contains(errMsg, "status 429") ||
		strings.Contains(errMsg, "status 500") ||
		strings.Contains(errMsg, "status 502") ||
		strings.Contains(errMsg, "status 503") ||
		strings.Contains(errMsg, "status 504")
}
