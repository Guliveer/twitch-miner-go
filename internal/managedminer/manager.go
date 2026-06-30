// Package managedminer provides dynamic lifecycle management for Miner goroutines.
// It replaces the static []minerEntry slice in main with a map that supports
// adding, stopping, and restarting miners at runtime.
package managedminer

import (
	"context"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/miner"
	"github.com/Guliveer/twitch-miner-go/internal/runtimecfg"
)

const stopTimeout = 15 * time.Second

type entry struct {
	cfg    *config.AccountConfig
	miner  *miner.Miner
	cancel context.CancelFunc
	done   chan struct{}
}

// Entry is an exported snapshot of a running miner, used by the analytics server.
type Entry struct {
	Cfg   *config.AccountConfig
	Miner *miner.Miner
}

// Manager owns the full lifecycle of all miner goroutines.
type Manager struct {
	mu      sync.RWMutex
	entries map[string]*entry

	parentCtx   context.Context
	rootLog     *logger.Logger
	twitchRT    *runtimecfg.Twitch
}

// NewManager creates a Manager. parentCtx is used as the base for all miner contexts.
func NewManager(parentCtx context.Context, rootLog *logger.Logger, twitchRT *runtimecfg.Twitch) *Manager {
	return &Manager{
		entries:  make(map[string]*entry),
		parentCtx: parentCtx,
		rootLog:   rootLog,
		twitchRT:  twitchRT,
	}
}

// Start creates and launches a miner for the given account config.
// Returns immediately if a miner for that username is already running.
func (m *Manager) Start(cfg *config.AccountConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[cfg.Username]; exists {
		return
	}
	m.startLocked(cfg)
}

// Stop gracefully shuts down the miner for the given username and waits for it
// to exit (up to stopTimeout). No-op if no miner exists for that username.
func (m *Manager) Stop(username string) {
	m.mu.Lock()
	e, ok := m.entries[username]
	if !ok {
		m.mu.Unlock()
		return
	}
	delete(m.entries, username)
	m.mu.Unlock()

	e.cancel()
	select {
	case <-e.done:
	case <-time.After(stopTimeout):
		m.rootLog.Warn("Miner stop timed out", "account", username)
	}
}

// Restart stops the miner for cfg.Username (if running) then starts a new one.
func (m *Manager) Restart(cfg *config.AccountConfig) {
	m.Stop(cfg.Username)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startLocked(cfg)
}

// Entries returns a snapshot of all currently running miners.
func (m *Manager) Entries() []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Entry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, Entry{Cfg: e.cfg, Miner: e.miner})
	}
	return out
}

// StopAll stops all running miners and waits for them to exit.
func (m *Manager) StopAll() {
	m.mu.Lock()
	usernames := make([]string, 0, len(m.entries))
	for u := range m.entries {
		usernames = append(usernames, u)
	}
	m.mu.Unlock()

	var wg sync.WaitGroup
	for _, u := range usernames {
		wg.Add(1)
		go func(username string) {
			defer wg.Done()
			m.Stop(username)
		}(u)
	}
	wg.Wait()
}

// startLocked creates and launches a miner. Must be called with m.mu held.
func (m *Manager) startLocked(cfg *config.AccountConfig) {
	ctx, cancel := context.WithCancel(m.parentCtx)
	accountLog := m.rootLog.WithAccount(cfg.Username)
	minerInstance := miner.NewMiner(cfg, accountLog, m.twitchRT)

	e := &entry{
		cfg:    cfg,
		miner:  minerInstance,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	m.entries[cfg.Username] = e

	go func() {
		defer close(e.done)
		if err := minerInstance.Run(ctx); err != nil && ctx.Err() == nil {
			accountLog.Error("Miner failed", "account", cfg.Username, "error", err)
		}
	}()
}
