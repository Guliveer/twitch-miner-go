package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/auth"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// logCapture wraps a bytes.Buffer and parses JSON log lines for assertions.
type logCapture struct {
	buf bytes.Buffer
}

func newLogCapture() *logCapture {
	return &logCapture{}
}

func (lc *logCapture) handler() slog.Handler {
	return slog.NewJSONHandler(&lc.buf, &slog.HandlerOptions{Level: slog.LevelDebug})
}

// entries returns all parsed log entries as map[string]any.
func (lc *logCapture) entries(t *testing.T) []map[string]any {
	t.Helper()
	var result []map[string]any
	dec := json.NewDecoder(&lc.buf)
	for dec.More() {
		var entry map[string]any
		if err := dec.Decode(&entry); err != nil {
			break
		}
		// Strip timing to make test output deterministic.
		delete(entry, "time")
		result = append(result, entry)
	}
	return result
}

// infoEntries returns log entries filtered to INFO level only, skipping
// internal GQL debug logs that are not the concern of these tests.
func (lc *logCapture) infoEntries(t *testing.T) []map[string]any {
	t.Helper()
	var info []map[string]any
	for _, e := range lc.entries(t) {
		if e["level"] == "INFO" {
			info = append(info, e)
		}
	}
	return info
}

func (lc *logCapture) reset() {
	lc.buf.Reset()
}

func newTestClientWithCapture(t *testing.T, transport *mockTransport) (*Client, *logCapture) {
	t.Helper()

	lc := newLogCapture()
	innerLog := slog.New(lc.handler())
	log := &logger.Logger{
		Logger: innerLog,
	}

	gqlClient := gql.NewClientForTest(&mockAuthProvider{}, log, &http.Client{Transport: transport})

	return &Client{
		Auth:      auth.NewForTest("12345"),
		Log:       log,
		GQL:       gqlClient,
		spadeURLs: &spadeCache{entries: make(map[string]spadeCacheEntry)},
	}, lc
}

// --- logDropStatuses tests ---

func TestLogDropStatuses_EmptyInventory(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	client, lc := newTestClientWithCapture(t, transport)

	client.logDropStatuses(context.Background())

	entries := lc.infoEntries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 info entry, got %d: %v", len(entries), entries)
	}
	if entries[0]["msg"] != "📦 No active drops" {
		t.Errorf("expected msg '📦 No active drops', got %v", entries[0]["msg"])
	}
	if _, ok := entries[0]["event"]; ok {
		t.Errorf("expected no event attribute for plain Info log, got event=%v", entries[0]["event"])
	}
}

func TestLogDropStatuses_NoCampaigns(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": {"currentUser": {"inventory": {"dropCampaignsInProgress": []}}}}`
	client, lc := newTestClientWithCapture(t, transport)

	client.logDropStatuses(context.Background())

	entries := lc.infoEntries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 info entry, got %d", len(entries))
	}
	if entries[0]["msg"] != "📦 No active drops" {
		t.Errorf("expected msg '📦 No active drops', got %v", entries[0]["msg"])
	}
}

func TestLogDropStatuses_DropInProgress(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{
			id:              "drop1",
			timeName:        "2 Hours",
			benefitName:     "Charm Pack",
			requiredMinutes: 120,
			currentMinutes:  15,
		},
	})
	client, lc := newTestClientWithCapture(t, transport)

	client.logDropStatuses(context.Background())

	entries := lc.infoEntries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 info entry, got %d: %v", len(entries), entries)
	}

	entry := entries[0]
	if entry["msg"] != "📦 Drop in progress" {
		t.Errorf("expected msg '📦 Drop in progress', got %v", entry["msg"])
	}
	if entry["progress"] != "15/120m" {
		t.Errorf("expected progress '15/120m', got %v", entry["progress"])
	}
	if entry["status"] != "IN_PROGRESS" {
		t.Errorf("expected status 'IN_PROGRESS', got %v", entry["status"])
	}
	if _, ok := entry["event"]; ok {
		t.Errorf("expected no event for non-claimable drop, got event=%v", entry["event"])
	}
}

func TestLogDropStatuses_ClaimableDropFiresEvent(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{
			id:              "drop1",
			timeName:        "2 Hours",
			benefitName:     "Charm Pack",
			instanceID:      "inst-1",
			requiredMinutes: 60,
			currentMinutes:  60,
		},
	})
	client, lc := newTestClientWithCapture(t, transport)

	client.logDropStatuses(context.Background())

	entries := lc.infoEntries(t)
	if len(entries) != 2 {
		t.Fatalf("expected 2 info entries (drop status + event), got %d: %v", len(entries), entries)
	}

	if entries[0]["msg"] != "📦 Drop in progress" {
		t.Errorf("expected first entry msg '📦 Drop in progress', got %v", entries[0]["msg"])
	}
	if _, ok := entries[0]["event"]; ok {
		t.Errorf("expected first entry to have no event, got event=%v", entries[0]["event"])
	}

	if entries[1]["msg"] != "📦 Drop available to claim: Charm Pack" {
		t.Errorf("expected event msg '📦 Drop available to claim: Charm Pack', got %v", entries[1]["msg"])
	}
	if entries[1]["event"] != "DROP_CLAIM_AVAILABLE" {
		t.Errorf("expected event=DROP_CLAIM_AVAILABLE, got %v", entries[1]["event"])
	}
	if entries[1]["reward"] != "Charm Pack" {
		t.Errorf("expected reward 'Charm Pack', got %v", entries[1]["reward"])
	}
}

func TestLogDropStatuses_ClaimableAlreadySeen_NoDuplicate(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{
			id:              "drop1",
			timeName:        "2 Hours",
			benefitName:     "Charm",
			instanceID:      "inst-1",
			requiredMinutes: 60,
			currentMinutes:  60,
		},
	})
	client, lc := newTestClientWithCapture(t, transport)

	// First call: should see claimable and fire event.
	client.logDropStatuses(context.Background())
	firstEntries := lc.infoEntries(t)
	if len(firstEntries) != 2 {
		t.Fatalf("expected 2 info entries on first call (info + event), got %d", len(firstEntries))
	}

	lc.reset()

	// Second call: should NOT fire the event again (already seen).
	client.logDropStatuses(context.Background())
	secondEntries := lc.infoEntries(t)
	if len(secondEntries) != 1 {
		t.Fatalf("expected 1 info entry on second call (just info, no duplicate event), got %d: %v",
			len(secondEntries), secondEntries)
	}
	if _, ok := secondEntries[0]["event"]; ok {
		t.Errorf("expected no event on second call (already seen), got event=%v", secondEntries[0]["event"])
	}
}

func TestLogDropStatuses_GQLError(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `not valid json`
	client, lc := newTestClientWithCapture(t, transport)

	client.logDropStatuses(context.Background())

	entries := lc.entries(t)
	if len(entries) == 0 {
		t.Fatal("expected at least 1 log entry, got 0")
	}

	found := false
	for _, e := range entries {
		if msg, ok := e["msg"].(string); ok && msg == "Failed to get drops inventory for status" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Failed to get drops inventory for status' debug entry, got: %v", entries)
	}
}

func TestLogDropStatuses_ClaimedDrop(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{
			id:              "drop1",
			timeName:        "2 Hours",
			benefitName:     "Charm",
			instanceID:      "inst-1",
			claimed:         true,
			requiredMinutes: 60,
			currentMinutes:  60,
		},
	})
	client, lc := newTestClientWithCapture(t, transport)

	client.logDropStatuses(context.Background())

	entries := lc.infoEntries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 info entry for claimed drop, got %d", len(entries))
	}
	if entries[0]["status"] != "CLAIMED" {
		t.Errorf("expected status 'CLAIMED', got %v", entries[0]["status"])
	}
	if _, ok := entries[0]["event"]; ok {
		t.Errorf("expected no event for claimed drop, got event=%v", entries[0]["event"])
	}
}

func TestLogDropStatuses_ClaimablePreSeen(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{
			id:              "drop1",
			timeName:        "2 Hours",
			benefitName:     "Charm",
			instanceID:      "inst-1",
			requiredMinutes: 60,
			currentMinutes:  60,
		},
	})
	client, lc := newTestClientWithCapture(t, transport)

	client.seenClaimable.Store("drop1", true)

	client.logDropStatuses(context.Background())

	entries := lc.infoEntries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 info entry (no event for pre-seen drop), got %d", len(entries))
	}
	if _, ok := entries[0]["event"]; ok {
		t.Errorf("expected no event for pre-seen drop, got event=%v", entries[0]["event"])
	}
}

// --- logDropProgress tests ---

// makeStreamerWithDrop returns a *model.Streamer with a single campaign containing
// one drop. The caller can further customize the drop's IsPrintable and
// HasPreconditionsMet fields as needed.
func makeStreamerWithDrop(username string, minutesRequired int) *model.Streamer {
	s := model.NewStreamer(username)

	game := &model.GameInfo{Name: "Test Game"}
	camp := model.NewCampaign("camp1", "Test Campaign", "ACTIVE", game,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour), nil)

	drop := model.NewDrop("drop1", "Test Drop", []string{"Reward"}, minutesRequired,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))

	camp.Drops = append(camp.Drops, drop)

	s.Stream.Campaigns = append(s.Stream.Campaigns, *camp)

	return s
}

func TestLogDropProgress_PrintableDrop(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())

	streamer := makeStreamerWithDrop("streamer1", 60)
	drop := streamer.Stream.Campaigns[0].Drops[0]
	drop.IsPrintable = true
	drop.HasPreconditionsMet = boolPtr(true)

	client.logDropProgress(context.Background(), streamer)

	entries := lc.entries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 event for printable drop, got %d: %v", len(entries), entries)
	}
	if entries[0]["event"] != "DROP_STATUS" {
		t.Errorf("expected event=DROP_STATUS, got %v", entries[0]["event"])
	}
}

func TestLogDropProgress_NonPrintableDrop(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())

	streamer := makeStreamerWithDrop("streamer1", 60)
	drop := streamer.Stream.Campaigns[0].Drops[0]
	drop.IsPrintable = false
	drop.HasPreconditionsMet = boolPtr(true)

	client.logDropProgress(context.Background(), streamer)

	entries := lc.entries(t)
	if len(entries) != 0 {
		t.Errorf("expected 0 events for non-printable drop, got %d: %v", len(entries), entries)
	}
}

func TestLogDropProgress_PreconditionsNotMet(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())

	streamer := makeStreamerWithDrop("streamer1", 60)
	drop := streamer.Stream.Campaigns[0].Drops[0]
	drop.IsPrintable = true
	drop.HasPreconditionsMet = boolPtr(false)

	client.logDropProgress(context.Background(), streamer)

	entries := lc.entries(t)
	if len(entries) != 0 {
		t.Errorf("expected 0 events when preconditions not met, got %d: %v", len(entries), entries)
	}
}

func TestLogDropProgress_NilPreconditions(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())

	streamer := makeStreamerWithDrop("streamer1", 60)
	drop := streamer.Stream.Campaigns[0].Drops[0]
	drop.IsPrintable = true
	drop.HasPreconditionsMet = nil

	client.logDropProgress(context.Background(), streamer)

	entries := lc.entries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 event when HasPreconditionsMet=nil (treated as met), got %d", len(entries))
	}
	if entries[0]["event"] != "DROP_STATUS" {
		t.Errorf("expected event=DROP_STATUS, got %v", entries[0]["event"])
	}
}

func TestLogDropProgress_NoCampaigns(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())

	streamer := model.NewStreamer("streamer1")

	client.logDropProgress(context.Background(), streamer)

	entries := lc.entries(t)
	if len(entries) != 0 {
		t.Errorf("expected 0 events for streamer with no campaigns, got %d", len(entries))
	}
}

func TestLogDropProgress_MixedPrintable(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())

	streamer := makeStreamerWithDrop("streamer1", 60)
	streamer.Stream.Campaigns[0].Drops[0].IsPrintable = true
	streamer.Stream.Campaigns[0].Drops[0].HasPreconditionsMet = boolPtr(true)

	// Add a second campaign with a non-printable drop.
	game2 := &model.GameInfo{Name: "Game B"}
	camp2 := model.NewCampaign("camp2", "Campaign 2", "ACTIVE", game2,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour), nil)
	drop2 := model.NewDrop("d2", "Drop 2", nil, 60,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	drop2.IsPrintable = false
	drop2.HasPreconditionsMet = boolPtr(true)
	camp2.Drops = append(camp2.Drops, drop2)

	streamer.Stream.Campaigns = append(streamer.Stream.Campaigns, *camp2)

	client.logDropProgress(context.Background(), streamer)

	entries := lc.entries(t)
	if len(entries) != 1 {
		t.Fatalf("expected 1 event (only printable drop fires), got %d: %v", len(entries), entries)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

// makeStreamerWithDropAtProgress returns a streamer with a drop configured
// at a specific percentage progress and DropInstanceID for milestone testing.
func makeStreamerWithDropAtProgress(username, dropID, instanceID string, minutesRequired, pct int) *model.Streamer {
	s := model.NewStreamer(username)
	s.Settings = &model.StreamerSettings{}

	game := &model.GameInfo{Name: "Test Game", Slug: "test-game"}
	camp := model.NewCampaign("camp1", "Test Campaign", "ACTIVE", game,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour), nil)

	drop := model.NewDrop(dropID, "Test Drop", []string{"Reward"}, minutesRequired,
		time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	drop.DropInstanceID = instanceID
	drop.PercentageProgress = pct
	drop.CurrentMinutesWatched = pct * minutesRequired / 100
	drop.IsPrintable = true

	camp.Drops = append(camp.Drops, drop)
	s.Stream.Campaigns = append(s.Stream.Campaigns, *camp)

	return s
}

func TestLogUpdatedDropProgress_MilestoneFiresAt25(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 25)

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	entries := lc.infoEntries(t)
	var found bool
	for _, e := range entries {
		if e["event"] == "DROP_MILESTONE" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DROP_MILESTONE event at 25%%, got entries: %v", entries)
	}
}

func TestLogUpdatedDropProgress_MilestoneFiresAt50(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 50)

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	entries := lc.infoEntries(t)
	var found bool
	for _, e := range entries {
		if e["event"] == "DROP_MILESTONE" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DROP_MILESTONE event at 50%%, got entries: %v", entries)
	}
}

func TestLogUpdatedDropProgress_MilestoneFiresAt75(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 75)

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	entries := lc.infoEntries(t)
	var found bool
	for _, e := range entries {
		if e["event"] == "DROP_MILESTONE" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DROP_MILESTONE event at 75%%, got entries: %v", entries)
	}
}

func TestLogUpdatedDropProgress_MilestoneFiresAt100(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 100)

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	entries := lc.infoEntries(t)
	var found bool
	for _, e := range entries {
		if e["event"] == "DROP_MILESTONE" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DROP_MILESTONE event at 100%%, got entries: %v", entries)
	}
}

func TestLogUpdatedDropProgress_MilestoneDedup(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 25)

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	count := milestoneEventCount(t, lc)
	if count != 1 {
		t.Fatalf("expected 1 DROP_MILESTONE on first call, got %d", count)
	}

	lc.reset()
	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	count = milestoneEventCount(t, lc)
	if count != 0 {
		t.Errorf("expected 0 DROP_MILESTONE on second call (dedup), got %d", count)
	}
}

func TestLogUpdatedDropProgress_NoMilestoneAtNonQuarter(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 30)

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	count := milestoneEventCount(t, lc)
	if count != 0 {
		t.Errorf("expected 0 DROP_MILESTONE at 30%%, got %d", count)
	}
}

func TestLogUpdatedDropProgress_NoMilestoneForNonPrintable(t *testing.T) {
	t.Parallel()

	client, lc := newTestClientWithCapture(t, newMockTransport())
	streamer := makeStreamerWithDropAtProgress("s1", "d1", "inst-1", 60, 25)
	streamer.Stream.Campaigns[0].Drops[0].IsPrintable = false

	client.logUpdatedDropProgress(context.Background(), []*model.Streamer{streamer})

	count := milestoneEventCount(t, lc)
	if count != 0 {
		t.Errorf("expected 0 DROP_MILESTONE for non-printable drop, got %d", count)
	}
}

func milestoneEventCount(t *testing.T, lc *logCapture) int {
	t.Helper()
	count := 0
	for _, e := range lc.infoEntries(t) {
		if e["event"] == "DROP_MILESTONE" {
			count++
		}
	}
	return count
}
