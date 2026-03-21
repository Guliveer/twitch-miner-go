package runtimecfg

import (
	"log/slog"
	"os"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/constants"
)

func TestLoadTwitchFromEnv_Defaults(t *testing.T) {
	for _, key := range []string{
		"TWITCH_CLIENT_ID_TV",
		"TWITCH_CLIENT_ID_BROWSER",
		"TWITCH_CLIENT_VERSION",
	} {
		t.Setenv(key, "")
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := LoadTwitchFromEnv(log)

	if cfg.ClientIDTV != constants.ClientID {
		t.Fatalf("expected default ClientIDTV %q, got %q", constants.ClientID, cfg.ClientIDTV)
	}
	if cfg.ClientIDBrowser != constants.ClientIDBrowser {
		t.Fatalf("expected default ClientIDBrowser %q, got %q", constants.ClientIDBrowser, cfg.ClientIDBrowser)
	}
	if cfg.ClientVersion != constants.ClientVersion {
		t.Fatalf("expected default ClientVersion %q, got %q", constants.ClientVersion, cfg.ClientVersion)
	}
}

func TestLoadTwitchFromEnv_EnvOverride(t *testing.T) {
	t.Setenv("TWITCH_CLIENT_ID_TV", "custom-tv")
	t.Setenv("TWITCH_CLIENT_ID_BROWSER", "custom-browser")
	t.Setenv("TWITCH_CLIENT_VERSION", "custom-version")

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := LoadTwitchFromEnv(log)

	if cfg.ClientIDTV != "custom-tv" {
		t.Fatalf("expected env override ClientIDTV %q, got %q", "custom-tv", cfg.ClientIDTV)
	}
	if cfg.ClientIDBrowser != "custom-browser" {
		t.Fatalf("expected env override ClientIDBrowser %q, got %q", "custom-browser", cfg.ClientIDBrowser)
	}
	if cfg.ClientVersion != "custom-version" {
		t.Fatalf("expected env override ClientVersion %q, got %q", "custom-version", cfg.ClientVersion)
	}
}

func TestClientIDsForGQL_Dedup(t *testing.T) {
	t.Parallel()

	cfg := &Twitch{
		ClientIDTV:      "same",
		ClientIDBrowser: "same",
		ClientIDMobile:  "mobile",
	}

	ids := cfg.ClientIDsForGQL()
	if len(ids) != 2 {
		t.Fatalf("expected 2 unique IDs, got %d: %v", len(ids), ids)
	}
}
