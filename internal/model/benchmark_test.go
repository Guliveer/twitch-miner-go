package model

import (
	"encoding/json"
	"testing"
)

// BenchmarkParseMessage benchmarks parsing a typical PubSub points-earned message.
func BenchmarkParseMessage(b *testing.B) {
	raw := map[string]any{
		"type": "points-earned",
		"data": map[string]any{
			"timestamp": "2025-01-15T12:00:00Z",
			"balance": map[string]any{
				"channel_id": "chan123",
			},
			"point_gain": map[string]any{
				"total_points": float64(50),
				"reason_code":  "WATCH",
			},
		},
	}
	rawJSON, _ := json.Marshal(raw)
	topic := "community-points-user-v1.userABC"

	for b.Loop() {
		_, _ = ParseMessage(topic, rawJSON)
	}
}

// BenchmarkBetCalculate benchmarks the SMART strategy bet calculation.
func BenchmarkBetCalculate(b *testing.B) {
	s := DefaultBetSettings()
	s.Strategy = StrategySmart
	s.PercentageGap = 20
	s.Percentage = 5
	s.MaxPoints = 50000
	s.StealthMode = false

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 1200, TotalPoints: 500000, TopPoints: 5000},
		{ID: "o2", Title: "No", TotalUsers: 800, TotalPoints: 300000, TopPoints: 3000},
	}

	bet := NewBet(outcomes, s)
	bet.UpdateOutcomes(outcomes)

	for b.Loop() {
		bet.Calculate(100000)
	}
}

// BenchmarkFilterConditionSkip benchmarks evaluating a filter condition.
func BenchmarkFilterConditionSkip(b *testing.B) {
	s := DefaultBetSettings()
	s.Strategy = StrategyNumber1
	s.FilterCondition = &FilterCondition{
		By:    OutcomeKeyTotalUsers,
		Where: ConditionGT,
		Value: 100,
	}

	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 1200, TotalPoints: 500000, TopPoints: 5000},
		{ID: "o2", Title: "No", TotalUsers: 800, TotalPoints: 300000, TopPoints: 3000},
	}

	bet := NewBet(outcomes, s)
	bet.UpdateOutcomes(outcomes)
	bet.Calculate(100000)

	for b.Loop() {
		bet.Skip()
	}
}

// BenchmarkUpdateOutcomes benchmarks recomputing outcome statistics.
func BenchmarkUpdateOutcomes(b *testing.B) {
	outcomes := []Outcome{
		{ID: "o1", Title: "Yes", TotalUsers: 1200, TotalPoints: 500000, TopPoints: 5000},
		{ID: "o2", Title: "No", TotalUsers: 800, TotalPoints: 300000, TopPoints: 3000},
	}

	bet := NewBet(outcomes, DefaultBetSettings())

	for b.Loop() {
		bet.UpdateOutcomes(outcomes)
	}
}

// BenchmarkUpdateHistory benchmarks recording earned points for a reason code,
// which runs for every points-earned event while watching streams.
func BenchmarkUpdateHistory(b *testing.B) {
	s := NewStreamer("bench")

	for b.Loop() {
		s.UpdateHistory(string(MsgTypePointsEarned), 50, 1)
	}
}
