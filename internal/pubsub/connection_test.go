package pubsub

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

// mockAuthProvider is a test double for auth.Provider that tracks token
// refresh calls and allows controlling success/failure.
type mockAuthProvider struct {
	mu           sync.Mutex
	token        string
	userID       string
	refreshErr   error
	refreshCount int
}

func (m *mockAuthProvider) Login(_ context.Context) error                         { return nil }
func (m *mockAuthProvider) AuthToken() string                                     { m.mu.Lock(); defer m.mu.Unlock(); return m.token }
func (m *mockAuthProvider) UserID() string                                        { return m.userID }
func (m *mockAuthProvider) GetAuthHeaders() map[string]string                     { return nil }
func (m *mockAuthProvider) FetchIntegrityToken(_ context.Context) (string, error) { return "", nil }

func (m *mockAuthProvider) RefreshToken(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshCount++
	if m.refreshErr != nil {
		return m.refreshErr
	}
	m.token = "refreshed-token"
	return nil
}

func newTestConnection(auth *mockAuthProvider) *Connection {
	log, _ := logger.Setup(logger.Config{Level: 100}) // suppress output
	return &Connection{
		index:        0,
		topics:       make([]*model.PubSubTopic, 0),
		messages:     make(chan *model.Message, 8),
		writeCh:      make(chan []byte, 64),
		auth:         auth,
		log:          log,
		nonceToTopic: make(map[string]string),
		isConnected:  true,
	}
}

func TestHandleResponse_ERR_BADAUTH_RefreshesAndResubscribes(t *testing.T) {
	mock := &mockAuthProvider{token: "old-token", userID: "123"}
	c := newTestConnection(mock)

	topic := &model.PubSubTopic{
		TopicType: model.PubSubTopicCommunityPoints,
		Streamer:  &model.Streamer{ChannelID: "456", Username: "teststreamer"},
	}
	c.topics = append(c.topics, topic)
	topicStr := topic.String()

	nonce := "test-nonce-1"
	c.nonceToTopic[nonce] = topicStr

	ctx := context.Background()
	resp := &Response{
		Type:  TypeResponse,
		Nonce: nonce,
		Error: "ERR_BADAUTH",
	}

	c.handleResponse(ctx, resp)

	mock.mu.Lock()
	if mock.refreshCount != 1 {
		t.Errorf("expected RefreshToken to be called once, got %d", mock.refreshCount)
	}
	if mock.token != "refreshed-token" {
		t.Errorf("expected token to be refreshed, got %q", mock.token)
	}
	mock.mu.Unlock()

	// Verify that a LISTEN was re-sent via the write channel.
	select {
	case data := <-c.writeCh:
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			t.Fatalf("failed to unmarshal re-subscribe request: %v", err)
		}
		if req.Type != TypeListen {
			t.Errorf("expected LISTEN request, got %q", req.Type)
		}
		if req.Data == nil || len(req.Data.Topics) == 0 {
			t.Fatal("re-subscribe request has no topics")
		}
		if req.Data.Topics[0] != topicStr {
			t.Errorf("expected topic %q, got %q", topicStr, req.Data.Topics[0])
		}
		if req.Data.AuthToken != "refreshed-token" {
			t.Errorf("expected refreshed token, got %q", req.Data.AuthToken)
		}
	default:
		t.Error("expected a LISTEN request on write channel, but none found")
	}
}

func TestHandleResponse_ERR_BADAUTH_RefreshFailsNoResubscribe(t *testing.T) {
	mock := &mockAuthProvider{
		token:      "old-token",
		userID:     "123",
		refreshErr: context.DeadlineExceeded,
	}
	c := newTestConnection(mock)

	topic := &model.PubSubTopic{
		TopicType: model.PubSubTopicCommunityPoints,
		Streamer:  &model.Streamer{ChannelID: "456", Username: "teststreamer"},
	}
	c.topics = append(c.topics, topic)

	nonce := "test-nonce-2"
	c.nonceToTopic[nonce] = topic.String()

	ctx := context.Background()
	resp := &Response{
		Type:  TypeResponse,
		Nonce: nonce,
		Error: "ERR_BADAUTH",
	}

	c.handleResponse(ctx, resp)

	mock.mu.Lock()
	if mock.refreshCount != 1 {
		t.Errorf("expected RefreshToken to be called once, got %d", mock.refreshCount)
	}
	// Token should still be old since refresh failed and no one else refreshed.
	if mock.token != "old-token" {
		t.Errorf("expected token to remain %q, got %q", "old-token", mock.token)
	}
	mock.mu.Unlock()

	// No re-subscribe should be attempted.
	select {
	case data := <-c.writeCh:
		t.Errorf("expected no re-subscribe, but got: %s", data)
	default:
		// OK — no write enqueued.
	}
}

func TestHandleResponse_ERR_BADAUTH_AlreadyRefreshedByAnother(t *testing.T) {
	mock := &mockAuthProvider{
		token:      "already-refreshed-token",
		userID:     "123",
		refreshErr: context.DeadlineExceeded, // refresh fails
	}
	c := newTestConnection(mock)

	topic := &model.PubSubTopic{
		TopicType: model.PubSubTopicCommunityPoints,
		Streamer:  &model.Streamer{ChannelID: "789", Username: "otherstreamer"},
	}
	c.topics = append(c.topics, topic)

	nonce := "test-nonce-3"
	c.nonceToTopic[nonce] = topic.String()

	// The connection was created with "already-refreshed-token" but internally
	// AuthToken() will return the mock's current token. We simulate the case
	// where the token changed between the initial read and the refresh attempt
	// by making oldToken differ from the current token.
	// We do this by overriding AuthToken to return a changing value.

	// Instead, set up the mock so the first AuthToken() call returns one value
	// and subsequent calls return another.
	// Simpler approach: just verify that when refresh fails but token changed,
	// the re-subscribe still happens.

	// Save oldToken as what sendListen used (the "expired" token).
	// Simulate: retryAfterRefresh is called, oldToken = "old-expired", but
	// mock now returns "already-refreshed-token" (changed by another goroutine).
	// refreshErr makes RefreshToken fail, but token differs → re-subscribe.

	// Reset to simulate the scenario properly.
	mock.mu.Lock()
	mock.token = "old-expired"
	mock.mu.Unlock()

	ctx := context.Background()

	// Now simulate another goroutine refreshing the token between
	// retryAfterRefresh reading oldToken and calling RefreshToken.
	origRefreshToken := mock.RefreshToken
	_ = origRefreshToken
	mock.refreshErr = nil
	// We override the mock to change the token on refresh even though it "fails"
	// Actually let me restructure: refresh succeeds but we check re-subscribe.
	// The "already refreshed by another" case: refresh fails, token changed.

	mock.mu.Lock()
	mock.token = "old-expired"
	mock.refreshErr = context.DeadlineExceeded
	mock.mu.Unlock()

	// We need oldToken to be "old-expired" and then after RefreshToken is called
	// (which fails), AuthToken() returns something different.
	// We achieve this by having a custom mock that changes token on RefreshToken call.
	specialMock := &tokenChangingMock{
		token:      "old-expired",
		newToken:   "already-refreshed-token",
		refreshErr: context.DeadlineExceeded,
	}
	c.auth = specialMock

	c.handleResponse(ctx, &Response{
		Type:  TypeResponse,
		Nonce: nonce,
		Error: "ERR_BADAUTH",
	})

	// Even though refresh failed, since token changed, re-subscribe should happen.
	select {
	case data := <-c.writeCh:
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if req.Type != TypeListen {
			t.Errorf("expected LISTEN, got %q", req.Type)
		}
		if req.Data.AuthToken != "already-refreshed-token" {
			t.Errorf("expected already-refreshed token, got %q", req.Data.AuthToken)
		}
	default:
		t.Error("expected re-subscribe after token was refreshed by another connection")
	}
}

// tokenChangingMock simulates a token that changes during a RefreshToken call
// (as if another goroutine refreshed it concurrently).
type tokenChangingMock struct {
	mu         sync.Mutex
	token      string
	newToken   string
	refreshErr error
}

func (m *tokenChangingMock) Login(_ context.Context) error                         { return nil }
func (m *tokenChangingMock) AuthToken() string                                     { m.mu.Lock(); defer m.mu.Unlock(); return m.token }
func (m *tokenChangingMock) UserID() string                                        { return "123" }
func (m *tokenChangingMock) GetAuthHeaders() map[string]string                     { return nil }
func (m *tokenChangingMock) FetchIntegrityToken(_ context.Context) (string, error) { return "", nil }

func (m *tokenChangingMock) RefreshToken(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simulate the token being refreshed by another goroutine.
	m.token = m.newToken
	return m.refreshErr
}

func TestHandleResponse_OtherErrors_NoRefresh(t *testing.T) {
	mock := &mockAuthProvider{token: "valid-token", userID: "123"}
	c := newTestConnection(mock)

	nonce := "test-nonce-4"
	c.nonceToTopic[nonce] = "some-topic.123"

	ctx := context.Background()
	resp := &Response{
		Type:  TypeResponse,
		Nonce: nonce,
		Error: "ERR_BADMESSAGE",
	}

	c.handleResponse(ctx, resp)

	mock.mu.Lock()
	if mock.refreshCount != 0 {
		t.Errorf("RefreshToken should not be called for non-BADAUTH errors, got %d calls", mock.refreshCount)
	}
	mock.mu.Unlock()
}
