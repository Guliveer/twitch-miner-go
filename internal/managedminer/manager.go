// Package managedminer provides dynamic lifecycle management for Miner goroutines.
// It replaces the static []minerEntry slice in main with a map that supports
// adding, stopping, and restarting miners at runtime.
package managedminer

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/auth"
	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/miner"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/runtimecfg"
)

const (
	stopTimeout         = 15 * time.Second
	initialRestartDelay = 10 * time.Second
	maxRestartDelay     = 5 * time.Minute
)

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

	parentCtx context.Context
	rootLog   *logger.Logger
	twitchRT  *runtimecfg.Twitch

	// launchFn starts the miner goroutine for an entry. Swappable in tests.
	launchFn func(e *entry, ctx context.Context, log *logger.Logger)

	suppressLifecycleNotify bool
	skipUnauth              bool
}

// SetSuppressLifecycleNotify controls whether all miners started by this
// Manager will suppress MINER_STARTED / MINER_STOPPED / MINER_CRASHED
// notifications. Must be called before the first Start().
func (m *Manager) SetSuppressLifecycleNotify(suppress bool) {
	m.suppressLifecycleNotify = suppress
}

func (m *Manager) SetSkipUnauth(skip bool) {
	m.skipUnauth = skip
}

// NewManager creates a Manager. parentCtx is used as the base for all miner contexts.
func NewManager(parentCtx context.Context, rootLog *logger.Logger, twitchRT *runtimecfg.Twitch) *Manager {
	m := &Manager{
		entries:   make(map[string]*entry),
		parentCtx: parentCtx,
		rootLog:   rootLog,
		twitchRT:  twitchRT,
	}
	m.launchFn = func(e *entry, ctx context.Context, log *logger.Logger) {
		go func() {
			defer close(e.done)
			delay := initialRestartDelay
			for {
				err := e.miner.Run(ctx)
				if ctx.Err() != nil {
					return
				}
				if err == nil {
					return
				}
				if errors.Is(err, auth.ErrSkippedUnauth) {
					return
				}
				log.Error("Miner crashed, restarting", "account", e.cfg.Username, "error", err, "retry_in", delay)
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
				}
				delay *= 2
				if delay > maxRestartDelay {
					delay = maxRestartDelay
				}
				newMiner := miner.NewMiner(e.cfg, log, m.twitchRT)
				newMiner.SetSuppressLifecycleNotify(m.suppressLifecycleNotify)
				newMiner.SetSkipUnauth(m.skipUnauth)
				m.mu.Lock()
				e.miner = newMiner
				m.mu.Unlock()
			}
		}()
	}
	return m
}

// Start creates and launches a miner for the given account config.
// Returns immediately if a miner for that username is already running.
func (m *Manager) Start(cfg *config.AccountConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[cfg.Username]; exists {
		return
	}
	m.startLocked(cfg, "")
}

// RestartChanged stops the miner for cfg.Username (if running) and starts a
// new one, firing an ACCOUNT_CONFIG_RELOADED notification once it is up.
func (m *Manager) RestartChanged(cfg *config.AccountConfig) {
	m.Stop(cfg.Username)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startLocked(cfg, model.EventAccountConfigReloaded)
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
	m.startLocked(cfg, "")
}

// LiveCount returns how many managed miners have not permanently exited.
// A miner counts as live while it is starting up, authenticating, running, or
// backing off before a restart, so the value does not flap during startup the
// way Miner.IsRunning does. It drops to zero once every miner has given up,
// which is what distinguishes an idle container from a working one.
func (m *Manager) LiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var live int
	for _, e := range m.entries {
		select {
		case <-e.done:
		default:
			live++
		}
	}
	return live
}

// Entries returns a snapshot of all managed miners, including any that have
// already exited. Use LiveCount to ask whether work is actually happening.
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
func (m *Manager) startLocked(cfg *config.AccountConfig, oneTimeEvent model.Event) {
	ctx, cancel := context.WithCancel(m.parentCtx)
	accountLog := m.rootLog.WithAccount(cfg.Username)
	minerInstance := miner.NewMiner(cfg, accountLog, m.twitchRT)
	minerInstance.SetSuppressLifecycleNotify(m.suppressLifecycleNotify)
	minerInstance.SetSkipUnauth(m.skipUnauth)
	if oneTimeEvent != "" {
		minerInstance.SetOneTimeEvent(oneTimeEvent)
	}

	e := &entry{
		cfg:    cfg,
		miner:  minerInstance,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	m.entries[cfg.Username] = e
	m.launchFn(e, ctx, accountLog)
}
