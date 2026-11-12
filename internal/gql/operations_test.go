package gql

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
)

// mockTransport intercepts HTTP requests and returns canned responses.
// It also captures the last request body for assertion.
type mockTransport struct {
	mu          sync.Mutex
	lastBody    []byte
	respBody    string
	statusCode  int
	callCount   int
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	m.mu.Lock()
	m.lastBody = body
	m.callCount++
	m.mu.Unlock()

	statusCode := m.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(m.respBody)),
		Header:     make(http.Header),
	}, nil
}

func (m *mockTransport) getCapturedVars(t *testing.T) map[string]any {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastBody == nil {
		t.Fatal("no request was captured")
	}

	var req struct {
		Variables map[string]any `json:"variables"`
	}
	if err := json.Unmarshal(m.lastBody, &req); err != nil {
		t.Fatalf("failed to parse captured request body: %v\nbody: %s", err, string(m.lastBody))
	}
	return req.Variables
}

func newTestGQLClient(transport *mockTransport) *Client {
	log, _ := logger.Setup(logger.Config{Level: 100})
	return NewClientForTest(&noopAuth{}, log, &http.Client{Transport: transport})
}

// noopAuth satisfies auth.Provider with no-op implementations.
type noopAuth struct{}

func (n *noopAuth) Login(_ context.Context) error       { return nil }
func (n *noopAuth) AuthToken() string                   { return "test-token" }
func (n *noopAuth) UserID() string                      { return "12345" }
func (n *noopAuth) GetAuthHeaders() map[string]string   { return map[string]string{} }
func (n *noopAuth) ClientVersion() string               { return "test" }
func (n *noopAuth) ClientIDsForGQL() []string           { return nil }
func (n *noopAuth) AndroidClientID() string             { return "android-test" }
func (n *noopAuth) FetchIntegrityToken(_ context.Context) (string, error) { return "", nil }
func (n *noopAuth) RefreshToken(_ context.Context) error                  { return nil }

// silence slog for tests
func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// assertVarPresent checks that a variable exists and is non-empty.
func assertVarPresent(t *testing.T, vars map[string]any, key string) {
	t.Helper()
	val, ok := vars[key]
	if !ok {
		t.Errorf("missing required variable %q in request", key)
		return
	}
	if val == nil {
		t.Errorf("variable %q is nil, expected non-nil value", key)
	}
}

// assertVarEquals checks that a variable equals the expected value.
func assertVarEquals(t *testing.T, vars map[string]any, key string, expected any) {
	t.Helper()
	val, ok := vars[key]
	if !ok {
		t.Errorf("missing variable %q in request", key)
		return
	}
	if val != expected {
		t.Errorf("variable %q = %v (%T), want %v (%T)", key, val, val, expected, expected)
	}
}

func TestGetPlaybackAccessToken_SendAllRequiredVariables(t *testing.T) {
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"streamPlaybackAccessToken":{"signature":"sig123","value":"token456"}}}`,
	}
	client := newTestGQLClient(transport)

	resp, err := client.GetPlaybackAccessToken(context.Background(), "testchannel")
	if err != nil {
		t.Fatalf("GetPlaybackAccessToken failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Signature != "sig123" {
		t.Errorf("Signature = %q, want %q", resp.Signature, "sig123")
	}

	vars := transport.getCapturedVars(t)

	// All required variables for PlaybackAccessToken query.
	assertVarPresent(t, vars, "login")
	assertVarPresent(t, vars, "isLive")
	assertVarPresent(t, vars, "isVod")
	assertVarPresent(t, vars, "vodID")
	assertVarPresent(t, vars, "playerType")
	assertVarPresent(t, vars, "platform") // This was missing before the fix

	assertVarEquals(t, vars, "login", "testchannel")
	assertVarEquals(t, vars, "isLive", true)
	assertVarEquals(t, vars, "isVod", false)
	assertVarEquals(t, vars, "vodID", "")
	assertVarEquals(t, vars, "playerType", "site")
	assertVarEquals(t, vars, "platform", "web")
}

func TestGetChannelPointsContext_SendsChannelLogin(t *testing.T) {
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"community":{"channel":{"self":{"communityPoints":{"balance":100,"activeMultipliers":[],"availableClaim":null}},"communityPointsSettings":{"goals":[]}}}}}`,
	}
	client := newTestGQLClient(transport)

	_, err := client.GetChannelPointsContext(context.Background(), "testchannel")
	if err != nil {
		t.Fatalf("GetChannelPointsContext failed: %v", err)
	}

	vars := transport.getCapturedVars(t)
	assertVarPresent(t, vars, "channelLogin")
	assertVarEquals(t, vars, "channelLogin", "testchannel")
}

func TestGetStreamInfo_SendsChannelLogin(t *testing.T) {
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"user":{"stream":null}}}`,
	}
	client := newTestGQLClient(transport)

	_, err := client.GetStreamInfo(context.Background(), "testchannel")
	if err != nil {
		t.Fatalf("GetStreamInfo failed: %v", err)
	}

	vars := transport.getCapturedVars(t)
	assertVarPresent(t, vars, "channel")
	assertVarEquals(t, vars, "channel", "testchannel")
}

func TestGetUserID_SendsLogin(t *testing.T) {
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"user":{"id":"99999"}}}`,
	}
	client := newTestGQLClient(transport)

	id, err := client.GetUserID(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("GetUserID failed: %v", err)
	}
	if id != "99999" {
		t.Errorf("UserID = %q, want %q", id, "99999")
	}

	vars := transport.getCapturedVars(t)
	assertVarPresent(t, vars, "login")
	assertVarEquals(t, vars, "login", "testuser")
}

func (m *mockTransport) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// TestGetFollowedStreamers_UsesDataAlongsidePartialErrors covers accounts that
// follow channels Twitch reports per-node "service error" entries for. The data
// is complete, so the errors must not discard it or trigger client-ID retries.
func TestGetFollowedStreamers_UsesDataAlongsidePartialErrors(t *testing.T) {
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"user":{"follows":{"edges":[` +
			`{"node":{"login":"alpha"},"cursor":"c1"},` +
			`{"node":{"login":"beta"},"cursor":"c2"}],` +
			`"pageInfo":{"hasNextPage":false}}}},` +
			`"errors":[{"message":"service error"},{"message":"service error"}]}`,
	}
	client := newTestGQLClient(transport)

	follows, err := client.GetFollowedStreamers(context.Background(), 100, "DESC")
	if err != nil {
		t.Fatalf("GetFollowedStreamers failed: %v", err)
	}

	want := []string{"alpha", "beta"}
	if !reflect.DeepEqual(follows, want) {
		t.Errorf("follows = %v, want %v", follows, want)
	}
	if got := transport.getCallCount(); got != 1 {
		t.Errorf("request count = %d, want 1 (partial errors must not trigger retries)", got)
	}
}

// TestGetFollowedStreamers_FailsWhenDataMissing verifies that errors still fail
// the call when the follows data itself is absent.
func TestGetFollowedStreamers_FailsWhenDataMissing(t *testing.T) {
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"user":null},"errors":[{"message":"service error"}]}`,
	}
	client := newTestGQLClient(transport)

	if _, err := client.GetFollowedStreamers(context.Background(), 100, "DESC"); err == nil {
		t.Fatal("expected an error when the follows data is missing")
	}
}

func TestGetPlaybackAccessToken_PlatformIsNonEmptyString(t *testing.T) {
	// Regression test: Twitch GQL API requires platform as a non-null String.
	// Before the fix, platform was not sent, causing:
	// Variable "platform" has invalid value null. Expected type "String!", found null.
	t.Parallel()

	transport := &mockTransport{
		respBody: `{"data":{"streamPlaybackAccessToken":{"signature":"sig","value":"tok"}}}`,
	}
	client := newTestGQLClient(transport)

	_, err := client.GetPlaybackAccessToken(context.Background(), "channel")
	if err != nil {
		t.Fatalf("GetPlaybackAccessToken failed: %v", err)
	}

	vars := transport.getCapturedVars(t)

	platform, ok := vars["platform"]
	if !ok {
		t.Fatal("platform variable is missing — Twitch GQL requires it as String!")
	}
	platformStr, ok := platform.(string)
	if !ok {
		t.Fatalf("platform is %T, want string", platform)
	}
	if platformStr == "" {
		t.Error("platform is empty string, Twitch GQL requires non-empty String!")
	}
}
