package config

import (
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
)

func TestApplyDefaultsSetsMaxWatchStreams(t *testing.T) {
	cfg := &AccountConfig{}

	applyDefaults(cfg)

	if cfg.MaxWatchStreams != constants.MaxWatchStreams {
		t.Fatalf("expected default max_watch_streams to be %d, got %d", constants.MaxWatchStreams, cfg.MaxWatchStreams)
	}
}

func TestValidateRejectsInvalidMaxWatchStreams(t *testing.T) {
	cfg := &AccountConfig{
		Username:        "tester",
		MaxWatchStreams: 0,
		Streamers: []StreamerConfig{
			{Username: "example"},
		},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error for max_watch_streams < 1")
	}
}
