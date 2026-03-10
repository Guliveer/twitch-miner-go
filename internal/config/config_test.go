package config

import "testing"

func TestApplyDefaultsSetsMaxWatchStreams(t *testing.T) {
	cfg := &AccountConfig{}

	applyDefaults(cfg)

	if cfg.MaxWatchStreams != 2 {
		t.Fatalf("expected default max_watch_streams to be 2, got %d", cfg.MaxWatchStreams)
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
