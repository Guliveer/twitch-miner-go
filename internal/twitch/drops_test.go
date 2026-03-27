package twitch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
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

func newTestClient(t *testing.T, transport *mockTransport) *Client {
	t.Helper()
	log, err := logger.Setup(logger.Config{Level: 100}) // suppress all log output
	if err != nil {
		t.Fatalf("logger setup: %v", err)
	}

	gqlClient := gql.NewClientForTest(&mockAuthProvider{}, log, &http.Client{Transport: transport})

	return &Client{
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

	inv := inventory{
		DropCampaignsInProgress: []campaign{{
			Game:           &game{Name: "Rainbow Six Siege", Slug: "tom-clancys-rainbow-six-siege"},
			TimeBasedDrops: entries,
		}},
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
