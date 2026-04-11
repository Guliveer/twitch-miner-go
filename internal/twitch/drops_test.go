package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/auth"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// mockTransport intercepts HTTP requests and returns canned GQL responses
// based on the operation name found in the request body.
type mockTransport struct {
	mu        sync.Mutex
	responses map[string]string // operationName -> JSON response body
	calls     map[string]int    // operationName -> call count
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		responses: make(map[string]string),
		calls:     make(map[string]int),
	}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	req.Body.Close()

	bodyStr := strings.TrimSpace(string(body))

	// Handle batch requests (JSON array) used by PostGQLBatch.
	if strings.HasPrefix(bodyStr, "[") {
		var batch []struct {
			OperationName string `json:"operationName"`
		}
		json.Unmarshal(body, &batch)

		var items []string
		m.mu.Lock()
		for _, op := range batch {
			m.calls[op.OperationName]++
			if resp, ok := m.responses[op.OperationName]; ok {
				items = append(items, resp)
			} else {
				items = append(items, `{"data": null}`)
			}
		}
		m.mu.Unlock()

		batchResp := "[" + strings.Join(items, ",") + "]"
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(batchResp)),
			Header:     make(http.Header),
		}, nil
	}

	// Single operation.
	var payload struct {
		OperationName string `json:"operationName"`
	}
	json.Unmarshal(body, &payload)

	m.mu.Lock()
	m.calls[payload.OperationName]++
	resp, ok := m.responses[payload.OperationName]
	m.mu.Unlock()

	if !ok {
		resp = `{"data": null}`
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(resp)),
		Header:     make(http.Header),
	}, nil
}

func (m *mockTransport) callCount(op string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[op]
}

// mockAuthProvider satisfies auth.Provider for tests.
type mockAuthProvider struct{}

func (m *mockAuthProvider) Login(_ context.Context) error                         { return nil }
func (m *mockAuthProvider) AuthToken() string                                     { return "test-token" }
func (m *mockAuthProvider) UserID() string                                        { return "12345" }
func (m *mockAuthProvider) GetAuthHeaders() map[string]string                     { return map[string]string{"Authorization": "OAuth test"} }
func (m *mockAuthProvider) FetchIntegrityToken(_ context.Context) (string, error) { return "", nil }
func (m *mockAuthProvider) RefreshToken(_ context.Context) error                  { return nil }
func (m *mockAuthProvider) ClientVersion() string                                 { return "test" }
func (m *mockAuthProvider) ClientIDsForGQL() []string                             { return nil }
func (m *mockAuthProvider) AndroidClientID() string                               { return "android-test" }

func newTestClient(t *testing.T, transport *mockTransport) *Client {
	t.Helper()
	log, err := logger.Setup(logger.Config{Level: 100}) // suppress all log output
	if err != nil {
		t.Fatalf("logger setup: %v", err)
	}

	gqlClient := gql.NewClientForTest(&mockAuthProvider{}, log, &http.Client{Transport: transport})

	return &Client{
		Auth:      auth.NewForTest("12345"),
		Log:       log,
		GQL:       gqlClient,
		spadeURLs: &spadeCache{entries: make(map[string]spadeCacheEntry)},
	}
}

// inventoryJSON builds a mock Inventory GQL response matching the real
// Twitch API shape: {"data": {"currentUser": {"inventory": {...}}}}.
func inventoryJSON(drops []inventoryDrop) string {
	type selfData struct {
		DropInstanceID string `json:"dropInstanceID"`
		IsClaimed      bool   `json:"isClaimed"`
	}
	type benefit struct {
		Name string `json:"name"`
	}
	type benefitEdge struct {
		Benefit benefit `json:"benefit"`
	}
	type dropEntry struct {
		ID           string        `json:"id"`
		Name         string        `json:"name"`
		BenefitEdges []benefitEdge `json:"benefitEdges"`
		Self         *selfData     `json:"self"`
	}
	type game struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	type campaign struct {
		Game           *game       `json:"game"`
		TimeBasedDrops []dropEntry `json:"timeBasedDrops"`
	}
	type inventory struct {
		DropCampaignsInProgress []campaign `json:"dropCampaignsInProgress"`
	}

	var entries []dropEntry
	for _, d := range drops {
		var edges []benefitEdge
		if d.benefitName != "" {
			edges = []benefitEdge{{Benefit: benefit{Name: d.benefitName}}}
		}
		entries = append(entries, dropEntry{
			ID:           d.id,
			Name:         d.timeName,
			BenefitEdges: edges,
			Self: &selfData{
				DropInstanceID: d.instanceID,
				IsClaimed:      d.claimed,
			},
		})
	}

	// Put each drop in its own campaign to simulate the real inventory
	// where the same drop can appear across multiple campaigns.
	var campaigns []campaign
	for _, entry := range entries {
		campaigns = append(campaigns, campaign{
			Game:           &game{Name: "The Finals", Slug: "the-finals"},
			TimeBasedDrops: []dropEntry{entry},
		})
	}

	inv := inventory{
		DropCampaignsInProgress: campaigns,
	}

	// Match the real GQL response shape that GetDropsInventory parses.
	wrapper := struct {
		Data struct {
			CurrentUser struct {
				Inventory inventory `json:"inventory"`
			} `json:"currentUser"`
		} `json:"data"`
	}{}
	wrapper.Data.CurrentUser.Inventory = inv

	b, _ := json.Marshal(wrapper)
	return string(b)
}

type inventoryDrop struct {
	id          string
	timeName    string
	benefitName string
	instanceID  string
	claimed     bool
}

func claimSuccessResponse() string {
	return `{"data": {"claimDropRewards": {"status": "ELIGIBLE_FOR_ALL"}}}`
}

func claimFailedResponse(status string) string {
	return `{"data": {"claimDropRewards": {"status": "` + status + `"}}}`
}

// --- Tests ---

func TestClaimAllDrops_DeduplicatesAcrossCalls(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm Pack", instanceID: "inst-1"},
		{id: "drop2", timeName: "4 Hours", benefitName: "Weapon Skin", instanceID: "inst-2"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimSuccessResponse()

	client := newTestClient(t, transport)

	ctx := context.Background()

	// First call should claim both drops.
	if err := client.ClaimAllDropsFromInventory(ctx); err != nil {
		t.Fatalf("first call: %v", err)
	}

	firstCallClaims := transport.callCount("DropsPage_ClaimDropRewards")
	if firstCallClaims != 2 {
		t.Fatalf("expected 2 claim calls on first run, got %d", firstCallClaims)
	}

	// Second call should skip both (already attempted).
	if err := client.ClaimAllDropsFromInventory(ctx); err != nil {
		t.Fatalf("second call: %v", err)
	}

	secondCallClaims := transport.callCount("DropsPage_ClaimDropRewards")
	if secondCallClaims != 2 {
		t.Fatalf("expected still 2 total claim calls after second run (dedup), got %d", secondCallClaims)
	}
}

func TestClaimAllDrops_SkipsAlreadyClaimed(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: "inst-1", claimed: true},
		{id: "drop2", timeName: "4 Hours", benefitName: "Skin", instanceID: "inst-2", claimed: false},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimSuccessResponse()

	client := newTestClient(t, transport)

	if err := client.ClaimAllDropsFromInventory(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 1 {
		t.Fatalf("expected 1 claim call (skip already claimed), got %d", got)
	}
}

func TestClaimAllDrops_SkipsEmptyInstanceID(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: ""},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimSuccessResponse()

	client := newTestClient(t, transport)

	if err := client.ClaimAllDropsFromInventory(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 0 {
		t.Fatalf("expected 0 claim calls (empty instance ID), got %d", got)
	}
}

func TestClaimAllDrops_HandlesUnexpectedStatus(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: "inst-1"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimFailedResponse("PRECONDITIONS_NOT_MET")

	client := newTestClient(t, transport)

	// Should not return error — the error is logged as a warning, not propagated.
	if err := client.ClaimAllDropsFromInventory(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have attempted the claim exactly once.
	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 1 {
		t.Fatalf("expected 1 claim call, got %d", got)
	}

	// Second run should skip (dedup even on failure).
	if err := client.ClaimAllDropsFromInventory(context.Background()); err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 1 {
		t.Fatalf("expected still 1 claim call after second run, got %d", got)
	}
}

func TestClaimAllDrops_DedupesSameDropAcrossCampaigns(t *testing.T) {
	t.Parallel()

	// Same drop ID ("drop1") appears 4 times with different instance IDs,
	// simulating the same drop available from 4 different campaigns.
	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Accounting Services", instanceID: "inst-1"},
		{id: "drop1", timeName: "2 Hours", benefitName: "Accounting Services", instanceID: "inst-2"},
		{id: "drop1", timeName: "2 Hours", benefitName: "Accounting Services", instanceID: "inst-3"},
		{id: "drop1", timeName: "2 Hours", benefitName: "Accounting Services", instanceID: "inst-4"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimSuccessResponse()

	client := newTestClient(t, transport)

	if err := client.ClaimAllDropsFromInventory(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only 1 claim call despite 4 instances — dedup by drop definition ID.
	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 1 {
		t.Fatalf("expected 1 claim call (dedup across campaigns), got %d", got)
	}
}

func TestClaimAllDrops_EmptyInventory(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`

	client := newTestClient(t, transport)

	if err := client.ClaimAllDropsFromInventory(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 0 {
		t.Fatalf("expected 0 claim calls for empty inventory, got %d", got)
	}
}

// --- SyncCampaigns / CAMPAIGN_REMINDER notification tests ---

// eventCapture collects notification events fired via Logger.Event.
type eventCapture struct {
	mu     sync.Mutex
	events []capturedEvent
}

type capturedEvent struct {
	event model.Event
	meta  map[string]string
}

func newEventCapture(log *logger.Logger) *eventCapture {
	ec := &eventCapture{}
	log.SetNotifyFunc(func(_ context.Context, _ string, event model.Event, meta map[string]string) {
		ec.mu.Lock()
		ec.events = append(ec.events, capturedEvent{event: event, meta: meta})
		ec.mu.Unlock()
	})
	return ec
}

func (ec *eventCapture) countEvent(event model.Event) int {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	count := 0
	for _, e := range ec.events {
		if e.event == event {
			count++
		}
	}
	return count
}

func (ec *eventCapture) reset() {
	ec.mu.Lock()
	ec.events = nil
	ec.mu.Unlock()
}

// dashboardJSON builds a ViewerDropsDashboard GQL response with the given campaign IDs.
func dashboardJSON(ids ...string) string {
	var items []string
	for _, id := range ids {
		items = append(items, fmt.Sprintf(`{"id":%q,"status":"ACTIVE"}`, id))
	}
	return fmt.Sprintf(`{"data":{"currentUser":{"dropCampaigns":[%s]}}}`, strings.Join(items, ","))
}

// campaignDetailJSON builds a DropCampaignDetails batch-element response for a single campaign.
func campaignDetailJSON(id, name, gameName, gameSlug string) string {
	now := time.Now()
	start := now.Add(-24 * time.Hour).Format(time.RFC3339)
	end := now.Add(24 * time.Hour).Format(time.RFC3339)
	return fmt.Sprintf(`{"data":{"user":{"dropCampaign":{`+
		`"id":%q,"name":%q,"status":"ACTIVE","startAt":%q,"endAt":%q,`+
		`"game":{"id":"game1","name":%q,"displayName":%q,"slug":%q},`+
		`"allow":{"channels":[{"id":"chan1"}]},`+
		`"timeBasedDrops":[{"id":"drop-%s","name":"Drop","requiredMinutesWatched":60,`+
		`"startAt":%q,"endAt":%q,"benefitEdges":[{"benefit":{"name":"Reward"}}]}]`+
		`}}}}`, id, name, start, end, gameName, gameName, gameSlug, id, start, end)
}

// testStreamer creates a minimal streamer that satisfies DropsCondition and matches the given campaign.
// withReminders controls whether the streamer opts in to CAMPAIGN_REMINDER notifications.
func testStreamer(gameName string, withReminders bool, campaignIDs ...string) *model.Streamer {
	s := model.NewStreamer("teststreamer")
	s.IsOnline = true
	if withReminders {
		s.CampaignReminders = &model.CampaignReminderConfig{OnDetection: true}
	}
	s.CategorySlug = strings.ToLower(strings.ReplaceAll(gameName, " ", "-"))
	s.Stream.Game = &model.GameInfo{Name: gameName, Slug: strings.ToLower(strings.ReplaceAll(gameName, " ", "-"))}
	s.Stream.CampaignIDs = campaignIDs
	s.Settings = &model.StreamerSettings{ClaimDrops: true}
	return s
}

func TestSyncCampaigns_SeedsOnFirstRun(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp1")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp1", "Season 5 Drops", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	streamers := []*model.Streamer{testStreamer("The Finals", true, "camp1")}

	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First call seeds knownCampaigns. No CAMPAIGN_REMINDER should fire because
	// this campaign is active (StartAt in the past), so catch-up is skipped.
	if got := capture.countEvent(model.EventCampaignReminder); got != 0 {
		t.Fatalf("expected 0 CAMPAIGN_REMINDER events on first sync (active campaign), got %d", got)
	}

	// Verify the campaign was stored in knownCampaigns.
	if _, ok := client.knownCampaigns.Load("camp1"); !ok {
		t.Fatal("expected camp1 to be stored in knownCampaigns after first sync")
	}
}

func TestSyncCampaigns_SkipsKnownCampaign(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp1")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp1", "Season 5 Drops", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	streamers := []*model.Streamer{testStreamer("The Finals", true, "camp1")}

	// First sync — seeds.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	capture.reset()

	// Second sync — same campaign, should NOT fire.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if got := capture.countEvent(model.EventCampaignReminder); got != 0 {
		t.Fatalf("expected 0 CAMPAIGN_REMINDER events for known campaign, got %d", got)
	}
}

func TestSyncCampaigns_NotifiesNewCampaign(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp1")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp1", "Season 5 Drops", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	streamers := []*model.Streamer{testStreamer("The Finals", true, "camp1", "camp2")}

	// First sync — seeds camp1.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	capture.reset()

	// Update mock: dashboard now returns camp2, details return camp2 data.
	transport.mu.Lock()
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp2")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp2", "New Event Drops", "The Finals", "the-finals")
	transport.mu.Unlock()

	// Second sync — should detect camp2 as new.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if got := capture.countEvent(model.EventCampaignReminder); got != 1 {
		t.Fatalf("expected 1 CAMPAIGN_REMINDER event for new campaign, got %d", got)
	}
}

func TestSyncCampaigns_SkipsWhenNotifyDisabled(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp1")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp1", "Season 5 Drops", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	// No reminders configured — should NOT fire CAMPAIGN_REMINDER even for new campaigns.
	streamers := []*model.Streamer{testStreamer("The Finals", false, "camp1", "camp2")}

	// First sync — seeds.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	capture.reset()

	// Update mock: new campaign appears.
	transport.mu.Lock()
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp2")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp2", "New Event Drops", "The Finals", "the-finals")
	transport.mu.Unlock()

	// Second sync — camp2 is new but notify is disabled.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if got := capture.countEvent(model.EventCampaignReminder); got != 0 {
		t.Fatalf("expected 0 CAMPAIGN_REMINDER events (notify disabled), got %d", got)
	}
}

// --- Upcoming campaign / reminder-specific helpers and tests ---

// upcomingDashboardJSON builds a dashboard response with UPCOMING status.
func upcomingDashboardJSON(ids ...string) string {
	var items []string
	for _, id := range ids {
		items = append(items, fmt.Sprintf(`{"id":%q,"status":"UPCOMING"}`, id))
	}
	return fmt.Sprintf(`{"data":{"currentUser":{"dropCampaigns":[%s]}}}`, strings.Join(items, ","))
}

// upcomingCampaignDetailJSON builds a campaign detail response with StartAt in the future.
func upcomingCampaignDetailJSON(id, name, gameName, gameSlug string, startIn, duration time.Duration) string {
	now := time.Now()
	start := now.Add(startIn).Format(time.RFC3339)
	end := now.Add(startIn + duration).Format(time.RFC3339)
	return fmt.Sprintf(`{"data":{"user":{"dropCampaign":{`+
		`"id":%q,"name":%q,"status":"UPCOMING","startAt":%q,"endAt":%q,`+
		`"game":{"id":"game1","name":%q,"displayName":%q,"slug":%q},`+
		`"allow":{"channels":[{"id":"chan1"}]},`+
		`"timeBasedDrops":[{"id":"drop-%s","name":"Drop","requiredMinutesWatched":60,`+
		`"startAt":%q,"endAt":%q,"benefitEdges":[{"benefit":{"name":"Reward"}}]}]`+
		`}}}}`, id, name, start, end, gameName, gameName, gameSlug, id, start, end)
}

// testStreamerWithDurations creates a streamer with time-based reminders (no on_detection).
func testStreamerWithDurations(gameName string, durations []time.Duration, campaignIDs ...string) *model.Streamer {
	s := model.NewStreamer("teststreamer")
	s.IsOnline = true
	s.CampaignReminders = &model.CampaignReminderConfig{Durations: durations}
	s.CategorySlug = strings.ToLower(strings.ReplaceAll(gameName, " ", "-"))
	s.Stream.Game = &model.GameInfo{Name: gameName, Slug: strings.ToLower(strings.ReplaceAll(gameName, " ", "-"))}
	s.Stream.CampaignIDs = campaignIDs
	s.Settings = &model.StreamerSettings{ClaimDrops: true}
	return s
}

func TestSyncCampaigns_CatchUpOnFirstSync(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	// Upcoming campaign starting in 2 hours.
	transport.responses["ViewerDropsDashboard"] = upcomingDashboardJSON("camp-upcoming")
	transport.responses["DropCampaignDetails"] = upcomingCampaignDetailJSON(
		"camp-upcoming", "Future Drops", "The Finals", "the-finals",
		2*time.Hour, 72*time.Hour,
	)

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	streamers := []*model.Streamer{testStreamer("The Finals", true, "camp-upcoming")}

	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First sync with an upcoming campaign should fire a catch-up notification.
	if got := capture.countEvent(model.EventCampaignReminder); got != 1 {
		t.Fatalf("expected 1 catch-up CAMPAIGN_REMINDER on first sync, got %d", got)
	}

	// Verify campaign was seeded in knownCampaigns.
	if _, ok := client.knownCampaigns.Load("camp-upcoming"); !ok {
		t.Fatal("expected camp-upcoming to be stored in knownCampaigns after first sync")
	}
}

func TestSyncCampaigns_NoCatchUpForActiveCampaign(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	// Active campaign (already started) — should NOT trigger catch-up.
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp-active")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp-active", "Current Drops", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	streamers := []*model.Streamer{testStreamer("The Finals", true, "camp-active")}

	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Active campaigns don't get catch-up (StartAt is in the past).
	if got := capture.countEvent(model.EventCampaignReminder); got != 0 {
		t.Fatalf("expected 0 CAMPAIGN_REMINDER for active campaign on first sync, got %d", got)
	}
}

func TestSyncCampaigns_OnDetectionAfterFirstSync(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp1")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp1", "Season 5", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	streamers := []*model.Streamer{testStreamer("The Finals", true, "camp1")}

	// First sync — seeds.
	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("first sync: %v", err)
	}
	capture.reset()

	// New upcoming campaign appears on second sync.
	transport.mu.Lock()
	transport.responses["ViewerDropsDashboard"] = upcomingDashboardJSON("camp-new")
	transport.responses["DropCampaignDetails"] = upcomingCampaignDetailJSON(
		"camp-new", "New Event", "The Finals", "the-finals",
		48*time.Hour, 72*time.Hour,
	)
	transport.mu.Unlock()

	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	// on_detection should fire for newly seen campaign.
	if got := capture.countEvent(model.EventCampaignReminder); got != 1 {
		t.Fatalf("expected 1 on_detection CAMPAIGN_REMINDER, got %d", got)
	}
}

func TestSyncCampaigns_CategoryMatchBySlug(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp1")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp1", "Drops", "The Finals", "the-finals")

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	// Streamer watching a different category — should NOT match.
	wrongCat := testStreamer("Valorant", true)
	wrongCat.CategorySlug = "valorant"
	wrongCat.Stream.Game = &model.GameInfo{Name: "Valorant", Slug: "valorant"}

	// First sync seeds.
	if err := client.SyncCampaigns(context.Background(), []*model.Streamer{wrongCat}); err != nil {
		t.Fatalf("first sync: %v", err)
	}
	capture.reset()

	// New campaign — wrong category, should not fire.
	transport.mu.Lock()
	transport.responses["ViewerDropsDashboard"] = dashboardJSON("camp2")
	transport.responses["DropCampaignDetails"] = campaignDetailJSON("camp2", "New Finals Drops", "The Finals", "the-finals")
	transport.mu.Unlock()

	if err := client.SyncCampaigns(context.Background(), []*model.Streamer{wrongCat}); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if got := capture.countEvent(model.EventCampaignReminder); got != 0 {
		t.Fatalf("expected 0 CAMPAIGN_REMINDER for mismatched category, got %d", got)
	}
}

func TestSyncCampaigns_CatchUpFiresWithDurationsOnly(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	// Upcoming campaign starting in 2 hours.
	transport.responses["ViewerDropsDashboard"] = upcomingDashboardJSON("camp-upcoming")
	transport.responses["DropCampaignDetails"] = upcomingCampaignDetailJSON(
		"camp-upcoming", "Future Drops", "The Finals", "the-finals",
		2*time.Hour, 72*time.Hour,
	)

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	// Streamer with ONLY time-based durations (no on_detection).
	streamers := []*model.Streamer{testStreamerWithDurations("The Finals", []time.Duration{24 * time.Hour, time.Hour})}

	if err := client.SyncCampaigns(context.Background(), streamers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First sync should still fire a catch-up even without on_detection,
	// because the streamer has reminders configured and the campaign is upcoming.
	if got := capture.countEvent(model.EventCampaignReminder); got != 1 {
		t.Fatalf("expected 1 catch-up CAMPAIGN_REMINDER with durations-only config, got %d", got)
	}
}

func TestSyncCampaigns_NoReminderWithoutConfig(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = `{"data": null}`
	transport.responses["ViewerDropsDashboard"] = upcomingDashboardJSON("camp-upcoming")
	transport.responses["DropCampaignDetails"] = upcomingCampaignDetailJSON(
		"camp-upcoming", "Future Drops", "The Finals", "the-finals",
		2*time.Hour, 72*time.Hour,
	)

	client := newTestClient(t, transport)
	capture := newEventCapture(client.Log)

	// Streamer with NO reminders configured (nil CampaignReminders).
	s := model.NewStreamer("teststreamer")
	s.IsOnline = true
	s.CategorySlug = "the-finals"
	s.Stream.Game = &model.GameInfo{Name: "The Finals", Slug: "the-finals"}
	s.Settings = &model.StreamerSettings{ClaimDrops: true}

	if err := client.SyncCampaigns(context.Background(), []*model.Streamer{s}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No reminders configured — no notifications at all.
	if got := capture.countEvent(model.EventCampaignReminder); got != 0 {
		t.Fatalf("expected 0 CAMPAIGN_REMINDER without config, got %d", got)
	}
}
