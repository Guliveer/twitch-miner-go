package managedminer

import (
	"context"
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
}

// fileManager is the subset of Manager that FileWatcher needs.
type fileManager interface {
	Start(cfg *config.AccountConfig)
	RestartChanged(cfg *config.AccountConfig)
	Stop(username string)
}

// NewFileWatcher creates a FileWatcher that polls dir every interval.
func NewFileWatcher(dir string, mgr fileManager, interval time.Duration, log *logger.Logger) *FileWatcher {
	return &FileWatcher{
		dir:       dir,
		manager:   mgr,
		interval:  interval,
		log:       log,
		watermark: make(map[string]int64),
	}
}

// Run starts the polling loop. It performs an initial sync and then blocks
// until ctx is cancelled.
func (w *FileWatcher) Run(ctx context.Context) {
	w.sync()

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
		// Only treat the directory as empty (and stop running miners) when
		// there genuinely are no YAML files. Any other error (parse failure,
		// permission issue, transient I/O) is logged and the last-known-good
		// set of miners is preserved.
		if isEmptyDirError(w.dir) {
			cfgs = nil
		} else {
			w.log.Error("File watcher: failed to load configs, keeping current miners", "dir", w.dir, "error", err)
			return
		}
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

// isEmptyDirError reports whether dir contains no YAML/YML files at all.
func isEmptyDirError(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext == ".yaml" || ext == ".yml" {
			return false
		}
	}
	return true
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
