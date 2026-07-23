package config

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
)

func TestApplyDefaultsSetsMaxWatchStreams(t *testing.T) {
	cfg := &AccountConfig{}

	applyDefaults(cfg)

	if cfg.MaxWatchStreams == nil || *cfg.MaxWatchStreams != constants.MaxWatchStreams {
		t.Fatalf("expected default max_watch_streams to be %d, got %v", constants.MaxWatchStreams, cfg.MaxWatchStreams)
	}
}

func TestValidateRejectsInvalidMaxWatchStreams(t *testing.T) {
	neg := -1
	cfg := &AccountConfig{
		Username:        "tester",
		MaxWatchStreams: &neg,
		Streamers: []StreamerConfig{
			{Username: "example"},
		},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error for max_watch_streams < 0")
	}
}

func TestValidateAllowsZeroMaxWatchStreams(t *testing.T) {
	zero := 0
	cfg := &AccountConfig{
		Username:        "tester",
		MaxWatchStreams: &zero,
		Streamers: []StreamerConfig{
			{Username: "example"},
		},
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("expected no error for max_watch_streams = 0, got %v", err)
	}
}
