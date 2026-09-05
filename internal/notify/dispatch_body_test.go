package notify

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// capturingNotifier records what the dispatcher hands to a provider.
type capturingNotifier struct {
	baseNotifier
	mu       sync.Mutex
	title    string
	message  string
	received chan struct{}
}

func newCapturingNotifier() *capturingNotifier {
	return &capturingNotifier{
		baseNotifier: baseNotifier{name: "capture", enabled: true},
		received:     make(chan struct{}, 1),
	}
}

func (c *capturingNotifier) Send(_ context.Context, _ model.Event, title, message string) error {
	c.mu.Lock()
	c.title = title
	c.message = message
	c.mu.Unlock()

	select {
	case c.received <- struct{}{}:
	default:
	}
	return nil
}

func (c *capturingNotifier) snapshot() (string, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.title, c.message
}

// TestNotifyFuncDeliversChannelInBetBody walks the full chain a bet result
// takes—logger.Event fields → Dispatcher.NotifyFunc → provider Send—and
// asserts the channel survives into the message body, not just the title.
// Matrix ignored the title for years, which is exactly how it went missing.
func TestNotifyFuncDeliversChannelInBetBody(t *testing.T) {
	capture := newCapturingNotifier()
	d := &Dispatcher{entries: []notifierEntry{{notifier: capture}}}

	fn := d.NotifyFunc("3jakec")
	fn(context.Background(), model.EventBetWin, "🏆 Prediction result", []logger.Field{
		{Key: "streamer", Value: "psychohypnotic"},
		{Key: "category", Value: "phasmophobia"},
		{Key: "title", Value: "Will it be Obambo, Thaye or Il Mimo?"},
		{Key: "choice", Value: "Yes (BLUE)"},
		{Key: "amount", Value: "881"},
		{Key: "result", Value: "WIN, Gained: +1762"},
	})

	// Dispatch fires unbatched sends in a goroutine.
	select {
	case <-capture.received:
	case <-time.After(2 * time.Second):
		t.Fatal("notifier was never called")
	}

	title, message := capture.snapshot()

	if title != "3jakec | 📺 psychohypnotic" {
		t.Errorf("unexpected title: %q", title)
	}
	if !strings.Contains(message, "Channel: psychohypnotic (phasmophobia)") {
		t.Errorf("body does not name the channel:\n%s", message)
	}
	if !strings.Contains(message, "Pick: Yes (BLUE) · 881 points") {
		t.Errorf("body does not carry the stake:\n%s", message)
	}
}