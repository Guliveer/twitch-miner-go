// Package notify provides notification dispatching to multiple providers
// (Telegram, Discord, Webhook, Matrix, Pushover, Gotify) based on event filtering.
package notify

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// defaultHTTPTimeout is the timeout for notification HTTP requests.
const defaultHTTPTimeout = 5 * time.Second

// Notifier is the interface that all notification providers must implement.
type Notifier interface {
	Send(ctx context.Context, event model.Event, title, message string) error
	Name() string
	IsEnabled() bool
	ShouldNotify(event model.Event) bool
}

// notifierEntry pairs a Notifier with an optional Batcher.
// When batcher is non-nil, notifications are routed through it.
type notifierEntry struct {
	notifier Notifier
	batcher  *Batcher
}

// Dispatcher manages multiple notifiers and dispatches notifications to all
// enabled notifiers that match the event.
type Dispatcher struct {
	entries []notifierEntry
	log     *logger.Logger
}

// NewDispatcher creates a Dispatcher from the notification configuration.
// It initialises all configured and enabled notification providers,
// optionally wrapping each in a Batcher when batching is enabled.
func NewDispatcher(cfg config.NotificationsConfig, log *logger.Logger) *Dispatcher {
	dispatcher := &Dispatcher{log: log}

	httpClient := &http.Client{
		Timeout: defaultHTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	// addEntry registers a notifier, wrapping it in a Batcher if the
	// resolved (global+provider) batch config is enabled.
	addEntry := func(n Notifier, providerBatch *config.BatchConfig) {
		resolved := config.ResolveBatchConfig(cfg.Batch, providerBatch)
		entry := notifierEntry{notifier: n}
		if resolved.IsBatchEnabled() {
			entry.batcher = NewBatcher(n, resolved, log)
			entry.batcher.Start()
		}
		dispatcher.entries = append(dispatcher.entries, entry)
	}

	if cfg.Telegram != nil && cfg.Telegram.Enabled {
		addEntry(&Telegram{
			baseNotifier:        baseNotifier{name: "Telegram", enabled: true, events: parseEvents(cfg.Telegram.Events)},
			token:               cfg.Telegram.Token,
			chatID:              cfg.Telegram.ChatID,
			disableNotification: cfg.Telegram.DisableNotification,
			httpClient:          httpClient,
		}, cfg.Telegram.Batch)
	}

	if cfg.Discord != nil && cfg.Discord.Enabled {
		addEntry(&Discord{
			baseNotifier: baseNotifier{name: "Discord", enabled: true, events: parseEvents(cfg.Discord.Events)},
			webhookURL:   cfg.Discord.WebhookURL,
			httpClient:   httpClient,
		}, cfg.Discord.Batch)
	}

	if cfg.Webhook != nil && cfg.Webhook.Enabled {
		method := cfg.Webhook.Method
		if method == "" {
			method = http.MethodPost
		}
		addEntry(&Webhook{
			baseNotifier: baseNotifier{name: "Webhook", enabled: true, events: parseEvents(cfg.Webhook.Events)},
			url:          cfg.Webhook.Endpoint,
			method:       method,
			httpClient:   httpClient,
		}, cfg.Webhook.Batch)
	}

	if cfg.Matrix != nil && cfg.Matrix.Enabled {
		addEntry(&Matrix{
			baseNotifier: baseNotifier{name: "Matrix", enabled: true, events: parseEvents(cfg.Matrix.Events)},
			homeserver:   cfg.Matrix.Homeserver,
			accessToken:  cfg.Matrix.AccessToken,
			roomID:       cfg.Matrix.RoomID,
			httpClient:   httpClient,
		}, cfg.Matrix.Batch)
	}

	if cfg.Pushover != nil && cfg.Pushover.Enabled {
		addEntry(&Pushover{
			baseNotifier: baseNotifier{name: "Pushover", enabled: true, events: parseEvents(cfg.Pushover.Events)},
			token:        cfg.Pushover.APIToken,
			userKey:      cfg.Pushover.UserKey,
			httpClient:   httpClient,
		}, cfg.Pushover.Batch)
	}

	if cfg.Gotify != nil && cfg.Gotify.Enabled {
		addEntry(&Gotify{
			baseNotifier: baseNotifier{name: "Gotify", enabled: true, events: parseEvents(cfg.Gotify.Events)},
			url:          cfg.Gotify.URL,
			token:        cfg.Gotify.Token,
			httpClient:   httpClient,
		}, cfg.Gotify.Batch)
	}

	return dispatcher
}

// Dispatch sends a notification to all enabled notifiers that match the event.
// For notifiers with batching enabled, the event is buffered; otherwise it is
// sent directly in a goroutine.
func (d *Dispatcher) Dispatch(ctx context.Context, event model.Event, title, message string) {
	for _, e := range d.entries {
		if !e.notifier.IsEnabled() || !e.notifier.ShouldNotify(event) {
			continue
		}
		if e.batcher != nil {
			e.batcher.Send(ctx, event, title, message)
		} else {
			go func(notifier Notifier) {
				sendCtx, cancel := context.WithTimeout(ctx, defaultHTTPTimeout)
				defer cancel()
				if err := notifier.Send(sendCtx, event, title, message); err != nil {
					d.log.Warn("notification send failed",
						"provider", notifier.Name(),
						"event", string(event),
						"error", err,
					)
				}
			}(e.notifier)
		}
	}
}

// Stop flushes all active batchers and releases resources.
// Must be called during graceful shutdown.
func (d *Dispatcher) Stop(ctx context.Context) {
	for _, e := range d.entries {
		if e.batcher != nil {
			e.batcher.Stop(ctx)
		}
	}
}

// DispatchSync sends a notification synchronously to all enabled notifiers
// matching the event, bypassing the batcher. It waits for all sends to complete.
// Use for lifecycle events (start/stop/crash) where delivery must be guaranteed
// before the dispatcher is shut down.
func (d *Dispatcher) DispatchSync(ctx context.Context, event model.Event, title, message string) {
	var wg sync.WaitGroup
	for _, e := range d.entries {
		if !e.notifier.IsEnabled() || !e.notifier.ShouldNotify(event) {
			continue
		}
		wg.Add(1)
		go func(notifier Notifier) {
			defer wg.Done()
			sendCtx, cancel := context.WithTimeout(ctx, defaultHTTPTimeout)
			defer cancel()
			if err := notifier.Send(sendCtx, event, title, message); err != nil {
				d.log.Warn("notification send failed",
					"provider", notifier.Name(),
					"event", string(event),
					"error", err,
				)
			}
		}(e.notifier)
	}
	wg.Wait()
}

// NotifyFunc returns a logger.NotifyFunc that dispatches notifications via this Dispatcher.
// The title is constructed dynamically from the account name and metadata map:
//   - "accountName | 📺 streamer" when a streamer context exists
//   - "accountName | 🎮 category" when only a category context exists (e.g. drop claims)
//   - "accountName" as a plain fallback
//   - "Twitch Miner" when no account name is available
func (d *Dispatcher) NotifyFunc(accountName string) logger.NotifyFunc {
	return func(ctx context.Context, message string, event model.Event, meta map[string]string) {
		streamer := meta["streamer"]
		category := meta["category"]

		var title string
		switch {
		case accountName != "" && streamer != "":
			title = fmt.Sprintf("%s | 📺 %s", accountName, streamer)
		case accountName != "" && category != "":
			title = fmt.Sprintf("%s | 🎮 %s", accountName, category)
		case accountName != "":
			title = accountName
		default:
			title = "Twitch Miner"
		}
		d.Dispatch(ctx, event, title, message)
	}
}

// TestAll sends a test notification to all enabled notifiers, bypassing event filters
// and batching.
func (d *Dispatcher) TestAll(ctx context.Context, title, message string) []error {
	var errs []error
	for _, e := range d.entries {
		if !e.notifier.IsEnabled() {
			continue
		}
		sendCtx, cancel := context.WithTimeout(ctx, defaultHTTPTimeout)
		if err := e.notifier.Send(sendCtx, model.EventTest, title, message); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.notifier.Name(), err))
		}
		cancel()
	}
	return errs
}

// HasNotifiers reports whether any notifiers are configured.
func (d *Dispatcher) HasNotifiers() bool {
	return len(d.entries) > 0
}

// parseEvents converts a slice of event name strings to model.Event values,
func parseEvents(names []string) []model.Event {
	events := make([]model.Event, 0, len(names))
	for _, name := range names {
		e := model.ParseEvent(name)
		if e != "" {
			events = append(events, e)
		}
	}
	return events
}

func containsEvent(events []model.Event, event model.Event) bool {
	for _, ev := range events {
		if ev == event {
			return true
		}
	}
	return false
}
