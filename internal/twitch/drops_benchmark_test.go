package twitch

import (
	"encoding/json"
	"testing"
)

// BenchmarkParseCampaign benchmarks parsing a drop campaign detail response,
// which runs for every campaign during the periodic drop sync (synced about
// every 10 minutes, then again on streamer updates).
func BenchmarkParseCampaign(b *testing.B) {
	// Mirrors the DropCampaignDetails GQL envelope used by
	// GetDropCampaignDetailsBatch (internal/gql/operations.go).
	payload := json.RawMessage(`{
		"id": "camp1",
		"name": "Example Game — Spring Drops",
		"status": "ACTIVE",
		"startAt": "2025-01-01T00:00:00Z",
		"endAt": "2099-01-01T00:00:00Z",
		"game": {
			"id": "g1",
			"name": "Example Game",
			"displayName": "Example Game",
			"slug": "example-game"
		},
		"allow": {
			"channels": [
				{"id": "chan1"},
				{"id": "chan2"}
			]
		},
		"timeBasedDrops": [
			{
				"id": "d1",
				"name": "Drop One",
				"requiredMinutesWatched": 60,
				"startAt": "2025-01-01T00:00:00Z",
				"endAt": "2099-01-01T00:00:00Z",
				"benefitEdges": [
					{"benefit": {"name": "Emote"}}
				]
			},
			{
				"id": "d2",
				"name": "Drop Two",
				"requiredMinutesWatched": 120,
				"startAt": "2025-01-01T00:00:00Z",
				"endAt": "2099-01-01T00:00:00Z",
				"benefitEdges": []
			}
		]
	}`)

	for b.Loop() {
		_, _ = parseCampaign(payload)
	}
}
