package config

import (
	"testing"
)

func benchConfig() *AccountConfig {
	enabled := true
	percentage := 5
	gap := 20
	maxPoints := 50000
	delay := 6.0
	return &AccountConfig{
		Username:        "bench",
		Enabled:         &enabled,
		MaxWatchStreams: ptr(2),
		Priority:        []string{"STREAK", "DROPS", "ORDER"},
		Streamers: []StreamerConfig{
			{Username: "s1"},
			{Username: "s2", Settings: &StreamerSettingsConfig{
				MakePredictions: &enabled,
				ClaimDrops:      &enabled,
				Bet: &BetSettingsConfig{
					Strategy:      "SMART",
					Percentage:    &percentage,
					PercentageGap: &gap,
					MaxPoints:     &maxPoints,
					Delay:         &delay,
				},
			}},
		},
		Blacklist: []string{"bad1", "bad2"},
	}
}

func ptr[T any](v T) *T { return &v }

// BenchmarkValidate benchmarks config validation, which runs every time a
// config file changes (file watcher) or an account is created/updated via the
// REST API in DB mode.
func BenchmarkValidate(b *testing.B) {
	cfg := benchConfig()

	for b.Loop() {
		_ = Validate(cfg)
	}
}

// BenchmarkAccountConfigToJSON benchmarks the JSON serialization used when
// storing account configs in PostgreSQL (DB mode).
func BenchmarkAccountConfigToJSON(b *testing.B) {
	cfg := benchConfig()

	for b.Loop() {
		_, _ = AccountConfigToJSON(cfg)
	}
}

// BenchmarkAccountConfigFromJSON benchmarks the JSON deserialization used when
// loading account configs from PostgreSQL (DB mode), including defaults and
// env-var overrides.
func BenchmarkAccountConfigFromJSON(b *testing.B) {
	blob, _ := AccountConfigToJSON(benchConfig())

	for b.Loop() {
		_, _ = AccountConfigFromJSON("bench", blob)
	}
}
