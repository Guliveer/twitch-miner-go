package managedminer

import (
	"context"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m := NewManager(context.Background(), discardLogger(), nil)
	// replace the launch function so tests don't start real miner goroutines
	m.launchFn = func(e *entry, _ context.Context, _ *logger.Logger) {
		close(e.done)
	}
	return m
}

func testConfig(username string) *config.AccountConfig {
	return &config.AccountConfig{Username: username}
}

// ── Start ──────────────────────────────────────────────────────────────────────

func TestManager_StartAddsEntry(t *testing.T) {
	m := newTestManager(t)
	m.Start(testConfig("alice"))

	entries := m.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Cfg.Username != "alice" {
		t.Errorf("expected alice, got %s", entries[0].Cfg.Username)
	}
}

func TestManager_StartIdempotent(t *testing.T) {
	m := newTestManager(t)
	m.Start(testConfig("alice"))
	m.Start(testConfig("alice")) // second call must be a no-op

	if len(m.Entries()) != 1 {
		t.Errorf("expected 1 entry after duplicate Start, got %d", len(m.Entries()))
	}
}

// ── Stop ───────────────────────────────────────────────────────────────────────

func TestManager_StopRemovesEntry(t *testing.T) {
	m := newTestManager(t)
	m.Start(testConfig("alice"))
	m.Stop("alice")

	if len(m.Entries()) != 0 {
		t.Errorf("expected 0 entries after Stop, got %d", len(m.Entries()))
	}
}

func TestManager_StopUnknownIsNoop(t *testing.T) {
	m := newTestManager(t)
	m.Stop("nobody") // must not panic
}

// ── Restart ────────────────────────────────────────────────────────────────────

func TestManager_RestartReplacesEntry(t *testing.T) {
	m := newTestManager(t)
	m.Start(testConfig("alice"))

	updated := testConfig("alice")
	m.Restart(updated)

	entries := m.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after Restart, got %d", len(entries))
	}
	if entries[0].Cfg != updated {
		t.Error("expected Restart to replace config pointer")
	}
}

// ── StopAll ────────────────────────────────────────────────────────────────────

func TestManager_StopAllClearsEntries(t *testing.T) {
	m := newTestManager(t)
	m.Start(testConfig("alice"))
	m.Start(testConfig("bob"))
	m.StopAll()

	if len(m.Entries()) != 0 {
		t.Errorf("expected 0 entries after StopAll, got %d", len(m.Entries()))
	}
}

// ── Entries ────────────────────────────────────────────────────────────────────

func TestManager_EntriesReturnsSnapshot(t *testing.T) {
	m := newTestManager(t)
	m.Start(testConfig("alice"))
	m.Start(testConfig("bob"))

	got := m.Entries()
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

// ── LiveCount ─────────────────────────────────────────────────────────────────

// newLiveManager keeps the launch goroutine "running" by never closing e.done,
// which is what an authenticating or actively mining account looks like.
func newLiveManager(t *testing.T) *Manager {
	t.Helper()
	m := NewManager(context.Background(), discardLogger(), nil)
	m.launchFn = func(_ *entry, _ context.Context, _ *logger.Logger) {}
	return m
}

func TestManager_LiveCountZeroWithoutEntries(t *testing.T) {
	if got := newLiveManager(t).LiveCount(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestManager_LiveCountCountsStartingMiners(t *testing.T) {
	m := newLiveManager(t)
	m.Start(testConfig("alice"))
	m.Start(testConfig("bob"))

	// Neither miner is "running" yet — they are still authenticating — but the
	// container is working, so health must not report it idle.
	if got := m.LiveCount(); got != 2 {
		t.Fatalf("expected 2 live miners, got %d", got)
	}
}

func TestManager_LiveCountExcludesExitedMiners(t *testing.T) {
	// The default test launchFn closes e.done immediately, i.e. the miner gave
	// up (skipped for missing credentials, or Run returned nil).
	m := newTestManager(t)
	m.Start(testConfig("alice"))

	if got := len(m.Entries()); got != 1 {
		t.Fatalf("expected the entry to remain, got %d", got)
	}
	if got := m.LiveCount(); got != 0 {
		t.Fatalf("expected 0 live miners for an exited entry, got %d", got)
	}
}
