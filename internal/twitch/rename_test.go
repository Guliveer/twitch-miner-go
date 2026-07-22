package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func TestCheckStreamerOnline_UpdatesUsernameOnRename(t *testing.T) {
	oldUsername := "oldname"
	newUsername := "newname"
	channelID := "123456"

	mt := newMockTransport()

	mt.setStatusFunc("VideoPlayerStreamInfoOverlayChannel", func(callCount int, vars map[string]any) int {
		login, _ := vars["channel"].(string)
		if login == oldUsername {
			return http.StatusBadRequest
		}
		return http.StatusOK
	})

	mt.setResponseFunc("VideoPlayerStreamInfoOverlayChannel", func(callCount int, vars map[string]any) string {
		login, _ := vars["channel"].(string)
		if login == oldUsername {
			return fmt.Sprintf(`{"error":"missing user data for %s"}`, login)
		}
		if login == newUsername {
			data, _ := json.Marshal(map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"stream": map[string]any{
							"id":          "stream-001",
							"title":       "Test Stream",
							"viewCount":   100,
							"broadcastId": "b-001",
							"game":        map[string]any{"id": "game-1", "displayName": "Just Chatting"},
						},
					},
				},
			})
			return string(data)
		}
		data, _ := json.Marshal(map[string]any{"data": map[string]any{"user": nil}})
		return string(data)
	})

	mt.setResponseFunc("GetLoginFromID", func(callCount int, vars map[string]any) string {
		id, _ := vars["id"].(string)
		if id == channelID {
			data, _ := json.Marshal(map[string]any{
				"data": map[string]any{
					"user": map[string]any{"login": newUsername},
				},
			})
			return string(data)
		}
		data, _ := json.Marshal(map[string]any{"data": map[string]any{"user": nil}})
		return string(data)
	})

	client := newTestClient(t, mt)

	streamer := model.NewStreamer(oldUsername)
	streamer.ChannelID = channelID

	ctx := context.Background()
	err := client.CheckStreamerOnline(ctx, streamer)
	if err != nil {
		t.Fatalf("CheckStreamerOnline returned error: %v", err)
	}

	streamer.Mu.RLock()
	got := streamer.Username
	streamer.Mu.RUnlock()

	if got != newUsername {
		t.Errorf("expected username %q after rename detection, got %q", newUsername, got)
	}
}

func TestIsStaleUsernameError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"missing user data", fmt.Errorf("getting stream info for oldname: GQL operation VideoPlayerStreamInfoOverlayChannel returned error: missing user data for oldname"), true},
		{"user not found", fmt.Errorf("GetUserID for oldname: user oldname not found"), true},
		{"nil error", nil, false},
		{"other error", fmt.Errorf("connection refused"), false},
		{"partial match missing", fmt.Errorf("stream info: something missing something"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaleUsernameError(tt.err)
			if got != tt.want {
				t.Errorf("isStaleUsernameError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestResolveLoginFromID(t *testing.T) {
	mt := newMockTransport()

	mt.setResponseFunc("GetLoginFromID", func(callCount int, vars map[string]any) string {
		id, _ := vars["id"].(string)
		if id == "123456" {
			data, _ := json.Marshal(map[string]any{
				"data": map[string]any{
					"user": map[string]any{"login": "newname"},
				},
			})
			return string(data)
		}
		data, _ := json.Marshal(map[string]any{
			"data": map[string]any{"user": nil},
		})
		return string(data)
	})

	client := newTestClient(t, mt)

	login, err := client.ResolveLoginFromID(context.Background(), "123456")
	if err != nil {
		t.Fatalf("ResolveLoginFromID returned error: %v", err)
	}
	if login != "newname" {
		t.Errorf("expected login %q, got %q", "newname", login)
	}

	login, err = client.ResolveLoginFromID(context.Background(), "999999")
	if err != nil {
		t.Fatalf("ResolveLoginFromID returned error for unknown ID: %v", err)
	}
	if login != "" {
		t.Errorf("expected empty login for unknown ID, got %q", login)
	}
}
