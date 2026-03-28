package pubsub

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	authpkg "github.com/Guliveer/twitch-miner-go/internal/auth"
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
func (m *mockAuthProvider) ClientVersion() string                                 { return "test-version" }
func (m *mockAuthProvider) ClientIDsForGQL() []string                             { return []string{"browser"} }
func (m *mockAuthProvider) AndroidClientID() string                               { return "android-test" }
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

func newTestConnection(authProvider authpkg.Provider) *Connection {
	log, _ := logger.Setup(logger.Config{Level: 100}) // suppress output
	if authProvider == nil {
		authProvider = &mockAuthProvider{}
	}
	return &Connection{
		index:        0,
		topics:       make([]*model.PubSubTopic, 0),
		messages:     make(chan *model.Message, 8),
		writeCh:      make(chan []byte, 64),
		auth:         authProvider,
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
	// Use a special mock that simulates the token being changed by another
	// goroutine during the RefreshToken call: RefreshToken returns an error,
	// but the token has changed, so re-subscribe should still happen.
	specialMock := &tokenChangingMock{
		token:      "old-expired",
		newToken:   "already-refreshed-token",
		refreshErr: context.DeadlineExceeded,
	}
	c := newTestConnection(nil)
	c.auth = specialMock

	topic := &model.PubSubTopic{
		TopicType: model.PubSubTopicCommunityPoints,
		Streamer:  &model.Streamer{ChannelID: "789", Username: "otherstreamer"},
	}
	c.topics = append(c.topics, topic)

	nonce := "test-nonce-3"
	c.nonceToTopic[nonce] = topic.String()

	ctx := context.Background()
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
func (m *tokenChangingMock) ClientVersion() string                                 { return "test-version" }
func (m *tokenChangingMock) ClientIDsForGQL() []string                             { return []string{"browser"} }
func (m *tokenChangingMock) AndroidClientID() string                               { return "android-test" }
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

func TestHandleResponse_ReconnectClosesConnection(t *testing.T) {
	mock := &mockAuthProvider{token: "valid-token", userID: "123"}
	c := newTestConnection(mock)

	ctx := context.Background()
	c.handleResponse(ctx, &Response{Type: TypeReconnect})

	if c.IsConnected() {
		t.Fatal("expected connection to be marked disconnected after reconnect request")
	}

	select {
	case _, ok := <-c.Messages():
		if ok {
			t.Fatal("expected messages channel to be closed after reconnect request")
		}
	default:
		t.Fatal("expected messages channel to close after reconnect request")
	}
}
