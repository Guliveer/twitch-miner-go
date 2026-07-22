package twitch

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/auth"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// mockTransport intercepts HTTP requests and returns canned GQL responses
// based on the operation name found in the request body.
type mockTransport struct {
	mu            sync.Mutex
	responses     map[string]string                                     // operationName -> JSON response body
	responseFuncs map[string]func(callCount int, vars map[string]any) string
	statusFuncs   map[string]func(callCount int, vars map[string]any) int
	calls         map[string]int // operationName -> call count
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		responses: make(map[string]string),
		calls:     make(map[string]int),
	}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

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
		OperationName string         `json:"operationName"`
		Variables     map[string]any `json:"variables"`
	}
	json.Unmarshal(body, &payload)

	m.mu.Lock()
	m.calls[payload.OperationName]++

	var resp string
	if fn, ok := m.responseFuncs[payload.OperationName]; ok {
		resp = fn(m.calls[payload.OperationName], payload.Variables)
	} else if r, ok := m.responses[payload.OperationName]; ok {
		resp = r
	} else {
		resp = `{"data": null}`
	}
	status := http.StatusOK
	if fn, ok := m.statusFuncs[payload.OperationName]; ok {
		status = fn(m.calls[payload.OperationName], payload.Variables)
	}
	m.mu.Unlock()

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(resp)),
		Header:     make(http.Header),
	}, nil
}

func (m *mockTransport) setResponseFunc(op string, fn func(callCount int, vars map[string]any) string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.responseFuncs == nil {
		m.responseFuncs = make(map[string]func(callCount int, vars map[string]any) string)
	}
	m.responseFuncs[op] = fn
}

func (m *mockTransport) setStatusFunc(op string, fn func(callCount int, vars map[string]any) int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.statusFuncs == nil {
		m.statusFuncs = make(map[string]func(callCount int, vars map[string]any) int)
	}
	m.statusFuncs[op] = fn
}

func (m *mockTransport) callCount(op string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[op]
}

// mockAuthProvider satisfies auth.Provider for tests.
type mockAuthProvider struct{}

func (m *mockAuthProvider) Login(_ context.Context) error { return nil }
func (m *mockAuthProvider) AuthToken() string             { return "test-token" }
func (m *mockAuthProvider) UserID() string                { return "12345" }
func (m *mockAuthProvider) GetAuthHeaders() map[string]string {
	return map[string]string{"Authorization": "OAuth test"}
}
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
		DropInstanceID        string `json:"dropInstanceID"`
		IsClaimed             bool   `json:"isClaimed"`
		HasPreconditionsMet   bool   `json:"hasPreconditionsMet"`
		CurrentMinutesWatched int    `json:"currentMinutesWatched"`
	}
	type benefit struct {
		Name string `json:"name"`
	}
	type benefitEdge struct {
		Benefit benefit `json:"benefit"`
	}
	type dropEntry struct {
		ID              string        `json:"id"`
		Name            string        `json:"name"`
		RequiredMinutes int           `json:"requiredMinutesWatched"`
		BenefitEdges    []benefitEdge `json:"benefitEdges"`
		Self            *selfData     `json:"self"`
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
			ID:              d.id,
			Name:            d.timeName,
			RequiredMinutes: d.requiredMinutes,
			BenefitEdges:    edges,
			Self: &selfData{
				DropInstanceID:        d.instanceID,
				IsClaimed:             d.claimed,
				HasPreconditionsMet:   d.hasPreconditionsMet,
				CurrentMinutesWatched: d.currentMinutes,
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
	id                  string
	timeName            string
	benefitName         string
	instanceID          string
	claimed             bool
	requiredMinutes     int
	currentMinutes      int
	hasPreconditionsMet bool
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

// vanishInventoryGQLResponse builds the full GQL envelope for GetDropsInventory
// with the given drop IDs present in the inventory.
func vanishInventoryGQLResponse(dropIDs []string) string {
	type self struct {
		HasPreconditionsMet   bool `json:"hasPreconditionsMet"`
		CurrentMinutesWatched int  `json:"currentMinutesWatched"`
		IsClaimed             bool `json:"isClaimed"`
	}
	type drop struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Self *self  `json:"self"`
	}
	type campaign struct {
		ID             string `json:"id"`
		TimeBasedDrops []drop `json:"timeBasedDrops"`
	}
	type inventory struct {
		DropCampaignsInProgress []campaign `json:"dropCampaignsInProgress"`
	}

	drops := make([]drop, 0, len(dropIDs))
	for _, id := range dropIDs {
		drops = append(drops, drop{
			ID:   id,
			Name: "test",
			Self: &self{HasPreconditionsMet: true, CurrentMinutesWatched: 5},
		})
	}
	inv := inventory{DropCampaignsInProgress: []campaign{{ID: "camp1", TimeBasedDrops: drops}}}

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

// setVanishMocks configures the mock transport for SyncCampaigns vanish tests.
// The inventory function is called for every "Inventory" single-request.
// Batch "DropCampaignDetails" uses the responses map (not responseFuncs).
func setVanishMocks(transport *mockTransport, inventoryResp func() string) {
	transport.responses["DropCampaignDetails"] = `{"data":{"user":{"dropCampaign":{"id":"camp1","game":{"name":"G","slug":"g"},"name":"C","startAt":"2025-01-01T00:00:00Z","endAt":"2099-01-01T00:00:00Z","detailsURL":"","status":"ACTIVE","timeBasedDrops":[{"id":"d1","name":"D1","requiredMinutesWatched":60,"benefitEdges":[],"startAt":"2025-01-01T00:00:00Z","endAt":"2099-01-01T00:00:00Z"}]}}}}`
	transport.setResponseFunc("ViewerDropsDashboard", func(_ int, _ map[string]any) string {
		return `{"data":{"currentUser":{"dropCampaigns":[{"id":"camp1"}]}}}`
	})
	transport.setResponseFunc("Inventory", func(_ int, _ map[string]any) string {
		return inventoryResp()
	})
}

func TestVanishDetection_DropVanishedAfterThreePolls(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	hasDrop := true
	setVanishMocks(transport, func() string {
		if hasDrop {
			return vanishInventoryGQLResponse([]string{"d1"})
		}
		return vanishInventoryGQLResponse([]string{})
	})

	ctx := context.Background()

	// Polls 1-3: drop present
	client.SyncCampaigns(ctx, nil)
	client.SyncCampaigns(ctx, nil)
	client.SyncCampaigns(ctx, nil)

	// Polls 4-6: drop missing
	hasDrop = false
	client.SyncCampaigns(ctx, nil)
	client.SyncCampaigns(ctx, nil)
	client.SyncCampaigns(ctx, nil)

	if _, vanished := client.dropVanished.Load("d1"); !vanished {
		t.Fatal("expected drop d1 to be detected as vanished after 3 missing polls")
	}
}

func TestVanishDetection_DropStillPresent(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	setVanishMocks(transport, func() string {
		return vanishInventoryGQLResponse([]string{"d1"})
	})

	ctx := context.Background()

	for i := 0; i < 6; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	if _, vanished := client.dropVanished.Load("d1"); vanished {
		t.Fatal("drop d1 should not be detected as vanished when always present")
	}
}

func TestVanishDetection_DropReappears(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	hasDrop := true
	setVanishMocks(transport, func() string {
		if hasDrop {
			return vanishInventoryGQLResponse([]string{"d1"})
		}
		return vanishInventoryGQLResponse([]string{})
	})

	ctx := context.Background()

	// Poll 1: present
	client.SyncCampaigns(ctx, nil)

	// Polls 2-3: missing (only 2, not enough for vanish)
	hasDrop = false
	client.SyncCampaigns(ctx, nil)
	client.SyncCampaigns(ctx, nil)

	// Poll 4: reappears
	hasDrop = true
	client.SyncCampaigns(ctx, nil)

	if _, vanished := client.dropVanished.Load("d1"); vanished {
		t.Fatal("drop d1 should not be detected as vanished after reappearing")
	}
}

func TestVanishDetection_DropNeverSeenNotTracked(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	setVanishMocks(transport, func() string {
		return vanishInventoryGQLResponse([]string{})
	})

	ctx := context.Background()

	// 6 polls with empty inventory — d1 was never present
	for i := 0; i < 6; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	if _, vanished := client.dropVanished.Load("d1"); vanished {
		t.Fatal("drop d1 should not be detected as vanished when never seen in inventory")
	}
}

func TestSynthSkip_DropNeverInInventoryMarkedSynthetic(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	setVanishMocks(transport, func() string {
		return vanishInventoryGQLResponse([]string{})
	})

	ctx := context.Background()

	for i := 0; i < 6; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	if _, synthetic := client.dropSynthetic.Load("d1"); !synthetic {
		t.Fatal("expected drop d1 to be detected as synthetic after 6 polls without appearing in inventory")
	}
}

func TestSynthSkip_DropAppearsInInventoryNotSynthetic(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	setVanishMocks(transport, func() string {
		return vanishInventoryGQLResponse([]string{"d1"})
	})

	ctx := context.Background()

	for i := 0; i < 6; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	if _, synthetic := client.dropSynthetic.Load("d1"); synthetic {
		t.Fatal("drop d1 should not be detected as synthetic when present in inventory")
	}
}

func TestSynthSkip_DropAppearsLateNotSynthetic(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	hasDrop := false
	setVanishMocks(transport, func() string {
		if hasDrop {
			return vanishInventoryGQLResponse([]string{"d1"})
		}
		return vanishInventoryGQLResponse([]string{})
	})

	ctx := context.Background()

	for i := 0; i < 4; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	hasDrop = true
	client.SyncCampaigns(ctx, nil)

	client.SyncCampaigns(ctx, nil)
	client.SyncCampaigns(ctx, nil)

	if _, synthetic := client.dropSynthetic.Load("d1"); synthetic {
		t.Fatal("drop d1 should not be detected as synthetic when it appeared in inventory before threshold")
	}
}

func TestSynthSkip_DropResetsWhenAppears(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	hasDrop := false
	setVanishMocks(transport, func() string {
		if hasDrop {
			return vanishInventoryGQLResponse([]string{"d1"})
		}
		return vanishInventoryGQLResponse([]string{})
	})

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	hasDrop = true
	client.SyncCampaigns(ctx, nil)

	hasDrop = false
	for i := 0; i < 4; i++ {
		client.SyncCampaigns(ctx, nil)
	}

	if _, synthetic := client.dropSynthetic.Load("d1"); synthetic {
		t.Fatal("drop d1 should not be detected as synthetic when counter was reset by reappearance")
	}
}

func TestLogUpdatedDropProgress_WhenClaimDropsFalse(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	client := newTestClient(t, transport)

	streamer := &model.Streamer{
		Username: "test_streamer",
		Settings: &model.StreamerSettings{
			ClaimDrops: false,
		},
		Stream: &model.Stream{
			Campaigns: []model.Campaign{
				{
					ID:   "campaign-1",
					Name: "Test Campaign",
					Game: &model.GameInfo{Name: "TestGame"},
					Drops: []*model.Drop{
						{
							ID:                    "drop-1",
							Name:                  "Test Drop",
							Benefit:               "Test Benefit",
							MinutesRequired:       60,
							CurrentMinutesWatched: 30,
							PercentageProgress:    50,
							IsPrintable:           true,
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	client.logUpdatedDropProgress(ctx, []*model.Streamer{streamer})

	entry := streamer.History[string(model.EventDropStatus)]
	if entry == nil {
		t.Fatal("expected DROP_STATUS history entry, got nil — logUpdatedDropProgress skipped streamer with ClaimDrops=false")
	}
	if entry.Counter != 1 {
		t.Fatalf("expected counter=1, got %d", entry.Counter)
	}
}

func TestClaimAllDrops_DetectsAccountLinkRequired(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: "inst-1"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimFailedResponse("PRECONDITIONS_NOT_MET")

	lc := newLogCapture()
	client, _ := newTestClientWithCapture(t, transport)
	_ = lc

	err := client.ClaimAllDropsFromInventory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 1 {
		t.Fatalf("expected 1 claim call, got %d", got)
	}
}

func TestClaimAllDrops_AccountLinkWarningLogged(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: "inst-1"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimFailedResponse("PRECONDITIONS_NOT_MET")

	lc := newLogCapture()
	innerLog := slog.New(lc.handler())
	log := &logger.Logger{Logger: innerLog}

	gqlClient := gql.NewClientForTest(&mockAuthProvider{}, log, &http.Client{Transport: transport})
	client := &Client{
		Auth:      auth.NewForTest("12345"),
		Log:       log,
		GQL:       gqlClient,
		spadeURLs: &spadeCache{entries: make(map[string]spadeCacheEntry)},
	}

	err := client.ClaimAllDropsFromInventory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := lc.entries(t)
	found := false
	for _, e := range entries {
		if e["level"] == "WARN" {
			if msg, ok := e["msg"].(string); ok {
				if strings.Contains(strings.ToLower(msg), "account") ||
					strings.Contains(strings.ToLower(msg), "link") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected a WARN log entry mentioning account linking, none found")
	}
}

func TestClaimAllDrops_NormalClaimingUnaffected(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: "inst-1"},
		{id: "drop2", timeName: "4 Hours", benefitName: "Skin", instanceID: "inst-2"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimSuccessResponse()

	client := newTestClient(t, transport)

	err := client.ClaimAllDropsFromInventory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := transport.callCount("DropsPage_ClaimDropRewards"); got != 2 {
		t.Fatalf("expected 2 claim calls, got %d", got)
	}
}

func TestClaimAllDrops_AccountLinkNotFalsePositive(t *testing.T) {
	t.Parallel()

	transport := newMockTransport()
	transport.responses["Inventory"] = inventoryJSON([]inventoryDrop{
		{id: "drop1", timeName: "2 Hours", benefitName: "Charm", instanceID: "inst-1"},
	})
	transport.responses["DropsPage_ClaimDropRewards"] = claimSuccessResponse()

	lc := newLogCapture()
	innerLog := slog.New(lc.handler())
	log := &logger.Logger{Logger: innerLog}

	gqlClient := gql.NewClientForTest(&mockAuthProvider{}, log, &http.Client{Transport: transport})
	client := &Client{
		Auth:      auth.NewForTest("12345"),
		Log:       log,
		GQL:       gqlClient,
		spadeURLs: &spadeCache{entries: make(map[string]spadeCacheEntry)},
	}

	err := client.ClaimAllDropsFromInventory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := lc.entries(t)
	for _, e := range entries {
		if e["level"] == "WARN" {
			if msg, ok := e["msg"].(string); ok {
				if strings.Contains(strings.ToLower(msg), "account") &&
					strings.Contains(strings.ToLower(msg), "link") {
					t.Error("unexpected account link warning for successful claim")
				}
			}
		}
	}
}
