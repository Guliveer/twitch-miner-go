package notify

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// mockNotifier records all Send calls for assertions.
type mockNotifier struct {
	baseNotifier
	mu    sync.Mutex
	sends []mockSend
}

type mockSend struct {
	event   model.Event
	title   string
	message string
}

func (m *mockNotifier) Send(_ context.Context, event model.Event, title, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sends = append(m.sends, mockSend{event: event, title: title, message: message})
	return nil
}

func (m *mockNotifier) getSends() []mockSend {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]mockSend, len(m.sends))
	copy(cp, m.sends)
	return cp
}

func boolPtr(b bool) *bool { return &b }

func newTestLogger() *logger.Logger {
	l, _ := logger.Setup(logger.DefaultConfig())
	return l
}

func TestBatcher_ImmediateEventsBypassBuffer(t *testing.T) {
	mock := &mockNotifier{baseNotifier: baseNotifier{name: "test", enabled: true}}
	cfg := &config.BatchConfig{
		Enabled:         boolPtr(true),
		Interval:        1 * time.Hour, // long interval — should not flush
		ImmediateEvents: []string{"BET_WIN", "STREAMER_ONLINE"},
	}

	b := NewBatcher(mock, cfg, newTestLogger())
	b.Start()
	defer b.Stop(context.Background())

	b.Send(context.Background(), model.EventBetWin, "user | xQc", "You won!")
	b.Send(context.Background(), model.EventStreamerOnline, "user | shroud", "Stream online")

	// Immediate events should be sent right away — small sleep for goroutine scheduling.
	time.Sleep(50 * time.Millisecond)

	sends := mock.getSends()
	if len(sends) != 2 {
		t.Fatalf("expected 2 immediate sends, got %d", len(sends))
	}
	if sends[0].message != "You won!" {
		t.Errorf("expected 'You won!', got %q", sends[0].message)
	}
	if sends[1].title != "user | shroud" {
		t.Errorf("expected 'user | shroud', got %q", sends[1].title)
	}
}

func TestBatcher_BuffersAndFlushes(t *testing.T) {
	mock := &mockNotifier{baseNotifier: baseNotifier{name: "test", enabled: true}}
	cfg := &config.BatchConfig{
		Enabled:  boolPtr(true),
		Interval: 1 * time.Hour,
	}

	b := NewBatcher(mock, cfg, newTestLogger())
	// Don't Start() — we'll flush manually.

	ctx := context.Background()
	b.Send(ctx, model.EventGainForWatch, "user | xQc", "+50 points")
	b.Send(ctx, model.EventBonusClaim, "user | xQc", "Claiming bonus")
	b.Send(ctx, model.EventGainForWatch, "user | shroud", "+50 points")

	// Nothing sent yet.
	if len(mock.getSends()) != 0 {
		t.Fatalf("expected 0 sends before flush, got %d", len(mock.getSends()))
	}

	b.flush(ctx)

	sends := mock.getSends()
	if len(sends) != 2 {
		t.Fatalf("expected 2 batched sends (one per title), got %d", len(sends))
	}

	// Find the xQc batch — should be two lines joined.
	var xqcMsg string
	for _, s := range sends {
		if s.title == "user | xQc" {
			xqcMsg = s.message
		}
	}
	expected := "+50 points\nClaiming bonus"
	if xqcMsg != expected {
		t.Errorf("expected batched message %q, got %q", expected, xqcMsg)
	}
}

func TestBatcher_SingleEntryNotBatched(t *testing.T) {
	mock := &mockNotifier{baseNotifier: baseNotifier{name: "test", enabled: true}}
	cfg := &config.BatchConfig{
		Enabled:  boolPtr(true),
		Interval: 1 * time.Hour,
	}

	b := NewBatcher(mock, cfg, newTestLogger())

	ctx := context.Background()
	b.Send(ctx, model.EventGainForWatch, "user | xQc", "+50 points")

	b.flush(ctx)

	sends := mock.getSends()
	if len(sends) != 1 {
		t.Fatalf("expected 1 send, got %d", len(sends))
	}
	// Single entry should be sent as-is, not wrapped in batch format.
	if sends[0].message != "+50 points" {
		t.Errorf("expected '+50 points', got %q", sends[0].message)
	}
}

func TestBatcher_MaxEntriesSplitsMessages(t *testing.T) {
	mock := &mockNotifier{baseNotifier: baseNotifier{name: "test", enabled: true}}
	cfg := &config.BatchConfig{
		Enabled:    boolPtr(true),
		Interval:   1 * time.Hour,
		MaxEntries: 3,
	}

	b := NewBatcher(mock, cfg, newTestLogger())

	ctx := context.Background()
	for i := 0; i < 7; i++ {
		b.Send(ctx, model.EventGainForWatch, "user | xQc", "+50 points")
	}

	b.flush(ctx)

	sends := mock.getSends()
	// 7 entries / 3 max = 3 messages (3 + 3 + 1)
	if len(sends) != 3 {
		t.Fatalf("expected 3 split messages, got %d", len(sends))
	}
}

func TestBatcher_StopFlushesRemaining(t *testing.T) {
	mock := &mockNotifier{baseNotifier: baseNotifier{name: "test", enabled: true}}
	cfg := &config.BatchConfig{
		Enabled:  boolPtr(true),
		Interval: 1 * time.Hour,
	}

	b := NewBatcher(mock, cfg, newTestLogger())
	b.Start()

	b.Send(context.Background(), model.EventGainForWatch, "user | xQc", "+50 points")

	// Stop should flush.
	b.Stop(context.Background())

	sends := mock.getSends()
	if len(sends) != 1 {
		t.Fatalf("expected 1 send after Stop, got %d", len(sends))
	}
}
