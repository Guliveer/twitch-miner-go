package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestLoadAllAccountConfigsReportsEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadAllAccountConfigs(dir)
	if !errors.Is(err, ErrNoUsableAccounts) {
		t.Fatalf("expected ErrNoUsableAccounts, got %v", err)
	}
}

func TestLoadAllAccountConfigsReportsOwnerAccountsSkipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ownerAccounts[0]+".yaml")
	if err := os.WriteFile(path, []byte("enabled: true\n"), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	_, err := LoadAllAccountConfigs(dir)
	if !errors.Is(err, ErrNoUsableAccounts) {
		t.Fatalf("expected ErrNoUsableAccounts, got %v", err)
	}
	// The operator must be able to tell this apart from an empty directory.
	if !strings.Contains(err.Error(), "RUN_OWNER_ACCOUNTS") {
		t.Fatalf("expected the error to name the opt-in variable, got %q", err)
	}

	t.Setenv("RUN_OWNER_ACCOUNTS", "true")
	cfgs, err := LoadAllAccountConfigs(dir)
	if err != nil {
		t.Fatalf("expected owner accounts to load with the opt-in set, got %v", err)
	}
	if len(cfgs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(cfgs))
	}
}
