package managedminer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

// FileWatcher polls a config directory for YAML file changes and drives a
// Manager accordingly. It mirrors the Poller API but sources configs from the
// filesystem instead of a database.
type FileWatcher struct {
	dir      string
	manager  fileManager
	interval time.Duration
	log      *logger.Logger

	// watermark: username → last seen mtime Unix seconds
	watermark map[string]int64

	// warnedNoAccounts suppresses the idle warning until accounts load again.
	warnedNoAccounts bool

	initialSyncDone chan struct{}
	// InitialSyncDone is closed after the first sync() completes.
	InitialSyncDone <-chan struct{}
}

// fileManager is the subset of Manager that FileWatcher needs.
type fileManager interface {
	Start(cfg *config.AccountConfig)
	RestartChanged(cfg *config.AccountConfig)
	Stop(username string)
}

// NewFileWatcher creates a FileWatcher that polls dir every interval.
func NewFileWatcher(dir string, mgr fileManager, interval time.Duration, log *logger.Logger) *FileWatcher {
	ch := make(chan struct{})
	return &FileWatcher{
		dir:             dir,
		manager:         mgr,
		interval:        interval,
		log:             log,
		watermark:       make(map[string]int64),
		initialSyncDone: ch,
		InitialSyncDone: ch,
	}
}

// Run starts the polling loop. It calls sync() immediately, closes the
// InitialSyncDone channel to signal the first load is complete, then
// continues polling on the configured interval. Blocks until ctx is cancelled.
func (w *FileWatcher) Run(ctx context.Context) {
	w.sync()
	close(w.initialSyncDone)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.sync()
		}
	}
}

func (w *FileWatcher) sync() {
	cfgs, err := config.LoadAllAccountConfigs(w.dir)
	if err != nil {
		// Only a directory that yields no runnable account stops the miners.
		// Any other error (parse failure, permission issue, transient I/O) is
		// logged and the last-known-good set of miners is preserved.
		if !errors.Is(err, config.ErrNoUsableAccounts) {
			w.log.Error("File watcher: failed to load configs, keeping current miners", "dir", w.dir, "error", err)
			return
		}
		// Warn once per transition, not once per poll: an operator who mounted
		// the wrong directory otherwise gets no signal at all, while a correct
		// idle state must not spam the log every interval.
		if !w.warnedNoAccounts {
			w.log.Warn("No account configs to run — no miners started", "dir", w.dir, "reason", err.Error())
			w.warnedNoAccounts = true
		}
		cfgs = nil
	} else {
		w.warnedNoAccounts = false
	}

	seen := make(map[string]struct{}, len(cfgs))

	for _, cfg := range cfgs {
		seen[cfg.Username] = struct{}{}

		mtime, err := w.fileMtime(cfg.Username)
		if err != nil {
			w.log.Error("File watcher: stat failed", "account", cfg.Username, "error", err)
			continue
		}

		if err := config.Validate(cfg); err != nil {
			w.log.Error("File watcher: invalid config, skipping", "account", cfg.Username, "error", err)
			continue
		}

		if !cfg.IsEnabled() {
			if _, known := w.watermark[cfg.Username]; known {
				w.log.Info("File watcher: account disabled, stopping miner", "account", cfg.Username)
				w.manager.Stop(cfg.Username)
				delete(w.watermark, cfg.Username)
			}
			continue
		}

		prev, known := w.watermark[cfg.Username]
		switch {
		case !known:
			w.log.Info("File watcher: new account detected, starting miner", "account", cfg.Username)
			w.manager.Start(cfg)
		case mtime > prev:
			w.log.Info("File watcher: config changed, restarting miner", "account", cfg.Username)
			w.manager.RestartChanged(cfg)
		}
		w.watermark[cfg.Username] = mtime
	}

	// Stop miners for accounts whose YAML was removed.
	for username := range w.watermark {
		if _, ok := seen[username]; !ok {
			w.log.Info("File watcher: config file removed, stopping miner", "account", username)
			w.manager.Stop(username)
			delete(w.watermark, username)
		}
	}
}

// fileMtime returns the modification time (Unix seconds) of the YAML file for
// the given username. Tries both .yaml and .yml extensions.
func (w *FileWatcher) fileMtime(username string) (int64, error) {
	for _, ext := range []string{".yaml", ".yml"} {
		info, err := os.Stat(filepath.Join(w.dir, username+ext))
		if err == nil {
			return info.ModTime().Unix(), nil
		}
	}
	// Fall back to directory mtime so the account is still tracked.
	info, err := os.Stat(w.dir)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}
