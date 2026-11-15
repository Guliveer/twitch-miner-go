package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Guliveer/twitch-miner-go/internal/gql"
	"github.com/Guliveer/twitch-miner-go/internal/model"
)

func TestBadgeWatcherRefreshesDeadlineFromInventory(t *testing.T) {
	t.Parallel()
	campaign := activeBadgeCampaign("catalog-id", "Game One", "game-one", "Badge One")
	campaign.WatchTimes = map[string]time.Duration{"Badge One": 45 * time.Minute}
	client := &stubBadgeGQL{streams: map[string][]gql.TopStream{
		"game-one": {{Username: "one", ChannelID: "1", GameName: "Game One"}},
	}}
	bw := testBadgeWatcher(t, client, 1, campaign)
	var streamers []*model.Streamer
	add := func(_ context.Context, s *model.Streamer) { streamers = append(streamers, s) }
	remove := func(string, string) { streamers = nil }
	get := func() []*model.Streamer { return streamers }

	bw.evaluate(context.Background(), add, remove, get)
	if len(streamers) != 1 || streamers[0].BadgeReward.Remaining != 45*time.Minute {
		t.Fatalf("expected one streamer with full catalog duration, got %v", streamers)
	}
	initial := streamers[0]
	dropEnd := campaign.EndsAt.Add(-10 * time.Minute)
	client.raw = json.RawMessage(fmt.Sprintf(`{"dropCampaignsInProgress":[{"id":"twitch-uuid","name":"Game One","endAt":%q,"game":{"slug":"game-one"},"timeBasedDrops":[{"endAt":%q,"requiredMinutesWatched":45,"benefitEdges":[{"benefit":{"name":"Badge One"}}],"self":{"hasPreconditionsMet":true,"currentMinutesWatched":30}}]}]}`, campaign.EndsAt.Format(time.RFC3339Nano), dropEnd.Format(time.RFC3339Nano)))
	bw.evaluate(context.Background(), add, remove, get)
	if len(streamers) != 1 || streamers[0] != initial || initial.BadgeReward.Remaining != 15*time.Minute || !initial.BadgeReward.EndAt.Equal(dropEnd) {
		t.Fatalf("tracker did not update in place: %+v", initial.BadgeReward)
	}
	client.owned = map[string]struct{}{"badge one": {}}
	bw.evaluate(context.Background(), add, remove, get)
	if len(streamers) != 0 {
		t.Fatal("owned badge still tracked")
	}
}

func TestBadgeWatcherSkipsPollWithMalformedProgress(t *testing.T) {
	t.Parallel()

	campaign := activeBadgeCampaign("catalog-id", "Game One", "game-one", "Badge One")
	campaign.WatchTimes = map[string]time.Duration{"Badge One": 30 * time.Minute}
	client := &stubBadgeGQL{
		raw: json.RawMessage(`{"dropCampaignsInProgress":[{"endAt":"invalid"}]}`),
		streams: map[string][]gql.TopStream{
			"game-one": {{Username: "one", ChannelID: "1", GameName: "Game One"}},
		},
	}
	bw := testBadgeWatcher(t, client, 1, campaign)
	bw.evaluate(context.Background(), func(context.Context, *model.Streamer) {
		t.Error("added a streamer using catalog estimates after an inventory parse error")
	}, func(string, string) {}, func() []*model.Streamer { return nil })
}

func TestBadgeWatchRequirement(t *testing.T) {
	t.Parallel()
	for input, want := range map[string]time.Duration{
		"Watch 1h 30m":  90 * time.Minute,
		"Watch 20m":     20 * time.Minute,
		"Subscribe":     0,
		"Watch unknown": 0,
		"Watch -1h":     0,
	} {
		if got := watchRequirement(input); got != want {
			t.Errorf("%q = %s, want %s", input, got, want)
		}
	}
}

func TestBadgeCampaignNextRewardUsesTightestSlack(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	campaign := badgeCampaign{
		ID:         "campaign",
		EndsAt:     now.Add(3 * time.Hour),
		BadgeNames: []string{"Short Badge", "Urgent Badge"},
	}
	var inventory []badgeInventoryCampaign
	if err := json.Unmarshal([]byte(`[{"id":"campaign","timeBasedDrops":[
		{"endAt":"2026-11-15T15:00:00Z","requiredMinutesWatched":30,"benefitEdges":[{"benefit":{"name":"Short Badge"}}]},
		{"endAt":"2026-11-15T13:50:00Z","requiredMinutesWatched":100,"benefitEdges":[{"benefit":{"name":"Urgent Badge"}}]}
	]}]`), &inventory); err != nil {
		t.Fatal(err)
	}
	if got, want := campaign.nextReward(nil, inventory, now).LatestStart(), now.Add(10*time.Minute); !got.Equal(want) {
		t.Fatalf("latest start = %s; want %s for the urgent longer badge", got, want)
	}
	inventory[0].Drops[1].RequiredMinutes = 106
	if got, want := campaign.nextReward(nil, inventory, now).LatestStart(), now.Add(150*time.Minute); !got.Equal(want) {
		t.Fatalf("latest start = %s; want %s when the urgent badge misses the safety margin", got, want)
	}
}
