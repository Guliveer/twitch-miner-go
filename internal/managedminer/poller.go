package managedminer

import (
	"context"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/store"
)

// minerManager is the subset of Manager that Poller needs.
type minerManager interface {
	Start(cfg *config.AccountConfig)
	RestartChanged(cfg *config.AccountConfig)
	Stop(username string)
}

// Poller watches a Store for changes and drives a Manager accordingly.
// On every tick (or LISTEN/NOTIFY signal from the store) it compares the DB
// state to its last-seen watermark and calls Start / Restart / Stop as needed.
type Poller struct {
	store    store.Store
	manager  minerManager
	interval time.Duration
	log      *logger.Logger

	// watermark: username → last seen updated_at Unix seconds
	watermark map[string]int64
}

// NewPoller creates a Poller that polls the given store on the given interval.
func NewPoller(st store.Store, mgr minerManager, interval time.Duration, log *logger.Logger) *Poller {
	return &Poller{
		store:     st,
		manager:   mgr,
		interval:  interval,
		log:       log,
		watermark: make(map[string]int64),
	}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	p.sync(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	changes := p.store.Changes()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.sync(ctx)
		case _, ok := <-changes:
			if !ok {
				changes = nil
				continue
			}
			p.sync(ctx)
		}
	}
}

func (p *Poller) touch(username string) {
	if err := p.store.TouchLastStartedAt(username); err != nil {
		p.log.Error("Failed to update last_started_at", "account", username, "error", err)
	}
}

func (p *Poller) sync(_ context.Context) {
	rows, err := p.store.ListAccounts()
	if err != nil {
		p.log.Error("DB sync failed", "error", err)
		return
	}

	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		seen[row.Username] = struct{}{}
		ts := row.UpdatedAt.UnixMilli()
		prev, known := p.watermark[row.Username]

		if !row.Enabled {
			if known {
				p.log.Info("Account disabled in DB, stopping miner", "account", row.Username)
				p.manager.Stop(row.Username)
				delete(p.watermark, row.Username)
			}
			continue
		}

		cfg, err := config.AccountConfigFromJSON(row.Username, row.ConfigJSON)
		if err != nil {
			p.log.Error("Failed to parse account config from DB", "account", row.Username, "error", err)
			continue
		}
		if err := config.Validate(cfg); err != nil {
			p.log.Error("Invalid account config from DB", "account", row.Username, "error", err)
			continue
		}

		switch {
		case !known:
			p.log.Info("New account detected in DB, starting miner", "account", row.Username)
			p.manager.Start(cfg)
			p.touch(row.Username)
		case ts > prev:
			p.log.Info("Account config changed in DB, restarting miner", "account", row.Username)
			p.manager.RestartChanged(cfg)
			p.touch(row.Username)
		}

		p.watermark[row.Username] = ts
	}

	// Stop miners for accounts no longer in DB.
	for username := range p.watermark {
		if _, ok := seen[username]; !ok {
			p.log.Info("Account removed from DB, stopping miner", "account", username)
			p.manager.Stop(username)
			delete(p.watermark, username)
		}
	}
}
