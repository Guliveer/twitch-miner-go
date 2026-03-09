package notify

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// batchEntry holds a single buffered notification.
type batchEntry struct {
	message string
	event   model.Event
}

// batchKey groups entries by title (e.g. "oliwer | xQc").
type batchKey = string

// Batcher buffers notifications and flushes them periodically as batched
// messages. Immediate events bypass the buffer entirely.
type Batcher struct {
	notifier        Notifier
	log             *logger.Logger
	interval        time.Duration
	maxEntries      int
	immediateEvents map[model.Event]bool

	mu      sync.Mutex
	pending map[batchKey][]batchEntry

	stopOnce sync.Once
	done     chan struct{}
}

// NewBatcher creates a Batcher that wraps a Notifier with the given batch config.
func NewBatcher(notifier Notifier, cfg *config.BatchConfig, log *logger.Logger) *Batcher {
	immediate := make(map[model.Event]bool, len(cfg.ImmediateEvents))
	for _, name := range cfg.ImmediateEvents {
		if e := model.ParseEvent(name); e != "" {
			immediate[e] = true
		}
	}

	interval := cfg.Interval
	if interval == 0 {
		interval = 15 * time.Minute
	}

	maxEntries := cfg.MaxEntries
	if maxEntries == 0 {
		maxEntries = 15
	}

	return &Batcher{
		notifier:        notifier,
		log:             log,
		interval:        interval,
		maxEntries:      maxEntries,
		immediateEvents: immediate,
		pending:         make(map[batchKey][]batchEntry),
		done:            make(chan struct{}),
	}
}

// Start begins the periodic flush loop. Call Stop to terminate it.
func (b *Batcher) Start() {
	go b.flushLoop()
}

// Stop flushes remaining entries and stops the flush loop.
func (b *Batcher) Stop(ctx context.Context) {
	b.stopOnce.Do(func() {
		close(b.done)
		b.flush(ctx)
	})
}

// Send either dispatches immediately or buffers for the next batch flush.
func (b *Batcher) Send(ctx context.Context, event model.Event, title, message string) {
	if b.immediateEvents[event] {
		b.sendDirect(ctx, event, title, message)
		return
	}

	b.mu.Lock()
	b.pending[title] = append(b.pending[title], batchEntry{
		message: message,
		event:   event,
	})
	b.mu.Unlock()
}

func (b *Batcher) sendDirect(ctx context.Context, event model.Event, title, message string) {
	sendCtx, cancel := context.WithTimeout(ctx, defaultHTTPTimeout)
	defer cancel()
	if err := b.notifier.Send(sendCtx, event, title, message); err != nil {
		b.log.Warn("notification send failed",
			"provider", b.notifier.Name(),
			"event", string(event),
			"error", err,
		)
	}
}

func (b *Batcher) flushLoop() {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flush(context.Background())
		case <-b.done:
			return
		}
	}
}

func (b *Batcher) flush(ctx context.Context) {
	b.mu.Lock()
	snapshot := b.pending
	b.pending = make(map[batchKey][]batchEntry)
	b.mu.Unlock()

	for title, entries := range snapshot {
		if len(entries) == 0 {
			continue
		}

		// Single entry — send as a regular (unbatched) message.
		if len(entries) == 1 {
			b.sendDirect(ctx, entries[0].event, title, entries[0].message)
			continue
		}

		// Split into chunks of maxEntries.
		for i := 0; i < len(entries); i += b.maxEntries {
			end := min(i+b.maxEntries, len(entries))
			chunk := entries[i:end]

			var sb strings.Builder
			for j, e := range chunk {
				if j > 0 {
					sb.WriteByte('\n')
				}
				sb.WriteString(e.message)
			}

			// Use the event from the last entry — the most recent context.
			lastEvent := chunk[len(chunk)-1].event
			b.sendDirect(ctx, lastEvent, title, sb.String())
		}
	}
}
