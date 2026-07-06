package miner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/auth"
	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/logger"
	"github.com/Guliveer/twitch-miner-go/internal/model"
	"github.com/Guliveer/twitch-miner-go/internal/config"
)

// mockTwitch is a test double for twitch.API that tracks JoinRaid calls.
type mockTwitch struct {
	mu         sync.Mutex
	joinCalls  []string // raid IDs passed to JoinRaid
}

func (m *mockTwitch) JoinRaid(_ context.Context, raidID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.joinCalls = append(m.joinCalls, raidID)
	return nil
}

func (m *mockTwitch) Login(_ context.Context) error                         { return nil }
func (m *mockTwitch) CheckStreamerOnline(_ context.Context, _ *model.Streamer) error { return nil }
func (m *mockTwitch) LoadChannelPointsContext(_ context.Context, _ *model.Streamer) error { return nil }
func (m *mockTwitch) SendMinuteWatchedEvents(_ context.Context, _ []*model.Streamer) error { return nil }
func (m *mockTwitch) MakePrediction(_ context.Context, _ *model.Streamer, _ *model.EventPrediction) error { return nil }
func (m *mockTwitch) ClaimChannelPoints(_ context.Context, _ *model.Streamer, _ string) error { return nil }
func (m *mockTwitch) ClaimMoment(_ context.Context, _ string) error { return nil }
func (m *mockTwitch) SyncCampaigns(_ context.Context, _ []*model.Streamer) error { return nil }
func (m *mockTwitch) ClaimAllDropsFromInventory(_ context.Context) error { return nil }
func (m *mockTwitch) GetChannelID(_ context.Context, _ string) (string, error) { return "", nil }
func (m *mockTwitch) GetFollowers(_ context.Context, _ int, _ string) ([]string, error) { return nil, nil }
func (m *mockTwitch) CheckViewerIsMod(_ context.Context, _ *model.Streamer) {}
func (m *mockTwitch) RefreshSpadeURL(_ context.Context, _ *model.Streamer) error { return nil }
func (m *mockTwitch) GQLClient() *gql.Client { return nil }
func (m *mockTwitch) AuthProvider() auth.Provider { return nil }

func newTestMiner(t *testing.T) (*Miner, *mockTwitch) {
	t.Helper()

	log, err := logger.Setup(logger.Config{Level: 100}) // suppress all output
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	mock := &mockTwitch{}

	m := &Miner{
		cfg:               &config.AccountConfig{},
		log:               log,
		twitch:            mock,
		eventsPredictions: make(map[string]*model.EventPrediction),
		pendingTimers:     make(map[string]*time.Timer),
		lastWatching:      make(map[string]bool),
	}

	return m, mock
}

// raidMessage builds a model.Message that simulates a raid_update_v2 PubSub event.
func raidMessage(channelID, raidID, targetLogin string) *model.Message {
	return &model.Message{
		Topic:     "raid",
		ChannelID: channelID,
		Type:      model.MsgTypeRaidUpdate,
		RawMessage: map[string]any{
			"type": string(model.MsgTypeRaidUpdate),
			"raid": map[string]any{
				"id":           raidID,
				"target_login": targetLogin,
			},
		},
	}
}

// TestHandleRaid_Dedup verifies that JOIN_RAID is emitted only once per raid ID.
func TestHandleRaid_Dedup(t *testing.T) {
	m, mock := newTestMiner(t)

	streamer := &model.Streamer{
		Username:  "teststreamer",
		ChannelID: "chan123",
		Settings: &model.StreamerSettings{
			FollowRaid: true,
		},
	}
	m.streamersMu.Lock()
	m.streamers = append(m.streamers, streamer)
	m.streamersMu.Unlock()

	ctx := context.Background()

	// First raid message — should call JoinRaid and set streamer.Raid.
	msg1 := raidMessage("chan123", "raid-001", "target_streamer")
	m.HandlePubSubMessage(ctx, msg1)

	mock.mu.Lock()
	if len(mock.joinCalls) != 1 {
		t.Errorf("expected 1 JoinRaid call after first message, got %d", len(mock.joinCalls))
	}
	if mock.joinCalls[0] != "raid-001" {
		t.Errorf("expected JoinRaid call with 'raid-001', got %q", mock.joinCalls[0])
	}
	mock.mu.Unlock()

	if streamer.Raid == nil {
		t.Fatal("expected streamer.Raid to be set after first raid message")
	}
	if streamer.Raid.RaidID != "raid-001" {
		t.Errorf("expected streamer.Raid.RaidID 'raid-001', got %q", streamer.Raid.RaidID)
	}
	if streamer.Raid.TargetLogin != "target_streamer" {
		t.Errorf("expected streamer.Raid.TargetLogin 'target_streamer', got %q", streamer.Raid.TargetLogin)
	}

	// Second raid message with the same raid ID — should be deduplicated.
	msg2 := raidMessage("chan123", "raid-001", "target_streamer")
	m.HandlePubSubMessage(ctx, msg2)

	mock.mu.Lock()
	if len(mock.joinCalls) != 1 {
		t.Errorf("expected no additional JoinRaid calls after duplicate raid ID, got %d total", len(mock.joinCalls))
	}
	mock.mu.Unlock()

	// Third raid message with a different raid ID — should trigger JoinRaid again.
	msg3 := raidMessage("chan123", "raid-002", "other_target")
	m.HandlePubSubMessage(ctx, msg3)

	mock.mu.Lock()
	if len(mock.joinCalls) != 2 {
		t.Errorf("expected 2 JoinRaid calls total after different raid ID, got %d", len(mock.joinCalls))
	}
	if mock.joinCalls[1] != "raid-002" {
		t.Errorf("expected second JoinRaid call with 'raid-002', got %q", mock.joinCalls[1])
	}
	mock.mu.Unlock()

	if streamer.Raid.RaidID != "raid-002" {
		t.Errorf("expected streamer.Raid.RaidID to be updated to 'raid-002', got %q", streamer.Raid.RaidID)
	}
	if streamer.Raid.TargetLogin != "other_target" {
		t.Errorf("expected streamer.Raid.TargetLogin 'other_target', got %q", streamer.Raid.TargetLogin)
	}
}

// TestHandleRaid_FollowRaidDisabled verifies that when FollowRaid is false,
// no JoinRaid call is made and streamer.Raid is not set.
func TestHandleRaid_FollowRaidDisabled(t *testing.T) {
	m, mock := newTestMiner(t)

	streamer := &model.Streamer{
		Username:  "teststreamer",
		ChannelID: "chan456",
		Settings: &model.StreamerSettings{
			FollowRaid: false,
		},
	}
	m.streamersMu.Lock()
	m.streamers = append(m.streamers, streamer)
	m.streamersMu.Unlock()

	ctx := context.Background()

	msg := raidMessage("chan456", "raid-001", "target_streamer")
	m.HandlePubSubMessage(ctx, msg)

	mock.mu.Lock()
	if len(mock.joinCalls) != 0 {
		t.Errorf("expected 0 JoinRaid calls when FollowRaid is disabled, got %d", len(mock.joinCalls))
	}
	mock.mu.Unlock()

	if streamer.Raid != nil {
		t.Error("expected streamer.Raid to be nil when FollowRaid is disabled")
	}
}

// TestHandleRaid_NoStreamer verifies that a raid message for an unknown
// channel ID is silently ignored.
func TestHandleRaid_NoStreamer(t *testing.T) {
	m, mock := newTestMiner(t)

	ctx := context.Background()

	// No streamers registered — message with unknown channel ID.
	msg := raidMessage("unknown-chan", "raid-001", "target_streamer")
	m.HandlePubSubMessage(ctx, msg)

	mock.mu.Lock()
	if len(mock.joinCalls) != 0 {
		t.Errorf("expected 0 JoinRaid calls for unknown streamer, got %d", len(mock.joinCalls))
	}
	mock.mu.Unlock()
}

// TestHandleRaid_MissingRaidData verifies that a raid_update_v2 message
// without a "raid" key is silently ignored.
func TestHandleRaid_MissingRaidData(t *testing.T) {
	m, mock := newTestMiner(t)

	streamer := &model.Streamer{
		Username:  "teststreamer",
		ChannelID: "chan789",
		Settings: &model.StreamerSettings{
			FollowRaid: true,
		},
	}
	m.streamersMu.Lock()
	m.streamers = append(m.streamers, streamer)
	m.streamersMu.Unlock()

	ctx := context.Background()

	// Message with correct type but no "raid" key in RawMessage.
	msg := &model.Message{
		Topic:     "raid",
		ChannelID: "chan789",
		Type:      model.MsgTypeRaidUpdate,
		RawMessage: map[string]any{
			"type": string(model.MsgTypeRaidUpdate),
		},
	}
	m.HandlePubSubMessage(ctx, msg)

	mock.mu.Lock()
	if len(mock.joinCalls) != 0 {
		t.Errorf("expected 0 JoinRaid calls when raid data is missing, got %d", len(mock.joinCalls))
	}
	mock.mu.Unlock()

	if streamer.Raid != nil {
		t.Error("expected streamer.Raid to be nil when raid data is missing")
	}
}

// TestHandleRaid_EmptyRaidID verifies that a raid message with an empty
// raid ID is silently ignored.
func TestHandleRaid_EmptyRaidID(t *testing.T) {
	m, mock := newTestMiner(t)

	streamer := &model.Streamer{
		Username:  "teststreamer",
		ChannelID: "chan101",
		Settings: &model.StreamerSettings{
			FollowRaid: true,
		},
	}
	m.streamersMu.Lock()
	m.streamers = append(m.streamers, streamer)
	m.streamersMu.Unlock()

	ctx := context.Background()

	msg := &model.Message{
		Topic:     "raid",
		ChannelID: "chan101",
		Type:      model.MsgTypeRaidUpdate,
		RawMessage: map[string]any{
			"type": string(model.MsgTypeRaidUpdate),
			"raid": map[string]any{
				"id":           "",
				"target_login": "target_streamer",
			},
		},
	}
	m.HandlePubSubMessage(ctx, msg)

	mock.mu.Lock()
	if len(mock.joinCalls) != 0 {
		t.Errorf("expected 0 JoinRaid calls with empty raid ID, got %d", len(mock.joinCalls))
	}
	mock.mu.Unlock()

	if streamer.Raid != nil {
		t.Error("expected streamer.Raid to be nil when raid ID is empty")
	}
}

// TestHandleRaid_WrongMessageType verifies that non-raid_update_v2 messages
// on the raid topic are ignored.
func TestHandleRaid_WrongMessageType(t *testing.T) {
	m, mock := newTestMiner(t)

	streamer := &model.Streamer{
		Username:  "teststreamer",
		ChannelID: "chan202",
		Settings: &model.StreamerSettings{
			FollowRaid: true,
		},
	}
	m.streamersMu.Lock()
	m.streamers = append(m.streamers, streamer)
	m.streamersMu.Unlock()

	ctx := context.Background()

	// raid_go_v2 should not trigger JoinRaid.
	msg := &model.Message{
		Topic:     "raid",
		ChannelID: "chan202",
		Type:      model.MsgTypeRaidGo,
		RawMessage: map[string]any{
			"type": string(model.MsgTypeRaidGo),
			"raid": map[string]any{
				"id":           "raid-001",
				"target_login": "target_streamer",
			},
		},
	}
	m.HandlePubSubMessage(ctx, msg)

	mock.mu.Lock()
	if len(mock.joinCalls) != 0 {
		t.Errorf("expected 0 JoinRaid calls for raid_go_v2 message, got %d", len(mock.joinCalls))
	}
	mock.mu.Unlock()

	if streamer.Raid != nil {
		t.Error("expected streamer.Raid to be nil for non-raid_update_v2 message")
	}
}
