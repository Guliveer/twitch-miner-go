package model

import (
	"encoding/json"
	"testing"
	"time"
)

// --- splitTopic ---

func TestSplitTopic_Normal(t *testing.T) {
	topic, user := splitTopic("community-points-user-v1.12345")
	if topic != "community-points-user-v1" {
		t.Errorf("expected topic 'community-points-user-v1', got %q", topic)
	}
	if user != "12345" {
		t.Errorf("expected user '12345', got %q", user)
	}
}

func TestSplitTopic_NoDot(t *testing.T) {
	topic, user := splitTopic("simpletopic")
	if topic != "simpletopic" {
		t.Errorf("expected topic 'simpletopic', got %q", topic)
	}
	if user != "" {
		t.Errorf("expected empty user, got %q", user)
	}
}

func TestSplitTopic_MultipleDots(t *testing.T) {
	// Should split on the LAST dot
	topic, user := splitTopic("video-playback-by-id.some.channel.99999")
	if topic != "video-playback-by-id.some.channel" {
		t.Errorf("expected topic 'video-playback-by-id.some.channel', got %q", topic)
	}
	if user != "99999" {
		t.Errorf("expected user '99999', got %q", user)
	}
}

// --- ParseMessage: basic types ---

func TestParseMessage_PointsEarned(t *testing.T) {
	raw := map[string]any{
		"type": "points-earned",
		"data": map[string]any{
			"timestamp": "2025-01-15T12:00:00Z",
			"balance": map[string]any{
				"channel_id": "chan123",
			},
			"point_gain": map[string]any{
				"total_points": float64(50),
				"reason_code":  "WATCH",
			},
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("community-points-user-v1.userABC", rawJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Topic != "community-points-user-v1" {
		t.Errorf("expected topic 'community-points-user-v1', got %q", msg.Topic)
	}
	if msg.TopicUser != "userABC" {
		t.Errorf("expected topic_user 'userABC', got %q", msg.TopicUser)
	}
	if msg.Type != MsgTypePointsEarned {
		t.Errorf("expected type points-earned, got %q", msg.Type)
	}
	// channel_id resolved from data.balance.channel_id
	if msg.ChannelID != "chan123" {
		t.Errorf("expected channel_id 'chan123', got %q", msg.ChannelID)
	}
	// Timestamp from data.timestamp (RFC3339)
	expected := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	if !msg.Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, msg.Timestamp)
	}
}

func TestParseMessage_PredictionEvent(t *testing.T) {
	raw := map[string]any{
		"type": "event-created",
		"data": map[string]any{
			"timestamp": "2025-06-01T08:30:00Z",
			"prediction": map[string]any{
				"channel_id": "pred_chan_456",
				"event_id":   "evt1",
			},
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("predictions-channel-v1.99999", rawJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != MsgTypePredictionEvent {
		t.Errorf("expected type event-created, got %q", msg.Type)
	}
	// channel_id from data.prediction.channel_id
	if msg.ChannelID != "pred_chan_456" {
		t.Errorf("expected channel_id 'pred_chan_456', got %q", msg.ChannelID)
	}
}

func TestParseMessage_ClaimAvailable(t *testing.T) {
	raw := map[string]any{
		"type": "claim-available",
		"data": map[string]any{
			"claim": map[string]any{
				"channel_id": "claim_chan_789",
				"id":         "claim1",
			},
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("community-points-user-v1.userXYZ", rawJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != MsgTypeClaimAvailable {
		t.Errorf("expected type claim-available, got %q", msg.Type)
	}
	// channel_id from data.claim.channel_id
	if msg.ChannelID != "claim_chan_789" {
		t.Errorf("expected channel_id 'claim_chan_789', got %q", msg.ChannelID)
	}
}

func TestParseMessage_StreamUp_NoDataField(t *testing.T) {
	// viewcount/stream-up messages may have no "data" field; type is in root
	raw := map[string]any{
		"type":        "stream-up",
		"server_time": float64(1700000000),
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("video-playback-by-id.stream123", rawJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != MsgTypeStreamUp {
		t.Errorf("expected type stream-up, got %q", msg.Type)
	}
	// No data → channel_id falls back to topicUser
	if msg.ChannelID != "stream123" {
		t.Errorf("expected channel_id 'stream123', got %q", msg.ChannelID)
	}
	// Timestamp from server_time
	expected := time.Unix(1700000000, 0).UTC()
	if !msg.Timestamp.Equal(expected) {
		t.Errorf("expected timestamp %v, got %v", expected, msg.Timestamp)
	}
}

func TestParseMessage_RaidMessage(t *testing.T) {
	// Raid messages have type in root, data field present
	raw := map[string]any{
		"type": "raid_update_v2",
		"data": map[string]any{
			"channel_id": "raid_chan_100",
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("raid.55555", rawJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != MsgTypeRaidUpdate {
		t.Errorf("expected type raid_update_v2, got %q", msg.Type)
	}
	// channel_id from data.channel_id
	if msg.ChannelID != "raid_chan_100" {
		t.Errorf("expected channel_id 'raid_chan_100', got %q", msg.ChannelID)
	}
}

// --- Timestamp resolution ---

func TestParseMessage_TimestampFromRFC3339(t *testing.T) {
	raw := map[string]any{
		"type": "points-earned",
		"data": map[string]any{
			"timestamp":   "2024-12-25T00:00:00Z",
			"server_time": float64(1000000000), // should be ignored
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("topic.user1", rawJSON)
	if err != nil {
		t.Fatal(err)
	}
	expected := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	if !msg.Timestamp.Equal(expected) {
		t.Errorf("expected RFC3339 timestamp %v, got %v", expected, msg.Timestamp)
	}
}

func TestParseMessage_TimestampFromServerTime(t *testing.T) {
	raw := map[string]any{
		"type": "stream-up",
		"data": map[string]any{
			"server_time": float64(1609459200),
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("topic.user1", rawJSON)
	if err != nil {
		t.Fatal(err)
	}
	expected := time.Unix(1609459200, 0).UTC()
	if !msg.Timestamp.Equal(expected) {
		t.Errorf("expected server_time timestamp %v, got %v", expected, msg.Timestamp)
	}
}

func TestParseMessage_TimestampFallbackToNow(t *testing.T) {
	raw := map[string]any{
		"type": "some-event",
		"data": map[string]any{
			"foo": "bar",
		},
	}
	rawJSON, _ := json.Marshal(raw)

	before := time.Now().Add(-time.Second)
	msg, err := ParseMessage("topic.user1", rawJSON)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now().Add(time.Second)

	if msg.Timestamp.Before(before) || msg.Timestamp.After(after) {
		t.Errorf("expected timestamp close to now, got %v", msg.Timestamp)
	}
}

// --- Channel ID resolution ---

func TestParseMessage_ChannelIDFromDataChannelID(t *testing.T) {
	raw := map[string]any{
		"type": "raid_go_v2",
		"data": map[string]any{
			"channel_id": "direct_chan",
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("topic.fallback_user", rawJSON)
	if err != nil {
		t.Fatal(err)
	}
	if msg.ChannelID != "direct_chan" {
		t.Errorf("expected 'direct_chan', got %q", msg.ChannelID)
	}
}

func TestParseMessage_ChannelIDFallbackToTopicUser(t *testing.T) {
	raw := map[string]any{
		"type": "some-event",
		"data": map[string]any{
			"unrelated": "field",
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("topic.fallback_user", rawJSON)
	if err != nil {
		t.Fatal(err)
	}
	if msg.ChannelID != "fallback_user" {
		t.Errorf("expected fallback 'fallback_user', got %q", msg.ChannelID)
	}
}

// --- Error cases ---

func TestParseMessage_MalformedJSON(t *testing.T) {
	_, err := ParseMessage("topic.user", []byte("{invalid json"))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseMessage_MissingType(t *testing.T) {
	raw := map[string]any{
		"data": map[string]any{
			"channel_id": "chan1",
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("topic.user1", rawJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "" {
		t.Errorf("expected empty type, got %q", msg.Type)
	}
}

func TestParseMessage_EmptyObject(t *testing.T) {
	msg, err := ParseMessage("topic.user1", []byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "" {
		t.Errorf("expected empty type, got %q", msg.Type)
	}
	// No data → channel_id falls back to topic user
	if msg.ChannelID != "user1" {
		t.Errorf("expected channel_id 'user1', got %q", msg.ChannelID)
	}
}

// --- Identifier ---

func TestParseMessage_Identifier(t *testing.T) {
	raw := map[string]any{
		"type": "claim-available",
		"data": map[string]any{
			"claim": map[string]any{
				"channel_id": "id_chan",
			},
		},
	}
	rawJSON, _ := json.Marshal(raw)

	msg, err := ParseMessage("community-points-user-v1.userZ", rawJSON)
	if err != nil {
		t.Fatal(err)
	}
	expected := "claim-available.community-points-user-v1.id_chan"
	if msg.Identifier != expected {
		t.Errorf("expected identifier %q, got %q", expected, msg.Identifier)
	}
}

// --- String ---

func TestMessage_String(t *testing.T) {
	msg := &Message{
		Type:      MsgTypeStreamUp,
		Topic:     "video-playback-by-id",
		ChannelID: "chan42",
	}
	s := msg.String()
	if s != "Message(type=stream-up, topic=video-playback-by-id, channel_id=chan42)" {
		t.Errorf("unexpected String(): %s", s)
	}
}
